package server

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/4396/tun/msg"
)

type agent struct {
	conn  net.Conn
	connc chan net.Conn
}

func (a *agent) Put(conn net.Conn) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%v", x)
		}
	}()
	a.connc <- conn
	return
}

var (
	ErrAgentClosed  = errors.New("Agent closed")
	ErrNoEnoughConn = errors.New("No enough work conn")
)

func (a *agent) dial() (conn net.Conn, err error) {
	select {
	case c, ok := <-a.connc:
		if !ok {
			err = ErrAgentClosed
			return
		}
		conn = c
	default:
		err = msg.Write(a.conn, &msg.Dial{})
	}
	if conn != nil || err != nil {
		return
	}

	select {
	case c, ok := <-a.connc:
		if !ok {
			err = ErrAgentClosed
			return
		}
		conn = c
	case <-time.After(time.Second):
		err = ErrNoEnoughConn
	}
	return
}

func (a *agent) Dial() (conn net.Conn, err error) {
	conn, err = a.dial()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.StartWork{})
	return
}

func (a *agent) Close() (err error) {
	fmt.Println("...close...")

	a.conn.Close()

	close(a.connc)
	for conn := range a.connc {
		conn.Close()
	}
	return
}
