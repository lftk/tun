package tcp

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/traffic"
	"github.com/4396/tun/transport"
)

type Proxy struct {
	sync.RWMutex
	net.Listener
	dialer transport.Dialer
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
	if p.dialer == d {
		p.Lock()
		p.dialer = nil
		p.Unlock()
		d.Close()
	}
}

func (p *Proxy) Bind(d transport.Dialer, m msg.Message) (ok bool) {
	// ...
	p.Lock()
	p.dialer = d
	p.Unlock()
	ok = true
	return
}

func (p *Proxy) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	p.RLock()
	dialer := p.dialer
	p.RUnlock()

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
