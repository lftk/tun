package client

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/4396/tun/proxy"
	"github.com/golang/sync/syncmap"
)

var (
	ErrDialerClosed  = errors.New("Dialer closed")
	ErrUnexpectedMsg = errors.New("Unexpected response")
)

type Client struct {
	addr     string
	service  proxy.Service
	handlers syncmap.Map
	errc     chan error
	sess     *session
}

func Dial(addr string) (c *Client, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	sess, err := newSession(conn)
	if err != nil {
		conn.Close()
		return
	}

	c = &Client{sess: sess}
	return
}

func (c *Client) Proxy(name, token, addr string) (err error) {
	err = c.sess.Auth(name, token)
	if err != nil {
		return
	}

	p := proxy.Wrap(name, c.sess)
	p.Bind(&dialer{Addr: addr})
	err = c.service.Proxy(p)
	return
}

func (c *Client) Run(ctx context.Context) (err error) {
	c.errc = make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(time.Second)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	go func() {
		err := c.service.Serve(ctx)
		if err != nil {
			c.errc <- err
		}
	}()

	for {
		select {
		case <-ticker.C:
			if c.sess.IsClosed() {
				err = ErrDialerClosed
				return
			}
		case err = <-c.errc:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}
