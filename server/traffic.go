package server

type traffic struct {
	TraffIn  TraffFunc
	TraffOut TraffFunc
}

func (t *traffic) In(id string, b []byte) {
	if t.TraffIn != nil {
		t.TraffIn(id, b)
	}
}

func (t *traffic) Out(id string, b []byte) {
	if t.TraffOut != nil {
		t.TraffOut(id, b)
	}
}
