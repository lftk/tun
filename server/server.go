package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/transport"
	"github.com/golang/sync/syncmap"
)

type Proxy interface {
	net.Listener
	Name() string
	Unbind(transport.Dialer)
	Bind(transport.Dialer, msg.Message) bool
	Handle(net.Conn, traffic.Traffic) error
}

type conn struct {
	Proxy
	net.Conn
}

type Server struct {
	Addr    string
	Traff   traffic.Traffic
	proxies syncmap.Map
	agents  syncmap.Map
	errc    chan error
	proxyc  chan Proxy
	connc   chan conn
	tconnc  chan net.Conn
	exitc   chan interface{}
}

func New(addr string) (s *Server) {
	s = &Server{Addr: addr}
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
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return
	}

	s.errc = make(chan error, 1)
	s.exitc = make(chan interface{})
	s.proxyc = make(chan Proxy, 16)
	s.connc = make(chan conn, 16)
	s.tconnc = make(chan net.Conn, 16)

	go s.listen(l)
	s.proxies.Range(func(key, val interface{}) bool {
		go s.listenProxy(val.(Proxy))
		return true
	})

	for {
		select {
		case p := <-s.proxyc:
			go s.listenProxy(p)
		case c := <-s.connc:
			go s.handleProxyConn(c)
		case t := <-s.tconnc:
			go s.handleConn(t)
		case err = <-s.errc:
			s.Shutdown()
			return
		case <-s.exitc:
			return
		}
	}
}

func (s *Server) listen(l net.Listener) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		select {
		case <-s.exitc:
		default:
			s.tconnc <- conn
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

func (s *Server) handleConn(c net.Conn) {
	fmt.Println("tran", c.RemoteAddr())

	// handshake message
	var m msg.Message
	if true {
		// login and auth
		a := NewAgent(c)
		name := c.RemoteAddr().String()
		_, loaded := s.agents.LoadOrStore(name, a)
		if loaded {
			return
		}

		s.proxies.Range(func(key, val interface{}) bool {
			ok := val.(Proxy).Bind(a, m)
			return !ok
		})
		return
	}

	// new workconn
	name := c.RemoteAddr().String()
	val, ok := s.agents.Load(name)
	if ok {
		val.(*Agent).Put(c)
	}
}

func (s *Server) handleProxyConn(c conn) {
	fmt.Println("conn", c.RemoteAddr())
	err := c.Handle(c.Conn, s.Traff)
	if err != nil {
		// ...
	}
}

func ListenAndServe(addr string, proxies ...Proxy) (err error) {
	s := &Server{Addr: addr}
	for _, p := range proxies {
		if err = s.Proxy(p); err != nil {
			return
		}
	}
	err = s.ListenAndServe()
	return
}
