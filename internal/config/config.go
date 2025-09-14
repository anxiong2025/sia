package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	App    AppConfig    `json:"app"`
	Server ServerConfig `json:"server"`
	Image  ImageConfig  `json:"image"`
	Log    LogConfig    `json:"log"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	GRPCPort int `json:"grpc_port"`
	HTTPPort int `json:"http_port"`
}

// ImageConfig 图片生成配置
type ImageConfig struct {
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url"`
	Model       string `json:"model"`
	DefaultSize string `json:"default_size"`
	Timeout     int    `json:"timeout"`
	MaxRetries  int    `json:"max_retries"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // json, text
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件（如果存在）
	loadEnvFile(".env")

	config := &Config{
		App: AppConfig{
			Name:        getEnvString("APP_NAME", "sia-image-service"),
			Version:     getEnvString("APP_VERSION", "1.0.0"),
			Environment: getEnvString("APP_ENVIRONMENT", "development"),
		},
		Server: ServerConfig{
			GRPCPort: getEnvInt("GRPC_PORT", 8080),
			HTTPPort: getEnvInt("HTTP_PORT", 9090),
		},
		Image: ImageConfig{
			APIKey:      getEnvString("IMAGE_API_KEY", ""),
			BaseURL:     getEnvString("IMAGE_BASE_URL", "https://ark.cn-beijing.volces.com"),
			Model:       getEnvString("IMAGE_MODEL", "doubao-seedream-4-0-250828"),
			DefaultSize: getEnvString("IMAGE_DEFAULT_SIZE", "2K"),
			Timeout:     getEnvInt("IMAGE_TIMEOUT", 300),
			MaxRetries:  getEnvInt("IMAGE_MAX_RETRIES", 3),
		},
		Log: LogConfig{
			Level:  getEnvString("LOG_LEVEL", "info"),
			Format: getEnvString("LOG_FORMAT", "json"),
		},
	}

	// 验证必需的配置
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate 验证配置
func (c *Config) validate() error {
	if c.Image.APIKey == "" {
		return fmt.Errorf("IMAGE_API_KEY is required")
	}

	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC_PORT: %d", c.Server.GRPCPort)
	}

	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP_PORT: %d", c.Server.HTTPPort)
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.Log.Level) {
		return fmt.Errorf("invalid LOG_LEVEL: %s, must be one of %v", c.Log.Level, validLogLevels)
	}

	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, c.Log.Format) {
		return fmt.Errorf("invalid LOG_FORMAT: %s, must be one of %v", c.Log.Format, validLogFormats)
	}

	return nil
}

// loadEnvFile 加载环境变量文件
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // 文件不存在时忽略
	}
	defer file.Close()

	// 简单的 .env 文件解析
	content, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 只有当环境变量不存在时才设置
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// getEnvString 获取字符串环境变量
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
