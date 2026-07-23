package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	storepb "server-client/pb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	mode := flag.String("mode", "", "pub, sub, report, or report-all")
	topic := flag.String("topic", "", "topic name")
	date := flag.String("date", "", "report date in YYYY-MM-DD (for report mode)")
	addr := flag.String("addr", "localhost:50053", "server address")
	message := flag.String("message", "", "message body (for publisher)")
	flag.Parse()

	if *mode == "" {
		log.Fatal("Flag - mode is required")
	}

	if *topic == "" && *mode != "report" && *mode != "report-all" {
		log.Fatal("Flag - topic is required")
	}

	cred := insecure.NewCredentials()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(cred))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer conn.Close()

	client := storepb.NewEventBusClient(conn)

	switch *mode {
	case "pub":
		publish(client, *topic, *message)
	case "sub":
		subscribe(client, *topic)
	case "report":
		getReportByDate(client, *date)
	case "report-all":
		getAllReports(client)
	default:
		log.Fatalf("Invalid mode: %s (supported: pub, sub, report, report-all)", *mode)
	}
}

func publish(client storepb.EventBusClient, topic string, message string) {
	if message == "" {
		log.Fatal("Flag - message is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &storepb.PublishRequest{
		Topic: topic,
		Body:  message,
	}

	resp, err := client.Publish(ctx, req)
	if err != nil {
		log.Fatalf("Publish failed: %v", err)
	}

	log.Printf("published message with ID: %s\n", resp.Id)
}

func subscribe(client storepb.EventBusClient, topic string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Subscribe(ctx, &storepb.SubscribeRequest{Topic: topic})
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	log.Printf("subscribed to topic: %s\n", topic)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed by server")
			return
		}
		if err != nil {
			log.Fatalf("Receive error: %v", err)
		}

		log.Printf("Topic: %s, Message: %s\n", msg.Topic, msg.Body)
	}
}

func getReportByDate(client storepb.EventBusClient, date string) {
	if date == "" {
		log.Fatal("Flag - date is required for report mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetReportByDate(ctx, &storepb.GetReportByDateRequest{Date: date})
	if err != nil {
		log.Fatalf("Get report failed: %v", err)
	}

	formatted, err := json.MarshalIndent(resp.GetReport(), "", "  ")
	if err != nil {
		log.Fatalf("Failed to format report: %v", err)
	}

	log.Println(string(formatted))
}

func getAllReports(client storepb.EventBusClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetAllReports(ctx, &storepb.GetAllReportsRequest{})
	if err != nil {
		log.Fatalf("Get all reports failed: %v", err)
	}

	formatted, err := json.MarshalIndent(resp.GetReports(), "", "  ")
	if err != nil {
		log.Fatalf("Failed to format reports: %v", err)
	}

	log.Println(string(formatted))
}
