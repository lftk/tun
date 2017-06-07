package traffic

type Traffic interface {
	In(string, []byte, int64)
	Out(string, []byte, int64)
}
