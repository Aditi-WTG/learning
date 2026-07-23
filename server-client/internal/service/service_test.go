package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
	storepb "server-client/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testCatalog() []models.Item {
	return []models.Item{
		{ItemID: "I001", Name: "apple", UnitPrice: 10},
		{ItemID: "I002", Name: "banana", UnitPrice: 5},
	}
}

func mustMarshalOrder(t *testing.T, order models.Order) string {
	t.Helper()

	b, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal order: %v", err)
	}
	return string(b)
}

func receiveReportMessage(t *testing.T, ch <-chan *storepb.Message) *storepb.Message {
	t.Helper()

	select {
	case msg := <-ch:
		return msg
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for report.daily message")
		return nil
	}
}

func TestPublishValidation(t *testing.T) {
	b := broker.NewBroker(4)
	a := reporting.NewReportAggregator(testCatalog())
	svc, err := NewEventBusService(b, a)
	if err != nil {
		t.Fatalf("unexpected service init error: %v", err)
	}

	validOrder := mustMarshalOrder(t, models.Order{
		OrderID:     "O1",
		CustomerID:  "C1",
		Items:       []models.OrderItem{{ItemID: "I001", Quantity: 1}},
		Destination: "Bangalore",
		Date:        "2026-07-07",
	})

	tests := []struct {
		name string
		req  *storepb.PublishRequest
		code codes.Code
	}{
		{name: "empty topic", req: &storepb.PublishRequest{Body: validOrder}, code: codes.InvalidArgument},
		{name: "empty body", req: &storepb.PublishRequest{Topic: topicOrderCreated}, code: codes.InvalidArgument},
		{name: "unsupported topic", req: &storepb.PublishRequest{Topic: "random.topic", Body: validOrder}, code: codes.InvalidArgument},
		{name: "invalid json", req: &storepb.PublishRequest{Topic: topicOrderCreated, Body: "{"}, code: codes.InvalidArgument},
		{name: "invalid order", req: &storepb.PublishRequest{Topic: topicOrderCreated, Body: mustMarshalOrder(t, models.Order{OrderID: "O2"})}, code: codes.InvalidArgument},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Publish(context.Background(), tc.req)
			if status.Code(err) != tc.code {
				t.Fatalf("expected code %v, got %v (err=%v)", tc.code, status.Code(err), err)
			}
		})
	}
}

func TestNewEventBusServiceRequiresAggregator(t *testing.T) {
	b := broker.NewBroker(4)

	svc, err := NewEventBusService(b, nil)
	if err == nil {
		t.Fatal("expected constructor error for nil aggregator")
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v (err=%v)", status.Code(err), err)
	}
	if svc != nil {
		t.Fatal("expected nil service when constructor fails")
	}
}

func TestPublishSuccessEmitsDailyReportEvent(t *testing.T) {
	b := broker.NewBroker(4)
	a := reporting.NewReportAggregator(testCatalog())
	svc, err := NewEventBusService(b, a)
	if err != nil {
		t.Fatalf("unexpected service init error: %v", err)
	}

	reportSub, err := b.AddSubscriber(topicReportDaily)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber(topicReportDaily, reportSub)

	ack, err := svc.Publish(context.Background(), &storepb.PublishRequest{
		Topic: topicOrderCreated,
		Body: mustMarshalOrder(t, models.Order{
			OrderID:     "O100",
			CustomerID:  "C100",
			Items:       []models.OrderItem{{ItemID: "I002", Quantity: 3}},
			Destination: "Mumbai",
			Date:        "2026-07-08",
		}),
	})
	if err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}
	if strings.TrimSpace(ack.GetId()) == "" {
		t.Fatal("expected non-empty ack id")
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(ack.GetId()) {
		t.Fatalf("expected UUIDv4 ack id, got %s", ack.GetId())
	}

	msg := receiveReportMessage(t, reportSub)
	if msg.GetTopic() != topicReportDaily {
		t.Fatalf("expected topic %s, got %s", topicReportDaily, msg.GetTopic())
	}
	if !strings.HasSuffix(msg.GetId(), "-report") {
		t.Fatalf("expected message id to end with -report, got %s", msg.GetId())
	}

	var report models.ShippingReport
	if err := json.Unmarshal([]byte(msg.GetBody()), &report); err != nil {
		t.Fatalf("failed to unmarshal report event body: %v", err)
	}
	if report.Date != "2026-07-08" {
		t.Fatalf("expected report date 2026-07-08, got %s", report.Date)
	}
	if report.TotalOrders != 1 {
		t.Fatalf("expected TotalOrders=1, got %d", report.TotalOrders)
	}
	if report.TotalCost != 15 {
		t.Fatalf("expected TotalCost=15, got %v", report.TotalCost)
	}
}

func TestGetReportByDate(t *testing.T) {
	b := broker.NewBroker(4)
	a := reporting.NewReportAggregator(testCatalog())
	svc, err := NewEventBusService(b, a)
	if err != nil {
		t.Fatalf("unexpected service init error: %v", err)
	}

	_, err = svc.GetReportByDate(context.Background(), &storepb.GetReportByDateRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for empty date, got %v", status.Code(err))
	}

	_, err = svc.GetReportByDate(context.Background(), &storepb.GetReportByDateRequest{Date: "2026-07-01"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound for missing report, got %v", status.Code(err))
	}

	_, err = svc.Publish(context.Background(), &storepb.PublishRequest{
		Topic: topicOrderCreated,
		Body: mustMarshalOrder(t, models.Order{
			OrderID:     "O200",
			CustomerID:  "C200",
			Items:       []models.OrderItem{{ItemID: "I001", Quantity: 2}},
			Destination: "Delhi",
			Date:        "2026-07-09",
		}),
	})
	if err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}

	resp, err := svc.GetReportByDate(context.Background(), &storepb.GetReportByDateRequest{Date: "2026-07-09"})
	if err != nil {
		t.Fatalf("unexpected get report error: %v", err)
	}
	if resp.GetReport().GetDate() != "2026-07-09" {
		t.Fatalf("expected report date 2026-07-09, got %s", resp.GetReport().GetDate())
	}
	if resp.GetReport().GetTotalOrders() != 1 {
		t.Fatalf("expected TotalOrders=1, got %d", resp.GetReport().GetTotalOrders())
	}
}

func TestGetAllReports(t *testing.T) {
	b := broker.NewBroker(4)
	a := reporting.NewReportAggregator(testCatalog())
	svc, err := NewEventBusService(b, a)
	if err != nil {
		t.Fatalf("unexpected service init error: %v", err)
	}

	emptyResp, err := svc.GetAllReports(context.Background(), &storepb.GetAllReportsRequest{})
	if err != nil {
		t.Fatalf("unexpected error for empty reports: %v", err)
	}
	if len(emptyResp.GetReports()) != 0 {
		t.Fatalf("expected 0 reports, got %d", len(emptyResp.GetReports()))
	}

	orders := []models.Order{
		{OrderID: "O301", CustomerID: "C1", Items: []models.OrderItem{{ItemID: "I001", Quantity: 1}}, Destination: "Bangalore", Date: "2026-07-11"},
		{OrderID: "O302", CustomerID: "C2", Items: []models.OrderItem{{ItemID: "I002", Quantity: 2}}, Destination: "Mumbai", Date: "2026-07-10"},
	}
	for _, order := range orders {
		_, err := svc.Publish(context.Background(), &storepb.PublishRequest{Topic: topicOrderCreated, Body: mustMarshalOrder(t, order)})
		if err != nil {
			t.Fatalf("unexpected publish error: %v", err)
		}
	}

	resp, err := svc.GetAllReports(context.Background(), &storepb.GetAllReportsRequest{})
	if err != nil {
		t.Fatalf("unexpected get all reports error: %v", err)
	}

	if len(resp.GetReports()) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(resp.GetReports()))
	}

	if resp.GetReports()[0].GetDate() != "2026-07-10" || resp.GetReports()[1].GetDate() != "2026-07-11" {
		t.Fatalf("expected sorted dates [2026-07-10, 2026-07-11], got [%s, %s]", resp.GetReports()[0].GetDate(), resp.GetReports()[1].GetDate())
	}
}
