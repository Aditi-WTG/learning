package main

import (
	"context"
	"flag"
	"io"
	"log"
	storepb "server-client/pb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	mode := flag.String("mode", "", "pub or sub")
	topic := flag.String("topic", "", "topic name")
	addr := flag.String("addr", "localhost:50053", "server address")
	message := flag.String("message", "", "message body (for publisher)")
	flag.Parse()

	if *mode == "" {
		log.Fatal("flag - mode is required")
	}

	if *topic == "" {
		log.Fatal("flag - topic is required")
	}

	cred := insecure.NewCredentials()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(cred))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer conn.Close()

	client := storepb.NewEventBusClient(conn)

	switch *mode {
	case "pub":
		publish(client, *topic, *message)
	case "sub":
		subscribe(client, *topic)
	default:
		log.Fatalf("invalid mode: %s", *mode)
	}
}

func publish(client storepb.EventBusClient, topic string, message string) {
	log.Printf("DEBUG topic=%q message=%q", topic, message)
	if message == "" {
		log.Fatal("flag - message is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &storepb.PublishRequest{
		Topic: topic,
		Body:  message,
	}

	resp, err := client.Publish(ctx, req)
	if err != nil {
		log.Fatalf("publish failed: %v", err)
	}

	log.Printf("published message with ID: %s\n", resp.Id)
}

func subscribe(client storepb.EventBusClient, topic string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Subscribe(ctx, &storepb.SubscribeRequest{Topic: topic})
	if err != nil {
		log.Fatalf("subscribe failed: %v", err)
	}

	log.Printf("subscribed to topic: %s\n", topic)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed by server")
			return
		}
		if err != nil {
			log.Fatalf("receive error: %v", err)
		}

		log.Printf("Topic: %s, Message: %s\n", msg.Topic, msg.Body)
	}
}
