package server

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	imagev1 "sia/api/image/v1"
	"sia/internal/config"
	"sia/internal/service"
	"sia/pkg/logger"
)

// NewGRPCServer 创建gRPC服务器
func NewGRPCServer(cfg *config.Config, logger *logger.Logger, imageService *service.ImageService) *grpc.Server {
	// gRPC服务器选项
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// 创建gRPC服务器
	server := grpc.NewServer(opts...)

	// 注册服务
	imagev1.RegisterImageServiceServer(server, imageService)

	logger.Info("gRPC server configured successfully")

	return server
}
