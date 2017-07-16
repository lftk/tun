package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/sync/syncmap"
)

// Service manages all proxies,
// and handles all user connections,
// and records all incoming and outgoing traffic.
type Service struct {
	Traff   Traffic
	proxies syncmap.Map
	proxyc  chan Proxy
	connc   chan proxyConn
	errc    chan error
}

// proxyConn combines Proxy and net.Conn.
type proxyConn struct {
	Proxy
	net.Conn
}

// Serve all proxies
func (s *Service) Serve(ctx context.Context) (err error) {
	s.errc = make(chan error, 1)
	s.connc = make(chan proxyConn, 16)
	s.proxyc = make(chan Proxy, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		// close channel
		close(s.connc)
		close(s.proxyc)
		s.proxyc = nil

		// close proxy
		s.proxies.Range(func(key, val interface{}) bool {
			s.proxies.Delete(key)
			val.(Proxy).Close()
			return true
		})
	}()

	s.proxies.Range(func(key, val interface{}) bool {
		go s.listenProxy(ctx, val.(Proxy))
		return true
	})

	for {
		select {
		case p := <-s.proxyc:
			go s.listenProxy(ctx, p)
		case c := <-s.connc:
			go s.handleConn(ctx, c)
		case err = <-s.errc:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Proxy listens for a new proxy.
func (s *Service) Proxy(p Proxy) (err error) {
	_, loaded := s.proxies.LoadOrStore(p.ID(), p)
	if loaded {
		err = fmt.Errorf("proxy '%s' already exists", p.ID())
		return
	}

	if s.proxyc != nil {
		s.proxyc <- p
	}
	return
}

// Proxies returns all proxies.
func (s *Service) Proxies() (proxies []Proxy) {
	s.proxies.Range(func(key, val interface{}) bool {
		proxies = append(proxies, val.(Proxy))
		return true
	})
	return
}

// Load returns a proxy.
func (s *Service) Load(id string) (p Proxy, ok bool) {
	val, ok := s.proxies.Load(id)
	if ok {
		p = val.(Proxy)
	}
	return
}

// Kill a proxy.
func (s *Service) Kill(id string) {
	val, ok := s.proxies.Load(id)
	if ok {
		s.proxies.Delete(id)
		val.(Proxy).Close()
	}
	return
}

// Register a dialer and bind to the proxy.
func (s *Service) Register(id string, dialer Dialer) (err error) {
	if val, ok := s.proxies.Load(id); ok {
		err = val.(Proxy).Bind(dialer)
	} else {
		err = fmt.Errorf("proxy '%s' not exists", id)
	}
	return
}

// Unregister a dialer and unbind from the proxy.
func (s *Service) Unregister(id string, dialer Dialer) (err error) {
	if val, ok := s.proxies.Load(id); ok {
		err = val.(Proxy).Unbind(dialer)
	} else {
		err = fmt.Errorf("proxy '%s' not exists", id)
	}
	return
}

// listens for a new proxy.
func (s *Service) listenProxy(ctx context.Context, p Proxy) {
	for {
		conn, err := p.Accept()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			s.connc <- proxyConn{p, conn}
		}
	}
}

// handle the user's connection.
func (s *Service) handleConn(ctx context.Context, pc proxyConn) {
	select {
	case <-ctx.Done():
	default:
		err := pc.Proxy.Handle(pc.Conn, s.Traff)
		if err != nil {
			pc.Conn.Close()
		}
	}
}
