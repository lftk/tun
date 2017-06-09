package proxy

import (
	"errors"
	"net"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/traffic"
)

var (
	ErrClosed = errors.New("Proxy closed")
)

type Proxy interface {
	Name() string
	Close() error
	Accept() (net.Conn, error)
	Bind(dialer.Dialer) error
	Unbind(dialer.Dialer) error
	Handle(net.Conn, traffic.Traffic) error
}
