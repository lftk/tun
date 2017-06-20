package main

import (
	"context"
	"flag"

	"github.com/4396/tun/cmd"
	"github.com/4396/tun/log"
	"github.com/4396/tun/server"
)

type tunServer struct {
	*server.Server
}

func Listen(addr, httpAddr string) (s *tunServer, err error) {
	s = new(tunServer)
	svr, err := server.Listen(addr, httpAddr, s.Auth)
	if err != nil {
		return
	}

	svr.Traffic(s)
	s.Server = svr
	return
}

func (s *tunServer) Auth(name, token, desc string) (err error) {
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
	flag.Parse()
	log.Info("Start tun server")

	s, err := Listen(":7000", ":7070")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run(context.Background()))
}
