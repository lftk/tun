package main

import (
	"context"
	"flag"

	"gopkg.in/ini.v1"

	"github.com/4396/tun/log"
	"github.com/4396/tun/server"
	"github.com/4396/tun/version"
)

type tunServer struct {
	*server.Server
}

func Listen(addr, httpAddr string) (s *tunServer, err error) {
	s = new(tunServer)
	svr, err := server.Listen(&server.Config{
		Addr:     addr,
		AddrHTTP: httpAddr,
		Auth:     s.Auth,
		Load:     s.Load,
		TraffIn:  s.TraffIn,
		TraffOut: s.TraffOut,
	})
	if err != nil {
		return
	}

	s.Server = svr
	return
}

func (s *tunServer) Auth(name, token string) (err error) {
	return
}

func (s *tunServer) Load(loader *server.Loader, name string) (err error) {
	if name == "ssh" {
		err = loader.ProxyTCP(name, 6060)
	} else {
		err = loader.ProxyHTTP(name, name)
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
	conf = flag.String("c", "conf/tuns.ini", "config file's path")
)

func main() {
	flag.Parse()
	log.Infof("Start tun server, version is %s", version.Version)

	cfg, err := ini.InsensitiveLoad(*conf)
	if err != nil {
		log.Fatalf("Load config file failed, err=%v", err)
		return
	}

	common := cfg.Section("common")
	addr := common.Key("addr").String()
	httpAddr := common.Key("http").String()

	s, err := Listen(addr, httpAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run(context.Background()))
}
