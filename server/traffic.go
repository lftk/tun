package server

type traffic struct {
	TraffIn  TraffFunc
	TraffOut TraffFunc
}

func (t *traffic) In(name string, b []byte) {
	if t.TraffIn != nil {
		t.TraffIn(name, b)
	}
}

func (t *traffic) Out(name string, b []byte) {
	if t.TraffOut != nil {
		t.TraffOut(name, b)
	}
}
