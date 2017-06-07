package main

import (
	"log"

	"github.com/4396/tun/proxy/tcp"
	"github.com/4396/tun/server"
)

func main() {
	p1, err := tcp.Proxy("tcp", ":7070")
	if err != nil {
		log.Fatal(err)
	}

	p2, err := tcp.Proxy("ssh", ":7071")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe(":8867", p1, p2))
}
