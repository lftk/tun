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

func main() {
	go webServer(":3456")
	go tcpServer(":4567")

	c := client.Client{
		ServerAddr: ":8867",
		LocalAddr:  ":3456",
		Name:       "web1",
		Token:      "token",
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second*3, func() {
		_ = cancel
		// cancel()
	})

	for {
		err := c.DialAndServe(ctx)
		if err == context.Canceled {
			return
		}
		time.Sleep(time.Second)
		fmt.Println("...reconnect...")
	}
}
