package udp

import (
	"bytes"
	"errors"
	"net"
	"sync/atomic"
)

type Listener struct {
	udpConn *net.UDPConn
	closed  int32
	connc   chan net.Conn
}

func ListenUDP(addr string) (l net.Listener, err error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}

	l = &Listener{
		udpConn: udpConn,
		connc:   make(chan net.Conn, 16),
	}

	go func() {
		for {
			b := make([]byte, 1024)
			_, udpAddr, err := udpConn.ReadFromUDP(b)
			if err != nil {
				return
			}

			l.(*Listener).connc <- &Conn{
				listener: l.(*Listener),
				addr:     udpAddr,
				data:     bytes.NewBuffer(b),
			}
		}
	}()
	return
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	conn = <-l.connc
	return
}

func (l *Listener) Addr() (addr net.Addr) {
	addr = l.udpConn.LocalAddr()
	return
}

func (l *Listener) Close() (err error) {
	err = l.udpConn.Close()
	return
}

func (l *Listener) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) {
	if atomic.LoadInt32(&l.closed) == 1 {
		err = errors.New("Listener closed")
		return
	}
	n, err = l.udpConn.WriteToUDP(b, addr)
	return
}
