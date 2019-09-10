###
# build stage
###

ARG GO_VERSION
ARG DEBIAN_VERSION

FROM golang:${GO_VERSION} as builder

# enable (new) go modules functionality
ENV GO111MODULE=on

# setup source directories, under $GOPATH
RUN mkdir -p /go/src/go.virtualstaticvoid.com/eventinator
WORKDIR /go/src/go.virtualstaticvoid.com/eventinator

# copy over sources
COPY . .

# build sources
RUN go install -i go.virtualstaticvoid.com/eventinator

###
# runtime stage
###

FROM debian:${DEBIAN_VERSION}

# copy builder binaries
COPY --from=builder /go/bin/eventinator /usr/bin/eventinator

# service
EXPOSE 5300

# metrics
EXPOSE 9000

ENTRYPOINT ["/usr/bin/eventinator"]
