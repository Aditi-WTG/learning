package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"server-client/internal/broker"
	"server-client/internal/models"
	"server-client/internal/reporting"
	"server-client/internal/service"
	storepb "server-client/pb"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	b := broker.NewBroker(16)
	catalog, err := loadCatalog()
	if err != nil {
		log.Fatalf("Failed to load catalog: %v", err)
	}

	aggregator := reporting.NewReportAggregator(catalog)
	svc, err := service.NewEventBusService(b, aggregator)
	if err != nil {
		log.Fatalf("Failed to initialize event bus service: %v", err)
	}

	storepb.RegisterEventBusServer(grpcServer, svc)

	log.Println("server listening on :50053")

	shutdownCtx, shutdown := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer shutdown()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(lis)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Println("shutdown signal received, stopping server")
		grpcServer.GracefulStop()
		if err := <-serveErr; err != nil {
			log.Fatalf("Failed to stop server cleanly: %v", err)
		}
	case err := <-serveErr:
		if err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}
}

func loadCatalog() ([]models.Item, error) {
	data, err := os.ReadFile(filepath.Join("data", "catalog.json"))
	if err != nil {
		return nil, err
	}

	var catalog []models.Item
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}

	if len(catalog) == 0 {
		return nil, errors.New("Catalog is empty")
	}

	return catalog, nil
}
