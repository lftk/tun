package udp

import (
	"net"
	"sync/atomic"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/traffic"
)

type Proxy struct {
	net.Listener
	name   string
	dialer atomic.Value
}

func Listen(name, addr string) (p *Proxy, err error) {
	listener, err := ListenUDP(addr)
	if err != nil {
		return
	}

	p = &Proxy{
		name:     name,
		Listener: listener,
	}
	return
}

func (p *Proxy) Name() (name string) {
	name = p.name
	return
}

func (p *Proxy) Bind(d dialer.Dialer) (err error) {
	p.dialer.Store(d)
	return
}

func (p *Proxy) Unbind(d dialer.Dialer) (err error) {
	if p.dialer.Load().(dialer.Dialer) == d {
		p.dialer.Store(d)
		d.Close()
	}
	return
}

func (p *Proxy) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	dialer, _ := p.dialer.Load().(dialer.Dialer)
	if dialer == nil {
		return
	}

	work, err := dialer.Dial()
	if err != nil {
		return
	}
	defer work.Close()

	in, out := traffic.Join(conn, work)
	if traff != nil {
		traff.In(p.name, in)
		traff.Out(p.name, out)
	}
	return
}
