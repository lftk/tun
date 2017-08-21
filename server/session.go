package server

import (
	"context"
	"errors"
	"net"

	"github.com/4396/tun/log"
	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/version"
	"github.com/golang/sync/syncmap"
)

// session describes the connection to the client.
// It also manages all of the proxies on that client.
type session struct {
	server  *Server
	session *mux.Session
	cmd     net.Conn
	proxies syncmap.Map
}

// newSession wrap a conn to create a session.
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
	}
	return
}

// Kill a proxy.
func (s *session) Kill(id string) (ok bool) {
	s.proxies.Range(func(key, val interface{}) bool {
		ok = (id == key.(string))
		if ok {
			log.Infof("proxy killed, id=%s", id)
			s.server.service.Kill(id)
		}
		return !ok
	})
	return
}

// Run this session and handle various commands.
func (s *session) Run(ctx context.Context) (err error) {
	defer func() {
		s.cmd.Close()
		s.session.Close()

		s.proxies.Range(func(key, val interface{}) bool {
			log.Infof("proxy exited, id=%s", key.(string))
			s.server.service.Kill(key.(string))
			return true
		})
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
			s.handleProxy(mm)
		default:
			log.Errorf("unexpected message, %+v", mm)
		}
	}
}

// handleProxy process the proxy login message.
func (s *session) handleProxy(proxy *msg.Proxy) {
	err := s.loadProxy(proxy)
	if err != nil {
		err = msg.Write(s.cmd, &msg.Error{Message: err.Error()})
		return
	}

	err = msg.Write(s.cmd, &msg.Version{Version: version.Version})
	if err != nil {
		log.Errorf("failed to reply version to client, proxy=%+v", proxy)
		s.server.service.Kill(proxy.ID)
		return
	}

	log.Infof("load proxy success, proxy=%+v", proxy)
	s.proxies.Store(proxy.ID, nil)
	return
}

// loadProxy auth and load a proxy.
func (s *session) loadProxy(proxy *msg.Proxy) (err error) {
	err = version.CompatClient(proxy.Version)
	if err != nil {
		log.Errorf("not compatible with client, proxy=%+v", proxy)
		return
	}

	if s.server.auth != nil {
		err = s.server.auth(proxy.ID, proxy.Token)
		if err != nil {
			log.Errorf("failed to auth proxy, err=%v, proxy=%+v", err, proxy)
			return
		}
	}

	_, ok := s.server.service.Load(proxy.ID)
	if !ok {
		if s.server.load != nil {
			err = s.server.load(&loader{s.server}, proxy.ID)
		} else {
			err = errors.New("no loader")
		}
	}
	if err != nil {
		log.Errorf("failed to load proxy, err=%v, proxy=%+v", err, proxy)
		return
	}

	err = s.server.service.Register(proxy.ID, &dialer{
		Session: s.session,
		ID:      proxy.ID,
	})
	if err != nil {
		log.Errorf("failed to register dialer, err=%v, proxy=%+v", err, proxy)
	}
	return
}
