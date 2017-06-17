package conn

import (
	"fmt"
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

func (l *Listener) Put(conn net.Conn) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%v", x)
		}
	}()
	l.connc <- conn
	return
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	conn = <-l.connc
	return
}

func (l *Listener) Addr() (addr net.Addr) {
	return
}

func (l *Listener) Close() (err error) {
	close(l.connc)
	for conn := range l.connc {
		conn.Close()
	}
	return
}
