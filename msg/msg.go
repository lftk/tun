package msg

import (
	"reflect"
)

type Message interface{}

type Login struct {
	Name  string
	Token string
}

type TcpLogin struct {
	Login
}

type UdpLogin struct {
	Login
}

type HttpLogin struct {
	Login
	Domain string
}

type Dial struct{}

type WorkConn struct {
	Name  string
	Token string
}

func IsLogin(msg Message) (l Login, b bool) {
	b = true
	switch mm := msg.(type) {
	case *TcpLogin:
		l = mm.Login
	case *UdpLogin:
		l = mm.Login
	case *HttpLogin:
		l = mm.Login
	default:
		b = false
	}
	return
}

func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	msgTypes = make(map[reflect.Type]int8)
	typeMsgs = []reflect.Type{
		typeof((*TcpLogin)(nil)),
		typeof((*UdpLogin)(nil)),
		typeof((*HttpLogin)(nil)),
		typeof((*Dial)(nil)),
		typeof((*WorkConn)(nil)),
	}
)

func init() {
	for i, t := range typeMsgs {
		msgTypes[t] = int8(i)
	}
}
