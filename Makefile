# Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

SHELL = bash

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	dep ensure -v -vendor-only

test:
	go test -cover ./...

vet:
	go vet ./...

fmt:
	[[ -z `go fmt ./... | tee -a /dev/stderr` ]]
