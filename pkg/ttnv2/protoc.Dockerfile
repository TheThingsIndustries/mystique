FROM ubuntu:18.04 AS protobuf_builder

RUN apt-get update && apt-get install -y wget unzip

ENV PROTOBUF_VERSION=3.7.1

RUN wget -O /tmp/protoc.zip https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip && \
    unzip /tmp/protoc.zip -d /usr/local

FROM golang:1.12 AS go_builder

ENV GOLANG_PROTOBUF_VERSION=1.3.1

RUN git clone https://github.com/golang/protobuf.git $GOPATH/src/github.com/golang/protobuf && \
    cd $GOPATH/src/github.com/golang/protobuf && \
    git checkout v${GOLANG_PROTOBUF_VERSION}

ENV GOGO_PROTOBUF_VERSION=1.2.1

RUN git clone https://github.com/gogo/protobuf.git $GOPATH/src/github.com/gogo/protobuf && \
    cd $GOPATH/src/github.com/gogo/protobuf && \
    git checkout v${GOGO_PROTOBUF_VERSION}

ENV GRPC_GO_VERSION=1.20.1

RUN git clone https://github.com/grpc/grpc-go.git $GOPATH/src/google.golang.org/grpc && \
    cd $GOPATH/src/google.golang.org/grpc && \
    git checkout v${GRPC_GO_VERSION}

ENV GRPC_GATEWAY_VERSION=1.9.0

RUN git clone https://github.com/grpc-ecosystem/grpc-gateway.git $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway && \
    cd $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway && \
    git checkout v${GRPC_GATEWAY_VERSION}

RUN go get -d \
    google.golang.org/genproto/... \
    github.com/golang/glog/... \
    github.com/ghodss/yaml/...

RUN go install \
    github.com/golang/protobuf/protoc-gen-go \
    github.com/gogo/protobuf/protoc-gen-combo \
    github.com/gogo/protobuf/protoc-gen-gofast \
    github.com/gogo/protobuf/protoc-gen-gogo \
    github.com/gogo/protobuf/protoc-gen-gogofast \
    github.com/gogo/protobuf/protoc-gen-gogofaster \
    github.com/gogo/protobuf/protoc-gen-gogoslick \
    github.com/gogo/protobuf/protoc-gen-gogotypes \
    github.com/gogo/protobuf/protoc-gen-gostring \
    github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

FROM ubuntu:18.04

COPY --from=protobuf_builder /usr/local/bin/protoc /usr/local/bin/protoc
COPY --from=protobuf_builder /usr/local/include/google/protobuf /usr/local/include/google/protobuf
COPY --from=go_builder /go/bin/protoc-gen-* /usr/local/bin/
COPY --from=go_builder /go/src/github.com/gogo/protobuf/gogoproto/gogo.proto /usr/local/include/github.com/gogo/protobuf/gogoproto/gogo.proto
COPY --from=go_builder /go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api /usr/local/include/google/api
COPY --from=go_builder /go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/rpc /usr/local/include/google/rpc

ENTRYPOINT ["/usr/local/bin/protoc", "-I/usr/local/include"]
