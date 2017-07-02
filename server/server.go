package server

import (
	"context"
	"fmt"
	"net"

	"github.com/4396/tun/conn"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/vhost"
)

type Server struct {
	listener     net.Listener
	httpListener net.Listener
	auth         AuthFunc
	muxer        vhost.Muxer
	service      proxy.Service
	connc        chan net.Conn
	httpConnc    chan net.Conn
	errc         chan error
}

type AuthFunc func(name, token string) error

func Listen(addr, httpAddr string, auth AuthFunc) (s *Server, err error) {
	var l, h net.Listener
	l, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	if httpAddr != "" {
		h, err = net.Listen("tcp", httpAddr)
		if err != nil {
			l.Close()
			return
		}
	}

	s = &Server{
		listener:     l,
		httpListener: h,
		auth:         auth,
	}
	return
}

func (s *Server) ProxyTCP(name string, port int) (err error) {
	addr := fmt.Sprintf(":%d", port)
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

func (s *Server) ProxyHTTP(name, domain string) (err error) {
	l := conn.NewListener()
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

func (s *Server) Traffic(traff proxy.Traffic) {
	s.service.Traff = traff
}

func (s *Server) Run(ctx context.Context) (err error) {
	s.errc = make(chan error, 1)
	s.connc = make(chan net.Conn, 16)
	s.httpConnc = make(chan net.Conn, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		close(s.connc)
		for conn := range s.connc {
			conn.Close()
		}

		close(s.httpConnc)
		for conn := range s.httpConnc {
			conn.Close()
		}
	}()

	go s.listen(ctx, s.listener, s.connc)
	if s.httpListener != nil {
		go s.listen(ctx, s.httpListener, s.httpConnc)
	}
	go func() {
		err := s.service.Serve(ctx)
		if err != nil {
			s.errc <- err
		}
	}()

	for {
		select {
		case c := <-s.connc:
			go s.handleConn(ctx, c)
		case c := <-s.httpConnc:
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
