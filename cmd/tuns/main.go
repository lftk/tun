package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"gopkg.in/ini.v1"

	"github.com/4396/tun/log"
	"github.com/4396/tun/server"
	"github.com/4396/tun/version"
)

type tunServer struct {
	*server.Server
	proxies *ini.File
}

func newServer(conf string) (s *tunServer, err error) {
	cfg, err := ini.Load(conf)
	if err != nil {
		return
	}

	sec := cfg.Section("tuns")
	addr := sec.Key("addr").MustString(":7000")
	http := sec.Key("http").MustString(":7070")

	s = &tunServer{proxies: cfg}
	svr, err := server.Listen(&server.Config{
		Addr:     addr,
		AddrHTTP: http,
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
	sec, err := s.proxies.GetSection(name)
	if err != nil {
		err = fmt.Errorf("proxy '%s' not exists", name)
		return
	}

	if token != sec.Key("token").String() {
		err = errors.New("token does not match")
	}
	return
}

func (s *tunServer) Load(loader server.Loader, name string) (err error) {
	sec, err := s.proxies.GetSection(name)
	if err != nil {
		err = fmt.Errorf("proxy '%s' not exists", name)
		return
	}

	switch sec.Key("type").String() {
	case "tcp":
		var port int
		port, err = sec.Key("port").Int()
		if err != nil {
			return
		}
		err = loader.ProxyTCP(name, port)
	case "http":
		domain := sec.Key("domain").String()
		err = loader.ProxyHTTP(name, domain)
	default:
		err = errors.New("unexpected proxy type")
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

	s, err := newServer(*conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run(context.Background()))
}
