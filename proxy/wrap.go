package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/traffic"
)

func Wrap(name string, l net.Listener) Proxy {
	return &proxy{
		Listener: l,
		name:     name,
	}
}

type proxy struct {
	net.Listener
	name   string
	closed bool
	mu     sync.RWMutex
	dialer dialer.Dialer
}

func (p *proxy) Name() string {
	return p.name
}

func (p *proxy) Close() (err error) {
	p.closed = true
	err = p.Listener.Close()
	if err != nil {
		return
	}

	p.mu.RLock()
	if p.dialer != nil {
		err = p.dialer.Close()
	}
	p.mu.RUnlock()
	return
}

func (p *proxy) Accept() (net.Conn, error) {
	conn, err := p.Listener.Accept()
	if err != nil {
		if p.closed {
			err = ErrClosed
		}
	}
	return conn, err
}

func (p *proxy) Bind(d dialer.Dialer) error {
	p.mu.Lock()
	p.dialer = d
	p.mu.Unlock()
	return nil
}

func (p *proxy) Unbind(d dialer.Dialer) error {
	p.mu.Lock()
	if p.dialer == d {
		p.dialer = nil
		d.Close()
	}
	p.mu.Unlock()
	return nil
}

func (p *proxy) Handle(conn net.Conn, traff traffic.Traffic) (err error) {
	var dialer dialer.Dialer
	p.mu.RLock()
	dialer = p.dialer
	p.mu.RUnlock()
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
		tc.Traffic.In(tc.name, b[:n], int64(n))
	}
	return
}

func (tc trafficConn) Write(b []byte) (n int, err error) {
	n, err = tc.Conn.Write(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.Out(tc.name, b[:n], int64(n))
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
