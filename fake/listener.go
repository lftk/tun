package fake

import (
	"errors"
	"net"
	"sync"
)

// Listener has been closed.
var errClosed = errors.New("listener closed")

// Listener is an implementation of net.Listener.
// It receives incoming net.Conn through the Put method
// and returns net.Conn through the Accept method.
type Listener struct {
	connc chan net.Conn
	die   chan interface{}
	mu    sync.Mutex
}

// NewListener returns a listener to build a net.Conn channel of pool size.
func NewListener(pool int) *Listener {
	return &Listener{
		die:   make(chan interface{}),
		connc: make(chan net.Conn, pool),
	}
}

// Put a net.Conn to listener.
func (l *Listener) Put(conn net.Conn) (err error) {
	select {
	case <-l.die:
		err = errClosed
	default:
		l.connc <- conn
	}
	return
}

// Accept returns a net.Conn.
func (l *Listener) Accept() (conn net.Conn, err error) {
	select {
	case <-l.die:
		err = errClosed
	default:
		var ok bool
		conn, ok = <-l.connc
		if !ok {
			err = errClosed
		}
	}
	return
}

// Addr returns a fake addr.
func (l *Listener) Addr() net.Addr {
	return NewAddr("")
}

// Close all the net.Conn.
func (l *Listener) Close() (err error) {
	l.mu.Lock()

	select {
	case <-l.die:
		l.mu.Unlock()
		err = errClosed
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

// IsClosed returns true if listener is closed.
func (l *Listener) IsClosed() bool {
	select {
	case <-l.die:
		return true
	default:
		return false
	}
}
