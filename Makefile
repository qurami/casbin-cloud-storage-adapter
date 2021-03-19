SHELL = /bin/bash
export PATH := $(shell yarn global bin):$(PATH)

default: test

test:
	go test -race -cover -v .

benchmark:
	go test -bench=.

release:
	yarn global add semantic-release@17.2.4
	semantic-release
