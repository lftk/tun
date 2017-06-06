package udp

import (
	"bytes"
	"io"
	"net"
	"time"
)

type Conn struct {
	listener *Listener
	addr     *net.UDPAddr
	data     *bytes.Buffer
}

func (c *Conn) Read(b []byte) (int, error) {
	return io.ReadFull(c.data, b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.listener.WriteToUDP(b, c.addr)
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return c.listener.Addr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
