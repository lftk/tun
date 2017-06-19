package proxy

import (
	"errors"
	"net"
)

var (
	ErrClosed = errors.New("Proxy closed")
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
	Name() string
	Close() error
	Accept() (net.Conn, error)
	Bind(Dialer) error
	Unbind(Dialer) error
	Handle(net.Conn, Traffic) error
}
