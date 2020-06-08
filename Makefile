SHELL := /bin/bash

.DEFAULT_GOAL: build 

build: vendor/ *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o etcd3-bootstrap-linux-amd64 -ldflags '-s'

fix:
	@echo "  >  Making sure go.mod matches the source code"
	go mod vendor
	go mod tidy
