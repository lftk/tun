package client

import (
	"context"
	"net"

	"github.com/4396/tun/conn"
	"github.com/4396/tun/msg"
)

type handler struct {
	*Client
	Name     string
	Conn     net.Conn
	Listener *conn.Listener
}

func (h handler) loopMessage(ctx context.Context) {
	msgc := make(chan msg.Message, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		close(msgc)
		h.Conn.Close()
	}()

	go h.processMessage(ctx, msgc)

	for {
		m, err := msg.Read(h.Conn)
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			msgc <- m
		}
	}
}

func (h *handler) processMessage(ctx context.Context, msgc <-chan msg.Message) {
	for m := range msgc {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		switch mm := m.(type) {
		case *msg.Dial:
			err = h.processDial(mm)
			if err != nil {
				// ...
			}
		default:
		}
	}
}

func (h *handler) processDial(dial *msg.Dial) (err error) {
	conn, err := h.Client.Dialer.Dial()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Worker{
		Name: h.Name,
	})
	if err != nil {
		conn.Close()
		return
	}

	go func() {
		var start msg.StartWork
		err := msg.ReadInto(conn, &start)
		if err != nil {
			conn.Close()
			return
		}

		err = h.Listener.Put(conn)
		if err != nil {
			conn.Close()
		}
	}()
	return
}
