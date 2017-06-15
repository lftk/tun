package mux

import (
	"net"

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

func (l *Listener) Accept() (net.Conn, error) {
	return l.sess.AcceptStream()
}

func (l *Listener) Close() error {
	return l.sess.Close()
}
