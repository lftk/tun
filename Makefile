.PHONY: build-tunc build-tuns dist

VERSION ?= `git describe --tags --always`
PACKAGE := github.com/4396/tun
DIST_DIRS := find * -type d -exec
LDFLAGS := -ldflags "-w -s"

build-tunc:
	go build ${LDFLAGS} -o ./bin/tunc ./cmd/tunc/*.go

build-tuns:
	go build ${LDFLAGS} -o ./bin/tuns ./cmd/tuns/*.go

dist:
	go get -u github.com/franciscocpg/gox

	gox -verbose ${LDFLAGS} \
	-os="linux windows" \
	-arch="amd64 386" \
	-osarch="darwin/amd64 linux/arm" \
	-output="dist/{{.OS}}_{{.Arch}}/{{.Dir}}" ${PACKAGE}/cmd/tuns ${PACKAGE}/cmd/tunc

	cd dist && \
	$(DIST_DIRS) cp ../conf/*.ini {} \; && \
	$(DIST_DIRS) tar -zcf tun_${VERSION}_{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r tun_${VERSION}_{}.zip {} \; && \
	cd ..

clean-tunc:
	rm -f ./bin/tunc

clean-tuns:
	rm -f ./bin/tuns

clean-dist:
	rm -rf ./dist

build: build-tunc build-tuns

clean: clean-tunc clean-tuns clean-dist

dep:
	go get github.com/golang/glog
	go get github.com/golang/snappy
	go get github.com/golang/sync/syncmap
	go get github.com/pkg/errors
	go get github.com/xtaci/smux
	go get gopkg.in/ini.v1
