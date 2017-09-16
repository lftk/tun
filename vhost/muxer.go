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
	"sync"

	"github.com/4396/tun/fake"
)

// Muxer is used to manage all domain listeners.
type Muxer struct {
	listener net.Listener
	domains  sync.Map
}

// Listen an address to create a muxer.
func Listen(addr string) (m *Muxer, err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	m = &Muxer{listener: l}
	return
}

// Close this muxer and all domain listeners.
func (m *Muxer) Close() error {
	m.domains.Range(func(key, val interface{}) bool {
		val.(*fake.Listener).Close()
		m.domains.Delete(key)
		return true
	})
	return m.listener.Close()
}

// Listen a domain to create a net.Listener.
func (m *Muxer) Listen(domain string) (l net.Listener, err error) {
	l = &listener{
		Muxer:    m,
		Domain:   domain,
		Listener: fake.NewListener(16),
	}
	_, loaded := m.domains.LoadOrStore(domain, l)
	if loaded {
		err = errors.New("already listen the domain")
	}
	return
}

// Serve this muxer, resolves an http requests
// and put the connection into the domain listener.
func (m *Muxer) Serve(ctx context.Context) (err error) {
	var conn net.Conn
	for {
		conn, err = m.listener.Accept()
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

// handleConn is used to resolve an http requests
// and put the connection into the bdomain listener.
func (m *Muxer) handleConn(conn net.Conn) {
	req, bc, err := readRequest(conn)
	if err != nil {
		conn.Close()
		return
	}

	ss := strings.Split(req.Host, ":")
	val, ok := m.domains.Load(ss[0])
	if !ok {
		// 404
		bc.Close()
		return
	}

	err = val.(*listener).Put(bc)
	if err != nil {
		// 500
		bc.Close()
	}
}

// bufferConn combines bytes.Buffer and net.Conn.
type bufferConn struct {
	net.Conn
	*bytes.Buffer
}

// Read data from bytes.Buffer and net.Conn.
func (c bufferConn) Read(b []byte) (int, error) {
	return io.MultiReader(c.Buffer, c.Conn).Read(b)
}

// Write data to net.Conn.
func (c bufferConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

// readRequest resolves net.Conn to an http request.
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
