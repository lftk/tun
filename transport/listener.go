package transport

import (
	"net"

	"github.com/xtaci/smux"
)

type muxListener struct {
	net.Listener
	errc  chan error
	stc   chan *smux.Stream
	donec chan interface{}
}

func MuxListen(addr string) (l Listener, err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	l = &muxListener{
		Listener: listener,
		errc:     make(chan error, 1),
		stc:      make(chan *smux.Stream, 1024),
		donec:    make(chan interface{}),
	}
	go l.(*muxListener).loopAccept()
	return
}

func (l *muxListener) loopAccept() {
	conf := smux.DefaultConfig()
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			l.errc <- err
			return
		}

		go func(conn net.Conn) {
			defer conn.Close()

			sess, err := smux.Server(conn, conf)
			if err != nil {
				return
			}

			for {
				st, err := sess.AcceptStream()
				if err != nil {
					return
				}
				l.stc <- st
			}
		}(conn)
	}
}

func (l *muxListener) Accept() (t Transport, err error) {
	select {
	case t = <-l.stc:
	case err = <-l.errc:
	case <-l.donec:
		err = ErrListenerClosed
	}
	return
}

func (l *muxListener) Close() (err error) {
	if l.Listener != nil {
		err = l.Listener.Close()
		close(l.donec)
	}
	return
}
