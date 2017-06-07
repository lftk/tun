package main

import (
	"fmt"
	"log"
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

func main() {
	go webServer(":3456")

	c, err := client.Dial(":8867")
	if err != nil {
		log.Fatal(err)
	}

	go c.Login("tcp", "token", ":3456")
	//go c.Login("ssh", "token", "")

	log.Fatal(c.Serve())
}
