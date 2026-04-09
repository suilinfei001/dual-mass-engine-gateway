package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// ResourcePoolProxyHandler 资源池 API 代理处理器
// 将 /api/resource-pool/* 的请求转发到 resource-pool-server
type ResourcePoolProxyHandler struct {
	configStorage ConfigStorage
	baseURL       string
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	GetEventReceiverIP() (string, error)
}

// NewResourcePoolProxyHandler 创建资源池代理处理器
func NewResourcePoolProxyHandler(configStorage ConfigStorage) *ResourcePoolProxyHandler {
	return &ResourcePoolProxyHandler{
		configStorage: configStorage,
		baseURL:       "http://resource-pool-server:5003", // Docker 网络内默认地址
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *ResourcePoolProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 移除 /api/resource-pool 前缀，获取实际路径
	prefix := "/api/resource-pool/"
	requestPath := r.URL.Path

	if strings.HasPrefix(requestPath, prefix) {
		requestPath = requestPath[len(prefix)-1:] // 保留开头的 /
	} else {
		http.Error(w, "Invalid resource pool proxy path", http.StatusBadRequest)
		return
	}

	// 构建目标 URL
	targetURL, err := url.Parse(h.baseURL + requestPath)
	if err != nil {
		http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
		log.Printf("[ResourcePoolProxy] Failed to parse target URL: %v", err)
		return
	}

	// 保留查询参数
	if r.URL.RawQuery != "" {
		targetURL.RawQuery = r.URL.RawQuery
	}

	// 创建代理请求
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		log.Printf("[ResourcePoolProxy] Failed to create proxy request: %v", err)
		return
	}

	// 复制请求头（排除一些不需要的头）
	for name, values := range r.Header {
		// 跳过一些不需要转发的头
		if name == "Host" || name == "Content-Length" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 设置 X-Forwarded-For 头
	if clientIP := r.RemoteAddr; clientIP != "" {
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// 发送代理请求
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reach resource-pool service: %v", err), http.StatusBadGateway)
		log.Printf("[ResourcePoolProxy] Failed to reach resource-pool service: %v", err)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 复制响应状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("[ResourcePoolProxy] Failed to copy response body: %v", err)
	}
}
