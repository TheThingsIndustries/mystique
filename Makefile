# Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

RELEASE_DIR ?= release

SHELL := bash

.PHONY: deps

deps:
	go mod download

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

$(RELEASE_DIR)/%-$(GOOS)-$(GOARCH): cmd/%/main.go $(wildcard pkg/*/*.go) $(wildcard pkg/*/*/*.go) go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w" -o $@$(shell go env GOEXE) $<

.PHONY: release

release: $(RELEASE_DIR)/mystique-server-$(GOOS)-$(GOARCH) $(RELEASE_DIR)/ttn-mqtt-$(GOOS)-$(GOARCH) $(RELEASE_DIR)/ttn-mqtt-bridge-$(GOOS)-$(GOARCH)

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
	docker build --build-arg bin_name=mystique-server -t thethingsindustries/mystique-server:latest .
	docker build --build-arg bin_name=ttn-mqtt -t thethingsindustries/ttn-mqtt:latest .
	docker build --build-arg bin_name=ttn-mqtt-bridge -t thethingsindustries/ttn-mqtt:bridge .

.PHONY: docker

DOCKER_TAG ?= $(shell date '+%Y%m%d%H%M')

docker-push:
	docker tag thethingsindustries/mystique-server:latest thethingsindustries/mystique-server:$(DOCKER_TAG)
	docker tag thethingsindustries/ttn-mqtt:latest thethingsindustries/ttn-mqtt:$(DOCKER_TAG)
	docker push thethingsindustries/mystique-server:$(DOCKER_TAG)
	docker push thethingsindustries/ttn-mqtt:$(DOCKER_TAG)

docker-push-latest:
	docker push thethingsindustries/mystique-server:latest
	docker push thethingsindustries/ttn-mqtt:latest
	docker push thethingsindustries/ttn-mqtt:bridge
