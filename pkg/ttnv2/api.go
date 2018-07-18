//go:generate docker build -t protoc -f protoc.Dockerfile .
//go:generate docker run --rm -v $PWD:/go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 -w /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 protoc -I /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 --gogofaster_out=plugins=grpc:/go/src api.proto
//go:generate docker run --rm -v $PWD:/go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 -w /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 protoc -I /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 --gogofaster_out=Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,plugins=grpc:/go/src discovery/discovery.proto
//go:generate docker run --rm -v $PWD:/go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 -w /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 protoc -I /go/src/github.com/TheThingsIndustries/mystique/pkg/ttnv2 --gogofaster_out=Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,plugins=grpc:/go/src router/router.proto

package ttnv2
