package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
)

// Wrap a net.Listener to generate a Proxy.
func Wrap(id string, listener net.Listener) Proxy {
	return &proxy{
		id:       id,
		listener: listener,
	}
}

type proxy struct {
	id       string
	dialer   Dialer
	mu       sync.RWMutex
	listener net.Listener
}

func (p *proxy) ID() string {
	return p.id
}

func (p *proxy) Close() error {
	p.listener.Close()
	p.mu.Lock()
	if p.dialer != nil {
		p.dialer.Close()
	}
	p.mu.Unlock()
	return nil
}

func (p *proxy) Accept() (net.Conn, error) {
	return p.listener.Accept()
}

// Bind a dialer, only one dialer can be bound.
func (p *proxy) Bind(dialer Dialer) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.dialer != nil {
		err = errors.New("already bound dialer")
		return
	}

	p.dialer = dialer
	return
}

func (p *proxy) Unbind(dialer Dialer) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.dialer == dialer {
		p.dialer = nil
		dialer.Close()
	} else {
		err = errors.New("the dialer is not bound")
	}
	return
}

func (p *proxy) Handle(conn net.Conn, traff Traffic) (err error) {
	var dialer Dialer
	p.mu.RLock()
	dialer = p.dialer
	p.mu.RUnlock()
	if dialer == nil {
		err = errors.New("unbund dialer")
		return
	}

	worker, err := dialer.Dial()
	if err != nil {
		return
	}

	trafficConn{
		ID:      p.id,
		Conn:    conn,
		Traffic: traff,
	}.Join(worker)
	return
}

type trafficConn struct {
	ID string
	net.Conn
	Traffic
}

func (tc trafficConn) Read(b []byte) (n int, err error) {
	n, err = tc.Conn.Read(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.In(tc.ID, b[:n])
	}
	return
}

func (tc trafficConn) Write(b []byte) (n int, err error) {
	n, err = tc.Conn.Write(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.Out(tc.ID, b[:n])
	}
	return
}

func (tc trafficConn) Join(conn net.Conn) {
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
	go pipe(tc, conn)
	go pipe(conn, tc)
	wg.Wait()
}
