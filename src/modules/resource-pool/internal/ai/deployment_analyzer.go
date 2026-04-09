package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// DeploymentLogAnalyzer 部署日志分析器
type DeploymentLogAnalyzer struct {
	configStorage storage.ConfigStorage
	client        *http.Client
}

// NewDeploymentLogAnalyzer 创建部署日志分析器
func NewDeploymentLogAnalyzer(configStorage storage.ConfigStorage) *DeploymentLogAnalyzer {
	return &DeploymentLogAnalyzer{
		configStorage: configStorage,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// DeploymentAnalysisResult 部署分析结果
type DeploymentAnalysisResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message"`
	MariaDBPort  int    `json:"mariadb_port"`
	MariaDBUser  string `json:"mariadb_user"`
	MariaDBPass  string `json:"mariadb_pass"`
	AppPort      int    `json:"app_port"`
	HealthStatus string `json:"health_status"`
	DeploymentID string `json:"deployment_id"`
	Summary      string `json:"summary"`
}

// ChatRequest AI 聊天请求
type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

// ChatResponse AI 聊天响应
type ChatResponse struct {
	Content string
}

// Chat 发送 AI 聊天请求
func (a *DeploymentLogAnalyzer) Chat(req *ChatRequest) (*ChatResponse, error) {
	config, err := a.configStorage.GetAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	if config.IP == "" || config.Model == "" || config.Token == "" {
		return nil, fmt.Errorf("AI not configured")
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8000
	}

	requestBody := map[string]interface{}{
		"model":             config.Model,
		"temperature":       temperature,
		"top_p":             float64(1),
		"max_tokens":        maxTokens,
		"top_k":             int(1),
		"presence_penalty":  float64(0),
		"frequency_penalty": float64(0),
		"stream":            false,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": req.SystemPrompt,
			},
			{
				"role":    "user",
				"content": req.UserPrompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s/api/mf-model-api/v1/chat/completions", config.IP)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.Token)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return &ChatResponse{
		Content: result.Choices[0].Message.Content,
	}, nil
}

// GetDeploymentAnalysisPrompt 获取部署日志分析的系统提示词
func (a *DeploymentLogAnalyzer) GetDeploymentAnalysisPrompt() string {
	return `## 角色

你是一个专业的 Azure DevOps 部署日志解析助手，负责从部署日志中提取关键信息并输出结构化的 JSON 结果。

## 任务

解析给定的部署日志内容，提取以下信息：

1. **部署状态 (success)**: 部署是否成功完成
2. **错误信息 (error_message)**: 如果部署失败，提取错误原因
3. **数据库配置**: MariaDB 端口、用户名、密码
4. **应用配置**: 应用端口
5. **健康检查 (health_status)**: 部署后的健康检查状态
6. **部署标识 (deployment_id)**: 任何部署相关的标识符
7. **摘要 (summary)**: 部署过程的简要总结

## 日志识别规则

### 部署成功识别
- 查找关键词：Successfully deployed、Deployment completed、deployed successfully、finished successfully
- 查找任务完成状态：Job completed、Task completed
- 所有主要步骤都成功执行

### 部署失败识别
- 查找关键词：Failed、Error、Exception、deployment failed
- 查找错误退出码：exit code 1、exit-code=1
- 任务超时：timeout、timed out

### 数据库配置提取
- 查找关键词：MariaDB、MySQL、database、db_port、3306
- 端口格式：PORT=3306、port: 3306、--port=3306
- 用户名格式：user=、username=、-u

### 应用配置提取
- 查找关键词：app_port、application port、8080、3000
- 端口格式：PORT=8080、port: 8080

### 健康检查提取
- 查找关键词：health check、healthcheck、curl、wget
- 状态：healthy、unhealthy、OK、FAIL

## 输出格式

请严格按照以下 JSON 格式输出结果，不要添加任何额外内容：

{
    "success": true|false,
    "error_message": "错误描述或null",
    "mariadb_port": 端口号或0,
    "mariadb_user": "用户名或null",
    "mariadb_pass": "密码或null",
    "app_port": 端口号或0,
    "health_status": "healthy|unhealthy|unknown",
    "deployment_id": "标识符或null",
    "summary": "部署摘要"
}

## 输出要求

1. 只输出 JSON 内容，不要添加任何解释性文字
2. success 字段必须为 true 或 false
3. 如果某项信息在日志中无法找到，使用 null（数值类型使用 0）
4. error_message 仅在部署失败时提供具体错误信息
5. 请直接输出 JSON 结果，不要使用代码块包裹`
}

// AnalyzeLog 分析单条部署日志
func (a *DeploymentLogAnalyzer) AnalyzeLog(logContent string) (*DeploymentAnalysisResult, error) {
	const maxLogContentSize = 10240 // 10KB

	originalSize := len(logContent)
	truncatedContent := logContent

	if originalSize > maxLogContentSize {
		truncatedContent = logContent[originalSize-maxLogContentSize:]
		log.Printf("[DeploymentAnalyzer] Log truncated from %d to %d bytes", originalSize, len(truncatedContent))
	}

	userPrompt := fmt.Sprintf("请解析以下部署日志并输出结果：\n\n%s", truncatedContent)

	chatResp, err := a.Chat(&ChatRequest{
		SystemPrompt: a.GetDeploymentAnalysisPrompt(),
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    8000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call AI for log analysis: %w", err)
	}

	log.Printf("[DeploymentAnalyzer] AI Raw Response: %s", chatResp.Content)

	var result DeploymentAnalysisResult
	if err := json.Unmarshal([]byte(chatResp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// AnalyzeLogDirectory 分析日志目录中的所有日志文件
func (a *DeploymentLogAnalyzer) AnalyzeLogDirectory(logDir string) (*DeploymentAnalysisResult, error) {
	log.Printf("[DeploymentAnalyzer] Analyzing log directory: %s", logDir)

	// 读取所有日志文件
	logFiles, err := a.readLogFiles(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log files: %w", err)
	}

	if len(logFiles) == 0 {
		return nil, fmt.Errorf("no log files found in directory: %s", logDir)
	}

	log.Printf("[DeploymentAnalyzer] Found %d log files to analyze", len(logFiles))

	// 获取排序后的日志文件名
	logFileNames := a.getSortedLogFileNames(logFiles)

	// 限制并发数
	maxConcurrency := 10
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// 收集所有分析结果
	resultChan := make(chan *DeploymentAnalysisResult, len(logFileNames))
	errorChan := make(chan error, len(logFileNames))

	// 并发处理每个日志文件
	for i, logFileName := range logFileNames {
		wg.Add(1)
		go func(idx int, fileName string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			logContent := logFiles[fileName]
			log.Printf("[DeploymentAnalyzer] [%d/%d] Analyzing: %s (size: %d bytes)",
				idx+1, len(logFileNames), fileName, len(logContent))

			result, err := a.AnalyzeLog(logContent)
			if err != nil {
				log.Printf("[DeploymentAnalyzer] Failed to analyze %s: %v", fileName, err)
				errorChan <- err
				return
			}

			resultChan <- result
		}(i, logFileName)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 合并结果
	var finalResult *DeploymentAnalysisResult
	resultCount := 0
	var errors []error

	for result := range resultChan {
		resultCount++
		if finalResult == nil {
			finalResult = result
		} else {
			// 合并策略：优先保留成功的结果和更完整的信息
			a.mergeResults(finalResult, result)
		}
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	if finalResult == nil {
		return nil, fmt.Errorf("all log analyses failed, errors: %v", errors)
	}

	log.Printf("[DeploymentAnalyzer] Analysis completed: %d/%d files successful, success=%v",
		resultCount, len(logFileNames), finalResult.Success)

	return finalResult, nil
}

// mergeResults 合并两个分析结果
func (a *DeploymentLogAnalyzer) mergeResults(base, newer *DeploymentAnalysisResult) {
	// 如果 newer 结果显示失败，更新错误信息
	if !newer.Success && base.Success {
		base.Success = false
		base.ErrorMessage = newer.ErrorMessage
	}

	// 合并端口信息（保留非零值）
	if newer.MariaDBPort > 0 && base.MariaDBPort == 0 {
		base.MariaDBPort = newer.MariaDBPort
	}
	if newer.AppPort > 0 && base.AppPort == 0 {
		base.AppPort = newer.AppPort
	}

	// 合并用户名
	if newer.MariaDBUser != "" && base.MariaDBUser == "" {
		base.MariaDBUser = newer.MariaDBUser
	}

	// 合并密码
	if newer.MariaDBPass != "" && base.MariaDBPass == "" {
		base.MariaDBPass = newer.MariaDBPass
	}

	// 合并健康状态（优先使用非 unknown 状态）
	if newer.HealthStatus != "unknown" && base.HealthStatus == "unknown" {
		base.HealthStatus = newer.HealthStatus
	}

	// 合并部署 ID
	if newer.DeploymentID != "" && base.DeploymentID == "" {
		base.DeploymentID = newer.DeploymentID
	}

	// 合并摘要
	if newer.Summary != "" {
		if base.Summary != "" {
			base.Summary += "; " + newer.Summary
		} else {
			base.Summary = newer.Summary
		}
	}
}

// readLogFiles 读取目录中的所有日志文件
func (a *DeploymentLogAnalyzer) readLogFiles(logDir string) (map[string]string, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	logFiles := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !a.isLogFile(filename) {
			continue
		}

		filePath := filepath.Join(logDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[DeploymentAnalyzer] Failed to read file %s: %v", filename, err)
			continue
		}

		// 限制单文件大小
		const maxSingleLogContentSize = 10240 // 10KB
		if len(content) > maxSingleLogContentSize {
			content = content[len(content)-maxSingleLogContentSize:]
			log.Printf("[DeploymentAnalyzer] Truncated file %s to %d bytes", filename, len(content))
		}

		logFiles[filename] = string(content)
	}

	return logFiles, nil
}

// isLogFile 检查文件名是否为有效的日志文件
func (a *DeploymentLogAnalyzer) isLogFile(filename string) bool {
	// 跳过元数据文件和错误文件
	if filename == "metadata.txt" {
		return false
	}
	if strings.HasSuffix(filename, "_error.txt") {
		return false
	}
	if !strings.HasSuffix(filename, ".txt") {
		return false
	}
	if !strings.HasPrefix(filename, "log_") {
		return false
	}
	// 检查中间部分是否为数字 (log_ID.txt 格式)
	idStr := filename[4 : len(filename)-4]
	_, err := strconv.Atoi(idStr)
	return err == nil
}

// getSortedLogFileNames 按日志 ID 排序返回日志文件名
func (a *DeploymentLogAnalyzer) getSortedLogFileNames(logFiles map[string]string) []string {
	names := make([]string, 0, len(logFiles))
	for name := range logFiles {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		idI := a.extractLogID(names[i])
		idJ := a.extractLogID(names[j])
		return idI < idJ
	})

	return names
}

// extractLogID 从日志文件名中提取数字 ID
func (a *DeploymentLogAnalyzer) extractLogID(filename string) int {
	if len(filename) < 9 {
		return 0
	}
	idStr := filename[4 : len(filename)-4]
	id, _ := strconv.Atoi(idStr)
	return id
}
