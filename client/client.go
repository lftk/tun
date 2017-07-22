package client

import (
	"context"
	"errors"
	"net"
	"os"
	"runtime"

	"github.com/4396/tun/fake"
	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
	"github.com/4396/tun/proxy"
	"github.com/4396/tun/version"
	"github.com/golang/sync/syncmap"
)

// Error constants
var (
	// The network exception or the server shutdown caused the session to be closed.
	ErrSessionClosed = errors.New("session closed")
)

// Client connects to the server through the Dial function.
// Multiple proxies can be established with the server.
type Client struct {
	service   proxy.Service
	session   *mux.Session
	listeners syncmap.Map
	cmd       net.Conn
	errc      chan error
}

// Dial returns a client that establishes a session with addr.
func Dial(addr string) (c *Client, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	sess, err := mux.Client(conn)
	if err != nil {
		conn.Close()
		return
	}

	cmd, err := sess.OpenConn()
	if err != nil {
		sess.Close()
		conn.Close()
		return
	}

	c = &Client{
		session: sess,
		cmd:     cmd,
		errc:    make(chan error, 1),
	}
	return
}

func (c *Client) authProxy(id, token string) (err error) {
	ver := version.Version
	hostname, _ := os.Hostname()
	err = msg.Write(c.cmd, &msg.Proxy{
		ID:       id,
		Token:    token,
		Version:  ver,
		Hostname: hostname,
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	})
	if err != nil {
		return
	}

	m, err := msg.Read(c.cmd)
	if err != nil {
		return
	}

	switch mm := m.(type) {
	case *msg.Version:
		err = version.CompatServer(mm.Version)
	case *msg.Error:
		err = errors.New(mm.Message)
	default:
		err = errors.New("unexpected message type")
	}
	return
}

// Proxy establishes a proxy with the server.
func (c *Client) Proxy(id, token, addr string) (err error) {
	err = c.authProxy(id, token)
	if err != nil {
		return
	}

	l := fake.NewListener(16)
	p := proxy.Wrap(id, l)
	p.Bind(&dialer{Addr: addr})
	err = c.service.Proxy(p)
	if err != nil {
		return
	}

	c.listeners.Store(id, l)
	return
}

// Run client handles many user requests that are forwarded over the server.
func (c *Client) Run(ctx context.Context) (err error) {
	connc := make(chan net.Conn, 16)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()

		c.cmd.Close()
		c.session.Close()

		close(connc)
		for conn := range connc {
			conn.Close()
		}
	}()

	go c.listen(ctx, connc)
	go func() {
		err := c.service.Serve(ctx)
		if err != nil {
			c.errc <- err
		}
	}()

	for {
		select {
		case conn := <-connc:
			c.handleConn(conn)
		case <-ctx.Done():
			err = ctx.Err()
			return
		case err = <-c.errc:
			return
		}
	}
}

func (c *Client) listen(ctx context.Context, connc chan<- net.Conn) {
	for {
		conn, err := c.session.AcceptConn()
		if err != nil {
			c.errc <- ErrSessionClosed
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			connc <- conn
		}
	}
}

func (c *Client) handleConn(conn net.Conn) {
	var worker msg.Worker
	err := msg.ReadInto(conn, &worker)
	if err != nil {
		conn.Close()
		return
	}

	val, ok := c.listeners.Load(worker.ID)
	if ok {
		err = val.(*fake.Listener).Put(conn)
		if err != nil {
			conn.Close()
		}
	}
}
