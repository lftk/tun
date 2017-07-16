package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
)

// Unpack []byte as a message. If `in` is nil,
// the packet is unpacked by type `t`.
func unpack(t byte, b []byte, in Message) (msg Message, err error) {
	if in == nil {
		if t >= byte(len(typeMsgs)) {
			err = errors.New("unknown message type")
			return
		}
		msg = reflect.New(typeMsgs[t]).Interface().(Message)
	} else {
		msg = in
	}
	err = json.Unmarshal(b, &msg)
	return
}

// UnPackInto will unpack the []byte to `msg`.
func UnPackInto(b []byte, msg Message) (err error) {
	_, err = unpack(' ', b, msg)
	return
}

// UnPack to a message according to type `t`.
func UnPack(t byte, b []byte) (Message, error) {
	return unpack(t, b, nil)
}

// Pack a message becomes []byte.
func Pack(msg Message) (b []byte, err error) {
	t, ok := msgTypes[typeof(msg)]
	if !ok {
		err = errors.New("unknown message type")
		return
	}

	var buf bytes.Buffer
	err = buf.WriteByte(byte(t))
	if err != nil {
		return
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return
	}

	n := int64(len(m))
	err = binary.Write(&buf, binary.BigEndian, n)
	if err != nil {
		return
	}

	_, err = buf.Write(m)
	if err != nil {
		return
	}

	b = buf.Bytes()
	return
}
