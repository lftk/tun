package server

import (
	"fmt"
	"net"

	"github.com/4396/tun/fake"
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

type httpProxy struct {
	proxy.Proxy
	domain string
}

func (l *loader) ProxyHTTP(name, domain string) (err error) {
	ln := &fake.Listener{}
	p := proxy.Wrap(name, ln)
	err = l.Proxy(httpProxy{p, domain})
	if err != nil {
		ln.Close()
		return
	}

	l.muxer.HandleFunc(domain, ln.Put)
	return
}

func (l *loader) Proxy(p proxy.Proxy) error {
	return l.service.Proxy(p)
}
