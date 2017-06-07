package traffic

import (
	"io"
	"net"
	"sync"
)

type Traffic interface {
	In(name string, num int64)
	Out(name string, num int64)
}

func Join(user, conn net.Conn) (in, out int64) {
	var wg sync.WaitGroup
	pipe := func(from net.Conn, to net.Conn, n *int64) {
		defer func() {
			from.Close()
			to.Close()
			wg.Done()
		}()
		*n, _ = io.Copy(to, from)
		return
	}

	wg.Add(2)
	go pipe(conn, user, &in)
	go pipe(user, conn, &out)
	wg.Wait()
	return
}
