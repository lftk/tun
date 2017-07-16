package server

import (
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
)

// dialer is an implementation of proxy.Dialer.
type dialer struct {
	*mux.Session
	ID string
}

// Dial the client to establish a reverse working connection.
func (d *dialer) Dial() (conn net.Conn, err error) {
	conn, err = d.Session.OpenConn()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Worker{ID: d.ID})
	return
}

// Close this dialer.
func (d *dialer) Close() error {
	return nil
}
