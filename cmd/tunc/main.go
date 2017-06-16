package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/4396/tun/client"
	"github.com/4396/tun/cmd"
	"github.com/4396/tun/mux"
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
			log.Println(err)
			return
		}

		r := bufio.NewReader(conn)
		req, err := http.ReadRequest(r)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("%+v\n", req)
		log.Println(req.Method, req.Host)
	}
}

func tcpProxy(c *client.Client, name, token, port, addr string) (err error) {
	desc, err := cmd.Encode(&cmd.Proxy{
		Type: cmd.TCP,
		Port: port,
	})
	if err != nil {
		return
	}

	err = c.Proxy(name, token, desc, addr)
	return
}

func httpProxy(c *client.Client, name, token, domain, addr string) (err error) {
	desc, err := cmd.Encode(&cmd.Proxy{
		Type:   cmd.HTTP,
		Domain: domain,
	})
	if err != nil {
		return
	}

	err = c.Proxy(name, token, desc, addr)
	return
}

func main() {
	go webServer(":3456")
	go tcpServer(":4567")

	dialer, err := mux.Dial(":7000")
	if err != nil {
		return
	}

	c := &client.Client{
		Dialer: dialer,
	}

	err = httpProxy(c, "web1", "token", "web1", ":3456")
	if err != nil {
		log.Println(err)
		return
	}

	err = tcpProxy(c, "tcp1", "token", ":6060", ":4567")
	if err != nil {
		log.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second*3, func() {
		_ = cancel
		// cancel()
	})

	log.Fatal(c.Serve(ctx))
}
