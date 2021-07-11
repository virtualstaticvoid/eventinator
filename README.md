# NATS Streaming via gRPC Prototype

Prototype to try out NATS Streaming via a gRPC service, using streaming request/response 
for publishing and subscribing to the messages.

## Usage

Build the docker image and run the service.

```
docker-compose build
docker-compose up nats service
```

Run two terminals, one for publishing messages,

```
docker-compose up publisher
```

And another for subscribing to the messages.

```
docker-compose up subscriber
```

To get metrics, visit `http://localhost:9000/metrics`.

E.g. Watch for any metric with the `eventinator` prefix

```
watch 'curl http://localhost:9000/metrics | grep ^eventinator'
```

## Hacking

On Linux, install the `protoc` tool and the golang compilers for `protobuf` and `grpc` compilers if you plan on changing any of the `*.proto` files.

```
sudo apt install protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

To (re)generate the protocol buffers and gRPC files.

```
go generate
```

## License

MIT License. Copyright (c) 2018 Chris Stefano. See LICENSE for details.
