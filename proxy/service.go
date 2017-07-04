package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/golang/sync/syncmap"
)

type Service struct {
	Traff   Traffic
	proxies syncmap.Map
	proxyc  chan Proxy
	connc   chan proxyConn
	errc    chan error
}

type proxyConn struct {
	Proxy
	net.Conn
}

func (s *Service) Serve(ctx context.Context) (err error) {
	s.errc = make(chan error, 1)
	s.connc = make(chan proxyConn, 16)
	s.proxyc = make(chan Proxy, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		// close channel
		//close(s.errc)
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

func (s *Service) Proxy(p Proxy) (err error) {
	_, loaded := s.proxies.LoadOrStore(p.Name(), p)
	if loaded {
		err = errors.New("Already existed")
		return
	}

	if s.proxyc != nil {
		s.proxyc <- p
	}
	return
}

func (s *Service) Proxies() (proxies []Proxy) {
	s.proxies.Range(func(key, val interface{}) bool {
		proxies = append(proxies, val.(Proxy))
		return true
	})
	return
}

func (s *Service) Load(name string) (p Proxy, ok bool) {
	val, ok := s.proxies.Load(name)
	if ok {
		p = val.(Proxy)
	}
	return
}

func (s *Service) Kill(name string) (p Proxy) {
	val, ok := s.proxies.Load(name)
	if ok {
		s.proxies.Delete(name)
		p = val.(Proxy)
		p.Close()
	}
	return
}

var ErrInvalidProxy = errors.New("Invalid proxy")

func (s *Service) Register(name string, dialer Dialer) (err error) {
	if val, ok := s.proxies.Load(name); ok {
		err = val.(Proxy).Bind(dialer)
	} else {
		err = ErrInvalidProxy
	}
	return
}

func (s *Service) Unregister(name string, dialer Dialer) (err error) {
	if val, ok := s.proxies.Load(name); ok {
		err = val.(Proxy).Unbind(dialer)
	} else {
		err = ErrInvalidProxy
	}
	return
}

func (s *Service) listenProxy(ctx context.Context, p Proxy) {
	for {
		conn, err := p.Accept()
		if err != nil {
			if err != ErrClosed {
				s.errc <- err
			}
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

func (s *Service) handleConn(ctx context.Context, pc proxyConn) {
	select {
	case <-ctx.Done():
	default:
		err := pc.Proxy.Handle(pc.Conn, s.Traff)
		if err != nil {
			fmt.Println(err)
			pc.Conn.Close()
		}
	}
}
