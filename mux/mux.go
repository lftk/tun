package mux

import (
	"net"

	"github.com/xtaci/smux"
)

func Server(conn net.Conn) (s *Session, err error) {
	sess, err := smux.Server(conn, nil)
	if err != nil {
		return
	}

	s = &Session{sess: sess}
	return
}

func Client(conn net.Conn) (s *Session, err error) {
	sess, err := smux.Client(conn, nil)
	if err != nil {
		return
	}

	s = &Session{sess: sess}
	return
}
