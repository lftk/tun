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
	handlers syncmap.Map
}

type HandlerFunc func(net.Conn) error

func (m *Muxer) Handler(domain string) (handler HandlerFunc) {
	val, ok := m.handlers.Load(domain)
	if ok {
		handler = val.(HandlerFunc)
	}
	return
}

func (m *Muxer) HandleFunc(domain string, handler HandlerFunc) {
	if handler != nil {
		m.handlers.Store(domain, handler)
	} else {
		m.handlers.Delete(domain)
	}
}

func (m *Muxer) Handle(c net.Conn) {
	domain, bc, err := subDomain(c)
	if err != nil {
		// 500
		fmt.Println("1-500", err)
		c.Close()
		return
	}

	val, ok := m.handlers.Load(domain)
	if !ok {
		// 400
		fmt.Println("400")
		bc.Close()
		return
	}

	//bc.Close()
	//return

	err = val.(func(net.Conn) error)(bc)
	if err != nil {
		// 500
		fmt.Println("2-500", err)
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
