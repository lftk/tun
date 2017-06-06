package dialer

import (
	"net"
)

type Dialer interface {
	Dial() (net.Conn, error)
	Close() error
}

type TcpDialer struct {
	Addr string
}

func (d *TcpDialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", d.Addr)
}

func (d *TcpDialer) Close() error {
	return nil
}

type UdpDialer struct {
	Addr string
}

func (d *UdpDialer) Dial() (net.Conn, error) {
	return net.Dial("udp", d.Addr)
}

func (d *UdpDialer) Close() error {
	return nil
}
