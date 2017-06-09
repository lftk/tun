package main

import (
	"fmt"
	"log"

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

	log.Fatal(s.ListenAndServe())
}
