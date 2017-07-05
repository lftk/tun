package fake

type Addr struct{}

func (a *Addr) Network() string {
	return ""
}

func (a *Addr) String() string {
	return ""
}
