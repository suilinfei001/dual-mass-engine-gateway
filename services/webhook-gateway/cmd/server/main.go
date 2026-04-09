// Webhook Gateway Service - Webhook 接收网关微服务
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/webhook-gateway/internal/api"
	"github.com/quality-gateway/webhook-gateway/internal/client"
	"github.com/quality-gateway/webhook-gateway/internal/service"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

var (
	port           string
	eventStoreURL  string
	apiToken       string
	githubSecret   string
	gitlabSecret   string
	logLevel       string
)

func init() {
	flag.StringVar(&port, "port", "4001", "Server port")
	flag.StringVar(&eventStoreURL, "event-store-url", "http://localhost:4002", "Event Store service URL")
	flag.StringVar(&apiToken, "api-token", "", "API token for Event Store authentication")
	flag.StringVar(&githubSecret, "github-secret", "", "GitHub webhook secret for signature verification")
	flag.StringVar(&gitlabSecret, "gitlab-secret", "", "GitLab webhook secret for signature verification")
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

func main() {
	flag.Parse()

	// 从环境变量覆盖配置
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	if v := os.Getenv("EVENT_STORE_URL"); v != "" {
		eventStoreURL = v
	}
	if v := os.Getenv("API_TOKEN"); v != "" {
		apiToken = v
	}
	if v := os.Getenv("GITHUB_SECRET"); v != "" {
		githubSecret = v
	}
	if v := os.Getenv("GITLAB_SECRET"); v != "" {
		gitlabSecret = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 创建日志器
	log := logger.New(logger.Config{
		Level: parseLogLevel(logLevel),
	})

	log.Info("Starting Webhook Gateway Service",
		logger.String("port", port),
		logger.String("event_store_url", eventStoreURL),
		logger.String("github_secret_set", fmt.Sprintf("%v", githubSecret != "")),
		logger.String("gitlab_secret_set", fmt.Sprintf("%v", gitlabSecret != "")),
	)

	// 创建 Webhook 配置
	config := client.DefaultWebhookConfig()
	config.EventStoreURL = eventStoreURL
	config.APIToken = apiToken
	if githubSecret != "" || gitlabSecret != "" {
		config.Verification = &client.WebhookVerificationConfig{
			GitHubSecret: githubSecret,
			GitLabSecret: gitlabSecret,
		}
	}

	// 创建 Event Store 客户端
	eventStoreClient := client.NewEventStoreClient(eventStoreURL, apiToken, log)

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventStoreClient.CheckHealth(ctx); err != nil {
		log.Warn("Event Store health check failed", logger.Err(err))
		log.Info("Continuing startup, will retry on first request...")
	}

	// 创建服务
	webhookService := service.NewWebhookGatewayService(eventStoreClient, config, log)

	// 创建服务器配置
	serverPort := 4001
	fmt.Sscanf(port, "%d", &serverPort)

	serverConfig := sharedapi.Config{
		Host:            "0.0.0.0",
		Port:            serverPort,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	// 创建服务器
	srv := api.NewServer(serverConfig, webhookService, log)

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server", logger.Err(err))
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx = context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Shutdown error", logger.Err(err))
	}

	log.Info("Server stopped")
}
