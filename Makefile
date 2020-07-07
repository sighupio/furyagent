.DEFAULT_GOAL := help
SHELL := /bin/bash

## help: show this help
.PHONY: help
help: Makefile
	@printf "\nChoose a command run in $(shell basename ${PWD}):\n"
	@sed -n 's/^##//p' $< | column -t -s ":" |  sed -e 's/^/ /'
	@echo

## require: errors on uninstalled commands.
require-%: 
	$(if $(shell command -v $* 2> /dev/null), , $(error Please install `$*` ***))

## build: generate binaries for osx and linux
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/linux/furyagent .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/darwin/furyagent .

## upload-to-s3: push to wasabisys 
.PHONY: upload-to-s3
upload-to-s3:
	aws s3 sync bin s3://sighup-releases --endpoint-url=https://s3.wasabisys.com --exclude '*' --include 'linux/furyagent' --include 'darwin/furyagent' 

## test: run tests
.PHONY: test
test:
	cd test && $(MAKE) test

## vendor: download the dependencies for go
.PHONY: vendor
vendor:
	go mod vendor

## bumpversion-test: Check code syntax
.PHONY: bumpversion-test
bumpversion-test: TARGET ?= minor
bumpversion-test: require-bumpversion
	@bumpversion ${TARGET} --verbose --dry-run --allow-dirty

## bumpversion-major: bump to the next major (i.e: 0.0.1 => 1.0.0)
## bumpversion-minor: bump to the next minor (i.e: 1.0.0 => 1.1.0)
## bumpversion-patch: bump to the next patch (i.e: 1.1.0 => 1.1.1)
bumpversion-%:
	@TARGET=$* ${MAKE} bumpversion-test
	@bumpversion $*
