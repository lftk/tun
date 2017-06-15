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
		b1 [1]byte
	)

	_, err = r.Read(b1[:])
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
	t, b, err := readMsg(r)
	if err != nil {
		return
	}

	msg, err = UnPack(t, b)
	return
}

func ReadInto(r io.Reader, msg Message) (err error) {
	_, b, err := readMsg(r)
	if err != nil {
		return
	}

	err = UnPackInto(b, msg)
	return
}

func Write(w io.Writer, msg interface{}) (err error) {
	b, err := Pack(msg)
	if err != nil {
		return
	}

	_, err = w.Write(b)
	return
}

var ok OK

func ReplyOK(w io.Writer) error {
	return Write(w, &ok)
}

func ReplyError(w io.Writer, msg string) error {
	return Write(w, &Error{msg})
}

func Okay(r io.Reader) (err error) {
	m, err := Read(r)
	if err != nil {
		return
	}

	switch mm := m.(type) {
	case *OK:
	case *Error:
		err = errors.New(mm.Message)
	default:
		err = errors.New("Unexpected error")
	}
	return
}
