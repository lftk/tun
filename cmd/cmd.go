package cmd

import (
	"encoding/json"
	"flag"

	"github.com/4396/tun/log"
	"github.com/golang/glog"
)

const (
	TCP = iota
	HTTP
)

type Proxy struct {
	Type   int
	Port   string
	Domain string
}

func Encode(p *Proxy) (desc string, err error) {
	b, err := json.Marshal(p)
	if err != nil {
		return
	}

	desc = string(b)
	return
}

func Decode(desc string, p *Proxy) (err error) {
	err = json.Unmarshal([]byte(desc), p)
	return
}

type logger struct{}

func (l *logger) Verbose(depth int, args ...interface{}) {
	if glog.V(20) {
		glog.InfoDepth(depth+1, args...)
	}
}

func (l *logger) Debug(depth int, args ...interface{}) {
	if glog.V(10) {
		glog.InfoDepth(depth+1, args...)
	}
}

func (l *logger) Info(depth int, args ...interface{}) {
	glog.InfoDepth(depth+1, args...)
}

func (l *logger) Warning(depth int, args ...interface{}) {
	glog.WarningDepth(depth+1, args...)
}

func (l *logger) Error(depth int, args ...interface{}) {
	glog.ErrorDepth(depth+1, args...)
}

func (l *logger) Fatal(depth int, args ...interface{}) {
	glog.FatalDepth(depth+1, args...)
}

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
	log.Logger = new(logger)
}
