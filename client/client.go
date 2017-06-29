package client

import (
	"context"
	"errors"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/4396/tun/conn"
	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/version"
	"github.com/golang/sync/syncmap"
)

var (
	ErrDialerClosed  = errors.New("Dialer closed")
	ErrUnexpectedMsg = errors.New("Unexpected response")
)

type Client struct {
	dialer   *mux.Dialer
	service  proxy.Service
	handlers syncmap.Map
	handlerc chan *handler
	errc     chan error
}

func Dial(addr string) (c *Client, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	dialer, err := mux.Dial(conn)
	if err != nil {
		return
	}

	c = &Client{dialer: dialer}
	return
}

func auth(conn net.Conn, name, token, desc, addr string) (err error) {
	ver := version.Version
	hostname, _ := os.Hostname()
	err = msg.Write(conn, &msg.Proxy{
		Name:     name,
		Token:    token,
		Desc:     desc,
		Version:  ver,
		Hostname: hostname,
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	})
	if err != nil {
		return
	}

	m, err := msg.Read(conn)
	if err != nil {
		return
	}

	switch mm := m.(type) {
	case *msg.Version:
		err = version.CompatServer(mm.Version)
	case *msg.Error:
		err = errors.New(mm.Message)
	default:
		err = ErrUnexpectedMsg
	}
	return
}

func (c *Client) Proxy(name, token, desc, addr string) (err error) {
	cc, err := c.dialer.Dial()
	if err != nil {
		return
	}

	err = auth(cc, name, token, desc, addr)
	if err != nil {
		cc.Close()
		return
	}

	l := conn.NewListener()
	p := proxy.Wrap(name, l)
	p.Bind(&dialer{addr})
	err = c.service.Proxy(p)
	if err != nil {
		cc.Close()
		return
	}

	h := &handler{
		Name:     name,
		Conn:     cc,
		Listener: l,
		Dialer:   c.dialer,
	}
	c.handlers.Store(name, h)
	if c.handlerc != nil {
		c.handlerc <- h
	}
	return
}

func (c *Client) Run(ctx context.Context) (err error) {
	c.errc = make(chan error, 1)
	c.handlerc = make(chan *handler, 16)
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(time.Second)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	c.handlers.Range(func(key, val interface{}) bool {
		go val.(*handler).LoopMessage(ctx)
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
		case <-ticker.C:
			if c.dialer.IsClosed() {
				err = ErrDialerClosed
				return
			}
		case h := <-c.handlerc:
			go h.LoopMessage(ctx)
		case err = <-c.errc:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}
