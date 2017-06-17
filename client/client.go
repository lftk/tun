package client

import (
	"context"

	"github.com/4396/tun/conn"
	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/proxy"
	"github.com/golang/sync/syncmap"
)

type Client struct {
	Dialer *mux.Dialer

	handlers syncmap.Map
	handlerc chan *handler
	service  proxy.Service
	errc     chan error
}

func (c *Client) Proxy(name, token, desc, addr string) (err error) {
	cc, err := c.Dialer.Dial()
	if err != nil {
		return
	}

	err = msg.Write(cc, &msg.Proxy{
		Name:  name,
		Token: token,
		Desc:  desc,
	})
	if err != nil {
		cc.Close()
		return
	}

	err = msg.Okay(cc)
	if err != nil {
		cc.Close()
		return
	}

	l := conn.NewListener()
	p := proxy.Wrap(name, l)
	p.Bind(&dialer{addr})
	err = c.service.Proxy(p)
	if err != nil {
		return
	}

	h := &handler{
		Client:   c,
		Name:     name,
		Conn:     cc,
		Listener: l,
	}
	c.handlers.Store(name, h)
	if c.handlerc != nil {
		c.handlerc <- h
	}
	return
}

func (c *Client) Serve(ctx context.Context) (err error) {
	c.errc = make(chan error, 1)
	c.handlerc = make(chan *handler, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	c.handlers.Range(func(key, val interface{}) bool {
		go val.(*handler).loopMessage(ctx)
		return true
	})

	go func() {
		err := c.service.Serve(ctx)
		if err != nil {
			c.errc <- err
		}
	}()

	for {
		select {
		case h := <-c.handlerc:
			go h.loopMessage(ctx)
		case err = <-c.errc:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}
