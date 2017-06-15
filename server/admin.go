package server

// Administrator ...
type Administrator interface {
	AuthCmder(token string) error
	AuthProxy(name, token string) error
	Command(b []byte) ([]byte, error)
}
