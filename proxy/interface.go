package proxy

import (
	"net"
)

type Traffic interface {
	In(string, []byte)
	Out(string, []byte)
}

type Dialer interface {
	Dial() (net.Conn, error)
	Close() error
}

type Proxy interface {
	ID() string
	Close() error
	Accept() (net.Conn, error)
	Bind(Dialer) error
	Unbind(Dialer) error
	Handle(net.Conn, Traffic) error
}
