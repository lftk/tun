package cmd

import (
	"encoding/json"
)

const (
	TCP = iota
	HTTP
)

type Proxy struct {
	Type   int
	Port   string
	Domain string
}

func Encode(p *Proxy) (desc string, err error) {
	b, err := json.Marshal(p)
	if err != nil {
		return
	}

	desc = string(b)
	return
}

func Decode(desc string, p *Proxy) (err error) {
	err = json.Unmarshal([]byte(desc), p)
	return
}
