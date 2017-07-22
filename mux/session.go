package mux

import (
	"net"
	"time"

	"github.com/xtaci/smux"
)

// Session wrap a smux.session, that is a multiplexed connection.
type Session struct {
	sess *smux.Session
}

// OpenConn creates a new connection.
func (s *Session) OpenConn() (conn net.Conn, err error) {
	st, err := s.sess.OpenStream()
	if err != nil {
		return
	}

	conn = withSnappy(st)
	return
}

// AcceptConn receives a new connection.
func (s *Session) AcceptConn() (conn net.Conn, err error) {
	st, err := s.sess.AcceptStream()
	if err != nil {
		return
	}

	conn = withSnappy(st)
	return
}

// NumConns returns the number of established connections.
func (s *Session) NumConns() int {
	return s.sess.NumStreams()
}

// SetDeadline sets a deadline used by Accept* calls.
func (s *Session) SetDeadline(t time.Time) error {
	return s.sess.SetDeadline(t)
}

// Close this session and all established connections.
func (s *Session) Close() error {
	return s.sess.Close()
}

// IsClosed to determine whether this session was closed.
func (s *Session) IsClosed() bool {
	return s.sess.IsClosed()
}
