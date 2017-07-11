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
	connc    chan net.Conn
	once     sync.Once
	die      chan interface{}
	dieMutex sync.Mutex
}

func (l *Listener) lazyInit() {
	l.once.Do(func() {
		l.connc = make(chan net.Conn, 16)
	})
}

func (l *Listener) Put(conn net.Conn) (err error) {
	select {
	case <-l.die:
		err = ErrClosed
	default:
		l.lazyInit()
		l.connc <- conn
	}
	return
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

func (l *Listener) Close() (err error) {
	l.dieMutex.Lock()

	select {
	case <-l.die:
		l.dieMutex.Unlock()
		err = ErrClosed
	default:
		close(l.die)
		l.dieMutex.Unlock()

		l.lazyInit()
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
