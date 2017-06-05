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
