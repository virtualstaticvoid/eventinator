ARG GO_VERSION
ARG DEBIAN_VERSION

FROM golang:${GO_VERSION} as builder

WORKDIR /app

RUN mkdir -p example/subscribe
WORKDIR /app/example/subscribe

COPY go.* ./
RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/example/subscribe
RUN CGO_ENABLED=0 go build -o subscribe

FROM debian:${DEBIAN_VERSION}

COPY --from=builder /app/example/subscribe/subscribe /usr/bin/subscribe

ENTRYPOINT ["/usr/bin/subscribe"]
