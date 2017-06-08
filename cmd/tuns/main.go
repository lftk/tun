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
	s := server.Server{Addr: ":8867"}

	s.Traffic(new(traffic))

	s.TCPProxy("tcp", ":7070")
	s.TCPProxy("ssh", ":7071")

	log.Fatal(s.ListenAndServe())
}
