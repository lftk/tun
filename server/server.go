package server

import (
	"context"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/vhost"
	"github.com/xtaci/smux"
)

type Server struct {
	Addr     string
	HttpAddr string

	Auth func(name, token string) error

	muxer   vhost.Muxer
	service proxy.Service

	errc        chan error
	httpConnc   chan net.Conn
	clientConnc chan net.Conn
	agentConnc  chan agentConn
}

type agentConn struct {
	*agent
	net.Conn
}

func (s *Server) Tcp(name, addr string) (err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	p := proxy.Wrap(name, l)
	err = s.Proxy(p)
	if err != nil {
		l.Close()
	}
	return
}

type httpProxy struct {
	proxy.Proxy
	domain string
}

func (s *Server) Http(name, domain string) (err error) {
	l := proxy.NewListener()
	p := proxy.Wrap(name, l)
	err = s.Proxy(httpProxy{p, domain})
	if err != nil {
		return
	}

	s.muxer.HandleFunc(domain, l.Put)
	return
}

func (s *Server) Proxy(p proxy.Proxy) error {
	return s.service.Proxy(p)
}

func (s *Server) Proxies() []proxy.Proxy {
	return s.service.Proxies()
}

func (s *Server) Kill(name string) {
	p := s.service.Kill(name)
	hp, ok := p.(httpProxy)
	if ok {
		s.muxer.HandleFunc(hp.domain, nil)
	}
}

func (s *Server) Traffic(traff traffic.Traffic) {
	s.service.Traff = traff
}

func (s *Server) ListenAndServe(ctx context.Context) (err error) {
	var l, h net.Listener
	l, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return
	}

	if s.HttpAddr != "" {
		h, err = net.Listen("tcp", s.HttpAddr)
		if err != nil {
			l.Close()
			return
		}
	}

	err = s.Serve(ctx, l, h)
	return
}

func (s *Server) Serve(ctx context.Context, l, h net.Listener) (err error) {
	s.errc = make(chan error, 1)
	s.httpConnc = make(chan net.Conn, 16)
	s.clientConnc = make(chan net.Conn, 16)
	s.agentConnc = make(chan agentConn, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		// close listener
		l.Close()
		h.Close()

		// close channel
		//close(s.errc)
		close(s.httpConnc)
		close(s.clientConnc)
		close(s.agentConnc)

		// close conn
		for conn := range s.httpConnc {
			conn.Close()
		}
		for conn := range s.clientConnc {
			conn.Close()
		}
		for ac := range s.agentConnc {
			ac.Conn.Close()
		}
	}()

	go s.listen(ctx, l, s.clientConnc)
	if h != nil {
		go s.listen(ctx, h, s.httpConnc)
	}

	go func() {
		s.errc <- s.service.Serve(ctx)
	}()

	for {
		select {
		case c := <-s.httpConnc:
			go s.handleHttpConn(ctx, c)
		case c := <-s.clientConnc:
			go s.handleClientConn(ctx, c)
		case c := <-s.agentConnc:
			go s.handleAgentConn(ctx, c)
		case err = <-s.errc:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) listen(ctx context.Context, l net.Listener, connc chan<- net.Conn) {
	for {
		conn, err := l.Accept()
		if err != nil {
			s.errc <- err
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			connc <- conn
		}
	}
}

func (s *Server) handleHttpConn(ctx context.Context, c net.Conn) {
	select {
	case <-ctx.Done():
	default:
		s.muxer.Handle(c)
	}
}

func (s *Server) handleClientConn(ctx context.Context, c net.Conn) {
	agent, err := s.authClientConn(c)
	if err != nil {
		c.Close()
		return
	}

	defer func() {
		s.service.Unregister(agent.name, agent)
	}()

	for {
		st, err := agent.sess.AcceptStream()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			s.agentConnc <- agentConn{agent, st}
		}
	}
}

func (s *Server) handleAgentConn(ctx context.Context, ac agentConn) {
	select {
	case <-ctx.Done():
	default:
		m, err := msg.Read(ac.Conn)
		if err != nil {
			ac.Conn.Close()
			return
		}

		switch m.(type) {
		case *msg.WorkConn:
			err := ac.agent.Put(ac.Conn)
			if err != nil {
				ac.Conn.Close()
			}
		default:
			fmt.Printf("other %+v\n", m)
		}
	}
}

func (s *Server) authClientConn(conn net.Conn) (a *agent, err error) {
	sess, err := smux.Server(conn, smux.DefaultConfig())
	if err != nil {
		return
	}

	st, err := sess.AcceptStream()
	if err != nil {
		return
	}

	var auth msg.Auth
	err = msg.ReadInto(st, &auth)
	if err != nil {
		return
	}

	if s.Auth != nil {
		err = s.Auth(auth.Name, auth.Token)
		if err != nil {
			return
		}
	}

	a = &agent{
		name:  auth.Name,
		sess:  sess,
		conn:  st,
		connc: make(chan net.Conn, 16),
	}
	err = s.service.Register(auth.Name, a)
	if err != nil {
		msg.Write(st, &msg.AuthResp{
			Error: err.Error(),
		})
		return
	}

	err = msg.Write(st, &msg.AuthResp{})
	return
}
