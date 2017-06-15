package mux

import (
	"net"

	"github.com/xtaci/smux"
)

func Dial(conn net.Conn) (d *Dialer, err error) {
	sess, err := smux.Client(conn, nil)
	if err != nil {
		return
	}

	d = &Dialer{sess: sess}
	return
}

type Dialer struct {
	sess *smux.Session
}

func (d *Dialer) Dial() (net.Conn, error) {
	return d.sess.OpenStream()
}

func (d *Dialer) Close() error {
	return d.sess.Close()
}
