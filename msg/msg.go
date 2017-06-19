package msg

import (
	"reflect"
)

type Message interface{}

type OK struct{}

type Error struct {
	Message string
}

type Proxy struct {
	Name  string
	Token string
	Desc  string
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
		typeof((*OK)(nil)),
		typeof((*Error)(nil)),
		typeof((*Proxy)(nil)),
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
