package client

import (
	"net"
)

type TcpDialer struct {
	Addr string
}

func (d *TcpDialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", d.Addr)
}

func (d *TcpDialer) Close() error {
	return nil
}
