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
	Desc     string
	Version  string
	Hostname string
	Os       string
	Arch     string
}

type Version struct {
	Version string
}

type Dial struct{}

type Worker struct {
	Name string
}

type StartWork struct{}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]byte)
	typeMsgs = []reflect.Type{
		typeof((*Error)(nil)),
		typeof((*Proxy)(nil)),
		typeof((*Version)(nil)),
		typeof((*Dial)(nil)),
		typeof((*Worker)(nil)),
		typeof((*StartWork)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = byte(i)
	}
}
