package proxy

import (
	"net"
)

// Traffic is used to record incoming and outgoing traffic.
type Traffic interface {
	In(string, []byte)
	Out(string, []byte)
}

// Dialer establishes a working connection with the remote service.
type Dialer interface {
	Dial() (net.Conn, error)
	Close() error
}

// Proxy combination listener and dialer.
type Proxy interface {
	ID() string
	Close() error
	Bind(Dialer) error
	Unbind(Dialer) error
	Accept() (net.Conn, error)
	Handle(net.Conn, Traffic) error
}
