package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/4396/tun/client"
	"github.com/4396/tun/log"
	"github.com/4396/tun/version"
)

var (
	conf   = flag.String("c", "conf/tunc.yaml", "config file's path")
	server = flag.String("server", "", "tun server addr")
	name   = flag.String("name", "", "tun proxy name")
	token  = flag.String("token", "", "tun proxy token")
	addr   = flag.String("addr", "", "tun proxy addr")
)

type proxy struct {
	Addr  string
	Token string
}

type config struct {
	Server  string
	Proxies map[string]proxy
}

func loadConfig() (cfg config, err error) {
	_, errSt := os.Stat(*conf)
	if errSt == nil {
		var b []byte
		b, err = ioutil.ReadFile(*conf)
		if err != nil {
			return
		}

		err = yaml.Unmarshal(b, &cfg)
		if err != nil {
			return
		}
	}

	if *server != "" {
		cfg.Server = *server
	}

	if *name != "" && *addr != "" {
		if cfg.Proxies == nil {
			cfg.Proxies = make(map[string]proxy)
		}
		cfg.Proxies[*name] = proxy{
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
