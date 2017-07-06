package server

import (
	"context"
	"net"

	"github.com/golang/sync/syncmap"

	"github.com/4396/tun/proxy"
	"github.com/4396/tun/vhost"
)

type Server struct {
	listener     net.Listener
	httpListener net.Listener
	auth         AuthFunc
	load         LoadFunc
	muxer        vhost.Muxer
	service      proxy.Service
	errc         chan error
	sessions     syncmap.Map
}

type (
	TraffFunc func(name string, b []byte)
	AuthFunc  func(name, token string) error
	LoadFunc  func(loader Loader, name string) error
)

type Config struct {
	Addr     string
	AddrHTTP string
	Auth     AuthFunc
	Load     LoadFunc
	TraffIn  TraffFunc
	TraffOut TraffFunc
}

func Listen(cfg *Config) (s *Server, err error) {
	var l, h net.Listener
	l, err = net.Listen("tcp", cfg.Addr)
	if err != nil {
		return
	}

	if cfg.AddrHTTP != "" {
		h, err = net.Listen("tcp", cfg.AddrHTTP)
		if err != nil {
			l.Close()
			return
		}
	}

	s = &Server{
		listener:     l,
		httpListener: h,
		auth:         cfg.Auth,
		load:         cfg.Load,
		errc:         make(chan error, 1),
	}
	s.service.Traff = &traffic{
		TraffIn:  cfg.TraffIn,
		TraffOut: cfg.TraffOut,
	}
	return
}

func (s *Server) Proxies() []proxy.Proxy {
	return s.service.Proxies()
}

func (s *Server) Kill(name string) {
	s.sessions.Range(func(key, val interface{}) bool {
		return !val.(*session).Kill(name)
	})
}

func (s *Server) Run(ctx context.Context) (err error) {
	connc := make(chan net.Conn, 16)
	httpConnc := make(chan net.Conn, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		close(connc)
		for conn := range connc {
			conn.Close()
		}

		close(httpConnc)
		for conn := range httpConnc {
			conn.Close()
		}
	}()

	go s.listen(ctx, s.listener, connc)
	if s.httpListener != nil {
		go s.listen(ctx, s.httpListener, httpConnc)
	}

	go func() {
		err := s.service.Serve(ctx)
		if err != nil {
			s.errc <- err
		}
	}()

	for {
		select {
		case c := <-connc:
			go s.handleConn(ctx, c)
		case c := <-httpConnc:
			go s.handleHttpConn(ctx, c)
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

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	select {
	case <-ctx.Done():
	default:
		sess, err := newSession(s, conn)
		if err != nil {
			conn.Close()
		} else {
			s.sessions.Store(sess, sess)
			defer s.sessions.Delete(sess)

			sess.Run(ctx)
		}
	}
}

func (s *Server) handleHttpConn(ctx context.Context, conn net.Conn) {
	select {
	case <-ctx.Done():
	default:
		s.muxer.Handle(conn)
	}
}
