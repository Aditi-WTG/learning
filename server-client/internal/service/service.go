package service

import (
	"context"
	"encoding/json"
	"fmt"
	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
	storepb "server-client/pb"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	topicOrderCreated = "order.created"
	topicReportDaily  = "report.daily"
)

type EventBusService struct {
	storepb.UnimplementedEventBusServer
	b          *broker.Broker
	aggregator *reporting.ReportAggregator
}

func NewEventBusService(b *broker.Broker, aggregator *reporting.ReportAggregator) *EventBusService {
	return &EventBusService{b: b, aggregator: aggregator}
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

	if topic == topicOrderCreated && s.aggregator != nil {
		var order models.Order
		if err := json.Unmarshal([]byte(body), &order); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid %s payload: %v", topicOrderCreated, err)
		}

		report, err := s.aggregator.ProcessOrder(order)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid order: %v", err)
		}

		if err := s.publishDailyReportEvent(id, report); err != nil {
			return nil, err
		}
	}

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

func (s *EventBusService) publishDailyReportEvent(baseID string, report models.ShippingReport) error {
	reportBody, err := json.Marshal(report)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to serialize daily report: %v", err)
	}

	s.b.Publish(topicReportDaily, &storepb.Message{
		Id:    baseID + "-report",
		Topic: topicReportDaily,
		Body:  string(reportBody),
	})

	return nil
}
