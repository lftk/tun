package fake

// Addr is an implementation of net.Addr.
type Addr struct {
	addr string
}

// NewAddr returns a Addr by receiving a custom network information.
func NewAddr(addr string) *Addr {
	return &Addr{addr}
}

// Network always returns "fake".
func (a *Addr) Network() string {
	return "fake"
}

// String to return network information.
func (a *Addr) String() string {
	return a.addr
}
