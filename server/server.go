package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/vhost"
	"github.com/xtaci/smux"
)

type Server struct {
	Addr     string
	HttpAddr string

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
		close(s.errc)
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
	fmt.Println("http", c.RemoteAddr())

	select {
	case <-ctx.Done():
	default:
		s.muxer.Handle(c)
	}
}

var pong msg.Pong

func (s *Server) handleClientConn(ctx context.Context, c net.Conn) {
	fmt.Println("client", c.RemoteAddr())
	defer fmt.Println("----------------------------")

	agent, err := s.authClientConn(c)
	if err != nil {
		c.Close()
		return
	}

	lastPing := time.Now()
	go func() {
		defer fmt.Println("..go1..")
		for {
			m, err := msg.Read(agent.conn)
			if err != nil {
				return
			}

			switch m.(type) {
			case *msg.Ping:
				lastPing = time.Now()
				msg.Write(agent.conn, &pong)
			default:
			}
		}
	}()

	var (
		dura   = time.Second * 10
		ticker = time.NewTicker(dura)
		exitc  = make(chan interface{})
	)

	defer func() {
		close(exitc)
		ticker.Stop()
	}()

	go func() {
		defer fmt.Println("..go2..")
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if lastPing.Add(3 * dura).Before(time.Now()) {
					s.service.Unregister(agent.name, agent)
					return
				}
			case <-exitc:
				s.service.Unregister(agent.name, agent)
				return
			}
		}
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
	fmt.Println("agent", ac.agent.name)

	select {
	case <-ctx.Done():
	default:
		m, err := msg.Read(ac.Conn)
		if err != nil {
			return
		}

		switch mm := m.(type) {
		case *msg.WorkConn:
			fmt.Println("work", mm)
			ac.agent.Put(ac.Conn)
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

	var m msg.Auth
	err = msg.ReadInto(st, &m)
	if err != nil {
		return
	}

	// auth client
	// ...

	a = &agent{
		name:  m.Name,
		sess:  sess,
		conn:  st,
		connc: make(chan net.Conn, 16),
	}
	err = s.service.Register(m.Name, a)
	if err != nil {
		msg.Write(st, &msg.AuthResp{
			Error: err.Error(),
		})
		return
	}

	err = msg.Write(st, &msg.AuthResp{})
	return
}
