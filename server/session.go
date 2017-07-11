package server

import (
	"container/list"
	"context"
	"errors"
	"net"
	"sync"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/version"
)

type session struct {
	server  *Server
	session *mux.Session
	cmd     net.Conn
	proxies *list.List
	locker  sync.Mutex
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
		server:  s,
		session: ms,
		cmd:     cmd,
		proxies: list.New(),
	}
	return
}

func (s *session) Kill(name string) (ok bool) {
	s.locker.Lock()
	defer s.locker.Unlock()

	for e := s.proxies.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == name {
			s.proxies.Remove(e)
			s.kill(name)
			ok = true
			return
		}
	}
	return
}

func (s *session) kill(name string) {
	s.server.service.Kill(name)
}

func (s *session) Run(ctx context.Context) (err error) {
	defer func() {
		s.cmd.Close()
		s.session.Close()

		s.locker.Lock()
		for e := s.proxies.Front(); e != nil; e = e.Next() {
			s.kill(e.Value.(string))
		}
		s.proxies.Init()
		s.locker.Unlock()
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
				s.kill(proxy.Name)
			} else {
				s.locker.Lock()
				s.proxies.PushBack(proxy.Name)
				s.locker.Unlock()
			}
		}
	}()

	err = version.CompatClient(proxy.Version)
	if err != nil {
		return
	}

	if s.server.auth != nil {
		err = s.server.auth(proxy.Name, proxy.Token)
		if err != nil {
			return
		}
	}

	_, ok := s.server.service.Load(proxy.Name)
	if !ok {
		if s.server.load != nil {
			err = s.server.load(&loader{s.server}, proxy.Name)
		} else {
			err = errors.New("Not exists proxy")
		}
	}

	d := &dialer{
		Session: s.session,
		Name:    proxy.Name,
	}
	err = s.server.service.Register(proxy.Name, d)
	return
}
