package vhost

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/golang/sync/syncmap"

	"github.com/4396/tun/fake"
)

type Muxer struct {
	net.Listener
	listeners syncmap.Map
}

func Listen(addr string) (m *Muxer, err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	m = &Muxer{Listener: l}
	return
}

func (m *Muxer) Listen(domain string) (l net.Listener, err error) {
	val, loaded := m.listeners.LoadOrStore(domain, &fake.Listener{})
	if loaded {
		err = errors.New("")
		return
	}

	l = val.(net.Listener)
	return
}

func (m *Muxer) Serve(ctx context.Context) (err error) {
	for {
		var conn net.Conn
		conn, err = m.Listener.Accept()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			m.handleConn(conn)
		}
	}
}

func (m *Muxer) handleConn(conn net.Conn) {
	domain, cc, err := subDomain(conn)
	if err != nil {
		conn.Close()
		return
	}

	val, ok := m.listeners.Load(domain)
	if ok {
		err = val.(*fake.Listener).Put(cc)
		if err == nil {
			return
		}
	}

	// 500
	conn.Close()
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
