package server

import (
	"context"
	"net"

	"github.com/4396/tun/msg"
	"github.com/xtaci/smux"
)

type session struct {
	*Server
	net.Conn
	agent map[string]*agent
}

type streamMessage struct {
	*smux.Stream
	msg.Message
}

func (s session) loopMessage(ctx context.Context) {
	sess, err := smux.Server(s.Conn, nil)
	if err != nil {
		s.Conn.Close()
		return
	}

	s.agent = make(map[string]*agent)
	stmc := make(chan streamMessage, 16)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		close(stmc)
		sess.Close()

		for name, a := range s.agent {
			s.Server.service.Unregister(name, a)
		}
	}()

	go s.processMessage(ctx, stmc)

	for {
		st, err := sess.AcceptStream()
		if err != nil {
			return
		}

		m, err := msg.Read(st)
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			stmc <- streamMessage{st, m}
		}
	}
}

func (s *session) processMessage(ctx context.Context, stmc <-chan streamMessage) {
	for stm := range stmc {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		switch m := stm.Message.(type) {
		case *msg.Proxy:
			err = s.proxy(stm.Stream, m)
		case *msg.Worker:
			err = s.worker(stm.Stream, m)
		case *msg.Cmder:
			err = s.cmder(stm.Stream, m)
		default:
			stm.Stream.Close()
		}
		if err != nil {
			stm.Stream.Close()
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
