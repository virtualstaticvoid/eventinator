package service

import (
	"fmt"
	"google.golang.org/grpc/metadata"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/nats-io/go-nats-streaming"
	"go.virtualstaticvoid.com/eventinator/metrics"

	log "github.com/sirupsen/logrus"
	pb "go.virtualstaticvoid.com/eventinator/protobuf"
)

type Service struct {
	pb.APIServer

	connection stan.Conn
	metric     *metrics.Instrumentation
}

func NewService(connection stan.Conn, metric *metrics.Instrumentation) *Service {
	service := Service{
		connection: connection,
		metric:     metric,
	}
	return &service
}

func (s *Service) Publish(stream pb.API_PublishServer) error {

	iid, _ := uuid.NewUUID()
	logctx := log.WithField("_call", iid).WithField("_method", "Publish")

	s.metric.Publishers.Inc()
	defer s.metric.Publishers.Dec()

	msgCount := 0
	var err error

	// publishers stream one or more publish requests
	for {

		start := time.Now()

		// receive request
		req, err := stream.Recv()
		if err == io.EOF {
			// done
			break
		}
		if err != nil {
			logctx.Errorf("Failed to receive publish request - %q", err.Error())
			break
		}

		logctx2 := logctx.WithField("RequestId", req.RequestId).WithField("Topic", req.Topic)
		logctx2.Debug(req)

		id, _ := uuid.NewUUID()
		createdAt, _ := ptypes.TimestampProto(time.Now().UTC())
		correlationId := req.CorrelationId
		if len(correlationId) == 0 {
			cid, _ := uuid.NewUUID()
			correlationId = cid.String()
		}

		event := &pb.Event{
			Id:            id.String(),
			Version:       req.Version,
			Source:        req.Source,
			Payload:       req.Payload,
			CorrelationId: req.CorrelationId,
			MetaData:      req.MetaData,
			ContentType:   fmt.Sprintf("application/x-protobuf; messageType=%q", req.Payload.TypeUrl),
			CreatedAt:     createdAt,
		}

		if ctx := req.GetSystemContext(); ctx != nil {
			event.Context = &pb.Event_SystemContext{
				SystemContext: ctx,
			}
		} else if ctx := req.GetTenantContext(); ctx != nil {
			event.Context = &pb.Event_TenantContext{
				TenantContext: ctx,
			}
		}

		data, err := proto.Marshal(event)
		if err != nil {
			logctx2.Errorf("Failed to marshal data - %q", err.Error())
			break
		}

		if err := s.connection.Publish(req.Topic, data); err != nil {
			logctx2.Errorf("Failed to publish data - %q", err.Error())
			break
		}
		logctx2.Debugf("Published %q", event.Id)

		res := &pb.PublishResponse{
			RequestId:     req.RequestId,
			MessageId:     id.String(),
			CorrelationId: correlationId,
		}

		err = stream.Send(res)
		if err != nil {
			logctx2.Errorf("Failed to send response - %q", err.Error())
			break
		}

		msgCount++

		s.metric.MessagesReceived.Inc()
		s.metric.PublisherPerformance.Observe(time.Since(start).Seconds())

	} // for

	logctx.Debugf("Processed %d messages", msgCount)

	return err
}

func (s *Service) Subscribe(req *pb.SubscribeRequest, stream pb.API_SubscribeServer) error {

	iid, _ := uuid.NewUUID()
	logctx := log.WithField("_call", iid).WithField("_method", "Subscribe").WithField("RequestId", req.RequestId).WithField("Topic", req.Topic)
	logctx.Debug(req)

	s.metric.Subscribers.Inc()
	defer s.metric.Subscribers.Dec()

	opt := []stan.SubscriptionOption{}
	opt = append(opt, stan.SetManualAckMode())

	// TODO: validate parameters

	switch req.DeliveryOption {
	case pb.DeliveryOption_NewOnly:
		// nop, this is the default

	case pb.DeliveryOption_DeliverAllAvailable:
		opt = append(opt, stan.DeliverAllAvailable())

	case pb.DeliveryOption_StartAtSequence:
		opt = append(opt, stan.StartAtSequence(req.StartAtSequence))

	case pb.DeliveryOption_StartAtTime:
		startAt, _ := ptypes.Timestamp(req.StartAtTime)
		opt = append(opt, stan.StartAtTime(startAt))

	case pb.DeliveryOption_StartAtDuration:
		startAt, _ := ptypes.Duration(req.StartAtDuration)
		opt = append(opt, stan.StartAtTimeDelta(startAt))

	case pb.DeliveryOption_StartWithLastReceived:
		opt = append(opt, stan.StartWithLastReceived())

	case pb.DeliveryOption_StartAfterLastProcessed:
		opt = append(opt, stan.DurableName(req.DurableName))
	}

	msgCount := 0

	handler := func(msg *stan.Msg) {

		start := time.Now()

		logctx2 := logctx.WithField("Sequence", msg.Sequence)
		logctx2.Debug(msg)

		event := new(pb.Event)
		err := proto.Unmarshal(msg.Data, event)
		if err != nil {
			logctx2.Errorf("Failed to unmarshal data - %q", err.Error())
			return
		}

		receivedAt, _ := ptypes.TimestampProto(time.Now().UTC())

		resp := &pb.SubscribeResponse{
			RequestId:     req.RequestId,
			MessageId:     event.Id,
			Topic:         msg.Subject,
			Version:       event.Version,
			Source:        event.Source,
			Payload:       event.Payload,
			CorrelationId: event.CorrelationId,
			MetaData:      event.MetaData,
			ContentType:   event.ContentType,
			CreatedAt:     event.CreatedAt,
			ReceivedAt:    receivedAt,
			Sequence:      msg.Sequence,
			Redelivered:   msg.Redelivered,
		}

		if ctx := event.GetSystemContext(); ctx != nil {
			resp.Context = &pb.SubscribeResponse_SystemContext{
				SystemContext: ctx,
			}
		} else if ctx := event.GetTenantContext(); ctx != nil {
			resp.Context = &pb.SubscribeResponse_TenantContext{
				TenantContext: ctx,
			}
		}

		// TODO: need to check if the client fails to process
		// the message, whether the error gets raised back
		// so that the message isn't acknowledged, to cause a redelivery

		header := metadata.MD{}
		header.Append("foo", fmt.Sprint(start))
		stream.SetHeader(header)

		err = stream.Send(resp)
		if err != nil {
			logctx2.Errorf("Error sending response - %q", err.Error())
			return
		}
		logctx2.Debugf("Sent %q", event.Id)

		// TODO
		// TODO: need to decide how messages can be acknowledged reliably
		// TODO: e.g. somehow the client needs to confirm receipt
		// TODO: but we don't want to have to block/wait here though...
		// TODO

		err = msg.Ack()
		if err != nil {
			logctx2.Errorf("Error acknowledging message - %q", err.Error())
			return
		}

		msgCount++

		s.metric.MessagesDelivered.Inc()
		s.metric.SubscriberPerformance.Observe(time.Since(start).Seconds())

	} // handler

	var subscription stan.Subscription
	var err error

	if len(req.Group) > 0 {
		logctx.Debugf("Subscribing to queue group %q", req.Group)
		subscription, err = s.connection.QueueSubscribe(req.Topic, req.Group, handler, opt...)
	} else {
		logctx.Debug("Subscribing to queue")
		subscription, err = s.connection.Subscribe(req.Topic, handler, opt...)
	}

	if err != nil {
		logctx.Errorf("Failed to subscribe to queue - %q", err.Error())

	} else {

		<-stream.Context().Done() // blocking

		logctx.Debugf("Processed %d messages", msgCount)

		// persistent clients must keep subscription active
		if req.DeliveryOption == pb.DeliveryOption_StartAfterLastProcessed {
			logctx.Debugf("Closing subscription")
			err = subscription.Close()
		} else {
			logctx.Debugf("Removing subscription")
			err = subscription.Unsubscribe()
		}

		if err != nil {
			logctx.Errorf("Failed to finalise subscription - %q", err.Error())
		}

	}

	return err
}
