package vhost

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/golang/sync/syncmap"
)

type Muxer struct {
	handler syncmap.Map
}

func (m *Muxer) HandleFunc(domain string, handle func(net.Conn) error) {
	if handle != nil {
		m.handler.Store(domain, handle)
	} else {
		m.handler.Delete(domain)
	}
}

func (m *Muxer) Handle(c net.Conn) {
	fmt.Println("...Muxer...")

	domain, bc, err := subDomain(c)
	if err != nil {
		// 500
		fmt.Println("500")
		c.Close()
		return
	}

	val, ok := m.handler.Load(domain)
	if !ok {
		// 400
		fmt.Println("400")
		bc.Close()
		return
	}

	err = val.(func(net.Conn) error)(bc)
	if err != nil {
		// 500
		fmt.Println("500")
		bc.Close()
	}
}

type bufferConn struct {
	net.Conn
	*bytes.Buffer
}

func (c bufferConn) Read(b []byte) (int, error) {
	return io.MultiReader(c.Buffer, c.Conn).Read(b)
}

func (c bufferConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c bufferConn) Close() error {
	return c.Conn.Close()
}

func readRequest(c net.Conn) (req *http.Request, bc net.Conn, err error) {
	buf := bytes.NewBuffer(nil)
	tr := io.TeeReader(c, buf)
	br := bufio.NewReader(tr)
	req, err = http.ReadRequest(br)
	if err != nil {
		return
	}

	bc = bufferConn{Conn: c, Buffer: buf}
	return
}

func subDomain(c net.Conn) (sub string, bc net.Conn, err error) {
	req, bc, err := readRequest(c)
	if err != nil {
		return
	}

	i := strings.Index(req.Host, ".")
	if i != -1 {
		sub = req.Host[:i]
	}
	return
}
