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
	name   = flag.String("name", "", "tun proxy name")
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
		name := sec.Name()
		if name == "tunc" {
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

		cfg.Proxies[name] = &proxy{
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

	if *name != "" && *addr != "" {
		cfg.Proxies[*name] = &proxy{
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

		for name, proxy := range cfg.Proxies {
			err = c.Proxy(name, proxy.Token, proxy.Addr)
			if err != nil {
				log.Fatalf("Load [%s] failed, %v", name, err)
				return
			}
			log.Infof("Load [%s] success", name)
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
