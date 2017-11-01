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

dev-cert:
	go run $$(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost

release/%: cmd/%/main.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-w" -o $@ $<

docker/%: GOOS=linux
docker/%: GOARCH=amd64
docker/%: release/%
	docker build --build-arg bin_name=$(patsubst release/%,%,$<) -t thethingsindustries/$(patsubst release/%,%,$<) .

docker: docker/mystique-server docker/ttn-mqtt
