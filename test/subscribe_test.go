package test

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "go.virtualstaticvoid.com/eventinator/protobuf"
	test "go.virtualstaticvoid.com/eventinator/test"
)

func TestSubscribe(t *testing.T) {

	t.Log("starting subscriber client...")

	serverAddr := "127.0.0.1:5300"

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		t.Fatalf("grpc.Dial(_) = _, %v", err)
	}
	defer conn.Close()
	client := pb.NewAPIClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	ctx := context.Background()

	// get topic and version for this message type
	md, _ := pb.GetMessageMetadata(&test.FeatureToggled{})
	requestId, _ := uuid.NewUUID()
	// deliveryOption := pb.DeliveryOption_NewOnly
	deliveryOption := pb.DeliveryOption_StartAfterLastProcessed

	req := &pb.SubscribeRequest{
		RequestId:      requestId.String(),
		Topic:          md.Topic,
		DeliveryOption: deliveryOption,
		DurableName:    "subscribe_test",
	}

	var callopts []grpc.CallOption
	callopts = append(callopts, grpc.FailFast(true))

	stream, err := client.Subscribe(ctx, req, callopts...)
	if err != nil {
		t.Fatalf("%v.Subscribe(_) = _, %v", client, err)
	}
	defer stream.CloseSend()

	count := 0

	// subscribers can stream responses, so loop until we don't want any more.
	for {

		// receive response
		res, err := stream.Recv()
		if err == io.EOF {
			// done, no more messages
			// this won't happen, since the server is subscribing
			// to a NATS queue/stream, unless the server terminates
			break
		}
		if err != nil {
			t.Fatalf("%v.Recv(_) = _, %v", stream, err)
		}

		header, _ := stream.Header()

		t.Log(header)
		t.Log(res)

		// TODO: simulate an error condition
		//stream.SendMsg(true)

		count++
		if count > 100 {
			t.Log("stopping...")
			break
		}

		//if count > 10 {
		//	panic(true)
		//}

	}

	t.Log("done...")
}
