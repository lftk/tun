package client

import (
	"errors"
	"net"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/msg"
	"github.com/4396/tun/traffic"
	"github.com/golang/sync/syncmap"
	"github.com/xtaci/smux"
)

type Client struct {
	sess  *smux.Session
	lns   syncmap.Map
	lnc   chan net.Listener
	connc chan net.Conn
	donec chan interface{}

	dialers syncmap.Map
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
		sess: sess,
	}
	return
}

func (c *Client) Dialer(name string, dialer dialer.Dialer) {
	c.dialers.Store(name, dialer)
}

func (c *Client) Login(name, token string) {
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

	for {
		m, err := msg.Read(st)
		if err != nil {
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

			go c.handleConn(name, st)
		}
	}
}

func (c *Client) handleConn(name string, conn net.Conn) {
	defer conn.Close()

	var start msg.StartWorkConn
	err := msg.ReadInto(conn, &start)
	if err != nil {
		return
	}

	val, ok := c.dialers.Load(name)
	if !ok {
		return
	}

	work, err := val.(dialer.Dialer).Dial()
	if err != nil {
		return
	}

	traffic.Join(conn, work)
}

func (c *Client) Listen(l net.Listener) (err error) {
	_, loaded := c.lns.LoadOrStore(l.Addr().String(), l)
	if loaded {
		err = errors.New("already existed")
		return
	}

	if c.lnc != nil {
		c.lnc <- l
	}
	return
}

func (c *Client) Serve() (err error) {
	for {
		select {
		case l := <-c.lnc:
			go c.listen(l)
		}
	}
}

func (c *Client) listen(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		select {
		case <-c.donec:
			return
		default:
			c.connc <- conn
		}
	}
}
