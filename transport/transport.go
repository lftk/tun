package transport

import (
	"net"
)

type Dialer interface {
	Dial() (net.Conn, error)
	Close() error
}

type Listener interface {
	Accept() (net.Conn, error)
	Close() error
}
