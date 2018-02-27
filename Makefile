# Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

RELEASE_DIR ?= release

GO_PATH := $(shell echo $(GOPATH) | awk -F':' '{ print $$1 }')

SHELL := bash

.PHONY: deps

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v -vendor-only

.PHONY: test

test:
	go test -cover ./...

.PHONY: vet

vet:
	go vet ./...

.PHONY: fmt

fmt:
	[[ -z `go fmt ./... | tee -a /dev/stderr` ]]

.PHONY: dev-cert

dev-cert:
	go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost

.PHONY: clean

clean:
	rm -rf release

$(RELEASE_DIR)/%-$(GOOS)-$(GOARCH): cmd/%/main.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -gcflags="all=-trimpath=$(GO_PATH)" -asmflags="all=-trimpath=$(GO_PATH)" -ldflags "-s -w" -o $@$(shell go env GOEXE) $<

.PHONY: release

release: $(RELEASE_DIR)/mystique-server-$(GOOS)-$(GOARCH) $(RELEASE_DIR)/ttn-mqtt-$(GOOS)-$(GOARCH)

releases:
	GOOS=linux GOARCH=amd64 make -j 2 release
	GOOS=linux GOARCH=386 make -j 2 release
	GOOS=linux GOARCH=arm make -j 2 release
	GOOS=darwin GOARCH=amd64 make -j 2 release
	GOOS=windows GOARCH=amd64 make -j 2 release
	GOOS=windows GOARCH=386 make -j 2 release

.PHONY: docker

docker:
	GOOS=linux GOARCH=amd64 make -j 2 release
	docker build --build-arg bin_name=mystique-server -t thethingsindustries/mystique-server .
	docker build --build-arg bin_name=ttn-mqtt -t thethingsindustries/ttn-mqtt .
