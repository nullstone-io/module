.PHONY:

.DEFAULT_GOAL: default

default: test

test:
	go fmt ./...
	go test -v ./...
