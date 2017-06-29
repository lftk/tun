package main

import (
	"context"
	"flag"
	"time"

	"gopkg.in/ini.v1"

	"github.com/4396/tun/client"
	"github.com/4396/tun/cmd"
	"github.com/4396/tun/log"
	"github.com/4396/tun/version"
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

func (c *tunClient) ProxyTCP(name, token, addr string, port int) (err error) {
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

func (c *tunClient) ProxyHTTP(name, token, addr, domain string) (err error) {
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

func (c *tunClient) LoadProxy(cfg *ini.File) (err error) {
	for _, sec := range cfg.Sections() {
		name := sec.Name()
		if name == "DEFAULT" || name == "common" {
			continue
		}

		var (
			typ   = sec.Key("type").String()
			token = sec.Key("token").String()
			addr  = sec.Key("addr").String()
		)
		switch typ {
		case "tcp":
			var port int
			port, err = sec.Key("port").Int()
			if err == nil {
				err = c.ProxyTCP(name, token, addr, port)
			}
		case "http":
			domain := sec.Key("domain").String()
			err = c.ProxyHTTP(name, token, addr, domain)
		default:
			log.Infof("Unknown proxy type, %s", name)
		}
		if err != nil {
			return
		}
	}
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

	var (
		idx  int64
		ctx  = context.Background()
		addr = cfg.Section("common").Key("addr").String()
	)
	for {
		c, err := Dial(addr)
		if err != nil {
			idx++
			time.Sleep(time.Second)
			log.Infof("Reconnect tun server %d", idx)
			continue
		}
		log.Info("Connect tun server success")

		err = c.LoadProxy(cfg)
		if err != nil {
			log.Fatalf("Load proxy failed, err=%v", err)
			return
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
