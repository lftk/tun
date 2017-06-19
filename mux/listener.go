package mux

import (
	"net"

	"github.com/4396/tun/conn"
	"github.com/xtaci/smux"
)

func Listen(conn net.Conn) (l *Listener, err error) {
	sess, err := smux.Server(conn, nil)
	if err != nil {
		return
	}

	l = &Listener{sess: sess}
	return
}

type Listener struct {
	sess *smux.Session
}

func (l *Listener) Accept() (c net.Conn, err error) {
	st, err := l.sess.AcceptStream()
	if err != nil {
		return
	}

	c = conn.WithSnappy(st)
	return
}

func (l *Listener) Close() error {
	return l.sess.Close()
}
