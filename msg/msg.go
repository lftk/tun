package msg

import (
	"reflect"
)

type Message interface{}

type Login struct {
	Name  string
	Token string
}

type Dial struct{}

type WorkConn struct {
	Name  string
	Token string
}

type StartWorkConn struct{}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]int8)
	typeMsgs = []reflect.Type{
		typeof((*Login)(nil)),
		typeof((*Dial)(nil)),
		typeof((*WorkConn)(nil)),
		typeof((*StartWorkConn)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = int8(i)
	}
}
