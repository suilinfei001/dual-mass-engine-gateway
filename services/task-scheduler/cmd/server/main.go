// Task Scheduler Service - 任务调度微服务
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/quality-gateway/shared/pkg/logger"
	sharedstorage "github.com/quality-gateway/shared/pkg/storage"
	"github.com/quality-gateway/task-scheduler/internal/api"
	"github.com/quality-gateway/task-scheduler/internal/scheduler"
	taskstorage "github.com/quality-gateway/task-scheduler/internal/storage"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

var (
	port           string
	dbHost         string
	dbPort         int
	dbUser         string
	dbPass         string
	dbName         string
	eventStoreURL  string
	logLevel       string
	staleInterval  time.Duration
	staleAge       time.Duration
)

func init() {
	flag.StringVar(&port, "port", "4003", "Server port")
	flag.StringVar(&dbHost, "db-host", "localhost", "Database host")
	flag.IntVar(&dbPort, "db-port", 3306, "Database port")
	flag.StringVar(&dbUser, "db-user", "root", "Database user")
	flag.StringVar(&dbPass, "db-pass", "", "Database password")
	flag.StringVar(&dbName, "db-name", "task_scheduler", "Database name")
	flag.StringVar(&eventStoreURL, "event-store-url", "http://localhost:4002", "Event Store service URL")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.DurationVar(&staleInterval, "stale-interval", 10*time.Minute, "Interval between stale task checks")
	flag.DurationVar(&staleAge, "stale-age", 30*time.Minute, "Age after which a task is considered stale")
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
	if v := os.Getenv("DB_PASS"); v != "" {
		dbPass = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		dbName = v
	}
	if v := os.Getenv("EVENT_STORE_URL"); v != "" {
		eventStoreURL = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 创建日志器
	log := logger.New(logger.Config{
		Level: parseLogLevel(logLevel),
	})

	log.Info("Starting Task Scheduler Service",
		logger.String("port", port),
		logger.String("db_host", dbHost),
		logger.String("db_name", dbName),
		logger.String("event_store_url", eventStoreURL),
	)

	// 创建数据库配置
	storageConfig := sharedstorage.Config{
		Driver:          "mysql",
		DSN:             fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}

	// 创建数据库连接
	db, err := sharedstorage.Open(storageConfig, log)
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Err(err))
	}
	defer db.Close()

	// 创建存储层（使用内部 sql.DB）
	taskStorage := taskstorage.NewMySQLStorage(db.DB)

	// 初始化 Event Store 客户端
	eventStoreClient := scheduler.NewHTTPEventStoreClient(eventStoreURL)
	log.Info("Event Store client initialized",
		logger.String("url", eventStoreURL),
	)

	// 初始化调度器
	sched := scheduler.NewScheduler(taskStorage, eventStoreClient)

	// 启动后台任务：重置过期的分析任务
	go runStaleAnalysisReset(taskStorage, log, staleInterval, staleAge)

	// 创建服务器配置
	serverPort := 4003
	fmt.Sscanf(port, "%d", &serverPort)

	serverConfig := sharedapi.Config{
		Host:            "0.0.0.0",
		Port:            serverPort,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	// 创建服务器
	srv := api.NewServer(serverConfig, sched, log)

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

// runStaleAnalysisReset 定期重置过期的分析任务
func runStaleAnalysisReset(store taskstorage.TaskStorage, log *logger.Logger, interval, age time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		count, err := store.ResetStaleAnalysisTasks(age)
		if err != nil {
			log.Error("Failed to reset stale analysis tasks", logger.Err(err))
		} else if count > 0 {
			log.Info("Reset stale analysis tasks", logger.Int("count", count))
		}
	}
}
