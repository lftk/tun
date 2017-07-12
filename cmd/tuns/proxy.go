package main

const (
	ProxyTCP  = "tcp"
	ProxyHTTP = "http"
)

type proxy struct {
	Type    string
	Name    string
	Token   string
	Reserve string
}
