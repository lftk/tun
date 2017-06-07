package dialer

import (
	"net"
)

type Dialer interface {
	Dial() (net.Conn, error)
	Close() error
}
