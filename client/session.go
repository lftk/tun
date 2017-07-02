package client

import (
	"errors"
	"net"
	"os"
	"runtime"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/version"
)

type session struct {
	*mux.Session
	cmd net.Conn
}

func newSession(conn net.Conn) (sess *session, err error) {
	ms, err := mux.Client(conn)
	if err != nil {
		return
	}

	cmd, err := ms.OpenConn()
	if err != nil {
		ms.Close()
		return
	}

	sess = &session{
		Session: ms,
		cmd:     cmd,
	}
	return
}

func (s *session) Accept() (net.Conn, error) {
	return s.Session.AcceptConn()
}

func (s *session) Addr() net.Addr {
	return nil
}

func (s *session) Close() error {
	return nil
}

func (s *session) Auth(name, token string) (err error) {
	ver := version.Version
	hostname, _ := os.Hostname()
	err = msg.Write(s.cmd, &msg.Proxy{
		Name:     name,
		Token:    token,
		Version:  ver,
		Hostname: hostname,
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	})
	if err != nil {
		return
	}

	m, err := msg.Read(s.cmd)
	if err != nil {
		return
	}

	switch mm := m.(type) {
	case *msg.Version:
		err = version.CompatServer(mm.Version)
	case *msg.Error:
		err = errors.New(mm.Message)
	default:
		err = ErrUnexpectedMsg
	}
	return
}
