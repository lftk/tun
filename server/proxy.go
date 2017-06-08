package server

import (
	"net"

	"github.com/4396/tun/proxy"
)

func tcpProxy(name, addr string) (p proxy.Proxy, err error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		p = proxy.Wrap(name, listener)
	}
	return
}

func httpProxy(name, domain string) (p proxy.Proxy, err error) {
	return
}
