package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	imagev1 "sia/api/image/v1"
)

func main() {
	// 连接到gRPC服务器
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := imagev1.NewImageServiceClient(conn)

	// 测试健康检查
	log.Println("=== 健康检查 ===")
	healthResp, err := client.HealthCheck(context.Background(), &imagev1.HealthCheckRequest{})
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		log.Printf("Health status: %v, Message: %s", healthResp.Status, healthResp.Message)
	}

	// 测试同步图片生成
	log.Println("\n=== 同步图片生成 ===")
	generateReq := &imagev1.GenerateImageRequest{
		Prompt:    "一只可爱的小猫在花园里玩耍",
		Model:     "doubao-seedream-4-0-250828",
		Size:      "2K",
		Watermark: true,
		Metadata: map[string]string{
			"user_id": "test_user",
			"source":  "example_client",
		},
	}

	generateResp, err := client.GenerateImage(context.Background(), generateReq)
	if err != nil {
		log.Printf("Generate image failed: %v", err)
	} else {
		log.Printf("Generated %d images:", len(generateResp.Images))
		for i, img := range generateResp.Images {
			log.Printf("  Image %d: %s", i+1, img.Url)
		}
	}

	// 测试异步图片生成
	log.Println("\n=== 异步图片生成 ===")
	asyncReq := &imagev1.GenerateImageRequest{
		Prompt:    "美丽的日落风景",
		Size:      "2K",
		Watermark: true,
	}

	asyncResp, err := client.GenerateImageAsync(context.Background(), asyncReq)
	if err != nil {
		log.Printf("Async generate image failed: %v", err)
	} else {
		log.Printf("Task created: %s, Status: %v", asyncResp.TaskId, asyncResp.Status)

		// 轮询任务状态
		taskId := asyncResp.TaskId
		for i := 0; i < 30; i++ { // 最多等待30秒
			time.Sleep(1 * time.Second)

			taskResp, err := client.GetImageTask(context.Background(), &imagev1.GetImageTaskRequest{
				TaskId: taskId,
			})
			if err != nil {
				log.Printf("Get task failed: %v", err)
				break
			}

			log.Printf("Task %s status: %v", taskId, taskResp.Status)

			if taskResp.Status == imagev1.TaskStatus_TASK_STATUS_COMPLETED {
				log.Printf("Task completed! Generated %d images:", len(taskResp.Result.Images))
				for i, img := range taskResp.Result.Images {
					log.Printf("  Image %d: %s", i+1, img.Url)
				}
				break
			} else if taskResp.Status == imagev1.TaskStatus_TASK_STATUS_FAILED {
				log.Printf("Task failed: %s", taskResp.ErrorMessage)
				break
			}
		}
	}

	// 测试序列图片生成
	log.Println("\n=== 序列图片生成 ===")
	sequentialReq := &imagev1.GenerateSequentialImagesRequest{
		Prompt:    "展示一天中不同时间的山景：早晨、中午、傍晚",
		MaxImages: 3,
		Size:      "2K",
		Watermark: true,
	}

	sequentialResp, err := client.GenerateSequentialImages(context.Background(), sequentialReq)
	if err != nil {
		log.Printf("Generate sequential images failed: %v", err)
	} else {
		log.Printf("Generated %d sequential images:", len(sequentialResp.Images))
		for i, img := range sequentialResp.Images {
			log.Printf("  Sequential Image %d: %s", i+1, img.Url)
		}
	}

	log.Println("\n=== 测试完成 ===")
}
