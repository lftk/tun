package transport

import (
	"net"

	"github.com/xtaci/smux"
)

type muxDialer struct {
	*smux.Session
}

func MuxDial(addr string) (d Dialer, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	sess, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return
	}

	d = &muxDialer{sess}
	return
}

func (d *muxDialer) Dial() (t Transport, err error) {
	if d.Session == nil {
		err = ErrDialerClosed
		return
	}
	t, err = d.Session.OpenStream()
	return
}

func (d *muxDialer) Close() (err error) {
	if d.Session != nil {
		err = d.Session.Close()
		d.Session = nil
	}
	return
}
