package mux

import (
	"io"
	"net"

	"github.com/golang/snappy"
)

type conn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (c *conn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func withSnappy(c net.Conn) net.Conn {
	r := snappy.NewReader(c)
	w := snappy.NewWriter(c)
	return &conn{c, r, w}
}
