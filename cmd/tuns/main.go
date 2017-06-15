package main

import (
	"context"
	"log"
	"time"

	"github.com/4396/tun/server"
)

type traffic struct{}

func (t *traffic) In(name string, b []byte, n int64) {
	//fmt.Println("in", name, string(b[:n]), n)
}

func (t *traffic) Out(name string, b []byte, n int64) {
	//fmt.Println("out", name, string(b[:n]), n)
}

type admin struct{}

func (a *admin) AuthProxy(name, token string) error {
	return nil
}

func (a *admin) AuthCmder(token string) error {
	return nil
}

func (a *admin) Command(b []byte) ([]byte, error) {
	return nil, nil
}

func main() {
	s := server.Server{
		Addr:     ":8867",
		HttpAddr: ":8082",
		Admin:    new(admin),
	}

	s.Traffic(new(traffic))

	s.ProxyTCP("tcp", ":7070")
	s.ProxyTCP("ssh", ":7071")
	s.ProxyHTTP("web1", "web1")
	s.ProxyHTTP("web2", "web2")

	ctx, cancel := context.WithCancel(context.Background())

	time.AfterFunc(time.Second*30, func() {
		_ = cancel
		//cancel()
		//s.Kill("web1")
		//time.Sleep(time.Second * 3)
		//s.HttpProxy("web1", "web1")
	})
	log.Fatal(s.ListenAndServe(ctx))
}
