package server

import (
	"context"
	"errors"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/version"
)

type session struct {
	*Server
	*mux.Session
	cmd     net.Conn
	proxies []string
}

func newSession(s *Server, conn net.Conn) (sess *session, err error) {
	ms, err := mux.Server(conn)
	if err != nil {
		return
	}

	cmd, err := ms.AcceptConn()
	if err != nil {
		ms.Close()
		return
	}

	sess = &session{
		Server:  s,
		Session: ms,
		cmd:     cmd,
	}
	return
}

func (s *session) Run(ctx context.Context) (err error) {
	defer func() {
		s.cmd.Close()
		s.Session.Close()
		for _, proxy := range s.proxies {
			s.Server.Kill(proxy)
		}
	}()

	for {
		var m msg.Message
		m, err = msg.Read(s.cmd)
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		switch mm := m.(type) {
		case *msg.Proxy:
			err = s.handleProxy(mm)
		default:
			// ...
		}

		if err != nil {
			return
		}
	}
}

func (s *session) handleProxy(proxy *msg.Proxy) (err error) {
	defer func() {
		if err != nil {
			err = msg.Write(s.cmd, &msg.Error{Message: err.Error()})
		} else {
			err = msg.Write(s.cmd, &msg.Version{Version: version.Version})
			if err != nil {
				s.Server.Kill(proxy.Name)
			} else {
				s.proxies = append(s.proxies, proxy.Name)
			}
		}
	}()

	err = version.CompatClient(proxy.Version)
	if err != nil {
		return
	}

	if s.Server.auth != nil {
		err = s.Server.auth(proxy.Name, proxy.Token)
		if err != nil {
			return
		}
	}

	_, ok := s.Server.service.Load(proxy.Name)
	if !ok {
		if s.Server.load != nil {
			err = s.Server.load(&loader{s.Server}, proxy.Name)
		} else {
			err = errors.New("Not exists proxy")
		}
	}

	d := &dialer{
		Session: s.Session,
		Name:    proxy.Name,
	}
	err = s.Server.service.Register(proxy.Name, d)
	return
}
