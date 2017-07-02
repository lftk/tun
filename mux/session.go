package mux

import (
	"net"
	"time"

	"github.com/xtaci/smux"
)

type Session struct {
	sess *smux.Session
}

func (s *Session) OpenConn() (conn net.Conn, err error) {
	st, err := s.sess.OpenStream()
	if err != nil {
		return
	}

	conn = withSnappy(st)
	return
}

func (s *Session) AcceptConn() (conn net.Conn, err error) {
	st, err := s.sess.AcceptStream()
	if err != nil {
		return
	}

	conn = withSnappy(st)
	return
}

func (s *Session) NumConns() int {
	return s.sess.NumStreams()
}

func (s *Session) SetDeadline(t time.Time) error {
	return s.sess.SetDeadline(t)
}

func (s *Session) Close() error {
	return s.sess.Close()
}

func (s *Session) IsClosed() bool {
	return s.sess.IsClosed()
}
