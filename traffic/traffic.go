package traffic

import (
	"time"

	"github.com/4396/tun/transport"
)

type Traffic interface {
	In(transport.Dialer, int64, time.Time)
	Out(transport.Dialer, int64, time.Time)
}
