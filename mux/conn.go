package mux

import (
	"io"
	"net"

	"github.com/golang/snappy"
)

// Use snappy wrap conn's read and write methods.
type conn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

// Read data from snappy.Reader.
func (c *conn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

// Write data to snappy.Writer.
func (c *conn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

// Use snappy wrap conn.
func withSnappy(c net.Conn) net.Conn {
	r := snappy.NewReader(c)
	w := snappy.NewWriter(c)
	return &conn{c, r, w}
}
