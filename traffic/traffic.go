package traffic

import (
	"io"
	"sync"
)

type Traffic interface {
	In(name string, num int64)
	Out(name string, num int64)
}

func Join(user, conn io.ReadWriter) (in, out int64) {
	var wg sync.WaitGroup
	pipe := func(src io.Reader, dst io.Writer, n *int64) {
		*n, _ = io.Copy(dst, src)
		wg.Done()
		return
	}
	wg.Add(2)
	go pipe(user, conn, &in)
	go pipe(conn, user, &out)
	wg.Wait()
	return
}
