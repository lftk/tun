package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/4396/tun/server"
)

type traffic struct{}

func (t *traffic) In(name string, b []byte, n int64) {
	fmt.Println("in", name, string(b[:n]), n)
}

func (t *traffic) Out(name string, b []byte, n int64) {
	fmt.Println("out", name, string(b[:n]), n)
}

func main() {
	s := server.Server{
		Addr:     ":8867",
		HttpAddr: ":8082",
	}
	s.Traffic(new(traffic))

	s.TcpProxy("tcp", ":7070")
	s.TcpProxy("ssh", ":7071")
	s.HttpProxy("web1", "web1")
	s.HttpProxy("web2", "web2")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	time.AfterFunc(time.Second*3, func() {
		//cancel()
		//s.Kill("web1")
	})
	log.Fatal(s.ListenAndServe(ctx))
}
