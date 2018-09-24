# NATS Streaming via gRPC Prototype

Prototype to try out NATS Streaming via a gRPC service, using streaming request/response for publishing and subscribing to messages.

## Usage

Build the docker image and run the service.

```
docker-compose build
docker-compose up
```

Run two terminals, one for the `publish_test.go` and the other for `subscribe_test.go`

```
go test test/publish_test.go -count=1 -v
```

And

```
go test test/subscribe_test.go -count=1 -v
```

To get metrics, visit `http://localhost:9000/metrics`.

E.g. Watch for any metric with the `eventinator` prefix

```
watch 'curl http://localhost:9000/metrics | grep ^eventinator'
```

## License

MIT License. Copyright (c) 2018 Chris Stefano. See LICENSE for details.
