package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
	"go.virtualstaticvoid.com/eventinator/config"
	"go.virtualstaticvoid.com/eventinator/metrics"
	"go.virtualstaticvoid.com/eventinator/protobuf"
	"go.virtualstaticvoid.com/eventinator/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"os/signal"
	"syscall"
)

//go:generate protoc --proto_path protobuf --go_out=protobuf --go_opt=paths=source_relative --go-grpc_out=protobuf --go-grpc_opt=paths=source_relative api.proto
//go:generate protoc --proto_path protobuf --go_out=protobuf --go_opt=paths=source_relative eventinator.proto
//go:generate protoc --proto_path protobuf --go_out=protobuf --go_opt=paths=source_relative internal.proto

func main() {

	cfg := config.ParseCommandLine()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	if cfg.VerboseLogging {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infof("Starting server on port %d", cfg.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Panicf("Failed %v", err)
	}

	metric := metrics.NewInstrumentation(cfg.MetricsEnabled)
	metric.Serve(cfg.MetricsPort)

	// connect to STAN
	log.Infof("Connecting to NATS on %s", cfg.ServerURL)
	conn, err := stan.Connect(cfg.ClusterId, cfg.ClientId, stan.NatsURL(cfg.ServerURL))
	if err != nil {
		log.Panicf("Failed to connect to NATS streaming server: %v", err)
	}
	defer conn.Close()

	svc := service.NewService(conn, metric)

	var opts []grpc.ServerOption
	if cfg.Secure {
		transCred, err := credentials.NewServerTLSFromFile(cfg.CertificateFile, cfg.CertificateKeyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(transCred)}
	}

	server := grpc.NewServer(opts...)
	protobuf.RegisterAPIServer(server, svc)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Gracefully shutting down server...")
		server.GracefulStop() // unblocks
	}()

	log.Info("Waiting for requests...")
	err = server.Serve(listener) // blocking
	if err != nil {
		log.Panicf("Failed to start server: %v", err)
	}

	log.Info("Stopped server.")
}
