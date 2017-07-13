package server

import (
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
)

type dialer struct {
	*mux.Session
	ID string
}

func (d *dialer) Dial() (conn net.Conn, err error) {
	conn, err = d.Session.OpenConn()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Worker{ID: d.ID})
	return
}

func (d *dialer) Close() error {
	return nil
}
