package msg

import (
	"reflect"
)

type Message interface{}

type Proxy struct {
	Name  string
	Token string
}

type ProxyResp struct {
	Error string
}

type Dial struct{}

type Worker struct {
	Name string
}

type StartWork struct{}

type Cmder struct {
	Token string
}

type CmderResp struct {
	Error string
}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]byte)
	typeMsgs = []reflect.Type{
		typeof((*Proxy)(nil)),
		typeof((*ProxyResp)(nil)),
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
