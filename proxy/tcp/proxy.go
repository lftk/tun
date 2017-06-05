package tcp

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/transport"
)

type Proxy struct {
	net.Listener
	dialer atomic.Value
}

func Listen(addr string) (p *Proxy, err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	p = &Proxy{Listener: listener}
	return
}

func (p *Proxy) Name() (name string) {
	name = p.Addr().String()
	return
}

func (p *Proxy) Unbind(d transport.Dialer) {
	if p.dialer.Load().(transport.Dialer) == d {
		p.dialer.Store(d)
		d.Close()
	}
}

func (p *Proxy) Bind(d transport.Dialer, m msg.Message) (ok bool) {
	// ...
	p.dialer.Store(d)
	ok = true
	return
}

func (p *Proxy) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	dialer := p.dialer.Load().(transport.Dialer)
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
		traff.In(dialer, in, time.Now())
		traff.Out(dialer, out, time.Now())
	}
	return
}
