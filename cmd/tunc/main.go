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

func main() {
	go webServer(":3456")
	go tcpServer(":4567")

	dialer, err := mux.Dial(":8867")
	if err != nil {
		return
	}

	c := client.Client{
		Dialer: dialer,
	}

	err = c.Proxy("web1", "token", ":3456")
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second*3, func() {
		_ = cancel
		// cancel()
	})

	log.Fatal(c.Serve(ctx))
}
