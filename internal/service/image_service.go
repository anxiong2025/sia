package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	imagev1 "sia/api/image/v1"
	"sia/internal/config"
	"sia/internal/domain"
	"sia/pkg/logger"
)

// ImageService 图片生成服务
type ImageService struct {
	imagev1.UnimplementedImageServiceServer
	config      *config.Config
	logger      *logger.Logger
	imageClient *domain.ImageClient
	taskManager *domain.TaskManager
}

// NewImageService 创建新的图片生成服务
func NewImageService(cfg *config.Config, logger *logger.Logger) *ImageService {
	imageClient := domain.NewImageClient(&domain.ImageClientConfig{
		APIKey:      cfg.Image.APIKey,
		BaseURL:     cfg.Image.BaseURL,
		Model:       cfg.Image.Model,
		DefaultSize: cfg.Image.DefaultSize,
		Timeout:     cfg.Image.Timeout,
		MaxRetries:  cfg.Image.MaxRetries,
	})

	taskManager := domain.NewTaskManager()

	return &ImageService{
		config:      cfg,
		logger:      logger,
		imageClient: imageClient,
		taskManager: taskManager,
	}
}

// GenerateImage 生成图片
func (s *ImageService) GenerateImage(ctx context.Context, req *imagev1.GenerateImageRequest) (*imagev1.GenerateImageResponse, error) {
	s.logger.Info("Generating image", "prompt", req.Prompt)

	// 验证请求
	if err := s.validateGenerateImageRequest(req); err != nil {
		s.logger.Error("Invalid request", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 创建域对象请求
	domainReq := &domain.ImageGenerationRequest{
		Model:          s.getModel(req.Model),
		Prompt:         req.Prompt,
		Image:          req.ImageUrls,
		ResponseFormat: "url",
		Size:           s.getSize(req.Size),
		Stream:         true,
		Watermark:      req.Watermark,
	}

	// 调用图片生成
	response, err := s.imageClient.GenerateImage(ctx, domainReq)
	if err != nil {
		s.logger.Error("Failed to generate image", "error", err)
		return nil, status.Error(codes.Internal, "Failed to generate image")
	}

	// 转换响应
	grpcResponse := s.convertToGRPCResponse(response)
	s.logger.Info("Image generated successfully", "image_count", len(grpcResponse.Images))

	return grpcResponse, nil
}

// GenerateImageAsync 异步生成图片
func (s *ImageService) GenerateImageAsync(ctx context.Context, req *imagev1.GenerateImageRequest) (*imagev1.GenerateImageAsyncResponse, error) {
	s.logger.Info("Starting async image generation", "prompt", req.Prompt)

	// 验证请求
	if err := s.validateGenerateImageRequest(req); err != nil {
		s.logger.Error("Invalid request", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 创建任务
	task := s.taskManager.CreateTask(req.Prompt)

	// 异步执行
	go func() {
		taskCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Image.Timeout)*time.Second)
		defer cancel()

		// 创建域对象请求
		domainReq := &domain.ImageGenerationRequest{
			Model:          s.getModel(req.Model),
			Prompt:         req.Prompt,
			Image:          req.ImageUrls,
			ResponseFormat: "url",
			Size:           s.getSize(req.Size),
			Stream:         true,
			Watermark:      req.Watermark,
		}

		// 更新任务状态为处理中
		s.taskManager.UpdateTaskStatus(task.ID, domain.TaskStatusProcessing)

		// 执行图片生成
		response, err := s.imageClient.GenerateImage(taskCtx, domainReq)
		if err != nil {
			s.logger.Error("Async image generation failed", "task_id", task.ID, "error", err)
			s.taskManager.UpdateTaskError(task.ID, err.Error())
		} else {
			s.logger.Info("Async image generation completed", "task_id", task.ID, "image_count", len(response.Data))
			s.taskManager.UpdateTaskResult(task.ID, response)
		}
	}()

	return &imagev1.GenerateImageAsyncResponse{
		TaskId:    task.ID,
		Status:    imagev1.TaskStatus_TASK_STATUS_PENDING,
		CreatedAt: timestamppb.New(task.CreatedAt),
	}, nil
}

// GetImageTask 获取图片生成任务状态
func (s *ImageService) GetImageTask(ctx context.Context, req *imagev1.GetImageTaskRequest) (*imagev1.GetImageTaskResponse, error) {
	s.logger.Debug("Getting task status", "task_id", req.TaskId)

	task, exists := s.taskManager.GetTask(req.TaskId)
	if !exists {
		return nil, status.Error(codes.NotFound, "Task not found")
	}

	response := &imagev1.GetImageTaskResponse{
		TaskId:    task.ID,
		Status:    s.convertTaskStatus(task.Status),
		CreatedAt: timestamppb.New(task.CreatedAt),
		UpdatedAt: timestamppb.New(task.UpdatedAt),
	}

	if task.Status == domain.TaskStatusCompleted && task.Result != nil {
		response.Result = s.convertToGRPCResponse(task.Result)
	}

	if task.Status == domain.TaskStatusFailed && task.Error != "" {
		response.ErrorMessage = task.Error
	}

	return response, nil
}

// GenerateSequentialImages 生成序列图片
func (s *ImageService) GenerateSequentialImages(ctx context.Context, req *imagev1.GenerateSequentialImagesRequest) (*imagev1.GenerateImageResponse, error) {
	s.logger.Info("Generating sequential images", "prompt", req.Prompt, "max_images", req.MaxImages)

	// 验证请求
	if err := s.validateSequentialImagesRequest(req); err != nil {
		s.logger.Error("Invalid request", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 创建域对象请求
	domainReq := &domain.ImageGenerationRequest{
		Model:                     s.getModel(req.Model),
		Prompt:                    req.Prompt,
		SequentialImageGeneration: "auto",
		SequentialImageGenerationOptions: &domain.SequentialImageGenerationOptions{
			MaxImages: int(req.MaxImages),
		},
		ResponseFormat: "url",
		Size:           s.getSize(req.Size),
		Stream:         true,
		Watermark:      req.Watermark,
	}

	// 调用图片生成
	response, err := s.imageClient.GenerateImage(ctx, domainReq)
	if err != nil {
		s.logger.Error("Failed to generate sequential images", "error", err)
		return nil, status.Error(codes.Internal, "Failed to generate sequential images")
	}

	// 转换响应
	grpcResponse := s.convertToGRPCResponse(response)
	s.logger.Info("Sequential images generated successfully", "image_count", len(grpcResponse.Images))

	return grpcResponse, nil
}

// HealthCheck 健康检查
func (s *ImageService) HealthCheck(ctx context.Context, req *imagev1.HealthCheckRequest) (*imagev1.HealthCheckResponse, error) {
	// 检查服务状态
	details := make(map[string]string)
	details["service"] = "image-service"
	details["version"] = s.config.App.Version
	details["environment"] = s.config.App.Environment

	// 可以添加更多健康检查逻辑，比如检查外部API连接等

	return &imagev1.HealthCheckResponse{
		Status:  imagev1.HealthStatus_HEALTH_STATUS_SERVING,
		Message: "Service is healthy",
		Details: details,
	}, nil
}

// validateGenerateImageRequest 验证生成图片请求
func (s *ImageService) validateGenerateImageRequest(req *imagev1.GenerateImageRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(req.Prompt) > 1000 {
		return fmt.Errorf("prompt too long, maximum 1000 characters")
	}

	return nil
}

// validateSequentialImagesRequest 验证序列图片请求
func (s *ImageService) validateSequentialImagesRequest(req *imagev1.GenerateSequentialImagesRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(req.Prompt) > 1000 {
		return fmt.Errorf("prompt too long, maximum 1000 characters")
	}

	if req.MaxImages <= 0 || req.MaxImages > 10 {
		return fmt.Errorf("max_images must be between 1 and 10")
	}

	return nil
}

// getModel 获取模型名称
func (s *ImageService) getModel(model string) string {
	if model == "" {
		return s.config.Image.Model
	}
	return model
}

// getSize 获取图片尺寸
func (s *ImageService) getSize(size string) string {
	if size == "" {
		return s.config.Image.DefaultSize
	}
	return size
}

// convertToGRPCResponse 转换为gRPC响应
func (s *ImageService) convertToGRPCResponse(response *domain.ImageGenerationResponse) *imagev1.GenerateImageResponse {
	images := make([]*imagev1.ImageData, len(response.Data))
	for i, img := range response.Data {
		images[i] = &imagev1.ImageData{
			Url:           img.URL,
			B64Json:       img.B64JSON,
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	return &imagev1.GenerateImageResponse{
		RequestId: response.ID,
		Images:    images,
		Usage: &imagev1.Usage{
			PromptTokens:     int32(response.Usage.PromptTokens),
			CompletionTokens: int32(response.Usage.CompletionTokens),
			TotalTokens:      int32(response.Usage.TotalTokens),
		},
		Model:     response.Model,
		CreatedAt: timestamppb.New(time.Unix(response.Created, 0)),
	}
}

// convertTaskStatus 转换任务状态
func (s *ImageService) convertTaskStatus(status domain.TaskStatus) imagev1.TaskStatus {
	switch status {
	case domain.TaskStatusPending:
		return imagev1.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusProcessing:
		return imagev1.TaskStatus_TASK_STATUS_PROCESSING
	case domain.TaskStatusCompleted:
		return imagev1.TaskStatus_TASK_STATUS_COMPLETED
	case domain.TaskStatusFailed:
		return imagev1.TaskStatus_TASK_STATUS_FAILED
	default:
		return imagev1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}
