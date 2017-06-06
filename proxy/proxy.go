package proxy

import (
	"io"
	"net"
	"sync"

	"github.com/4396/tun/dialer"
	"github.com/4396/tun/traffic"
)

type Proxy interface {
	Name() string
	Close() error
	Accept() (net.Conn, error)
	Bind(dialer.Dialer) error
	Unbind(dialer.Dialer) error
	Handle(net.Conn, traffic.Traffic) error
}

func Join(dialer dialer.Dialer, conn net.Conn) (in, out int64, err error) {
	work, err := dialer.Dial()
	if err != nil {
		return
	}

	var (
		errs [2]error
		wg   sync.WaitGroup
	)

	pipe := func(src io.Reader, dst io.Writer, n *int64, err *error) {
		*n, *err = io.Copy(dst, src)
		wg.Done()
		return
	}

	wg.Add(2)
	go pipe(conn, work, &in, &errs[0])
	go pipe(work, conn, &out, &errs[1])
	wg.Wait()

	for _, err = range errs {
		if err != nil {
			return
		}
	}
	return
}
