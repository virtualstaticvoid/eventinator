###
# build stage
###

FROM golang:1.11 as builder

# need unzip
RUN apt-get -q update \
 && apt-get -qy install \
 	unzip \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

# install protoc compiler
ARG PB_VER=3.1.0
ARG PB_URL=https://github.com/google/protobuf/releases/download/v${PB_VER}/protoc-${PB_VER}-linux-x86_64.zip

RUN mkdir -p /tmp/protoc \
 && curl -L ${PB_URL} > /tmp/protoc/protoc.zip \
 && cd /tmp/protoc \
 && unzip protoc.zip \
 && cp /tmp/protoc/bin/protoc /usr/local/bin \
 && cp -R /tmp/protoc/include/* /usr/local/include \
 && chmod go+rx /usr/local/bin/protoc \
 && cd /tmp \
 && rm -rf protoc

# install go plugin for protoc compiler
RUN go get github.com/golang/protobuf/protoc-gen-go

# setup source directories, under $GOPATH
RUN mkdir -p /go/src/go.virtualstaticvoid.com/eventinator
WORKDIR /go/src/go.virtualstaticvoid.com/eventinator

# copy over sources
COPY . .

# compile proto files
RUN cd protobuf && protoc --proto_path=. --proto_path=include --go_out=plugins=grpc:. *.proto

ENV GO111MODULE=on

# build sources
RUN go install go.virtualstaticvoid.com/eventinator

###
# runtime stage
###

FROM debian:stretch

# copy builder binaries
COPY --from=builder /go/bin/eventinator /usr/bin/eventinator

# service
EXPOSE 5300

# metrics
EXPOSE 9000

ENTRYPOINT ["/usr/bin/eventinator"]
