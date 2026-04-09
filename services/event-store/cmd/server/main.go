// Event Store Service - 事件存储微服务
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quality-gateway/event-store/internal/api"
	"github.com/quality-gateway/event-store/internal/service"
	"github.com/quality-gateway/event-store/internal/storage"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
	"github.com/quality-gateway/shared/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

var (
	port      string
	dbHost    string
	dbPort    int
	dbUser    string
	dbPass    string
	dbName    string
	logLevel  string
)

func init() {
	flag.StringVar(&port, "port", "4002", "Server port")
	flag.StringVar(&dbHost, "db-host", "localhost", "Database host")
	flag.IntVar(&dbPort, "db-port", 3306, "Database port")
	flag.StringVar(&dbUser, "db-user", "root", "Database user")
	flag.StringVar(&dbPass, "db-pass", "", "Database password")
	flag.StringVar(&dbName, "db-name", "event_store", "Database name")
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
	if v := os.Getenv("DB_HOST"); v != "" {
		dbHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &dbPort)
	}
	if v := os.Getenv("DB_USER"); v != "" {
		dbUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		dbPass = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		dbName = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 创建日志器
	log := logger.New(logger.Config{
		Level: parseLogLevel(logLevel),
	})

	log.Info("Starting Event Store Service",
		logger.String("port", port),
		logger.String("db_host", dbHost),
		logger.String("db_name", dbName),
	)

	// 创建数据库配置
	storageConfig := storage.Config{
		Driver:   "mysql",
		Host:     dbHost,
		Port:     dbPort,
		Database: dbName,
		Username: dbUser,
		Password: dbPass,
	}

	// 创建数据库连接
	db, err := storage.Open(storageConfig, log)
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Err(err))
	}
	defer db.Close()

	// 创建存储层
	eventStorage := storage.NewEventStorage(db)
	qualityCheckStorage := storage.NewQualityCheckStorage(db)

	// 创建服务
	eventStoreService := service.NewEventStoreService(
		eventStorage,
		qualityCheckStorage,
		log,
	)

	// 创建服务器配置
	serverPort := 4002
	fmt.Sscanf(port, "%d", &serverPort)

	serverConfig := sharedapi.Config{
		Host:            "0.0.0.0",
		Port:            serverPort,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	// 创建服务器
	srv := api.NewServer(serverConfig, eventStoreService, log)

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
