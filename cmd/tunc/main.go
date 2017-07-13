package main

import (
	"context"
	"flag"
	"os"
	"time"

	"gopkg.in/ini.v1"

	"github.com/4396/tun/client"
	"github.com/4396/tun/log"
	"github.com/4396/tun/version"
)

var (
	conf   = flag.String("c", "conf/tunc.ini", "config file's path")
	server = flag.String("server", "", "tun server addr")
	id     = flag.String("id", "", "tun proxy id")
	token  = flag.String("token", "", "tun proxy token")
	addr   = flag.String("addr", "", "tun proxy local addr")
)

type proxy struct {
	Addr  string
	Token string
}

type config struct {
	Server  string
	Proxies map[string]*proxy
}

func parse(filename string, cfg *config) (err error) {
	_, errSt := os.Stat(*conf)
	if errSt != nil {
		return
	}

	f, err := ini.Load(filename)
	if err != nil {
		return
	}

	for _, sec := range f.Sections() {
		id := sec.Name()
		if id == "tunc" {
			cfg.Server = sec.Key("server").String()
			continue
		}

		token := sec.Key("token").String()
		if token == "" {
			continue
		}

		addr := sec.Key("addr").String()
		if addr == "" {
			continue
		}

		cfg.Proxies[id] = &proxy{
			Addr:  addr,
			Token: token,
		}
	}
	return
}

func loadConfig() (cfg *config, err error) {
	cfg = &config{
		Proxies: make(map[string]*proxy),
	}

	err = parse(*conf, cfg)
	if err != nil {
		return
	}

	if *server != "" {
		cfg.Server = *server
	}

	if *id != "" && *addr != "" {
		cfg.Proxies[*id] = &proxy{
			Addr:  *addr,
			Token: *token,
		}
	}
	return
}

func main() {
	flag.Parse()
	log.Infof("Start tun client, version is %s", version.Version)

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Load config failed, %v", err)
		return
	}

	var (
		idx int64
		ctx = context.Background()
	)
	for {
		c, err := client.Dial(cfg.Server)
		if err != nil {
			idx++
			time.Sleep(time.Second)
			log.Infof("Reconnect tun server %d", idx)
			continue
		}
		log.Info("Connect tun server success")

		for id, proxy := range cfg.Proxies {
			err = c.Proxy(id, proxy.Token, proxy.Addr)
			if err != nil {
				log.Fatalf("Load [%s] failed, %v", id, err)
				return
			}
			log.Infof("Load [%s] success", id)
		}

		idx = 0
		err = c.Run(ctx)
		if err != nil {
			if err != client.ErrSessionClosed {
				log.Errorf("Run client failed, err=%v", err)
				return
			}
		}
	}
}
