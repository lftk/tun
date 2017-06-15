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
			err = s.proxy(m.Conn, msg)
		case *msg.Worker:
			err = s.worker(m.Conn, msg)
		case *msg.Cmder:
			err = s.cmder(m.Conn, msg)
		default:
			m.Conn.Close()
		}
		if err != nil {
			m.Conn.Close()
		}
	}
}

func (s *session) proxy(conn net.Conn, proxy *msg.Proxy) (err error) {
	if s.Server.Admin != nil {
		err = s.Server.Admin.AuthProxy(proxy.Name, proxy.Token)
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

func (s *session) cmder(conn net.Conn, cmder *msg.Cmder) (err error) {
	if s.Server.Admin == nil {
		// err = ...
		msg.ReplyError(conn, "")
		return
	}

	err = s.Server.Admin.AuthCmder(cmder.Token)
	if err != nil {
		msg.ReplyError(conn, err.Error())
		return
	}

	err = msg.ReplyOK(conn)
	if err != nil {
		return
	}

	go func() {
		defer conn.Close()
		for {
			var (
				cmd  msg.Command
				resp msg.CommandResp
			)

			err := msg.ReadInto(conn, &cmd)
			if err != nil {
				return
			}

			data, err := s.Server.Admin.Command(cmd.Data)
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Data = data
			}

			err = msg.Write(conn, &resp)
			if err != nil {
				return
			}
		}
	}()
	return
}

func (s *session) worker(conn net.Conn, worker *msg.Worker) (err error) {
	a, ok := s.agent[worker.Name]
	if !ok {
		// ...
		return
	}

	err = a.Put(conn)
	return
}
