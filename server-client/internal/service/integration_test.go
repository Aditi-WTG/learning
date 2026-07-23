package service

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
	storepb "server-client/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestEventBusEndToEnd(t *testing.T) {
	const bufSize = 1024 * 1024

	catalog := []models.Item{
		{ItemID: "I001", Name: "apple", UnitPrice: 10},
	}

	b := broker.NewBroker(16)
	a := reporting.NewReportAggregator(catalog)
	svc, err := NewEventBusService(b, a)
	if err != nil {
		t.Fatalf("unexpected service init error: %v", err)
	}

	listener := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	storepb.RegisterEventBusServer(grpcServer, svc)

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		_ = grpcServer.Serve(listener)
	}()
	defer func() {
		grpcServer.Stop()
		<-serverDone
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn server: %v", err)
	}
	defer conn.Close()

	client := storepb.NewEventBusClient(conn)

	subCtx, subCancel := context.WithCancel(context.Background())
	defer subCancel()

	subStream, err := client.Subscribe(subCtx, &storepb.SubscribeRequest{Topic: topicReportDaily})
	if err != nil {
		t.Fatalf("failed to subscribe report.daily: %v", err)
	}

	order := models.Order{
		OrderID:     "O9001",
		CustomerID:  "C9001",
		Items:       []models.OrderItem{{ItemID: "I001", Quantity: 3}},
		Destination: "Bangalore",
		Date:        "2026-07-12",
	}
	bodyBytes, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal order payload: %v", err)
	}

	pubResp, err := client.Publish(context.Background(), &storepb.PublishRequest{
		Topic: topicOrderCreated,
		Body:  string(bodyBytes),
	})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if pubResp.GetId() == "" {
		t.Fatal("expected non-empty publish ack id")
	}

	recvCtx, recvCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer recvCancel()

	recvDone := make(chan *storepb.Message, 1)
	recvErr := make(chan error, 1)
	go func() {
		msg, recvErrInner := subStream.Recv()
		if recvErrInner != nil {
			recvErr <- recvErrInner
			return
		}
		recvDone <- msg
	}()

	var reportMsg *storepb.Message
	select {
	case <-recvCtx.Done():
		t.Fatal("timed out waiting for report.daily event")
	case err := <-recvErr:
		t.Fatalf("failed receiving report.daily event: %v", err)
	case msg := <-recvDone:
		reportMsg = msg
	}

	if reportMsg.GetTopic() != topicReportDaily {
		t.Fatalf("expected topic %s, got %s", topicReportDaily, reportMsg.GetTopic())
	}

	reportByDate, err := client.GetReportByDate(context.Background(), &storepb.GetReportByDateRequest{Date: "2026-07-12"})
	if err != nil {
		t.Fatalf("GetReportByDate failed: %v", err)
	}
	if reportByDate.GetReport().GetTotalOrders() != 1 {
		t.Fatalf("expected TotalOrders=1, got %d", reportByDate.GetReport().GetTotalOrders())
	}
	if reportByDate.GetReport().GetTotalCost() != 30 {
		t.Fatalf("expected TotalCost=30, got %v", reportByDate.GetReport().GetTotalCost())
	}

	allReports, err := client.GetAllReports(context.Background(), &storepb.GetAllReportsRequest{})
	if err != nil {
		t.Fatalf("GetAllReports failed: %v", err)
	}
	if len(allReports.GetReports()) != 1 {
		t.Fatalf("expected 1 report, got %d", len(allReports.GetReports()))
	}
	if allReports.GetReports()[0].GetDate() != "2026-07-12" {
		t.Fatalf("expected report date 2026-07-12, got %s", allReports.GetReports()[0].GetDate())
	}
}
