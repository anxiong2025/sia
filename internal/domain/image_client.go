package domain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ImageClient 图片生成API客户端
type ImageClient struct {
	config     *ImageClientConfig
	httpClient *http.Client
}

// ImageClientConfig 图片客户端配置
type ImageClientConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	DefaultSize string
	Timeout     int
	MaxRetries  int
}

// NewImageClient 创建新的图片生成客户端
func NewImageClient(config *ImageClientConfig) *ImageClient {
	return &ImageClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// GenerateImage 生成图片
func (c *ImageClient) GenerateImage(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error) {
	// 设置默认值
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.ResponseFormat == "" {
		req.ResponseFormat = "url"
	}
	if req.Size == "" {
		req.Size = c.config.DefaultSize
	}
	if req.SequentialImageGeneration == "" {
		req.SequentialImageGeneration = "auto"
	}
	if req.SequentialImageGenerationOptions == nil {
		req.SequentialImageGenerationOptions = &SequentialImageGenerationOptions{
			MaxImages: 3,
		}
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/v3/images/generations", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析SSE流式响应
	imageResp, err := c.parseSSEResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSE response: %w", err)
	}

	return imageResp, nil
}

// parseSSEResponse 解析SSE流式响应
func (c *ImageClient) parseSSEResponse(body io.ReadCloser) (*ImageGenerationResponse, error) {
	scanner := bufio.NewScanner(body)
	var imageResp ImageGenerationResponse
	var images []ImageData

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和非data行
		if line == "" || line == "data: [DONE]" {
			continue
		}

		// 处理data行
		if strings.HasPrefix(line, "data: ") {
			jsonData := strings.TrimPrefix(line, "data: ")

			// 解析JSON数据
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &eventData); err != nil {
				continue // 跳过无法解析的行
			}

			// 处理不同类型的事件
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			switch eventType {
			case "image_generation.partial_succeeded":
				// 提取图片信息
				if url, ok := eventData["url"].(string); ok {
					imageData := ImageData{
						URL: url,
					}
					if revisedPrompt, ok := eventData["revised_prompt"].(string); ok {
						imageData.RevisedPrompt = revisedPrompt
					}
					images = append(images, imageData)
				}

				// 设置基本响应信息
				if imageResp.ID == "" {
					if id, ok := eventData["id"].(string); ok {
						imageResp.ID = id
					}
					if model, ok := eventData["model"].(string); ok {
						imageResp.Model = model
					}
					if created, ok := eventData["created"].(float64); ok {
						imageResp.Created = int64(created)
					}
					imageResp.Object = "list"
				}

			case "image_generation.completed":
				// 提取使用统计信息
				if usageData, ok := eventData["usage"].(map[string]interface{}); ok {
					var usage Usage
					if generatedImages, ok := usageData["generated_images"].(float64); ok {
						usage.PromptTokens = int(generatedImages)
					}
					if outputTokens, ok := usageData["output_tokens"].(float64); ok {
						usage.CompletionTokens = int(outputTokens)
					}
					if totalTokens, ok := usageData["total_tokens"].(float64); ok {
						usage.TotalTokens = int(totalTokens)
					}
					imageResp.Usage = usage
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SSE stream: %w", err)
	}

	// 设置图片数据
	imageResp.Data = images

	// 如果没有获取到图片，返回错误
	if len(images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return &imageResp, nil
}
