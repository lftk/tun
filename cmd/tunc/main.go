package main

import (
	"log"

	"github.com/4396/tun/client"
)

func main() {
	c, err := client.Dial(":8867")
	if err != nil {
		log.Fatal(err)
	}
	c.Login("name", "token")
	log.Fatal(c.Serve())
}
