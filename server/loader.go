package server

import (
	"fmt"
	"net"

	"github.com/4396/tun/proxy"
)

// Loader is used to load proxy.
type Loader interface {
	Proxy(p proxy.Proxy) error
	ProxyTCP(id string, port int) error
	ProxyHTTP(id, domain string) error
}

// Loader has a specific implementation class is loader.
// But we only expose the Loader interface to the outside.
type loader struct {
	*Server
}

// Proxy is used to load a generic proxy.
func (l *loader) Proxy(p proxy.Proxy) error {
	return l.service.Proxy(p)
}

// ProxyTCP is used to load TCP proxy.
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

// ProxyHTTP is used to load HTTP proxy.
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
