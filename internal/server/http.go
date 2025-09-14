package server

import (
	"encoding/json"
	"net/http"
	"time"

	"sia/internal/config"
	"sia/pkg/logger"
)

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(cfg *config.Config, logger *logger.Logger) *http.Server {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":      "healthy",
			"service":     cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// 就绪检查端点
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// 指标端点（可以集成Prometheus）
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// 这里可以集成Prometheus指标
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint - integrate with Prometheus\n"))
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("HTTP server configured successfully")

	return server
}
