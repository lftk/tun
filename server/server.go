package server

import (
	"context"
	"net"

	"github.com/4396/tun/proxy"
	"github.com/4396/tun/vhost"
	"github.com/golang/sync/syncmap"
)

// Server combination net.Listener and vhost.Muxer, manage all proxies.
// Listen to the client's connection and the user's http connection.
// The load function loads the proxy using lazy load mode.
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

// Config describes the configuration information for the server.
type Config struct {
	Addr     string
	AddrHTTP string
	Auth     AuthFunc
	Load     LoadFunc
	TraffIn  TraffFunc
	TraffOut TraffFunc
}

// TraffFunc records traffic for a proxy.
type TraffFunc func(id string, b []byte)

// AuthFunc using id and token to authorize a proxy.
type AuthFunc func(id, token string) error

// LoadFunc using loader to load a proxy with id.
type LoadFunc func(loader Loader, id string) error

// Listen tun address and http address, create a server.
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

// Proxies returns all proxies.
func (s *Server) Proxies() []proxy.Proxy {
	return s.service.Proxies()
}

// Kill a proxy.
func (s *Server) Kill(id string) {
	s.sessions.Range(func(key, val interface{}) bool {
		return !val.(*session).Kill(id)
	})
}

// Run this server, receive various user connections
// and client connections, and handle events.
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

	s.goctx(ctx, s.listen)
	if s.muxer != nil {
		s.goctx(ctx, s.muxer.Serve)
	}
	s.goctx(ctx, s.service.Serve)

	for {
		select {
		case c := <-s.connc:
			s.handleConn(ctx, c)
		case <-ctx.Done():
			err = ctx.Err()
			return
		case err = <-s.errc:
			return
		}
	}
}

// listen tun address, receive client connections.
func (s *Server) listen(ctx context.Context) (err error) {
	defer s.listener.Close()
	for {
		var conn net.Conn
		conn, err = s.listener.Accept()
		if err != nil {
			s.errc <- err
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

// goctx using goroutine execute `do` function.
func (s *Server) goctx(ctx context.Context, do func(context.Context) error) {
	go func() {
		if err := do(ctx); err != nil {
			s.errc <- err
		}
	}()
}

// handle the client's connection.
func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	sess, err := newSession(s, conn)
	if err != nil {
		conn.Close()
		return
	}

	s.sessions.Store(sess, sess)
	go func() {
		defer s.sessions.Delete(sess)
		sess.Run(ctx)
	}()
}
