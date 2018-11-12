

build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
	mv furyagent bin

test-furyagent:
	cd test && $(MAKE) test

vendor:
	go mod vendor

install_deps:
	go get github.com/mitchellh/gox
