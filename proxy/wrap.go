package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
)

// Wrap a net.Listener to create a Proxy.
func Wrap(id string, listener net.Listener) Proxy {
	return &proxy{
		id:       id,
		listener: listener,
	}
}

// proxy is an implementation of Proxy.
type proxy struct {
	id       string
	dialer   Dialer
	mu       sync.RWMutex
	listener net.Listener
}

// ID returns the id of proxy.
func (p *proxy) ID() string {
	return p.id
}

// Close the proxy and close the bound dialer.
func (p *proxy) Close() error {
	p.listener.Close()
	p.mu.Lock()
	if p.dialer != nil {
		p.dialer.Close()
	}
	p.mu.Unlock()
	return nil
}

// Accept returns the user's connection.
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

// Unbind a dialer that is already bound.
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

// Handle the user's connection.
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

// trafficConn combines Traffic and net.Conn.
type trafficConn struct {
	ID string
	net.Conn
	Traffic
}

// Read data from conn and records traffic.
func (tc trafficConn) Read(b []byte) (n int, err error) {
	n, err = tc.Conn.Read(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.In(tc.ID, b[:n])
	}
	return
}

// Write data to conn and records traffic.
func (tc trafficConn) Write(b []byte) (n int, err error) {
	n, err = tc.Conn.Write(b)
	if tc.Traffic != nil && n > 0 {
		tc.Traffic.Out(tc.ID, b[:n])
	}
	return
}

// Join working connection and exchange data.
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
