package server

import (
	"fmt"
	"net"

	"github.com/4396/tun/proxy"
)

type Loader interface {
	ProxyTCP(name string, port int) error
	ProxyHTTP(name, domain string) error
	Proxy(proxy.Proxy) error
}

type loader struct {
	*Server
}

func (l *loader) ProxyTCP(name string, port int) (err error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	p := proxy.Wrap(name, ln)
	err = l.Proxy(p)
	if err != nil {
		ln.Close()
	}
	return
}

func (l *loader) ProxyHTTP(name, domain string) (err error) {
	ln, err := l.muxer.Listen(domain)
	if err != nil {
		return
	}

	p := proxy.Wrap(name, ln)
	err = l.Proxy(p)
	if err != nil {
		ln.Close()
	}
	return
}

func (l *loader) Proxy(p proxy.Proxy) error {
	return l.service.Proxy(p)
}