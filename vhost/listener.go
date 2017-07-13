package vhost

import (
	"github.com/4396/tun/fake"
)

type listener struct {
	*Muxer
	*fake.Listener
	Domain string
}

func (l *listener) Close() error {
	l.Muxer.domains.Delete(l.Domain)
	return l.Listener.Close()
}
