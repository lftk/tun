package server

import (
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
	proxy    proxy.Service
	muxer    vhost.Muxer
	agents   syncmap.Map
	errc     chan error
	connc    chan net.Conn
	stc      chan *smux.Stream
	donec    chan interface{}
}

func (s *Server) TcpProxy(name, addr string) (err error) {
	p, err := tcpProxy(name, addr)
	if err == nil {
		err = s.Proxy(p)
	}
	return
}

func (s *Server) HttpProxy(name, domain string) (err error) {
	l := s.muxer.Route(domain)
	p := proxy.Wrap(name, l)
	err = s.Proxy(p)
	return
}

func (s *Server) Proxy(p proxy.Proxy) error {
	return s.proxy.Proxy(p)
}

func (s *Server) Proxies() []proxy.Proxy {
	return s.proxy.Proxies()
}

func (s *Server) Traffic(traff traffic.Traffic) {
	s.proxy.Traff = traff
}

func (s *Server) Shutdown() {
	if s.donec != nil {
		close(s.donec)
		s.donec = nil
	}
	s.proxy.Shutdown()
}

func (s *Server) ListenAndServe() (err error) {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return
	}

	m, err := net.Listen("tcp", s.HttpAddr)
	if err != nil {
		l.Close()
		return
	}

	err = s.Serve(l, m)
	return
}

func (s *Server) Serve(l, m net.Listener) (err error) {
	s.errc = make(chan error, 1)
	s.donec = make(chan interface{})
	s.connc = make(chan net.Conn, 16)
	s.stc = make(chan *smux.Stream, 16)

	go s.listen(l)
	go s.muxer.Serve(m)
	go s.proxy.Serve()

	for {
		select {
		case c := <-s.connc:
			go s.handleConn(c)
		case st := <-s.stc:
			go s.handleStream(st)
		case err = <-s.errc:
			s.Shutdown()
			return
		case <-s.donec:
			return
		}
	}
}

func (s *Server) listen(l net.Listener) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			s.errc <- err
			return
		}

		select {
		case <-s.donec:
		default:
			s.connc <- conn
		}
	}
}

func (s *Server) handleConn(c net.Conn) {
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
		case <-s.donec:
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

		a := NewAgent(st)
		err = s.proxy.Register(mm.Name, a)
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
		fmt.Printf("%+v\n", mm)
	}
}
