package msg

import (
	"reflect"
)

type Message interface{}

type Error struct {
	Message string
}

type Proxy struct {
	Name     string
	Token    string
	Version  string
	Hostname string
	Os       string
	Arch     string
}

type Version struct {
	Version string
}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]byte)
	typeMsgs = []reflect.Type{
		typeof((*Error)(nil)),
		typeof((*Proxy)(nil)),
		typeof((*Version)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = byte(i)
	}
}
