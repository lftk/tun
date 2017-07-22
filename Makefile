.PHONY: build-tunc build-tuns

build-tunc:
	go build -ldflags "-w -s" -o bin/tunc cmd/tunc/*.go

build-tuns:
	go build -ldflags "-w -s" -o bin/tuns cmd/tuns/*.go

clean-tunc:
	rm -f bin/tunc

clean-tuns:
	rm -f bin/tuns

build: build-tunc build-tuns

clean: clean-tunc clean-tuns

dep:
	go get github.com/golang/glog
	go get github.com/golang/snappy
	go get github.com/golang/sync/syncmap
	go get github.com/pkg/errors
	go get github.com/xtaci/smux
	go get gopkg.in/ini.v1
