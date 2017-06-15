package server

import (
	"context"
	"net"

	"github.com/4396/tun/proxy"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/vhost"
)

type Server struct {
	Addr     string
	HttpAddr string

	Admin Administrator

	muxer   vhost.Muxer
	service proxy.Service

	errc      chan error
	connc     chan net.Conn
	httpConnc chan net.Conn
}

func (s *Server) ProxyTCP(name, addr string) (err error) {
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
	s.connc = make(chan net.Conn, 16)
	s.httpConnc = make(chan net.Conn, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		close(s.connc)
		close(s.httpConnc)

		// close conn
		for conn := range s.httpConnc {
			conn.Close()
		}
		for conn := range s.connc {
			conn.Close()
		}
	}()

	go s.listen(ctx, l, s.connc)
	if h != nil {
		go s.listen(ctx, h, s.httpConnc)
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

func (s *Server) handleConn(ctx context.Context, c net.Conn) {
	select {
	case <-ctx.Done():
	default:
		session{Server: s, Conn: c}.loopMessage(ctx)
	}
}

func (s *Server) handleHttpConn(ctx context.Context, c net.Conn) {
	select {
	case <-ctx.Done():
	default:
		s.muxer.Handle(c)
	}
}
