package server

import (
	"net"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/mux"
)

type dialer struct {
	*mux.Session
	Name string
}

func (d *dialer) Dial() (conn net.Conn, err error) {
	conn, err = d.Session.OpenConn()
	if err != nil {
		return
	}

	err = msg.Write(conn, &msg.Worker{Name: d.Name})
	return
}

func (d *dialer) Close() error {
	return nil
}
