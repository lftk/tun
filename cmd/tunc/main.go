package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/4396/tun/client"
	"github.com/4396/tun/dialer"
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

	c.Dialer("name", &dialer.TcpDialer{Addr: ":3456"})

	c.Login("name", "token")
	log.Fatal(c.Serve())
}
