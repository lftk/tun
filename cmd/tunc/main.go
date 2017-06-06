package main

import (
	"fmt"
	"log"
	"net"
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

func udpServer(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}

	b := make([]byte, 1024)
	for {
		n, addr, err := udpConn.ReadFromUDP(b)
		if err != nil {
			return
		}
		fmt.Println(string(b[:n]))

		_, err = udpConn.WriteToUDP([]byte("hello"), addr)
		if err != nil {
			return
		}
	}
}

func main() {
	go webServer(":3456")
	go udpServer(":4567")

	c, err := client.Dial(":8867")
	if err != nil {
		log.Fatal(err)
	}

	c.Dialer("tcp", &dialer.TcpDialer{Addr: ":3456"})
	c.Dialer("udp", &dialer.UdpDialer{Addr: ":4567"})

	go c.Login("tcp", "token")
	go c.Login("udp", "token")

	log.Fatal(c.Serve())
}
