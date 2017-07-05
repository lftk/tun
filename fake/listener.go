package fake

import (
	"errors"
	"net"
	"sync"
)

type Listener struct {
	connc chan net.Conn
	once  sync.Once
}

func (l *Listener) lazyInit() {
	l.once.Do(func() {
		l.connc = make(chan net.Conn, 16)
	})
}

func (l *Listener) Put(conn net.Conn) error {
	l.lazyInit()
	l.connc <- conn
	return nil
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	l.lazyInit()
	conn, ok := <-l.connc
	if !ok {
		err = errors.New("Listener closed")
	}
	return
}

func (l *Listener) Addr() net.Addr {
	return &Addr{}
}

func (l *Listener) Close() error {
	l.lazyInit()
	close(l.connc)
	for conn := range l.connc {
		conn.Close()
	}
	return nil
}
