package msg

import (
	"reflect"
)

// The Message is defined as interface{},
// which is the base class for all messages.
type Message interface{}

// Proxy describes a proxy message for a client
// to connect to the server.
type Proxy struct {
	ID       string
	Token    string
	Version  string
	Hostname string
	Os       string
	Arch     string
}

// Error describes an error message that is primarily
// used by the server to return to the client.
type Error struct {
	Message string
}

// Version contains the server release information back
// to the client when the client tries to establish a
// proxy with the server.
type Version struct {
	Version string
}

// Worker indicates that the server initiates a
// reverse connection to the client.
type Worker struct {
	ID string
}

// Get the type of an interface.
func typeof(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

var (
	// Records the byte corresponding to the message type.
	msgTypes = make(map[reflect.Type]byte)
	// Registers all message types.
	typeMsgs = []reflect.Type{
		typeof((*Proxy)(nil)),
		typeof((*Error)(nil)),
		typeof((*Version)(nil)),
		typeof((*Worker)(nil)),
	}
)

func init() {
	// Associative message type with byte.
	for i, t := range typeMsgs {
		msgTypes[t] = byte(i)
	}
}
