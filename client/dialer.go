package client

import (
	"net"
)

type Dialer struct {
	addr string
}

func NewDialer(addr string) *Dialer {
	return &Dialer{addr: addr}
}

func (d *Dialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", d.addr)
}

func (d *Dialer) Close() error {
	return nil
}
