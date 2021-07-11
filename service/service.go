package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
	"go.virtualstaticvoid.com/eventinator/metrics"
	pb "go.virtualstaticvoid.com/eventinator/protobuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type service struct {
	pb.UnimplementedAPIServer

	connection stan.Conn
	metric     *metrics.Instrumentation
}

func NewService(connection stan.Conn, metric *metrics.Instrumentation) *service {
	service := service{
		connection: connection,
		metric:     metric,
	}
	return &service
}

func (s *service) Publish(stream pb.API_PublishServer) error {

	iid, _ := uuid.NewUUID()
	logCtx := log.
		WithField("_call", iid).
		WithField("_method", "Publish")

	s.metric.Publishers.Inc()
	defer s.metric.Publishers.Dec()

	var err error

	// publishers stream in one or more publish requests
	for {

		req, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			logCtx.Errorf("Failed to receive publish request - %q", err.Error())
			break
		}

		innerLogCtx := logCtx.
			WithField("RequestId", req.RequestId).
			WithField("Topic", req.Topic)
		innerLogCtx.Debug(req)

		id, _ := uuid.NewUUID()
		createdAt := timestamppb.New(time.Now().UTC())
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

		data, err := proto.Marshal(event)
		if err != nil {
			innerLogCtx.Errorf("Failed to marshal data - %q", err.Error())
			break
		}

		if err := s.connection.Publish(req.Topic, data); err != nil {
			innerLogCtx.Errorf("Failed to publish data - %q", err.Error())
			break
		}
		innerLogCtx.Debugf("Published %q", event.Id)

		res := &pb.PublishResponse{
			RequestId:     req.RequestId,
			MessageId:     id.String(),
			CorrelationId: correlationId,
		}

		err = stream.Send(res)
		if err != nil {
			innerLogCtx.Errorf("Failed to send response - %q", err.Error())
			break
		}

		s.metric.MessagesReceived.Inc()

	} // for

	return err
}

func (s *service) Subscribe(req *pb.SubscribeRequest, stream pb.API_SubscribeServer) error {

	iid, _ := uuid.NewUUID()
	logCtx := log.
		WithField("_call", iid).
		WithField("_method", "Subscribe").
		WithField("RequestId", req.RequestId).
		WithField("Topic", req.Topic)

	logCtx.Debug(req)

	s.metric.Subscribers.Inc()
	defer s.metric.Subscribers.Dec()

	var opt []stan.SubscriptionOption
	opt = append(opt, stan.SetManualAckMode())

	// TODO: validate parameters

	switch req.DeliveryOption {
	case pb.DeliveryOption_NewOnly:
		// no-op, this is the default

	case pb.DeliveryOption_DeliverAllAvailable:
		opt = append(opt, stan.DeliverAllAvailable())

	case pb.DeliveryOption_StartAtSequence:
		opt = append(opt, stan.StartAtSequence(req.StartAtSequence))

	case pb.DeliveryOption_StartAtTime:
		startAt := req.StartAtTime.AsTime()
		opt = append(opt, stan.StartAtTime(startAt))

	case pb.DeliveryOption_StartAtDuration:
		startAt := req.StartAtDuration.AsDuration()
		opt = append(opt, stan.StartAtTimeDelta(startAt))

	case pb.DeliveryOption_StartWithLastReceived:
		opt = append(opt, stan.StartWithLastReceived())

	case pb.DeliveryOption_StartAfterLastProcessed:
		opt = append(opt, stan.DurableName(req.DurableName))
	}

	handler := func(msg *stan.Msg) {

		innerLogCtx := logCtx.
			WithField("Sequence", msg.Sequence)
		innerLogCtx.Debug(msg)

		event := new(pb.Event)
		err := proto.Unmarshal(msg.Data, event)
		if err != nil {
			innerLogCtx.Errorf("Failed to unmarshal data - %q", err.Error())
			return
		}

		receivedAt := timestamppb.New(time.Now().UTC())

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

		// header := metadata.MD{}
		// header.Append("foo", fmt.Sprint(start))
		// _ = stream.SetHeader(header)

		// TODO: need to check if the client fails to process
		// the message, whether the error gets raised back
		// so that the message isn't acknowledged, to cause a redelivery

		err = stream.Send(resp)
		if err != nil {
			innerLogCtx.Errorf("Error sending response - %q", err.Error())
			return
		}
		innerLogCtx.Debugf("Sent %q", event.Id)

		// TODO
		// TODO: need to decide how messages can be acknowledged reliably
		// TODO: e.g. somehow the client needs to confirm receipt
		// TODO: but we don't want to have to block/wait here though...
		// TODO

		err = msg.Ack()
		if err != nil {
			innerLogCtx.Errorf("Error acknowledging message - %q", err.Error())
			return
		}

		s.metric.MessagesDelivered.Inc()

	} // handler

	var subscription stan.Subscription
	var err error

	if len(req.Group) > 0 {
		logCtx.Debugf("Subscribing to queue group %q", req.Group)
		subscription, err = s.connection.QueueSubscribe(req.Topic, req.Group, handler, opt...)
	} else {
		logCtx.Debug("Subscribing to queue")
		subscription, err = s.connection.Subscribe(req.Topic, handler, opt...)
	}

	if err != nil {
		logCtx.Errorf("Failed to subscribe to queue - %q", err.Error())

	} else {

		// block here
		<-stream.Context().Done()

		// persistent clients must keep subscription active
		if req.DeliveryOption == pb.DeliveryOption_StartAfterLastProcessed {
			logCtx.Debugf("Closing subscription")
			err = subscription.Close()
		} else {
			logCtx.Debugf("Removing subscription")
			err = subscription.Unsubscribe()
		}

		if err != nil {
			logCtx.Errorf("Failed to finalise subscription - %q", err.Error())
		}

	}

	return err
}
