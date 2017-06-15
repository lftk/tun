package client

import (
	"net"
)

type dialer struct {
	Addr string
}

func (d *dialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", d.Addr)
}

func (d *dialer) Close() error {
	return nil
}
