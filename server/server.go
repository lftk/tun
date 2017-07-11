package server

import (
	"context"
	"net"

	"github.com/golang/sync/syncmap"

	"github.com/4396/tun/proxy"
	"github.com/4396/tun/vhost"
)

type Server struct {
	listener net.Listener
	connc    chan net.Conn
	muxer    *vhost.Muxer
	auth     AuthFunc
	load     LoadFunc
	service  proxy.Service
	sessions syncmap.Map
	errc     chan error
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
	var (
		listener net.Listener
		muxer    *vhost.Muxer
	)

	listener, err = net.Listen("tcp", cfg.Addr)
	if err != nil {
		return
	}

	if cfg.AddrHTTP != "" {
		muxer, err = vhost.Listen(cfg.AddrHTTP)
		if err != nil {
			return
		}
	}

	s = &Server{
		listener: listener,
		muxer:    muxer,
		auth:     cfg.Auth,
		load:     cfg.Load,
		errc:     make(chan error, 1),
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
	s.connc = make(chan net.Conn, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		close(s.connc)
		for conn := range s.connc {
			conn.Close()
		}

		if s.muxer != nil {
			s.muxer.Close()
		}
	}()

	s.ctxDo(ctx, s.listen)
	s.ctxDo(ctx, s.muxer.Serve)
	s.ctxDo(ctx, s.service.Serve)

	for {
		select {
		case c := <-s.connc:
			go s.handleConn(ctx, c)
		case <-ctx.Done():
			err = ctx.Err()
			return
		case err = <-s.errc:
			return
		}
	}
}

func (s *Server) listen(ctx context.Context) (err error) {
	defer s.listener.Close()
	for {
		var conn net.Conn
		conn, err = s.listener.Accept()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			s.connc <- conn
		}
	}
}

func (s *Server) ctxDo(ctx context.Context, do func(context.Context) error) {
	go func() {
		if err := do(ctx); err != nil {
			s.errc <- err
		}
	}()
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
