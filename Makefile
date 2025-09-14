.PHONY: proto build run test clean docker-build docker-run

# 项目配置
PROJECT_NAME := sia
BINARY_NAME := sia-server
DOCKER_IMAGE := $(PROJECT_NAME):latest

# Go 配置
GO_VERSION := 1.21
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Proto 文件路径
PROTO_DIR := proto
API_DIR := api

# 生成 protobuf 代码
proto:
	@echo "Generating protobuf code..."
	@mkdir -p $(API_DIR)/image/v1
	@protoc --go_out=. --go_opt=module=sia \
		--go-grpc_out=. --go-grpc_opt=module=sia \
		$(PROTO_DIR)/image_service.proto
	@echo "Protobuf code generated successfully"

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# 构建项目
build: proto
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s" -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "Build completed: bin/$(BINARY_NAME)"

# 运行项目
run: build
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

# 运行开发模式
dev:
	@echo "Running in development mode..."
	@go run cmd/server/main.go

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	@golangci-lint run

# 清理构建文件
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -rf $(API_DIR)/
	@rm -f coverage.out coverage.html
	@echo "Clean completed"

# 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

# 运行 Docker 容器
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -p 9090:9090 --env-file .env $(DOCKER_IMAGE)

# 推送 Docker 镜像
docker-push: docker-build
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE)

# 生成 API 文档
docs:
	@echo "Generating API documentation..."
	@protoc --doc_out=./docs --doc_opt=html,index.html $(PROTO_DIR)/image_service.proto
	@echo "API documentation generated: docs/index.html"

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	@echo "Development tools installed"

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  proto          - Generate protobuf code"
	@echo "  deps           - Install dependencies"
	@echo "  build          - Build the project"
	@echo "  run            - Build and run the project"
	@echo "  dev            - Run in development mode"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build files"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-push    - Push Docker image"
	@echo "  docs           - Generate API documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help message"