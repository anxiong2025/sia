# SIA 企业级图片生成服务 - 项目结构

## 项目概述

本项目已成功重构为标准的企业级Go项目，支持gRPC接口供外部调用。项目采用现代化的微服务架构，具备完整的配置管理、日志系统、错误处理和健康检查功能。

## 项目结构

```
sia/
├── api/                          # API定义和生成的代码
│   └── image/
│       └── v1/
│           ├── image_service.pb.go      # protobuf生成的消息类型
│           ├── image_service_types.pb.go # 扩展的消息类型
│           └── image_service_grpc.pb.go # gRPC服务接口
├── cmd/                          # 应用程序入口
│   └── server/
│       └── main.go              # 服务器主程序
├── internal/                     # 内部包（不对外暴露）
│   ├── config/
│   │   └── config.go           # 配置管理
│   ├── domain/                 # 业务领域层
│   │   ├── image_client.go     # 图片生成客户端
│   │   ├── task_manager.go     # 任务管理器
│   │   └── types.go           # 领域类型定义
│   ├── server/                 # 服务器实现
│   │   ├── grpc.go            # gRPC服务器
│   │   └── http.go            # HTTP服务器（健康检查）
│   └── service/                # 业务服务层
│       └── image_service.go    # 图片生成服务实现
├── pkg/                         # 可复用的公共包
│   └── logger/
│       └── logger.go           # 结构化日志
├── proto/                       # Protocol Buffers定义
│   └── image_service.proto     # 服务接口定义
├── examples/                    # 示例代码
│   └── client/
│       └── main.go             # gRPC客户端示例
├── scripts/                     # 脚本文件
│   └── test.sh                 # 测试脚本
├── bin/                         # 编译输出目录
│   └── sia-server              # 编译后的服务器可执行文件
├── .env                         # 环境变量配置（本地）
├── .env.example                 # 环境变量示例
├── .gitignore                   # Git忽略文件
├── docker-compose.yml           # Docker Compose配置
├── Dockerfile                   # Docker镜像构建
├── go.mod                       # Go模块定义
├── go.sum                       # Go模块校验和
├── Makefile                     # 构建脚本
├── PROJECT_STRUCTURE.md         # 项目结构说明（本文件）
└── README.md                    # 项目说明文档
```

## 🗑️ 已清理的旧文件

以下旧文件已被删除，因为功能已整合到新的企业级架构中：

- ~~`image_agent.go`~~ → 功能整合到 `internal/service/image_service.go`
- ~~`image_client.go`~~ → 功能整合到 `internal/domain/image_client.go`
- ~~`image_types.go`~~ → 功能整合到 `internal/domain/types.go`
- ~~`main.go`~~ → 替换为 `cmd/server/main.go`
- ~~`sia`~~ → 旧的可执行文件，现在使用 `bin/sia-server`

## 核心功能

### 1. gRPC服务接口

- **GenerateImage**: 同步生成图片
- **GenerateImageAsync**: 异步生成图片
- **GetImageTask**: 查询异步任务状态
- **GenerateSequentialImages**: 生成序列图片
- **HealthCheck**: 健康检查

### 2. 企业级特性

- **配置管理**: 支持环境变量和.env文件
- **结构化日志**: 使用slog进行JSON格式日志记录
- **健康检查**: HTTP和gRPC双重健康检查
- **优雅关闭**: 支持信号处理和优雅关闭
- **错误处理**: 统一的错误处理和状态码
- **任务管理**: 异步任务状态跟踪
- **容器化**: Docker和Docker Compose支持

### 3. 安全性

- **API密钥管理**: 安全的API密钥配置
- **输入验证**: 请求参数验证
- **超时控制**: 请求超时和重试机制
- **资源限制**: 并发控制和资源管理

## 快速开始

### 1. 环境配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件，设置必要的API密钥
vim .env
```

### 2. 构建和运行

```bash
# 构建项目
make build

# 运行服务器
./bin/sia-server

# 或者直接运行
go run cmd/server/main.go
```

### 3. 测试服务

```bash
# 运行测试脚本
./scripts/test.sh

# 或者运行客户端示例
go run examples/client/main.go
```

### 4. Docker部署

```bash
# 构建Docker镜像
make docker-build

# 使用Docker Compose启动
docker-compose up -d
```

## API使用示例

### gRPC客户端调用

```go
// 连接到服务器
conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := imagev1.NewImageServiceClient(conn)

// 生成图片
resp, err := client.GenerateImage(context.Background(), &imagev1.GenerateImageRequest{
    Prompt: "一只可爱的小猫",
    Size:   "2K",
    Watermark: true,
})
```

### HTTP健康检查

```bash
# 健康检查
curl http://localhost:9090/health

# 就绪检查
curl http://localhost:9090/ready
```

## 配置说明

### 环境变量

- `APP_NAME`: 应用名称
- `APP_VERSION`: 应用版本
- `APP_ENVIRONMENT`: 运行环境（development/production）
- `GRPC_PORT`: gRPC服务端口（默认8080）
- `HTTP_PORT`: HTTP服务端口（默认9090）
- `IMAGE_API_KEY`: 图片生成API密钥（必需）
- `IMAGE_BASE_URL`: 图片生成API地址
- `IMAGE_MODEL`: 默认图片生成模型
- `LOG_LEVEL`: 日志级别（debug/info/warn/error）
- `LOG_FORMAT`: 日志格式（json/text）

## 监控和运维

### 健康检查端点

- `GET /health`: 服务健康状态
- `GET /ready`: 服务就绪状态
- `GET /metrics`: 服务指标（可扩展）

### 日志格式

项目使用结构化JSON日志，便于日志收集和分析：

```json
{
  "time": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "Request processed",
  "request_id": "req-123",
  "duration": "100ms"
}
```

## 扩展指南

### 添加新的gRPC方法

1. 在`proto/image_service.proto`中定义新方法
2. 重新生成protobuf代码：`make proto`
3. 在`internal/service/image_service.go`中实现方法
4. 更新客户端示例和测试

### 添加新的配置项

1. 在`internal/config/config.go`中添加配置结构
2. 在`Load()`函数中添加环境变量读取
3. 在`validate()`函数中添加验证逻辑
4. 更新`.env.example`文件

### 集成新的存储后端

1. 在`internal/domain/`中定义存储接口
2. 实现具体的存储适配器
3. 在配置中添加存储相关配置
4. 在服务中注入存储依赖

## 部署建议

### 生产环境

- 使用容器化部署（Docker/Kubernetes）
- 配置负载均衡器
- 启用TLS/SSL加密
- 配置监控和告警
- 设置日志收集和分析
- 配置备份和恢复策略

### 性能优化

- 启用gRPC连接池
- 配置适当的超时时间
- 使用缓存减少API调用
- 监控资源使用情况
- 优化并发处理能力

## 故障排除

### 常见问题

1. **端口冲突**: 检查端口是否被占用
2. **API密钥错误**: 验证环境变量配置
3. **网络连接问题**: 检查防火墙和网络配置
4. **内存不足**: 调整容器资源限制
5. **日志级别**: 调整日志级别获取更多信息

### 调试模式

```bash
# 启用调试日志
export LOG_LEVEL=debug

# 运行服务器
go run cmd/server/main.go
```

## 贡献指南

1. Fork项目仓库
2. 创建功能分支
3. 提交代码变更
4. 运行测试确保通过
5. 提交Pull Request

## 许可证

本项目采用MIT许可证，详见LICENSE文件。