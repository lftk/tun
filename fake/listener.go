package fake

import (
	"errors"
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

func (l *Listener) Put(conn net.Conn) error {
	l.connc <- conn
	return nil
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	conn, ok := <-l.connc
	if !ok {
		err = errors.New("Listener closed")
	}
	return
}

func (l *Listener) Addr() net.Addr {
	return nil
}

func (l *Listener) Close() error {
	close(l.connc)
	for conn := range l.connc {
		conn.Close()
	}
	return nil
}
