package domain

import "time"

// ImageGenerationRequest 图片生成请求
type ImageGenerationRequest struct {
	Model                            string                            `json:"model"`
	Prompt                           string                            `json:"prompt"`
	Image                            []string                          `json:"image,omitempty"`
	SequentialImageGeneration        string                            `json:"sequential_image_generation,omitempty"`
	SequentialImageGenerationOptions *SequentialImageGenerationOptions `json:"sequential_image_generation_options,omitempty"`
	ResponseFormat                   string                            `json:"response_format"`
	Size                             string                            `json:"size"`
	Stream                           bool                              `json:"stream"`
	Watermark                        bool                              `json:"watermark"`
}

// SequentialImageGenerationOptions 序列图片生成选项
type SequentialImageGenerationOptions struct {
	MaxImages int `json:"max_images"`
}

// ImageGenerationResponse 图片生成响应
type ImageGenerationResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Data    []ImageData `json:"data"`
	Usage   Usage       `json:"usage"`
}

// ImageData 图片数据
type ImageData struct {
	URL           string `json:"url"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusProcessing
	TaskStatusCompleted
	TaskStatusFailed
)

// Task 任务
type Task struct {
	ID        string                   `json:"id"`
	Status    TaskStatus               `json:"status"`
	Prompt    string                   `json:"prompt"`
	Result    *ImageGenerationResponse `json:"result,omitempty"`
	Error     string                   `json:"error,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}
