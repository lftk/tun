package proxy

import (
	"net"

	"github.com/4396/tun/traffic"
	"github.com/4396/tun/transport"
)

type Proxy interface {
	Name() string
	Close() error
	Accept() (net.Conn, error)
	Bind(transport.Dialer) error
	Unbind(transport.Dialer) error
	Handle(net.Conn, traffic.Traffic) error
}
