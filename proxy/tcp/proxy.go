package tcp

import (
	"net"

	"github.com/4396/tun/proxy"
)

func Proxy(name, addr string) (p proxy.Proxy, err error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		p = proxy.Wrap(name, listener)
	}
	return
}
