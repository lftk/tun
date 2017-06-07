package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
)

func unpack(t byte, b []byte, in Message) (msg Message, err error) {
	if in == nil {
		if t >= byte(len(typeMsgs)) {
			err = errors.New("Unknown msg type")
			return
		}
		msg = reflect.New(typeMsgs[t]).Interface().(Message)
	} else {
		msg = in
	}
	err = json.Unmarshal(b, &msg)
	return
}

func UnPackInto(b []byte, msg Message) (err error) {
	_, err = unpack(0, b, msg)
	return
}

func UnPack(t byte, b []byte) (msg Message, err error) {
	return unpack(t, b, nil)
}

func Pack(msg Message) (b []byte, err error) {
	t, ok := msgTypes[typeof(msg)]
	if !ok {
		err = errors.New("Unknown msg type")
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
