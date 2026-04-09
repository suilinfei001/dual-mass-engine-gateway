// AI Analyzer Service - AI 日志分析微服务
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quality-gateway/ai-analyzer/internal/analyzer"
	"github.com/quality-gateway/ai-analyzer/internal/api"
	"github.com/quality-gateway/ai-analyzer/internal/client"
	"github.com/quality-gateway/ai-analyzer/internal/pool"
	"github.com/quality-gateway/ai-analyzer/internal/types"
	"github.com/quality-gateway/shared/pkg/logger"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

var (
	port           string
	aiIP           string
	aiModel        string
	aiToken        string
	poolSize       int
	logLevel       string
)

func init() {
	flag.StringVar(&port, "port", "4005", "Server port")
	flag.StringVar(&aiIP, "ai-ip", "", "AI server IP address")
	flag.StringVar(&aiModel, "ai-model", "", "AI model name")
	flag.StringVar(&aiToken, "ai-token", "", "AI server token")
	flag.IntVar(&poolSize, "pool-size", 50, "AI request pool size")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

func parseLogLevel(level string) logger.Level {
	switch level {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}

// MockConfigProvider is a mock implementation of AIConfigProvider
// In production, this would be replaced by a database-backed implementation
type MockConfigProvider struct {
	config *types.AIConfig
}

func (m *MockConfigProvider) GetAIConfig() (*types.AIConfig, error) {
	return m.config, nil
}

func main() {
	flag.Parse()

	// 从环境变量覆盖配置
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	if v := os.Getenv("AI_IP"); v != "" {
		aiIP = v
	}
	if v := os.Getenv("AI_MODEL"); v != "" {
		aiModel = v
	}
	if v := os.Getenv("AI_TOKEN"); v != "" {
		aiToken = v
	}
	if v := os.Getenv("POOL_SIZE"); v != "" {
		fmt.Sscanf(v, "%d", &poolSize)
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 创建日志器
	log := logger.New(logger.Config{
		Level: parseLogLevel(logLevel),
	})

	log.Info("Starting AI Analyzer Service",
		logger.String("port", port),
		logger.Int("pool_size", poolSize),
	)

	// 初始化 AI 配置
	config := &types.AIConfig{
		IP:    aiIP,
		Model: aiModel,
		Token: aiToken,
	}

	// 检查是否配置了 AI
	if !config.IsConfigured() {
		log.Warn("AI is not configured. Set AI_IP, AI_MODEL, and AI_TOKEN environment variables.",
			logger.String("ai_ip_set", fmt.Sprintf("%v", aiIP != "")),
			logger.String("ai_model_set", fmt.Sprintf("%v", aiModel != "")),
			logger.String("ai_token_set", fmt.Sprintf("%v", aiToken != "")),
		)
		log.Info("Service will start but analysis requests will fail until configured.")
	} else {
		log.Info("AI configured",
			logger.String("ip", config.IP),
			logger.String("model", config.Model),
		)
	}

	// 初始化组件
	configProvider := &MockConfigProvider{config: config}
	aiClient := client.NewAIClient(configProvider)
	requestPool := pool.NewAIRequestPool(poolSize)
	logAnalyzer := analyzer.NewLogAnalyzer(aiClient, requestPool)

	// 创建服务器配置
	serverPort := 4005
	fmt.Sscanf(port, "%d", &serverPort)

	serverConfig := sharedapi.Config{
		Host:            "0.0.0.0",
		Port:            serverPort,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    10 * time.Minute, // 批量分析可能需要较长时间
		ShutdownTimeout: 10 * time.Second,
	}

	// 创建服务器
	srv := api.NewServer(serverConfig, logAnalyzer, requestPool, log)

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server", logger.Err(err))
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Shutdown error", logger.Err(err))
	}

	log.Info("Server stopped")
}
