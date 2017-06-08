package server

import (
	"net"

	"github.com/4396/tun/proxy"
)

func tcpProxy(name, addr string) (p proxy.Proxy, err error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		p = proxy.Wrap(name, listener)
	}
	return
}

type Listener struct {
	connc chan net.Conn
}

func NewListener() *Listener {
	return &Listener{
		connc: make(chan net.Conn, 16),
	}
}

func (l *Listener) Put(conn net.Conn) {
	l.connc <- conn
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	conn = <-l.connc
	return
}

func (l *Listener) Addr() (addr net.Addr) {
	return
}

func (l *Listener) Close() (err error) {
	for {
		select {
		case conn := <-l.connc:
			conn.Close()
		default:
			close(l.connc)
			return
		}
	}
}

func httpProxy(name string) (p proxy.Proxy, err error) {
	p = proxy.Wrap(name, NewListener())
	return
}
