package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/service"
)

func main() {
	cfg := config.NewConfig()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Register oracle service
	oracleSvc := service.NewOracleService(cfg)
	_ = oracleSvc // Will register with gRPC once proto is generated

	log.Printf("Oracle service listening on port %s", cfg.Port)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("Received shutdown signal, stopping gracefully...")
		grpcServer.GracefulStop()
		cancel()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	<-ctx.Done()
	log.Println("Oracle service stopped.")
}
