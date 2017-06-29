package main

import (
	"context"
	"flag"
	"time"

	"github.com/4396/tun/client"
	"github.com/4396/tun/cmd"
	"github.com/4396/tun/log"
	"github.com/4396/tun/version"
	"gopkg.in/ini.v1"
)

type tunClient struct {
	*client.Client
}

func Dial(addr string) (c *tunClient, err error) {
	cli, err := client.Dial(addr)
	if err != nil {
		return
	}

	c = &tunClient{Client: cli}
	return
}

func (c *tunClient) ProxyTCP(name, token, port, addr string) (err error) {
	desc, err := cmd.Encode(&cmd.Proxy{
		Type: cmd.TCP,
		Port: port,
	})
	if err != nil {
		return
	}

	err = c.Client.Proxy(name, token, desc, addr)
	return
}

func (c *tunClient) ProxyHTTP(name, token, domain, addr string) (err error) {
	desc, err := cmd.Encode(&cmd.Proxy{
		Type:   cmd.HTTP,
		Domain: domain,
	})
	if err != nil {
		return
	}

	err = c.Client.Proxy(name, token, desc, addr)
	return
}

var (
	conf = flag.String("c", "conf/tunc.ini", "config file's path")
)

func main() {
	flag.Parse()
	log.Infof("Start tun client, version is %s", version.Version)

	cfg, err := ini.InsensitiveLoad(*conf)
	if err != nil {
		log.Fatalf("Load config file failed, err=%v", err)
		return
	}
	_ = cfg

	var (
		idx int64
		ctx = context.Background()
	)
	for {
		c, err := Dial(":7000")
		if err != nil {
			idx++
			time.Sleep(time.Second)
			log.Infof("Reconnect tun server %d", idx)
			continue
		}
		log.Info("Connect tun server success")

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
