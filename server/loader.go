package server

import (
	"fmt"
	"net"

	"github.com/4396/tun/proxy"
)

type Loader interface {
	Proxy(p proxy.Proxy) error
	ProxyTCP(id string, port int) error
	ProxyHTTP(id, domain string) error
}

type loader struct {
	*Server
}

func (l *loader) Proxy(p proxy.Proxy) error {
	return l.service.Proxy(p)
}

func (l *loader) ProxyTCP(id string, port int) (err error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	p := proxy.Wrap(id, ln)
	err = l.Proxy(p)
	if err != nil {
		ln.Close()
	}
	return
}

func (l *loader) ProxyHTTP(id, domain string) (err error) {
	ln, err := l.muxer.Listen(domain)
	if err != nil {
		return
	}

	p := proxy.Wrap(id, ln)
	err = l.Proxy(p)
	if err != nil {
		ln.Close()
	}
	return
}
