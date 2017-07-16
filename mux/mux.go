package mux

import (
	"net"

	"github.com/xtaci/smux"
)

// Server wrap a conn to create a server-side session.
func Server(conn net.Conn) (s *Session, err error) {
	sess, err := smux.Server(conn, nil)
	if err != nil {
		return
	}

	s = &Session{sess: sess}
	return
}

// Client wrap a conn to create a client-side session.
func Client(conn net.Conn) (s *Session, err error) {
	sess, err := smux.Client(conn, nil)
	if err != nil {
		return
	}

	s = &Session{sess: sess}
	return
}
