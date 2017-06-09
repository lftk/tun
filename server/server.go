package server

import (
	"context"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/vhost"
	"github.com/golang/sync/syncmap"
	"github.com/xtaci/smux"
)

type Server struct {
	Addr     string
	HttpAddr string

	agents  syncmap.Map
	muxer   vhost.Muxer
	service proxy.Service

	errc   chan error
	connc  chan net.Conn
	hconnc chan net.Conn
	stc    chan *smux.Stream
}

func (s *Server) TcpProxy(name, addr string) (err error) {
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

func (s *Server) HttpProxy(name, domain string) (err error) {
	l := proxy.NewListener()
	p := proxy.Wrap(name, l)
	err = s.Proxy(p)
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
	s.connc = make(chan net.Conn, 16)
	s.hconnc = make(chan net.Conn, 16)
	s.stc = make(chan *smux.Stream, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.listen(ctx, l, s.connc)
	if h != nil {
		go s.listen(ctx, h, s.hconnc)
	}

	go func() {
		s.errc <- s.service.Serve(ctx)
	}()

	for {
		select {
		case c := <-s.hconnc:
			go s.handleHttpConn(ctx, c)
		case c := <-s.connc:
			go s.handleConn(ctx, c)
		case c := <-s.stc:
			go s.handleStream(ctx, c)
		case err = <-s.errc:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) listen(ctx context.Context, l net.Listener, connc chan<- net.Conn) {
	defer l.Close()
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

func (s *Server) handleConn(ctx context.Context, c net.Conn) {
	fmt.Println("client", c.RemoteAddr())

	sess, err := smux.Server(c, smux.DefaultConfig())
	if err != nil {
		return
	}

	defer sess.Close()
	for {
		st, err := sess.AcceptStream()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			s.stc <- st
		}
	}
}

func (s *Server) handleStream(ctx context.Context, st *smux.Stream) {
	fmt.Println("stream", st.RemoteAddr())

	select {
	case <-ctx.Done():
	default:
		m, err := msg.Read(st)
		if err != nil {
			return
		}

		switch mm := m.(type) {
		case *msg.Login:
			fmt.Println("login", mm)

			a := NewAgent(st)
			err = s.service.Register(mm.Name, a)
			if err != nil {
				return
			}

			s.agents.Store(mm.Name, a)
		case *msg.WorkConn:
			fmt.Println("work", mm)

			val, ok := s.agents.Load(mm.Name)
			if ok {
				val.(*Agent).Put(st)
			}
		default:
			fmt.Printf("other %+v\n", mm)
		}
	}
}
