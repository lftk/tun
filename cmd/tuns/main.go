package main

import (
	"context"
	"log"
	"time"

	"github.com/4396/tun/cmd"
	"github.com/4396/tun/server"
)

type tunServer struct {
	*server.Server
}

func newTunServer(addr, httpAddr string) (s *tunServer) {
	s = new(tunServer)
	s.Server = &server.Server{
		Addr:     addr,
		HttpAddr: httpAddr,
		Auth:     s.AuthProxy,
	}
	s.Traffic(s)
	return
}

func (s *tunServer) AuthProxy(name, token, desc string) (err error) {
	var p cmd.Proxy
	err = cmd.Decode(desc, &p)
	if err != nil {
		return
	}

	for _, p := range s.Server.Proxies() {
		if p.Name() == name {
			return
		}
	}

	switch p.Type {
	case cmd.TCP:
		err = s.Server.ProxyTCP(name, p.Port)
	case cmd.HTTP:
		err = s.Server.ProxyHTTP(name, p.Domain)
	}
	return
}

func (s *tunServer) In(name string, b []byte) {
	//fmt.Println("in", name, string(b[:n]))
}

func (s *tunServer) Out(name string, b []byte) {
	//fmt.Println("out", name, string(b[:n]))
}

func main() {
	s := newTunServer(":7000", ":7070")

	ctx, cancel := context.WithCancel(context.Background())

	time.AfterFunc(time.Second*30, func() {
		_ = cancel
		//cancel()
		//s.Kill("web1")
		//time.Sleep(time.Second * 3)
		//s.HttpProxy("web1", "web1")
	})
	log.Fatal(s.ListenAndServe(ctx))
}
