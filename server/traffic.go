package server

// traffic is an implementation of proxy.Traffic.
type traffic struct {
	TraffIn  TraffFunc
	TraffOut TraffFunc
}

// In is used to record incoming traffic for proxy.
func (t *traffic) In(id string, b []byte) {
	if t.TraffIn != nil {
		t.TraffIn(id, b)
	}
}

// Out is used to record outgoing traffic for proxy.
func (t *traffic) Out(id string, b []byte) {
	if t.TraffOut != nil {
		t.TraffOut(id, b)
	}
}
