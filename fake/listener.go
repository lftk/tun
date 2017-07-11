package fake

import (
	"errors"
	"net"
	"sync"
)

var (
	ErrClosed = errors.New("listener closed")
)

type Listener struct {
	connc chan net.Conn
	die   chan interface{}
	mu    sync.Mutex
}

func NewListener() *Listener {
	return &Listener{
		die:   make(chan interface{}),
		connc: make(chan net.Conn, 16),
	}
}

func (l *Listener) Put(conn net.Conn) (err error) {
	select {
	case <-l.die:
		err = ErrClosed
	default:
		l.connc <- conn
	}
	return
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	select {
	case <-l.die:
		err = ErrClosed
	default:
		var ok bool
		conn, ok = <-l.connc
		if !ok {
			err = ErrClosed
		}
	}
	return
}

func (l *Listener) Addr() net.Addr {
	return &Addr{}
}

func (l *Listener) Close() (err error) {
	l.mu.Lock()

	select {
	case <-l.die:
		l.mu.Unlock()
		err = ErrClosed
	default:
		close(l.die)
		l.mu.Unlock()

		close(l.connc)
		for conn := range l.connc {
			conn.Close()
		}
	}
	return
}

func (l *Listener) IsClosed() bool {
	select {
	case <-l.die:
		return true
	default:
		return false
	}
}
