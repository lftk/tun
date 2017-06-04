package tcp

import (
	"io"
	"net"
	"sync"

	"github.com/4396/tun/msg"
	"github.com/4396/tun/transport"
)

type Proxy struct {
	net.Listener
	tran transport.Transport
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

func (p *Proxy) Bind(t transport.Transport, m msg.Message) (ok bool) {
	// ...
	p.tran = t
	ok = true
	return
}

func (p *Proxy) Handle(conn net.Conn) (err error) {
	if p.tran == nil {
		return
	}

	var wg sync.WaitGroup
	pipe := func(src, dst io.ReadWriter) {
		io.Copy(dst, src)
		wg.Done()
	}

	wg.Add(2)
	go pipe(conn, p.tran)
	go pipe(p.tran, conn)
	wg.Wait()
	return
}
