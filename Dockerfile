ARG GO_VERSION
ARG DEBIAN_VERSION

FROM golang:${GO_VERSION} as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o eventinator

FROM debian:${DEBIAN_VERSION}

COPY --from=builder /app/eventinator /usr/bin/eventinator

# grpc
EXPOSE 5000

# metrics
EXPOSE 9000

ENTRYPOINT ["/usr/bin/eventinator"]
