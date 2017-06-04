package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/transport"
	"github.com/golang/sync/syncmap"
)

type Proxy interface {
	net.Listener
	Name() string
	Handle(net.Conn) error
	Bind(transport.Transport, msg.Message) bool
}

type conn struct {
	Proxy
	net.Conn
}

type Server struct {
	Listener transport.Listener
	proxies  syncmap.Map
	errc     chan error
	connc    chan conn
	proxyc   chan Proxy
	tranc    chan transport.Transport
	exitc    chan interface{}
}

func New(listener transport.Listener) (s *Server) {
	s = &Server{Listener: listener}
	return
}

func (s *Server) Proxy(p Proxy) (err error) {
	_, loaded := s.proxies.LoadOrStore(p.Name(), p)
	if loaded {
		err = errors.New("already existed")
		return
	}

	if s.proxyc != nil {
		s.proxyc <- p
	}
	return
}

func (s *Server) Proxies() (proxies []Proxy) {
	s.proxies.Range(func(key, val interface{}) bool {
		proxies = append(proxies, val.(Proxy))
		return true
	})
	return
}

func (s *Server) Shutdown() {
	if s.exitc != nil {
		close(s.exitc)
		s.exitc = nil
	}
}

func (s *Server) ListenAndServe() (err error) {
	s.errc = make(chan error, 1)
	s.exitc = make(chan interface{})
	s.connc = make(chan conn, 16)
	s.proxyc = make(chan Proxy, 16)
	s.tranc = make(chan transport.Transport, 16)

	go s.listenTransport()
	s.proxies.Range(func(key, val interface{}) bool {
		go s.listenProxy(val.(Proxy))
		return true
	})

	for {
		select {
		case p := <-s.proxyc:
			go s.listenProxy(p)
		case t := <-s.tranc:
			go s.handleTransport(t)
		case c := <-s.connc:
			go s.handleConn(c)
		case err = <-s.errc:
			s.Shutdown()
			return
		case <-s.exitc:
			return
		}
	}
}

func (s *Server) listenTransport() {
	defer s.Listener.Close()
	for {
		tran, err := s.Listener.Accept()
		if err != nil {
			return
		}

		select {
		case <-s.exitc:
		default:
			s.tranc <- tran
		}
	}
}

func (s *Server) listenProxy(p Proxy) {
	defer p.Close()
	for {
		c, err := p.Accept()
		if err != nil {
			s.errc <- err
			return
		}

		select {
		case <-s.exitc:
		default:
			s.connc <- conn{Proxy: p, Conn: c}
		}
	}
}

func (s *Server) handleTransport(t transport.Transport) {
	fmt.Println("tran", t.RemoteAddr())
	var m msg.Message
	s.proxies.Range(func(key, val interface{}) bool {
		ok := val.(Proxy).Bind(t, m)
		return !ok
	})
}

func (s *Server) handleConn(c conn) {
	fmt.Println("conn", c.RemoteAddr())
	err := c.Handle(c.Conn)
	if err != nil {
		// ...
	}
}

func ListenAndServe(listener transport.Listener, proxies ...Proxy) (err error) {
	s := New(listener)
	for _, p := range proxies {
		if err = s.Proxy(p); err != nil {
			return
		}
	}
	err = s.ListenAndServe()
	return
}
