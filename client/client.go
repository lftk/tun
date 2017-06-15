package client

import (
	"context"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/proxy"
	"github.com/golang/sync/syncmap"
)

type Client struct {
	Dialer *mux.Dialer

	proxys  syncmap.Map
	proxyc  chan *Proxy
	service proxy.Service
	errc    chan error
}

func (c *Client) Proxy(name, token, addr string) (err error) {
	conn, err := c.Dialer.Dial()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Proxy{
		Name:  name,
		Token: token,
	})
	if err != nil {
		conn.Close()
		return
	}

	err = msg.Okay(conn)
	if err != nil {
		conn.Close()
		return
	}

	l := proxy.NewListener()
	p := proxy.Wrap(name, l)
	p.Bind(&dialer{addr})
	err = c.service.Proxy(p)
	if err != nil {
		return
	}

	pp := &Proxy{
		Client:   c,
		Name:     name,
		Conn:     conn,
		Listener: l,
	}
	c.proxys.Store(name, pp)
	if c.proxyc != nil {
		c.proxyc <- pp
	}
	return
}

func (c *Client) Serve(ctx context.Context) (err error) {
	c.errc = make(chan error, 1)
	c.proxyc = make(chan *Proxy, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	c.proxys.Range(func(key, val interface{}) bool {
		go val.(*Proxy).loopMessage(ctx)
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
		case p := <-c.proxyc:
			go p.loopMessage(ctx)
		case err = <-c.errc:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}
