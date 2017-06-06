package traffic

type Traffic interface {
	In(name string, num int64)
	Out(name string, num int64)
}
