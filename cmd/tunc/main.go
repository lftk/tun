package main

import (
	"log"

	"github.com/4396/tun/transport"
)

func main() {
	d, err := transport.MuxDial(":8867")
	if err != nil {
		log.Fatal(err)
	}

	t, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()
	log.Println(t.LocalAddr())

	b := make([]byte, 100)
	for {
		_, err = t.Read(b)
		if err != nil {
			return
		}
		log.Println(string(b))

		_, err = t.Write([]byte("hello"))
		if err != nil {
			return
		}
	}
}
