# SIA Image Service

一个基于gRPC的企业级图片生成服务，支持文本生成图片、图片到图片转换和序列图片生成。

## 功能特性

- 🚀 **高性能gRPC服务**：基于gRPC协议，支持高并发请求
- 🎨 **多种图片生成模式**：支持文本生成图片、图片到图片、序列图片生成
- ⚡ **异步处理**：支持异步图片生成，提高系统吞吐量
- 🔧 **企业级架构**：标准的项目结构，易于维护和扩展
- 📊 **监控和健康检查**：内置健康检查和指标监控
- 🐳 **容器化部署**：支持Docker和Docker Compose部署
- 📝 **结构化日志**：JSON格式日志，便于日志分析
- 🛡️ **优雅关闭**：支持优雅关闭，确保请求完整处理

## 项目结构

```
sia/
├── api/                    # 生成的protobuf代码
│   └── image/v1/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── domain/            # 业务领域
│   ├── server/            # 服务器实现
│   └── service/           # gRPC服务实现
├── pkg/                   # 公共包
│   └── logger/            # 日志包
├── proto/                 # protobuf定义
│   └── image_service.proto
├── monitoring/            # 监控配置
├── docs/                  # 文档
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21+
- Protocol Buffers编译器 (protoc)
- Docker (可选)

### 安装开发工具

```bash
make install-tools
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置你的API密钥
```

### 生成protobuf代码

```bash
make proto
```

### 构建和运行

```bash
# 开发模式运行
make dev

# 或者构建后运行
make build
make run
```

### 使用Docker运行

```bash
# 构建并运行
make docker-run

# 或者使用docker-compose
docker-compose up -d
```

## API接口

### gRPC服务

服务运行在端口8080，提供以下接口：

#### 1. 生成图片
```protobuf
rpc GenerateImage(GenerateImageRequest) returns (GenerateImageResponse);
```

#### 2. 异步生成图片
```protobuf
rpc GenerateImageAsync(GenerateImageRequest) returns (GenerateImageAsyncResponse);
```

#### 3. 获取任务状态
```protobuf
rpc GetImageTask(GetImageTaskRequest) returns (GetImageTaskResponse);
```

#### 4. 生成序列图片
```protobuf
rpc GenerateSequentialImages(GenerateSequentialImagesRequest) returns (GenerateImageResponse);
```

#### 5. 健康检查
```protobuf
rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
```

### HTTP端点

服务在端口9090提供HTTP端点：

- `GET /health` - 健康检查
- `GET /ready` - 就绪检查
- `GET /metrics` - 指标监控

## 使用示例

### 使用grpcurl测试

```bash
# 健康检查
grpcurl -plaintext localhost:8080 image.v1.ImageService/HealthCheck

# 生成图片
grpcurl -plaintext -d '{
  "prompt": "一只可爱的小猫在花园里玩耍",
  "model": "doubao-seedream-4-0-250828",
  "size": "2K",
  "watermark": true
}' localhost:8080 image.v1.ImageService/GenerateImage

# 异步生成图片
grpcurl -plaintext -d '{
  "prompt": "美丽的日落风景",
  "size": "2K"
}' localhost:8080 image.v1.ImageService/GenerateImageAsync

# 查询任务状态
grpcurl -plaintext -d '{
  "task_id": "task_1234567890"
}' localhost:8080 image.v1.ImageService/GetImageTask
```

### 使用Go客户端

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    imagev1 "sia/api/image/v1"
)

func main() {
    // 连接到服务器
    conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // 创建客户端
    client := imagev1.NewImageServiceClient(conn)
    
    // 生成图片
    resp, err := client.GenerateImage(context.Background(), &imagev1.GenerateImageRequest{
        Prompt: "一只可爱的小猫",
        Size:   "2K",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Generated %d images", len(resp.Images))
    for i, img := range resp.Images {
        log.Printf("Image %d: %s", i+1, img.Url)
    }
}
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `APP_NAME` | 应用名称 | `sia-image-service` |
| `APP_VERSION` | 应用版本 | `1.0.0` |
| `APP_ENVIRONMENT` | 运行环境 | `development` |
| `GRPC_PORT` | gRPC服务端口 | `8080` |
| `HTTP_PORT` | HTTP服务端口 | `9090` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LOG_FORMAT` | 日志格式 | `json` |
| `IMAGE_API_KEY` | 图片生成API密钥 | **必需** |
| `IMAGE_BASE_URL` | API基础URL | `https://ark.cn-beijing.volces.com` |
| `IMAGE_MODEL` | 默认模型 | `doubao-seedream-4-0-250828` |
| `IMAGE_DEFAULT_SIZE` | 默认图片尺寸 | `2K` |
| `IMAGE_TIMEOUT` | 请求超时时间(秒) | `300` |
| `IMAGE_MAX_RETRIES` | 最大重试次数 | `3` |

## 开发指南

### 添加新功能

1. 在`proto/image_service.proto`中定义新的RPC方法
2. 运行`make proto`生成代码
3. 在`internal/service/image_service.go`中实现方法
4. 添加相应的测试

### 代码规范

```bash
# 格式化代码
make fmt

# 运行linter
make lint

# 运行测试
make test

# 生成测试覆盖率报告
make test-coverage
```

## 部署

### Docker部署

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -p 8080:8080 -p 9090:9090 --env-file .env sia:latest
```

### Docker Compose部署

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f sia-server

# 停止服务
docker-compose down
```

### Kubernetes部署

参考`k8s/`目录中的Kubernetes配置文件。

## 监控

### 健康检查

- gRPC健康检查：使用grpc-health-probe
- HTTP健康检查：`GET /health`

### 指标监控

服务支持Prometheus指标收集，可以通过`/metrics`端点获取指标数据。

### 日志

服务使用结构化JSON日志，便于日志聚合和分析。

## 故障排除

### 常见问题

1. **连接失败**
   - 检查端口是否被占用
   - 确认防火墙设置
   - 验证网络连接

2. **API密钥错误**
   - 确认`IMAGE_API_KEY`环境变量设置正确
   - 检查API密钥是否有效

3. **图片生成失败**
   - 检查网络连接到图片生成API
   - 验证请求参数是否正确
   - 查看服务日志获取详细错误信息

### 日志分析

```bash
# 查看实时日志
docker-compose logs -f sia-server

# 过滤错误日志
docker-compose logs sia-server | grep '"level":"error"'
```

## 贡献

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请创建Issue或联系维护者。