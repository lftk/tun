package proxy

import (
	"net"
)

type Listener struct {
	connc  chan net.Conn
	closed chan interface{}
}

func NewListener() *Listener {
	return &Listener{
		connc:  make(chan net.Conn, 16),
		closed: make(chan interface{}),
	}
}

func (l *Listener) Put(conn net.Conn) {
	select {
	case <-l.closed:
	default:
		l.connc <- conn
	}
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
			close(l.closed)
			return
		}
	}
}
