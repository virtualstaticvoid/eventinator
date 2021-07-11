package main

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.virtualstaticvoid.com/eventinator/config"
	ex "go.virtualstaticvoid.com/eventinator/example/publish/protobuf"
	pb "go.virtualstaticvoid.com/eventinator/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:generate protoc --proto_path=../../protobuf --proto_path=protobuf --go_out=protobuf --go_opt=paths=source_relative test.proto

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	cfg := config.ParseCommandLine()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	if cfg.VerboseLogging {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	serverAddr := fmt.Sprintf("%s:%d", cfg.ServerURL, cfg.Port)
	log.Infof("Starting publisher, connecting to %q", serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAPIClient(conn)

	// start stream for publishing messages
	ctx := contextWithSignal(context.Background())
	stream, err := client.Publish(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// wire up response stream
	waitChan := make(chan struct{})
	go func() {
		for {
			// receive response to Publish
			res, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// server closed channel
					log.Info("completed")
				} else {
					log.Error(err)
				}
				close(waitChan)
				return
			}
			log.Info(res)
		} // for..ever
	}()

	err = publishMessages(ctx, stream)
	if err != nil {
		log.Error(err)
	}

	err = stream.CloseSend()
	if err != nil {
		log.Error(err)
	}

	<-waitChan

	log.Info("done...")
}

func publishMessages(ctx context.Context, stream pb.API_PublishClient) error {
	log.Info("publishing messages...")

	// get topic and version for this message type
	msgOpt := (&ex.OrderCreated{}).ProtoReflect().Descriptor().Options()
	topic := proto.GetExtension(msgOpt, pb.E_Topic).(string)
	version := proto.GetExtension(msgOpt, pb.E_Version).(string)

	metadata := make(map[string]string)
	metadata["foo"] = "bar"
	orderNum := 0

	for {

		orderNum++
		requestId, _ := uuid.NewUUID()
		payload, _ := anypb.New(&ex.OrderCreated{
			Id:             fmt.Sprintf("order-%04d", orderNum),
			OrderReference: "foobar",
		})

		req := pb.PublishRequest{
			RequestId:     requestId.String(),
			Topic:         topic,
			Version:       version,
			Source:        "publish-test-app",
			Payload:       payload,
			CorrelationId: requestId.String(),
			MetaData:      metadata,
		}

		err := stream.Send(&req)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(req.String())

		select {

		// has the context been cancelled?
		case <-ctx.Done():
			log.Info("exiting publish loop")
			return ctx.Err()

		// wait for a random time (up to 20s) to simulate a publisher
		// sending messages sporadically
		case <-time.After(time.Duration(rand.Intn(20)) * time.Second):
			// continue
		}

	}

	return nil
}

func contextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			cancel()
		}
	}()
	return newCtx
}
