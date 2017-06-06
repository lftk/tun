package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/traffic"
	"github.com/golang/sync/syncmap"
	"github.com/xtaci/smux"
)

type Server struct {
	Addr    string
	Traff   traffic.Traffic
	proxies syncmap.Map
	agents  syncmap.Map
	errc    chan error
	proxyc  chan proxy.Proxy
	connc   chan proxyConn
	tconnc  chan net.Conn
	stc     chan *smux.Stream
	exitc   chan interface{}
}

type proxyConn struct {
	proxy.Proxy
	net.Conn
}

func New(addr string) (s *Server) {
	s = &Server{Addr: addr}
	return
}

func (s *Server) Proxy(p proxy.Proxy) (err error) {
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

func (s *Server) Proxies() (proxies []proxy.Proxy) {
	s.proxies.Range(func(key, val interface{}) bool {
		proxies = append(proxies, val.(proxy.Proxy))
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
	s.proxyc = make(chan proxy.Proxy, 16)
	s.connc = make(chan proxyConn, 16)
	s.tconnc = make(chan net.Conn, 16)
	s.stc = make(chan *smux.Stream, 16)

	go s.listen(l)
	s.proxies.Range(func(key, val interface{}) bool {
		go s.listenProxy(val.(proxy.Proxy))
		return true
	})

	for {
		select {
		case p := <-s.proxyc:
			go s.listenProxy(p)
		case c := <-s.connc:
			go s.handleProxyConn(c)
		case t := <-s.tconnc:
			go s.handleClientConn(t)
		case st := <-s.stc:
			go s.handleStream(st)
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

func (s *Server) listenProxy(p proxy.Proxy) {
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
			s.connc <- proxyConn{Proxy: p, Conn: c}
		}
	}
}

func (s *Server) handleClientConn(c net.Conn) {
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
		case <-s.exitc:
		default:
			s.stc <- st
		}
	}
}

func (s *Server) handleStream(st *smux.Stream) {
	fmt.Println("stream", st.RemoteAddr())

	m, err := msg.Read(st)
	if err != nil {
		return
	}

	switch mm := m.(type) {
	case *msg.Login:
		fmt.Println("login", mm)

		val, ok := s.proxies.Load(mm.Name)
		if !ok {
			return
		}

		a := NewAgent(st)
		_, loaded := s.agents.LoadOrStore(mm.Name, a)
		if loaded {
			return
		}

		err = val.(proxy.Proxy).Bind(a)
		// ...
	case *msg.WorkConn:
		fmt.Println("work", mm)
		val, ok := s.agents.Load(mm.Name)
		if ok {
			val.(*Agent).Put(st)
		}
	default:
		fmt.Printf("%+v\n", mm)
	}
}

func (s *Server) handleProxyConn(c proxyConn) {
	fmt.Println("conn", c.RemoteAddr())
	err := c.Handle(c.Conn, s.Traff)
	if err != nil {
		// ...
	}
}

func ListenAndServe(addr string, proxies ...proxy.Proxy) (err error) {
	s := &Server{Addr: addr}
	for _, p := range proxies {
		if err = s.Proxy(p); err != nil {
			return
		}
	}
	err = s.ListenAndServe()
	return
}
