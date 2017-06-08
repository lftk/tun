package vhost

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/4396/tun/proxy"
	"github.com/golang/sync/syncmap"
)

type Muxer struct {
	connc     chan net.Conn
	listeners syncmap.Map
}

func (m *Muxer) Route(host string) net.Listener {
	listener := proxy.NewListener()
	actucal, _ := m.listeners.LoadOrStore(host, listener)
	return actucal.(net.Listener)
}

func (m *Muxer) Serve(l net.Listener) (err error) {
	m.connc = make(chan net.Conn, 16)

	go m.listen(l)
	for {
		select {
		case c := <-m.connc:
			go m.handleConn(c)
		}
	}
}

func (m *Muxer) listen(l net.Listener) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		m.connc <- conn
	}
}

func (m *Muxer) handleConn(c net.Conn) {
	domain, cc, err := ParseConn(c)
	if err == nil {
		val, ok := m.listeners.Load(domain)
		if ok {
			val.(*proxy.Listener).Put(cc)
			return
		}
	}
	// ...
	c.Close()
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

func ParseConn(c net.Conn) (host string, cc net.Conn, err error) {
	req, cc, err := ReadRequest(c)
	if err != nil {
		return
	}

	ss := strings.Split(req.Host, ":")
	if len(ss) > 0 {
		host = ss[0]
	}
	return
}
