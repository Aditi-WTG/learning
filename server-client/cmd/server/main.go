package main

import (
	"log"
	"net"
	"server-client/internal/broker"
	"server-client/internal/service"
	storepb "server-client/pb"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	b := broker.NewBroker(16)
	svc := service.NewEventBusService(b)

	storepb.RegisterEventBusServer(grpcServer, svc)

	log.Println("server listening on :50053")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
