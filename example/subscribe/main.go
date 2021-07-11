package main

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.virtualstaticvoid.com/eventinator/config"
	ex "go.virtualstaticvoid.com/eventinator/example/subscribe/protobuf"
	pb "go.virtualstaticvoid.com/eventinator/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/anypb"
	"io"
	"os"
	"os/signal"
	"syscall"
)

//go:generate protoc --proto_path=../../protobuf --proto_path=protobuf --go_out=protobuf --go_opt=paths=source_relative test.proto

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

	log.Infof("Starting subscriber, connecting to %q", serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAPIClient(conn)

	ctx := contextWithSignal(context.Background())

	// get topic and version for this message type
	msgOpt := (&ex.OrderCreated{}).ProtoReflect().Descriptor().Options()
	topic := proto.GetExtension(msgOpt, pb.E_Topic).(string)

	requestId, _ := uuid.NewUUID()
	req := &pb.SubscribeRequest{
		RequestId:      requestId.String(),
		Topic:          topic,
		DeliveryOption: pb.DeliveryOption_StartWithLastReceived,
		DurableName:    "subscribe_test",
	}

	var callOpts []grpc.CallOption
	callOpts = append(callOpts, grpc.WaitForReady(true))

	stream, err := client.Subscribe(ctx, req, callOpts...)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.CloseSend()

	// subscribers can stream responses, so loop until we don't want any more.
	for {

		// receive response
		res, err := stream.Recv()

		if err == io.EOF {
			// done, no more messages
			// this won't happen, since the server is subscribing
			// to a NATS queue/stream, unless the server terminates
			// in which case subscriber needs to be fault tolerant
			// and attempt retry subscription
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		var orderCreated ex.OrderCreated
		err = res.Payload.UnmarshalTo(&orderCreated)
		if err != nil {
			log.Fatal(err)
		}
		header, _ := stream.Header()

		log.Infof("Order {id: %q, ref: %q} [%s]\n", orderCreated.Id, orderCreated.OrderReference, header)

		// TODO: error conditions

	} // for..ever

	log.Info("done...")
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
