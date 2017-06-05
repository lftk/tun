package msg

import (
	"encoding/binary"
	"errors"
	"io"
)

var MaxMsgLength int64 = 10240

func readMsg(r io.Reader) (t byte, b []byte, err error) {
	var (
		n  int64
		b1 = make([]byte, 1)
	)

	_, err = r.Read(b1)
	if err != nil {
		return
	}

	if t = b1[0]; int(t) >= len(typeMsgs) {
		err = errors.New("Message type error")
		return
	}

	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}

	if n > MaxMsgLength {
		err = errors.New("Message length exceed the limit")
		return
	}

	b = make([]byte, n)
	rn, err := io.ReadFull(r, b)
	if err != nil {
		return
	}

	if int64(rn) != n {
		err = errors.New("Message format error")
	}
	return
}

func Read(r io.Reader) (msg Message, err error) {
	if t, b, err := readMsg(r); err == nil {
		msg, err = UnPack(t, b)
	}
	return
}

func ReadInto(r io.Reader, msg Message) (err error) {
	if _, b, err := readMsg(r); err == nil {
		err = UnPackInto(b, msg)
	}
	return
}

func Write(w io.Writer, msg interface{}) (err error) {
	if b, err := Pack(msg); err == nil {
		_, err = w.Write(b)
	}
	return
}
