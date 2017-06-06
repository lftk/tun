package tcp

import (
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/4396/tun/traffic"
	"github.com/4396/tun/transport"
)

type Proxy struct {
	net.Listener
	name   string
	dialer atomic.Value
}

func Listen(name, addr string) (p *Proxy, err error) {
	listener, err := net.Listen("tcp", addr)
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

func (p *Proxy) Bind(d transport.Dialer) (err error) {
	p.dialer.Store(d)
	return
}

func (p *Proxy) Unbind(d transport.Dialer) (err error) {
	if p.dialer.Load().(transport.Dialer) == d {
		p.dialer.Store(d)
		d.Close()
	}
	return
}

func (p *Proxy) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	dialer, _ := p.dialer.Load().(transport.Dialer)
	if dialer == nil {
		return
	}

	work, err := dialer.Dial()
	if err != nil {
		return
	}
	defer work.Close()

	var (
		in, out int64
		wg      sync.WaitGroup
	)
	pipe := func(src, dst io.ReadWriter, written *int64) {
		n, _ := io.Copy(dst, src)
		if written != nil {
			*written = n
		}
		wg.Done()
	}

	wg.Add(2)
	go pipe(conn, work, &in)
	go pipe(work, conn, &out)
	wg.Wait()

	if traff != nil {
		traff.In(p.name, in)
		traff.Out(p.name, out)
	}
	return
}
