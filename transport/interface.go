package transport

import (
	"errors"
	"net"
)

type Transport interface {
	net.Conn
}

type Listener interface {
	Accept() (Transport, error)
	Close() error
}

type Dialer interface {
	Dial() (Transport, error)
	Close() error
}

var (
	ErrListenerClosed = errors.New("Listener closed")
	ErrDialerClosed   = errors.New("Dialer closed")
)
