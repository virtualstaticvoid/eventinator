package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/go-nats-streaming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"vsv.io/eventinator/config"
	"vsv.io/eventinator/metrics"
	"vsv.io/eventinator/protobuf"
	"vsv.io/eventinator/service"

	log "github.com/sirupsen/logrus"
)

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
		panic(err)
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
		creds, err := credentials.NewServerTLSFromFile(cfg.CertificateFile, cfg.CertificateKeyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	server := grpc.NewServer(opts...)
	protobuf.RegisterAPIServer(server, svc)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
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
