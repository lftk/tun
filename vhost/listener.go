package vhost

import (
	"github.com/4396/tun/fake"
)

// listener is a subdomain listener.
// That combines Muxer and fake.Listenr.
type listener struct {
	*Muxer
	*fake.Listener
	Domain string
}

// Close this listener.
func (l *listener) Close() error {
	l.Muxer.domains.Delete(l.Domain)
	return l.Listener.Close()
}
