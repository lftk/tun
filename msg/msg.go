package msg

import (
	"reflect"
)

type Message interface{}

type Proxy struct {
	Name     string
	Token    string
	Version  string
	Hostname string
	Os       string
	Arch     string
}

type Error struct {
	Message string
}

type Version struct {
	Version string
}

type Worker struct {
	Name string
}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]byte)
	typeMsgs = []reflect.Type{
		typeof((*Proxy)(nil)),
		typeof((*Error)(nil)),
		typeof((*Version)(nil)),
		typeof((*Worker)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = byte(i)
	}
}
