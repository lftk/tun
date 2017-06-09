package vhost

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/golang/sync/syncmap"
)

type Muxer struct {
	connc   chan net.Conn
	handler syncmap.Map
}

func (m *Muxer) HandleFunc(domain string, handle func(net.Conn)) {
	if handle != nil {
		m.handler.Store(domain, handle)
	} else {
		m.handler.Delete(domain)
	}
}

func (m *Muxer) Handle(c net.Conn) {
	domain, cc, err := SubDomain(c)
	if err != nil {
		// 500
		c.Close()
		return
	}

	val, ok := m.handler.Load(domain)
	if !ok {
		// 400
		cc.Close()
		return
	}

	val.(func(net.Conn))(cc)
}

type bufferConn struct {
	net.Conn
	buf *bytes.Buffer
}

func (c bufferConn) Read(b []byte) (int, error) {
	return io.MultiReader(c.buf, c.Conn).Read(b)
}

func (c bufferConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c bufferConn) Close() error {
	return c.Conn.Close()
}

func ReadRequest(c net.Conn) (req *http.Request, cc net.Conn, err error) {
	buf := bytes.NewBuffer(nil)
	tr := io.TeeReader(c, buf)
	br := bufio.NewReader(tr)
	req, err = http.ReadRequest(br)
	if err != nil {
		return
	}

	cc = bufferConn{
		Conn: c,
		buf:  buf,
	}
	return
}

func SubDomain(c net.Conn) (sub string, cc net.Conn, err error) {
	req, cc, err := ReadRequest(c)
	if err != nil {
		return
	}

	i := strings.Index(req.Host, ".")
	if i != -1 {
		sub = req.Host[:i]
	}
	return
}
