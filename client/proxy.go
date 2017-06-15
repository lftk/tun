package client

import (
	"context"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
)

type Proxy struct {
	*Client
	Name     string
	Conn     net.Conn
	Listener *proxy.Listener
}

func (p Proxy) loopMessage(ctx context.Context) {
	msgc := make(chan msg.Message, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		close(msgc)
		p.Conn.Close()
	}()

	go p.processMessage(ctx, msgc)

	for {
		m, err := msg.Read(p.Conn)
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

func (p *Proxy) processMessage(ctx context.Context, msgc <-chan msg.Message) {
	for m := range msgc {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		switch mm := m.(type) {
		case *msg.Dial:
			err = p.dial(mm)
			if err != nil {
				// ...
			}
		default:
		}
	}
}

func (p *Proxy) dial(dial *msg.Dial) (err error) {
	conn, err := p.Client.Dialer.Dial()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Worker{
		Name: p.Name,
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

		err = p.Listener.Put(conn)
		if err != nil {
			conn.Close()
		}
	}()
	return
}
