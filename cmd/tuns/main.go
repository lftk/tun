package main

import (
	"context"
	"errors"
	"flag"
	"strconv"

	"github.com/4396/tun/log"
	"github.com/4396/tun/server"
	"github.com/4396/tun/version"
)

type tunServer struct {
	storage
	*server.Server
}

func Listen(addr, httpAddr string) (s *tunServer, err error) {
	s = new(tunServer)
	cfg := &server.Config{
		Addr:     addr,
		AddrHTTP: httpAddr,
		Auth:     s.Auth,
		Load:     s.Load,
		TraffIn:  s.TraffIn,
		TraffOut: s.TraffOut,
	}
	svr, err := server.Listen(cfg)
	if err != nil {
		return
	}

	s.Server = svr
	return
}

func (s *tunServer) Auth(name, token string) (err error) {
	return
}

func (s *tunServer) Load(loader server.Loader, name string) (err error) {
	p, err := s.storage.Load(name)
	if err != nil {
		return
	}

	switch p.Type {
	case ProxyTCP:
		var port int
		port, err = strconv.Atoi(p.Reserve)
		if err != nil {
			return
		}
		err = loader.ProxyTCP(p.Name, port)
	case ProxyHTTP:
		err = loader.ProxyHTTP(p.Name, p.Reserve)
	default:
		err = errors.New("Unexpected proxy")
	}
	return
}

func (s *tunServer) TraffIn(name string, b []byte) {
	log.Infof("[IN] %s %d", name, len(b))
}

func (s *tunServer) TraffOut(name string, b []byte) {
	log.Infof("[OUT] %s %d", name, len(b))
}

var (
	addr     = flag.String("addr", ":7000", "tun server listen addr")
	httpAddr = flag.String("http", ":7070", "web server listen addr")
)

func main() {
	flag.Parse()
	log.Infof("Start tun server, version is %s", version.Version)

	s, err := Listen(*addr, *httpAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run(context.Background()))
}
