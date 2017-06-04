package main

import (
	"log"

	"github.com/4396/tun/proxy/tcp"
	"github.com/4396/tun/server"
	"github.com/4396/tun/transport"
)

func main() {
	l, err := transport.MuxListen(":8867")
	if err != nil {
		log.Fatal(err)
	}

	p, err := tcp.Listen(":7070")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe(l, p))
}
