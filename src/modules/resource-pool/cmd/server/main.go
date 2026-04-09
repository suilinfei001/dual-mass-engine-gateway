package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hugoh/go-designs/resource-pool/internal/api"
	_ "github.com/go-sql-driver/mysql"
)

var (
	port           string
	dbHost         string
	dbPort         string
	dbUser         string
	dbPass         string
	dbName         string
	deployerType   string
	tencentAPIKey  string
	tencentSecret  string
	processorAPI   string
)

func init() {
	flag.StringVar(&port, "port", "5003", "Server port")
	flag.StringVar(&dbHost, "db-host", "event-processor-mysql", "Database host")
	flag.StringVar(&dbPort, "db-port", "3306", "Database port")
	flag.StringVar(&dbUser, "db-user", "root", "Database user")
	flag.StringVar(&dbPass, "db-pass", "root123456", "Database password")
	flag.StringVar(&dbName, "db-name", "event_processor", "Database name")
	flag.StringVar(&deployerType, "deployer", "", "Deployer type (mock, ssh, or tencent)")
	flag.StringVar(&tencentAPIKey, "tencent-api-key", "", "Tencent API key")
	flag.StringVar(&tencentSecret, "tencent-secret", "", "Tencent secret key")
	flag.StringVar(&processorAPI, "processor-api", "http://event-processor-server:5002", "Event processor API")
}

func main() {
	flag.Parse()

	// 从环境变量读取 deployer 类型（如果命令行参数未设置）
	if deployerType == "" {
		deployerType = os.Getenv("DEPLOYER_TYPE")
		if deployerType == "" {
			deployerType = "mock" // 默认使用 mock
		}
	}

	log.Printf("Starting Resource Pool Server...")
	log.Printf("Configuration:")
	log.Printf("  Port: %s", port)
	log.Printf("  Database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)
	log.Printf("  Deployer: %s", deployerType)

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 部署器配置
	deployerConfig := make(map[string]string)
	if deployerType == "tencent" {
		if tencentAPIKey == "" {
			tencentAPIKey = os.Getenv("TENCENT_API_KEY")
		}
		if tencentSecret == "" {
			tencentSecret = os.Getenv("TENCENT_SECRET_KEY")
		}
		deployerConfig["tencent_api_key"] = tencentAPIKey
		deployerConfig["tencent_secret_key"] = tencentSecret
	}

	// 创建服务器（不传入 userStorage，因为资源池不管理用户）
	// 注意：需要从 event-processor 获取用户存储用于 session 验证
	// 这里简化处理，暂时传入 nil
	server, err := api.NewServer(port, dsn, deployerType, deployerConfig, nil)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
