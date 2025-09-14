#!/bin/bash

# 测试脚本
set -e

echo "=== SIA Image Service 测试脚本 ==="

# 检查依赖
echo "检查依赖..."
if ! command -v grpcurl &> /dev/null; then
    echo "警告: grpcurl 未安装，跳过 gRPC 测试"
    SKIP_GRPC_TEST=true
fi

# 构建项目
echo "构建项目..."
make build

# 启动服务器（后台运行）
echo "启动服务器..."
./bin/sia-server &
SERVER_PID=$!

# 等待服务器启动
echo "等待服务器启动..."
sleep 5

# 测试HTTP健康检查
echo "测试HTTP健康检查..."
if curl -f http://localhost:9090/health > /dev/null 2>&1; then
    echo "✓ HTTP健康检查通过"
else
    echo "✗ HTTP健康检查失败"
fi

# 测试HTTP就绪检查
echo "测试HTTP就绪检查..."
if curl -f http://localhost:9090/ready > /dev/null 2>&1; then
    echo "✓ HTTP就绪检查通过"
else
    echo "✗ HTTP就绪检查失败"
fi

# 测试gRPC接口（如果grpcurl可用）
if [ "$SKIP_GRPC_TEST" != "true" ]; then
    echo "测试gRPC健康检查..."
    if grpcurl -plaintext localhost:8080 image.v1.ImageService/HealthCheck > /dev/null 2>&1; then
        echo "✓ gRPC健康检查通过"
    else
        echo "✗ gRPC健康检查失败"
    fi

    echo "测试gRPC图片生成..."
    if grpcurl -plaintext -d '{
        "prompt": "测试图片生成",
        "size": "2K",
        "watermark": true
    }' localhost:8080 image.v1.ImageService/GenerateImage > /dev/null 2>&1; then
        echo "✓ gRPC图片生成测试通过"
    else
        echo "✗ gRPC图片生成测试失败"
    fi
fi

# 运行Go客户端测试
echo "运行Go客户端测试..."
if go run examples/client/main.go > /dev/null 2>&1; then
    echo "✓ Go客户端测试通过"
else
    echo "✗ Go客户端测试失败"
fi

# 清理
echo "清理..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo "=== 测试完成 ==="