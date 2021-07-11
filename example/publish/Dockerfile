ARG GO_VERSION
ARG DEBIAN_VERSION

FROM golang:${GO_VERSION} as builder

WORKDIR /app

RUN mkdir -p example/publish
WORKDIR /app/example/publish

COPY go.* ./
RUN go mod download

WORKDIR /app
COPY . .

WORKDIR /app/example/publish
RUN CGO_ENABLED=0 go build -o publish

FROM debian:${DEBIAN_VERSION}

COPY --from=builder /app/example/publish/publish /usr/bin/publish

ENTRYPOINT ["/usr/bin/publish"]
