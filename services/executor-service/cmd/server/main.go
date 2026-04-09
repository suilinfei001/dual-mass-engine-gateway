// Executor Service - Azure DevOps Pipeline 执行器微服务
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
	"github.com/quality-gateway/executor-service/internal/api"
	"github.com/quality-gateway/executor-service/internal/azure"
	"github.com/quality-gateway/executor-service/internal/service"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

var (
	port           string
	azureOrgURL    string
	azurePAT       string
	logLevel       string
	cleanupInterval time.Duration
	cleanupAge     time.Duration
)

func init() {
	flag.StringVar(&port, "port", "4004", "Server port")
	flag.StringVar(&azureOrgURL, "azure-org-url", "", "Azure DevOps organization URL")
	flag.StringVar(&azurePAT, "azure-pat", "", "Azure DevOps Personal Access Token")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.DurationVar(&cleanupInterval, "cleanup-interval", 1*time.Hour, "Execution cleanup interval")
	flag.DurationVar(&cleanupAge, "cleanup-age", 24*time.Hour, "Age after which executions are cleaned up")
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
	if v := os.Getenv("AZURE_ORG_URL"); v != "" {
		azureOrgURL = v
	}
	if v := os.Getenv("AZURE_PAT"); v != "" {
		azurePAT = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 创建日志器
	log := logger.New(logger.Config{
		Level: parseLogLevel(logLevel),
	})

	log.Info("Starting Executor Service",
		logger.String("port", port),
		logger.String("azure_org", azureOrgURL),
	)

	// Azure 配置为可选，如果没有配置则使用默认值
	if azureOrgURL == "" {
		azureOrgURL = os.Getenv("AZURE_ORG_URL")
	}
	if azurePAT == "" {
		azurePAT = os.Getenv("AZURE_PAT")
	}

	// 如果仍然没有配置，使用默认值并记录警告
	if azureOrgURL == "" {
		azureOrgURL = "https://dev.azure.com"
		log.Warn("Azure DevOps organization URL not configured, using default",
			logger.String("default", azureOrgURL))
	}
	if azurePAT == "" {
		log.Warn("Azure DevOps PAT not configured, API calls will require authentication",
			logger.String("advice", "configure AZURE_ORG_URL and AZURE_PAT environment variables"))
	}

	// 创建 Azure 认证配置
	azureAuth := &azure.AzureAuthConfig{
		OrganizationURL: azureOrgURL,
		PAT:             azurePAT,
	}

	// 创建 Azure 客户端
	azureClient := azure.NewAzureClient(azureAuth, log)

	// 创建服务
	executorService := service.NewExecutorService(azureClient, log)

	// 启动清理协程
	go runCleanupLoop(executorService, log, cleanupInterval, cleanupAge)

	// 创建服务器配置
	serverPort := 4004
	fmt.Sscanf(port, "%d", &serverPort)

	serverConfig := sharedapi.Config{
		Host:            "0.0.0.0",
		Port:            serverPort,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	// 创建服务器
	srv := api.NewServer(serverConfig, executorService, log)

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

// runCleanupLoop 运行清理循环
func runCleanupLoop(svc *service.ExecutorService, log *logger.Logger, interval, age time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		count := svc.CleanupOldExecutions(age)
		if count > 0 {
			log.Info("Cleanup completed", logger.Int("cleaned", count))
		}
	}
}
