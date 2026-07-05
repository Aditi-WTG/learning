package service

import (
	"context"
	"fmt"
	"server-client/internal/broker"
	storepb "server-client/pb"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventBusService struct {
	storepb.UnimplementedEventBusServer
	b *broker.Broker
}

func NewEventBusService(b *broker.Broker) *EventBusService {
	return &EventBusService{b: b}
}

func (s *EventBusService) Publish(ctx context.Context, req *storepb.PublishRequest) (*storepb.PublishAck, error) {
	topic := strings.TrimSpace(req.GetTopic())
	body := strings.TrimSpace(req.GetBody())

	if topic == "" {
		return nil, status.Error(codes.InvalidArgument, "topic is required")
	}

	if body == "" {
		return nil, status.Error(codes.InvalidArgument, "body is required")
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	msg := &storepb.Message{
		Id:    id,
		Topic: topic,
		Body:  body,
	}

	s.b.Publish(topic, msg)

	return &storepb.PublishAck{Id: id}, nil
}

func (s *EventBusService) Subscribe(req *storepb.SubscribeRequest, stream grpc.ServerStreamingServer[storepb.Message]) error {
	topic := strings.TrimSpace(req.Topic)

	if topic == "" {
		return status.Error(codes.InvalidArgument, "topic is required")
	}

	ch, err := s.b.AddSubscriber(topic)

	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	defer s.b.RemoveSubscriber(topic, ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}

			err := stream.Send(msg)
			if err != nil {
				return err
			}
		}
	}
}
