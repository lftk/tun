package conn

import (
	"errors"
	"fmt"
	"net"
)

type Listener struct {
	connc chan net.Conn
}

func NewListener() *Listener {
	return &Listener{
		connc: make(chan net.Conn, 16),
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
	var ok bool
	conn, ok = <-l.connc
	if !ok {
		err = errors.New("Listener closed")
	}
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
