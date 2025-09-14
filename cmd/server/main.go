package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"sia/internal/config"
	"sia/internal/server"
	"sia/internal/service"
	"sia/pkg/logger"
)

func main() {
	// 初始化日志
	logger := logger.New()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", "error", err)
	}

	logger.Info("Starting SIA Image Service",
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
		"grpc_port", cfg.Server.GRPCPort,
		"http_port", cfg.Server.HTTPPort,
	)

	// 创建服务
	imageService := service.NewImageService(cfg, logger)

	// 创建gRPC服务器
	grpcServer := server.NewGRPCServer(cfg, logger, imageService)

	// 创建HTTP服务器（用于健康检查和指标）
	httpServer := server.NewHTTPServer(cfg, logger)

	// 启动服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动gRPC服务器
	go func() {
		if err := startGRPCServer(grpcServer, cfg.Server.GRPCPort, logger); err != nil {
			logger.Error("gRPC server failed", "error", err)
			cancel()
		}
	}()

	// 启动HTTP服务器
	go func() {
		if err := startHTTPServer(httpServer, cfg.Server.HTTPPort, logger); err != nil {
			logger.Error("HTTP server failed", "error", err)
			cancel()
		}
	}()

	// 等待中断信号
	waitForShutdown(ctx, cancel, grpcServer, httpServer, logger)
}

func startGRPCServer(server *grpc.Server, port int, logger *logger.Logger) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	logger.Info("gRPC server starting", "port", port)

	// 启用反射（开发环境）
	reflection.Register(server)

	// 注册健康检查
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	return server.Serve(lis)
}

func startHTTPServer(server *http.Server, port int, logger *logger.Logger) error {
	server.Addr = fmt.Sprintf(":%d", port)
	logger.Info("HTTP server starting", "port", port)
	return server.ListenAndServe()
}

func waitForShutdown(ctx context.Context, cancel context.CancelFunc, grpcServer *grpc.Server, httpServer *http.Server, logger *logger.Logger) {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	// 优雅关闭
	logger.Info("Shutting down servers...")

	// 关闭HTTP服务器
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer httpCancel()

	if err := httpServer.Shutdown(httpCtx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	} else {
		logger.Info("HTTP server shutdown completed")
	}

	// 关闭gRPC服务器
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server shutdown completed")
	case <-time.After(30 * time.Second):
		logger.Warn("gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	logger.Info("Shutdown completed")
}
