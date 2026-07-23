package service

import (
	"context"
	"encoding/json"
	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
	storepb "server-client/pb"
	"strings"

	"github.com/google/uuid"
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

func NewEventBusService(b *broker.Broker, aggregator *reporting.ReportAggregator) (*EventBusService, error) {
	if aggregator == nil {
		return nil, status.Error(codes.FailedPrecondition, "Report aggregator is not configured")
	}

	return &EventBusService{b: b, aggregator: aggregator}, nil
}

func (s *EventBusService) Publish(ctx context.Context, req *storepb.PublishRequest) (*storepb.PublishAck, error) {
	topic := strings.TrimSpace(req.GetTopic())
	body := strings.TrimSpace(req.GetBody())

	if topic == "" {
		return nil, status.Error(codes.InvalidArgument, "Topic is required")
	}

	if body == "" {
		return nil, status.Error(codes.InvalidArgument, "Body is required")
	}

	id := uuid.NewString()

	if topic != topicOrderCreated {
		return nil, status.Errorf(codes.InvalidArgument, "Unsupported topic: %s", topic)
	}

	var order models.Order
	if err := json.Unmarshal([]byte(body), &order); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid %s payload: %v", topicOrderCreated, err)
	}

	report, err := s.aggregator.ProcessOrder(order)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid order: %v", err)
	}

	if err := s.publishDailyReportEvent(id, report); err != nil {
		return nil, err
	}

	return &storepb.PublishAck{Id: id}, nil
}

func (s *EventBusService) Subscribe(req *storepb.SubscribeRequest, stream grpc.ServerStreamingServer[storepb.Message]) error {
	topic := strings.TrimSpace(req.Topic)

	if topic == "" {
		return status.Error(codes.InvalidArgument, "Topic is required")
	}

	ch, err := s.b.AddSubscriber(topic)

	if err != nil {
		return status.Errorf(codes.Internal, "Failed to subscribe: %v", err)
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

func (s *EventBusService) GetReportByDate(ctx context.Context, req *storepb.GetReportByDateRequest) (*storepb.GetReportByDateResponse, error) {
	date := strings.TrimSpace(req.GetDate())
	if date == "" {
		return nil, status.Error(codes.InvalidArgument, "Date is required")
	}

	report, ok := s.aggregator.Snapshot(date)
	if !ok {
		return nil, status.Error(codes.NotFound, "Report not found for date")
	}

	return &storepb.GetReportByDateResponse{Report: toProtoReport(report)}, nil
}

func (s *EventBusService) GetAllReports(ctx context.Context, req *storepb.GetAllReportsRequest) (*storepb.GetAllReportsResponse, error) {
	all := s.aggregator.SnapshotAll()
	reports := make([]*storepb.Report, 0, len(all))
	for _, report := range all {
		reports = append(reports, toProtoReport(report))
	}

	return &storepb.GetAllReportsResponse{Reports: reports}, nil
}

func (s *EventBusService) publishDailyReportEvent(baseID string, report models.ShippingReport) error {
	reportBody, err := json.Marshal(report)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to serialize daily report: %v", err)
	}

	s.b.Publish(topicReportDaily, &storepb.Message{
		Id:    baseID + "-report",
		Topic: topicReportDaily,
		Body:  string(reportBody),
	})

	return nil
}

func toProtoReport(report models.ShippingReport) *storepb.Report {
	items := make(map[string]*storepb.ItemSummary, len(report.Items))
	for itemID, summary := range report.Items {
		items[itemID] = &storepb.ItemSummary{
			Name:     summary.Name,
			Quantity: int32(summary.Quantity),
		}
	}

	ordersByDestination := make(map[string]int32, len(report.OrdersByDestination))
	for destination, count := range report.OrdersByDestination {
		ordersByDestination[destination] = int32(count)
	}

	ordersByCustomer := make(map[string]int32, len(report.OrdersByCustomer))
	for customerID, count := range report.OrdersByCustomer {
		ordersByCustomer[customerID] = int32(count)
	}

	duplicateOrdersByCustomer := make(map[string]int32, len(report.DuplicateOrdersByCustomer))
	for customerID, count := range report.DuplicateOrdersByCustomer {
		duplicateOrdersByCustomer[customerID] = int32(count)
	}

	return &storepb.Report{
		Date:                      report.Date,
		TotalOrders:               int32(report.TotalOrders),
		TotalCost:                 report.TotalCost,
		Items:                     items,
		OrdersByDestination:       ordersByDestination,
		DuplicateOrders:           int32(report.DuplicateOrders),
		OrdersByCustomer:          ordersByCustomer,
		DuplicateOrdersByCustomer: duplicateOrdersByCustomer,
	}
}
