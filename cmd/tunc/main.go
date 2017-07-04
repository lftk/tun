package main

import (
	"context"
	"flag"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/4396/tun/client"
	"github.com/4396/tun/log"
	"github.com/4396/tun/version"
)

type config struct {
	Server  string
	Proxies map[string]struct {
		Addr  string
		Token string
	}
}

func loadConfig(path string) (cfg config, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, &cfg)
	return
}

var (
	conf = flag.String("c", "conf/tunc.yaml", "config file's path")
)

func main() {
	flag.Parse()
	log.Infof("Start tun client, version is %s", version.Version)

	cfg, err := loadConfig(*conf)
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
			if err != client.ErrDialerClosed {
				log.Errorf("c.Run failed, err=%v", err)
				return
			}
		}
	}
}
