package client

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/xtaci/smux"
)

type Client struct {
	ServerAddr string
	LocalAddr  string
	Name       string
	Token      string

	service  proxy.Service
	listener *proxy.Listener
	sess     *smux.Session
	conn     *smux.Stream
	errc     chan error
	msgc     chan msg.Message
}

func (c *Client) handleWorkConn(conn net.Conn) {
	var start msg.StartWorkConn
	err := msg.ReadInto(conn, &start)
	if err != nil {
		fmt.Println("msg.StartWorkConn..", err)
		return
	}

	err = c.listener.Put(conn)
	if err != nil {
		conn.Close()
		// ...
	}
}

func (c *Client) DialAndServe(ctx context.Context) (err error) {
	conn, err := net.Dial("tcp", c.ServerAddr)
	if err != nil {
		return
	}
	err = c.Serve(ctx, conn)
	return
}

func (c *Client) init(conn net.Conn) (err error) {
	sess, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return
	}

	st, err := sess.OpenStream()
	if err != nil {
		sess.Close()
		return
	}

	err = auth(st, c.Name, c.Token)
	if err != nil {
		st.Close()
		sess.Close()
		return
	}

	l := proxy.NewListener()
	p := proxy.Wrap(c.Name, l)
	p.Bind(NewDialer(c.LocalAddr))
	c.service.Proxy(p)

	c.conn = st
	c.sess = sess
	c.listener = l
	c.errc = make(chan error, 1)
	c.msgc = make(chan msg.Message, 16)
	return
}

func (c *Client) uninit() {
	c.conn.Close()
	c.sess.Close()

	// close channel
	close(c.errc)
	close(c.msgc)
}

func (c *Client) Serve(ctx context.Context, conn net.Conn) (err error) {
	err = c.init(conn)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		c.uninit()
	}()

	go c.loopMessage(ctx)
	go func() {
		err := c.service.Serve(ctx)
		if err != nil {
			c.errc <- err
		}
	}()

	for {
		select {
		case m := <-c.msgc:
			go c.handleMessage(ctx, m)
		case err = <-c.errc:
			return
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	}
}

func (c *Client) loopMessage(ctx context.Context) {
	for {
		m, err := msg.Read(c.conn)
		if err != nil {
			c.errc <- err
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			c.msgc <- m
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, m msg.Message) {
	select {
	case <-ctx.Done():
	default:
		switch m.(type) {
		case *msg.Dial:
			st, err := c.sess.OpenStream()
			if err != nil {
				return
			}
			err = msg.Write(st, &msg.WorkConn{})
			if err != nil {
				st.Close()
				return
			}
			go c.handleWorkConn(st)
		default:
		}
	}
}

func auth(conn net.Conn, name, token string) (err error) {
	err = msg.Write(conn, &msg.Auth{
		Name:  name,
		Token: token,
	})
	if err != nil {
		return
	}

	var resp msg.AuthResp
	err = msg.ReadInto(conn, &resp)
	if err != nil {
		return
	}

	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return
}
