

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/linux/furyagent .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/darwin/furyagent .

upload-to-s3:
	aws s3 sync bin s3://sighup-releases --endpoint-url=https://s3.wasabisys.com --exclude '*' --include 'linux/furyagent' --include 'darwin/furyagent' 

test-furyagent:
	cd test && $(MAKE) test

vendor:
	go mod vendor