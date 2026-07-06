package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
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
	catalog, err := loadCatalog()
	if err != nil {
		log.Fatalf("failed to load catalog: %v", err)
	}

	aggregator := reporting.NewReportAggregator(catalog)
	svc := service.NewEventBusService(b, aggregator)

	storepb.RegisterEventBusServer(grpcServer, svc)

	log.Println("server listening on :50053")

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

func loadCatalog() ([]models.Item, error) {
	data, err := os.ReadFile("../online-store/data/catalog.json")
	if err != nil {
		return nil, err
	}

	var catalog []models.Item
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}

	if len(catalog) == 0 {
		return nil, os.ErrInvalid
	}

	return catalog, nil
}
