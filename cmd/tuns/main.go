package main

import (
	"fmt"
	"log"

	"github.com/4396/tun/proxy/tcp"
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
	p1, err := tcp.Proxy("tcp", ":7070")
	if err != nil {
		log.Fatal(err)
	}

	p2, err := tcp.Proxy("ssh", ":7071")
	if err != nil {
		log.Fatal(err)
	}

	s := server.Server{Addr: ":8867"}
	s.Traffic(new(traffic))
	s.Proxy(p1)
	s.Proxy(p2)

	log.Fatal(s.ListenAndServe())
}
