package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
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
	dialer, _ := p.dialer.Load().(dialer.Dialer)
	if dialer == nil {
		err = errors.New("Not bind dialer")
		return
	}

	work, err := dialer.Dial()
	if err != nil {
		return
	}

	trafficConn{
		name:    p.name,
		Conn:    conn,
		Traffic: traff,
	}.Join(work)

	fmt.Println("Handle succ...")
	return
}

type trafficConn struct {
	name string
	net.Conn
	traffic.Traffic
}

func (tc trafficConn) Read(b []byte) (n int, err error) {
	n, err = tc.Conn.Read(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.In(tc.name, b, int64(n))
	}
	return
}

func (tc trafficConn) Write(b []byte) (n int, err error) {
	n, err = tc.Conn.Write(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.Out(tc.name, b, int64(n))
	}
	return
}

func (tc trafficConn) Join(work net.Conn) {
	var wg sync.WaitGroup
	pipe := func(from, to net.Conn) {
		defer func() {
			from.Close()
			to.Close()
			wg.Done()
		}()
		io.Copy(to, from)
		return
	}
	wg.Add(2)
	go pipe(tc, work)
	go pipe(work, tc)
	wg.Wait()
}

func Join(dialer dialer.Dialer, conn net.Conn) (in, out int64) {
	work, err := dialer.Dial()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	pipe := func(from, to net.Conn, n *int64) {
		defer func() {
			from.Close()
			to.Close()
			wg.Done()
		}()
		*n, _ = io.Copy(to, from)
		return
	}

	wg.Add(2)
	go pipe(conn, work, &in)
	go pipe(work, conn, &out)
	wg.Wait()
	return
}
