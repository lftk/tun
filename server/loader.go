package server

import (
	"fmt"
	"net"

	"github.com/4396/tun/fake"
	"github.com/4396/tun/proxy"
)

type Loader struct {
	server *Server
}

func (l *Loader) ProxyTCP(name string, port int) (err error) {
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

func (l *Loader) ProxyHTTP(name, domain string) (err error) {
	ln := &fake.Listener{}
	p := proxy.Wrap(name, ln)
	err = l.Proxy(httpProxy{p, domain})
	if err != nil {
		ln.Close()
		return
	}

	l.server.muxer.HandleFunc(domain, ln.Put)
	return
}

func (l *Loader) Proxy(p proxy.Proxy) error {
	return l.server.service.Proxy(p)
}
