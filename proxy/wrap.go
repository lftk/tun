package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
)

func Wrap(id string, listener net.Listener) Proxy {
	return &proxy{
		id:       id,
		listener: listener,
	}
}

type proxy struct {
	id       string
	dialer   Dialer
	locker   sync.RWMutex
	listener net.Listener
}

func (p *proxy) ID() string {
	return p.id
}

func (p *proxy) Close() error {
	p.listener.Close()
	p.locker.Lock()
	if p.dialer != nil {
		p.dialer.Close()
	}
	p.locker.Unlock()
	return nil
}

func (p *proxy) Accept() (net.Conn, error) {
	return p.listener.Accept()
}

func (p *proxy) Bind(dialer Dialer) (err error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.dialer != nil {
		err = errors.New("exists dialer")
		return
	}

	p.dialer = dialer
	return
}

func (p *proxy) Unbind(dialer Dialer) (err error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.dialer == dialer {
		p.dialer = nil
		dialer.Close()
	}
	return
}

func (p *proxy) Handle(conn net.Conn, traff Traffic) (err error) {
	var dialer Dialer
	p.locker.RLock()
	dialer = p.dialer
	p.locker.RUnlock()
	if dialer == nil {
		err = errors.New("not bind dialer")
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
