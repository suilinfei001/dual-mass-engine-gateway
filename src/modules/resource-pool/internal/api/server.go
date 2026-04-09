package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/jobs"
	"github.com/hugoh/go-designs/resource-pool/internal/service"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// Server 资源池服务器
type Server struct {
	Port             string
	db               *storage.MySQLStorage
	service          service.ResourcePoolService
	jobManager       *jobs.JobManager
	internalHandler  *InternalAPIHandler
	externalHandler  *ExternalAPIHandler
	adminHandler     *AdminAPIHandler
	userStorage      UserStorage // event-processor 的用户存储
}

// UserStorage 用户存储接口（用于 session 验证）
type UserStorage interface {
	GetSessionWithUser(sessionID string) (map[string]interface{}, error)
}

// NewServer 创建资源池服务器
func NewServer(port string, dsn string, deployerType string, deployerConfig map[string]string, userStorage UserStorage) (*Server, error) {
	// 创建数据库连接
	db, err := storage.NewMySQLStorage(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// 如果没有传入 userStorage，创建默认的 MySQLUserStorage
	if userStorage == nil {
		userStorage, err = NewMySQLUserStorage(dsn)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create user storage: %w", err)
		}
	}

	// 创建存储层
	testbedStorage := storage.NewMySQLTestbedStorage(db.DB())
	allocationStorage := storage.NewMySQLAllocationStorage(db.DB())
	categoryStorage := storage.NewMySQLCategoryStorage(db.DB())
	quotaStorage := storage.NewMySQLQuotaPolicyStorage(db.DB())
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db.DB())
	taskStorage := storage.NewMySQLResourceInstanceTaskStorage(db.DB())
	configStorage := storage.NewMySQLConfigStorage(db.DB())
	deploymentTaskStorage := storage.NewMySQLDeploymentTaskStorage(db.DB())
	pipelineTemplateStorage := storage.NewMySQLDeploymentPipelineTemplateStorage(db.DB())

	// 创建 Deployer
	factory := deployer.NewDeployerFactory(deployerType)
	deployer, err := factory.CreateDeployer(deployerConfig)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create deployer: %w", err)
	}

	// 如果是 SSH Deployer，设置 Azure 配置
	if deployerType == "ssh" {
		// 使用接口类型断言来访问 SSHDeployService 的方法
		// 创建一个设置器接口
		type azureSetter interface {
			SetAzureConfig(config interface{})
			SetDefaultPipelineID(int)
			SetCMPCredentials(accessKey, secretKey string)
		}
		if azureSetter, ok := deployer.(azureSetter); ok {
			// 获取 Azure 配置
			storageAzureConfig, err := configStorage.GetAzureConfig()
			if err == nil && storageAzureConfig != nil {
				azureConfigInfo := map[string]interface{}{
					"Organization": storageAzureConfig.Organization,
					"Project":      storageAzureConfig.Project,
					"PAT":          storageAzureConfig.PAT,
					"BaseURL":      storageAzureConfig.BaseURL,
				}
				if storageAzureConfig.BaseURL == "" {
					azureConfigInfo["BaseURL"] = "https://devops.aishu.cn"
				}
				azureSetter.SetAzureConfig(azureConfigInfo)
				log.Printf("[Server] Azure config loaded for SSH deployer: %s/%s", storageAzureConfig.Organization, storageAzureConfig.Project)
			} else {
				log.Printf("[Server] Warning: Failed to get Azure config: %v", err)
			}

			// 从部署管道模板中获取默认 Pipeline ID
			templates, err := pipelineTemplateStorage.ListEnabledTemplates()
			defaultPipelineID := 0
			if err == nil && len(templates) > 0 {
				// 使用第一个启用的模板的 Pipeline ID
				defaultPipelineID = templates[0].PipelineID
				log.Printf("[Server] Using pipeline ID %d from template '%s'", defaultPipelineID, templates[0].Name)
			} else {
				// 如果没有模板，尝试从命令行配置获取
				if pipelineID, ok := deployerConfig["default_pipeline_id"]; ok {
					var pid int
					if _, err := fmt.Sscanf(pipelineID, "%d", &pid); err == nil {
						defaultPipelineID = pid
						log.Printf("[Server] Using pipeline ID %d from command line config", defaultPipelineID)
					}
				}
			}

			// 设置默认 Pipeline ID
			if defaultPipelineID > 0 {
				azureSetter.SetDefaultPipelineID(defaultPipelineID)
			}

			// 获取并设置 CMP 凭证
			cmpAccessKey, err := configStorage.GetConfig("cmp_access_key")
			if err != nil {
				log.Printf("[Server] Warning: Failed to get cmp_access_key: %v", err)
			}
			cmpSecretKey, err := configStorage.GetConfig("cmp_secret_key")
			if err != nil {
				log.Printf("[Server] Warning: Failed to get cmp_secret_key: %v", err)
			}
			if cmpAccessKey != "" && cmpSecretKey != "" {
				azureSetter.SetCMPCredentials(cmpAccessKey, cmpSecretKey)
				log.Printf("[Server] CMP credentials configured for SSH deployer")
			} else {
				log.Printf("[Server] Warning: CMP credentials not configured, snapshot restore will fail")
			}
		}
	}

	// 创建服务层
	poolService := service.NewResourcePoolService(
		testbedStorage,
		allocationStorage,
		categoryStorage,
		quotaStorage,
		resourceStorage,
		taskStorage,
		deployer,
		86400, // 默认 24 小时
	)

	// 创建后台任务（通过服务接口实现）
	jobManager := jobs.NewJobManager(
		allocationStorage,
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	// 创建 API 处理器
	internalHandler := NewInternalAPIHandler(poolService, resourceStorage)
	externalHandler := NewExternalAPIHandler(poolService, userStorage, resourceStorage, categoryStorage, testbedStorage, allocationStorage)
	adminHandler := NewAdminAPIHandler(poolService, testbedStorage, allocationStorage, resourceStorage, categoryStorage, quotaStorage)
	adminHandler.SetUserStorage(userStorage) // 设置用户存储用于认证
	adminHandler.SetDeployer(deployer)       // 设置部署服务用于健康检查
	adminHandler.SetTaskStorage(taskStorage) // 设置任务存储
	adminHandler.SetConfigStorage(configStorage) // 设置配置存储
	adminHandler.SetPipelineTemplateStorage(pipelineTemplateStorage) // 设置部署管道模板存储

	// 创建部署服务
	deploymentService := service.NewDeploymentService(deploymentTaskStorage, configStorage, allocationStorage)
	deploymentService.SetTestbedStorage(testbedStorage) // 设置 Testbed 存储
	deploymentService.SetResourceInstanceStorage(resourceStorage) // 设置 ResourceInstance 存储
	adminHandler.SetDeploymentService(deploymentService) // 设置部署服务

	server := &Server{
		Port:            port,
		db:              db,
		service:         poolService,
		jobManager:      jobManager,
		internalHandler: internalHandler,
		externalHandler: externalHandler,
		adminHandler:    adminHandler,
		userStorage:     userStorage,
	}

	// 初始化资源池（创建主线分类和robot用户）
	if err := server.initializeResources(userStorage, categoryStorage, quotaStorage); err != nil {
		log.Printf("[Server] Warning: Failed to initialize resources: %v", err)
		// 不阻塞服务启动，只记录警告
	}

	return server, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	router := mux.NewRouter()

	// 注册路由
	s.internalHandler.RegisterRoutes(router)
	s.externalHandler.RegisterRoutes(router)
	s.adminHandler.RegisterRoutes(router)

	// 静态文件（如果有）
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/app/static")))

	// CORS 中间件
	corsMiddleware := createCORSMiddleware()

	// 启动后台任务
	s.jobManager.Start()
	defer s.jobManager.Stop()

	// 启动 HTTP 服务器
	addr := ":" + s.Port
	log.Printf("[Server] Starting resource pool server on %s", addr)

	handler := corsMiddleware(router)
	return http.ListenAndServe(addr, handler)
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.jobManager.Stop()
	// 关闭 userStorage（如果是 MySQLUserStorage）
	if mysqlUserStorage, ok := s.userStorage.(*MySQLUserStorage); ok {
		mysqlUserStorage.Close()
	}
	return s.db.Close()
}

// createCORSMiddleware 创建 CORS 中间件
func createCORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// 如果有 Origin，则使用它；否则不设置（同源请求不需要 CORS）
			// 注意：当使用 Allow-Credentials: true 时，不能用通配符 *
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// initializeResources 初始化资源池（创建主线分类和robot用户）
func (s *Server) initializeResources(userStorage UserStorage, categoryStorage storage.CategoryStorage, quotaStorage storage.QuotaPolicyStorage) error {
	log.Printf("[Server] Initializing resource pool...")

	// 创建适配器，将 api.UserStorage 转换为 storage.UserStorage
	storageUserStorage := &userStorageAdapter{userStorage: userStorage}

	// 创建初始化器
	initializer := service.NewInitializer(
		s.service.(*service.ResourcePoolServiceImpl),
		storageUserStorage,
		categoryStorage,
		quotaStorage,
	)

	// 执行初始化
	return initializer.Initialize()
}

// userStorageAdapter 适配器，将 api.UserStorage 转换为 storage.UserStorage
type userStorageAdapter struct {
	userStorage UserStorage
}

// GetUserByUsername 根据用户名获取用户
func (a *userStorageAdapter) GetUserByUsername(username string) (*storage.User, error) {
	// 由于 api.UserStorage 没有这个方法，我们需要直接从数据库获取
	// 这里需要访问底层的 MySQLUserStorage
	if mysqlUserStorage, ok := a.userStorage.(*MySQLUserStorage); ok {
		return mysqlUserStorage.GetUserByUsername(username)
	}
	return nil, fmt.Errorf("user storage does not support GetUserByUsername")
}

// CreateUser 创建新用户
func (a *userStorageAdapter) CreateUser(username, hashedPassword, role string) error {
	if mysqlUserStorage, ok := a.userStorage.(*MySQLUserStorage); ok {
		return mysqlUserStorage.CreateUser(username, hashedPassword, role)
	}
	return fmt.Errorf("user storage does not support CreateUser")
}

// WaitForShutdown 等待关闭信号
func (s *Server) WaitForShutdown() {
	// 阻塞主线程
	select {}
}
