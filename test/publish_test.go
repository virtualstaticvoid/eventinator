package test

import (
	"io"
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "go.virtualstaticvoid.com/eventinator/protobuf"
)

func TestPublish(t *testing.T) {

	t.Log("starting publisher client...")

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

	stream, err := client.Publish(ctx)
	if err != nil {
		t.Fatalf("%v.Publish(_) = _, %v", client, err)
	}

	waitchan := make(chan struct{})

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				t.Log("completed")
				close(waitchan)
				return
			}
			if err != nil {
				t.Fatalf("%v.Recv(_) = _, %v", stream, err)
			} else {
				t.Log(res)
			}
		}
	}()

	// get topic and version for this message type
	md, _ := pb.GetMessageMetadata(&FeatureToggled{})

	metadata := make(map[string]string)
	metadata["foo"] = "bar"

	sc := pb.PublishRequest_SystemContext{
		SystemContext: &pb.SystemContext{
			UserIdentifier: "vSv",
		},
	}

	t.Log("publishing messages...")
	count := 0

	for {

		requestId, _ := uuid.NewUUID()
		payload, _ := ptypes.MarshalAny(&FeatureToggled{
			Id:      "feature-1",
			Enabled: true,
		})

		req := pb.PublishRequest{
			RequestId:     requestId.String(),
			Topic:         md.Topic,
			Version:       md.Version,
			Context:       &sc,
			Source:        "publish-test-app",
			Payload:       payload,
			CorrelationId: requestId.String(),
			MetaData:      metadata,
		}

		err := stream.Send(&req)
		if err != nil {
			t.Fatalf("%v.Send(_) = _, %v", stream, err)
		}
		t.Log(req)

		count++
		if count > 101 {
			t.Log("stopping...")
			break
		}
	}

	stream.CloseSend()
	<-waitchan
	t.Log("done...")
}
