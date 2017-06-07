package client

import (
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/proxy"
	"github.com/golang/sync/syncmap"
	"github.com/xtaci/smux"
)

type Client struct {
	Addr  string
	proxy proxy.Service
	sess  *smux.Session
	lns   syncmap.Map
	donec chan interface{}
}

func Dial(addr string) (c *Client, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	sess, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return
	}

	c = &Client{
		Addr:  addr,
		sess:  sess,
		donec: make(chan interface{}),
	}
	return
}

func (c *Client) Proxy(name, token, addr string) {
	st, err := c.sess.OpenStream()
	if err != nil {
		return
	}

	err = msg.Write(st, &msg.Login{
		Name:  name,
		Token: token,
	})
	if err != nil {
		return
	}

	l := NewListener()
	p := proxy.Wrap(name, l)
	err = c.proxy.Proxy(p)
	if err != nil {
		return
	}

	c.lns.Store(name, l)
	p.Bind(NewDialer(addr))

	for {
		m, err := msg.Read(st)
		if err != nil {
			fmt.Println("Proxy", err)
			return
		}

		switch m.(type) {
		case *msg.Dial:
			st, err := c.sess.OpenStream()
			if err != nil {
				return
			}
			err = msg.Write(st, &msg.WorkConn{
				Name:  name,
				Token: token,
			})
			if err != nil {
				return
			}

			go c.handleConn(name, st, l)
		}
	}
}

func (c *Client) handleConn(name string, conn net.Conn, l *Listener) {
	var start msg.StartWorkConn
	err := msg.ReadInto(conn, &start)
	if err != nil {
		return
	}
	l.Put(conn)
}

func (c *Client) Serve() (err error) {
	err = c.proxy.Serve()
	return
}

func (c *Client) Shutdown() {
	c.proxy.Shutdown()
	c.lns.Range(func(key, val interface{}) bool {
		val.(*Listener).Close()
		return true
	})
}
