package server

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/4396/tun/msg"
	"github.com/xtaci/smux"
)

type agent struct {
	name  string
	sess  *smux.Session
	conn  *smux.Stream
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

	err = msg.Write(conn, &msg.StartWorkConn{})
	return
}

func (a *agent) Close() (err error) {
	fmt.Println("...close...")

	// close channel
	close(a.connc)

	// close conn and sess
	a.conn.Close()
	a.sess.Close()

	// close conn
	for conn := range a.connc {
		conn.Close()
	}
	return
}
