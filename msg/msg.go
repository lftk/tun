package msg

import (
	"reflect"
)

type Message interface{}

type Auth struct {
	Name  string
	Token string
}

type AuthResp struct {
	Error string
}

type Dial struct{}

type WorkConn struct{}

type StartWorkConn struct{}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]byte)
	typeMsgs = []reflect.Type{
		typeof((*Auth)(nil)),
		typeof((*AuthResp)(nil)),
		typeof((*Dial)(nil)),
		typeof((*WorkConn)(nil)),
		typeof((*StartWorkConn)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = byte(i)
	}
}
