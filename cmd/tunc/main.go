package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/4396/tun/client"
	"github.com/4396/tun/cmd"
	"github.com/4396/tun/log"
)

func webServer(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, time.Now().String())
	})
	http.ListenAndServe(addr, nil)
}

func tcpServer(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		r := bufio.NewReader(conn)
		req, err := http.ReadRequest(r)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("%+v\n", req)
		log.Info(req.Method, req.Host)
	}
}

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

func main() {
	flag.Parse()
	log.Info("Start tun client")

	go webServer(":3456")
	go tcpServer(":4567")

	c, err := Dial(":7000")
	if err != nil {
		return
	}

	err = c.ProxyTCP("tcp1", "token", ":6060", ":4567")
	if err != nil {
		log.Fatal(err)
	}

	err = c.ProxyHTTP("web1", "token", "web1", ":3456")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(c.Serve(context.Background()))
}
