// Auth Service - 认证服务
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/quality-gateway/shared/pkg/auth"
)

// corsMiddleware CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 允许的源（生产环境应该从配置读取）
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:8081",
			"http://localhost:8082",
			"http://localhost:8083",
			"http://localhost:8084",
			"http://10.4.111.141:8081",
			"http://10.4.111.141",
			"http://10.4.174.125:8083",
			"http://10.4.174.125:8084",
		}

		// 检查源是否在允许列表中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var (
	port    string
	logLevel string
)

func init() {
	flag.StringVar(&port, "port", "4007", "Server port")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

func main() {
	flag.Parse()

	// 从环境变量覆盖配置
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		logLevel = v
	}

	// 获取管理员凭证
	adminUser, adminPass := auth.GetAdminCredentials()
	log.Printf("Starting Auth Service on port %s", port)
	log.Printf("Default credentials: %s / %s", adminUser, adminPass)
	log.Printf("You can set custom credentials via ADMIN_USER and ADMIN_PASS environment variables")

	// 创建会话存储
	sessionStore := auth.NewSessionStore()

	// 创建路由
	mux := http.NewServeMux()

	// 注册认证 API - 标准路由
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(sessionStore, w, r)
	})
	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		handleLogout(sessionStore, w, r)
	})
	mux.HandleFunc("/api/check-login", func(w http.ResponseWriter, r *http.Request) {
		handleCheckLogin(sessionStore, w, r)
	})

	// 注册认证 API - /api/auth/* 路由（兼容 event-processor）
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(sessionStore, w, r)
	})
	mux.HandleFunc("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		handleLogout(sessionStore, w, r)
	})
	mux.HandleFunc("/api/auth/check-login", func(w http.ResponseWriter, r *http.Request) {
		handleCheckLogin(sessionStore, w, r)
	})
	mux.HandleFunc("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		handleAuthStatus(sessionStore, w, r)
	})
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		handleRegister(sessionStore, w, r)
	})

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// 应用 CORS 中间件
	handler := corsMiddleware(mux)

	// 启动服务器
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleLogin 处理登录请求
func handleLogin(sessionStore *auth.SessionStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证凭证
	if !auth.ValidateCredentials(loginReq.Username, loginReq.Password) {
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": "Invalid username or password",
		}, http.StatusOK)
		return
	}

	// 创建会话
	session, err := sessionStore.CreateSession(loginReq.Username)
	if err != nil {
		writeJSONError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// 设置会话 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.SessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24小时
		SameSite: http.SameSiteLaxMode,
	})

	writeJSONResponse(w, map[string]interface{}{
		"success":  true,
		"message":  "登录成功",
		"user": map[string]interface{}{
			"username": loginReq.Username,
			"role":     "admin",
		},
	}, http.StatusOK)
}

// handleLogout 处理登出请求
func handleLogout(sessionStore *auth.SessionStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取并删除会话
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionStore.DeleteSession(cookie.Value)
	}

	// 清除 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "登出成功",
	}, http.StatusOK)
}

// handleCheckLogin 处理登录状态检查请求
func handleCheckLogin(sessionStore *auth.SessionStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取会话 cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		writeJSONResponse(w, map[string]interface{}{
			"is_logged_in": false,
		}, http.StatusOK)
		return
	}

	// 验证会话
	session, exists := sessionStore.GetSession(cookie.Value)
	if !exists {
		writeJSONResponse(w, map[string]interface{}{
			"is_logged_in": false,
		}, http.StatusOK)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"is_logged_in": true,
		"username":     session.Username,
	}, http.StatusOK)
}

// handleAuthStatus 处理认证状态检查请求 (event-processor 前端使用)
func handleAuthStatus(sessionStore *auth.SessionStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取会话 cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		writeJSONResponse(w, map[string]interface{}{
			"loggedIn": false,
		}, http.StatusOK)
		return
	}

	// 验证会话
	session, exists := sessionStore.GetSession(cookie.Value)
	if !exists {
		writeJSONResponse(w, map[string]interface{}{
			"loggedIn": false,
		}, http.StatusOK)
		return
	}

	// 判断用户角色 (简化实现：默认 admin 用户角色为 admin)
	role := "user"
	if session.Username == "admin" {
		role = "admin"
	}

	writeJSONResponse(w, map[string]interface{}{
		"loggedIn": true,
		"user": map[string]interface{}{
			"username": session.Username,
			"role":     role,
		},
	}, http.StatusOK)
}

// handleRegister 处理注册请求
func handleRegister(sessionStore *auth.SessionStore, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var registerReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 简单实现：只允许管理员用户名注册
	// 生产环境应该有更完善的用户管理
	if registerReq.Username == "" || registerReq.Password == "" {
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": "用户名和密码不能为空",
		}, http.StatusOK)
		return
	}

	// 检查是否是保留的管理员用户名
	adminUser, _ := auth.GetAdminCredentials()
	if registerReq.Username == adminUser {
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": "该用户名已被保留",
		}, http.StatusOK)
		return
	}

	// 在实际生产环境中，这里应该将用户保存到数据库
	// 当前简化实现：只允许默认管理员登录，其他用户注册后也无法使用
	writeJSONResponse(w, map[string]interface{}{
		"success": false,
		"message": "当前只支持默认管理员账号登录，请联系管理员获取账号",
	}, http.StatusOK)
}

// writeJSONResponse 写入 JSON 响应
func writeJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError 写入 JSON 错误响应
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
