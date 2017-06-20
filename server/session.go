package server

import (
	"context"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
)

type session struct {
	*Server
	net.Conn
	agent map[string]*agent
}

type message struct {
	net.Conn
	msg.Message
}

func (s session) loopMessage(ctx context.Context) {
	l, err := mux.Listen(s.Conn)
	if err != nil {
		s.Conn.Close()
		return
	}

	msgc := make(chan message, 16)
	s.agent = make(map[string]*agent)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		l.Close()
		close(msgc)

		for name, a := range s.agent {
			s.Server.service.Unregister(name, a)
		}
	}()

	go s.processMessage(ctx, msgc)

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		m, err := msg.Read(conn)
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			msgc <- message{conn, m}
		}
	}
}

func (s *session) processMessage(ctx context.Context, msgc <-chan message) {
	for m := range msgc {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		switch msg := m.Message.(type) {
		case *msg.Proxy:
			err = s.processProxy(m.Conn, msg)
		case *msg.Worker:
			err = s.processWorker(m.Conn, msg)
		default:
			m.Conn.Close()
		}
		if err != nil {
			m.Conn.Close()
		}
	}
}

func (s *session) processProxy(conn net.Conn, proxy *msg.Proxy) (err error) {
	if s.Server.auth != nil {
		err = s.Server.auth(proxy.Name, proxy.Token, proxy.Desc)
		if err != nil {
			msg.ReplyError(conn, err.Error())
			return
		}
	}

	a := &agent{
		conn:  conn,
		connc: make(chan net.Conn, 16),
	}
	err = s.Server.service.Register(proxy.Name, a)
	if err != nil {
		msg.ReplyError(conn, err.Error())
		return
	}

	err = msg.ReplyOK(conn)
	if err != nil {
		return
	}

	s.agent[proxy.Name] = a
	return
}

func (s *session) processWorker(conn net.Conn, worker *msg.Worker) (err error) {
	a, ok := s.agent[worker.Name]
	if !ok {
		// ...
		return
	}

	err = a.Put(conn)
	return
}
