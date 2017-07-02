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
	svr, err := server.Listen(addr, httpAddr, s.Auth)
	if err != nil {
		return
	}

	svr.Traffic(s)
	s.Server = svr
	return
}

func (s *tunServer) Auth(name, token string) (err error) {
	if name == "ssh" {
		err = s.Server.ProxyTCP(name, 6060)
	} else {
		err = s.Server.ProxyHTTP(name, name)
	}
	return
}

func (s *tunServer) In(name string, b []byte) {
	//fmt.Println("in", name, string(b[:n]))
}

func (s *tunServer) Out(name string, b []byte) {
	//fmt.Println("out", name, string(b[:n]))
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
