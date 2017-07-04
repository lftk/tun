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
	lsn, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	p := proxy.Wrap(name, lsn)
	err = l.Proxy(p)
	if err != nil {
		lsn.Close()
	}
	return
}

type httpProxy struct {
	proxy.Proxy
	domain string
}

func (l *Loader) ProxyHTTP(name, domain string) (err error) {
	lsn := fake.NewListener()
	p := proxy.Wrap(name, lsn)
	err = l.Proxy(httpProxy{p, domain})
	if err != nil {
		lsn.Close()
		return
	}

	l.server.muxer.HandleFunc(domain, lsn.Put)
	return
}

func (l *Loader) Proxy(p proxy.Proxy) error {
	return l.server.service.Proxy(p)
}
