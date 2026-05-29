package main

import (
	"context"
	"fmt"
	"log/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.NewConfig()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		logger.Error("failed to listen", "error", err, "port", cfg.Port)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Create oracle service
	oracleSvc := service.NewOracleService(cfg, logger)

	// Start heartbeat
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	oracleSvc.StartHeartbeat(ctx)

	_ = oracleSvc // Will register with gRPC once proto is generated

	logger.Info("oracle service listening", "port", cfg.Port)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		logger.Info("received shutdown signal, stopping gracefully")
		grpcServer.GracefulStop()
		cancel()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}

	logger.Info("oracle service stopped")
}
