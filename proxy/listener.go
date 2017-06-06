package proxy

import (
	"errors"
	"net"
	"sync/atomic"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/traffic"
)

type listener struct {
	net.Listener
	name   string
	dialer atomic.Value
}

func Wrap(name string, l net.Listener) Proxy {
	return &listener{name: name, Listener: l}
}

func (p *listener) Name() string {
	return p.name
}

func (p *listener) Bind(d dialer.Dialer) error {
	p.dialer.Store(d)
	return nil
}

func (p *listener) Unbind(d dialer.Dialer) error {
	if p.dialer.Load().(dialer.Dialer) == d {
		p.dialer.Store(d)
		d.Close()
	}
	return nil
}

func (p *listener) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	defer conn.Close()
	dialer, _ := p.dialer.Load().(dialer.Dialer)
	if dialer == nil {
		err = errors.New("Not bind dialer")
		return
	}

	in, out, err := Join(dialer, conn)
	if traff != nil {
		traff.In(p.name, in)
		traff.Out(p.name, out)
	}
	return
}
