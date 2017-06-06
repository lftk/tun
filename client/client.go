package client

import (
	"errors"
	"fmt"
	"net"

	"github.com/4396/tun/msg"
	"github.com/golang/sync/syncmap"
	"github.com/xtaci/smux"
)

type Client struct {
	sess  *smux.Session
	lns   syncmap.Map
	lnc   chan net.Listener
	connc chan net.Conn
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
		sess: sess,
	}
	return
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

			go func(conn net.Conn) {
				defer conn.Close()

				var start msg.StartWorkConn
				err = msg.ReadInto(st, &start)
				if err != nil {
					return
				}

				b := make([]byte, 100)
				for {
					n, err := conn.Read(b)
					if err != nil {
						return
					}
					fmt.Println(string(b[:n]))

					_, err = conn.Write([]byte("hello"))
					if err != nil {
						return
					}
				}
			}(st)
		}
	}
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
