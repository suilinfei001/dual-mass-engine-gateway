package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
)

const (
	// MaxSingleLogContentSize is the maximum size for a single log file (10KB)
	MaxSingleLogContentSize = 10240
)

// LogAnalyzer handles AI-based log analysis
type LogAnalyzer struct {
	aiClient      *AIClient
	configStorage *storage.MySQLConfigStorage
	requestPool   *AIRequestPool
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer(configStorage *storage.MySQLConfigStorage) *LogAnalyzer {
	return &LogAnalyzer{
		aiClient:      NewAIClient(configStorage),
		configStorage: configStorage,
		requestPool:   GetGlobalRequestPool(configStorage),
	}
}

// LogAnalysisResult represents the AI response for log analysis
type LogAnalysisResult struct {
	Results []models.TaskResult `json:"results"`
}

// AnalyzeLogsRequest represents the request for log analysis
type AnalyzeLogsRequest struct {
	LogContent string `json:"log_content"`
}

// GetPromptByTaskName returns the appropriate prompt based on task name
func (la *LogAnalyzer) GetPromptByTaskName(taskName string) string {
	switch taskName {
	case "deployment_deployment":
		return la.GetDeploymentLogPrompt()
	case "specialized_tests":
		return la.GetSpecializedTestsLogPrompt()
	case "basic_ci_all", "basic_ci":
		return la.GetBuildLogPrompt()
	default:
		log.Printf("[LogAnalyzer] No specific prompt for task '%s', using default build log prompt", taskName)
		return la.GetBuildLogPrompt()
	}
}

// GetBuildLogPrompt returns the system prompt for build log analysis (basic_ci_all)
func (la *LogAnalyzer) GetBuildLogPrompt() string {
	return `## 角色

你是一个专业的CI/CD构建日志解析助手，负责从构建日志中提取关键信息并输出结构化的JSON结果。

## 任务

解析给定的构建日志内容，识别以下四类检查项的状态：

1. compilation（编译）：检查镜像构建、Docker构建等编译相关任务
2. code_lint（代码检查）：检查代码静态分析、lint、sonarqube等
3. security_scan（安全扫描）：检查容器安全扫描、Trivy等
4. unit_test（单元测试）：检查测试执行、覆盖率等

## 日志识别规则

### compilation
- 识别关键词：镜像构建、BuildImage、docker build、Build task、Chart构建、Chart 包已成功上传、.tgz
- 提取信息：构建是否成功、生成的镜像地址（amd64/arm64）、Chart 下载链接
- 特殊规则：
  - 查找 docker build 成功的镜像地址，填充到 extra.image.amd64 或 extra.image.arm64
  - 查找 Chart 上传成功的下载链接（如 https://xxx/charts/xxx.tgz），填充到 extra.chart
  - 如果构建任务执行成功，结果为 pass
  - 如果构建任务执行失败，结果为 fail

### code_lint
- 识别关键词：sonarqube、lint、pylint、code check、static analysis、代码检查
- 提取信息：是否执行了代码检查任务
- 特殊规则：
  - 如果日志中没有找到任何代码检查相关的任务执行记录，结果为 skipped，detail 为空字符串
  - 如果找到 sonarqube 或 lint 任务且执行成功，结果为 pass
  - 如果找到任务但执行失败，结果为 fail
  - **重要**：Azure DevOps 流水线中的 "Publish Test Results" 任务输出是报告格式问题，不是代码检查失败，以下信息都不是真正的失败：
    - "not available"、"Timestamp is not available"
    - "It was not possible to find any installed .NET Core SDKs"
    - "dotnet" 相关警告
    - 任何关于 SDK 安装、测试报告格式的警告
  - 判断 code_lint 结果时，应关注实际的 lint 任务执行结果，而非流水线任务发布格式
  - **关键**：只有当日志中明确出现 lint 执行失败、代码检查未通过、静态分析报错等情况时才能判定为 fail

### security_scan
- 识别关键词：trivy、security scan、vulnerability、CVE-、Total:、HIGH、CRITICAL、Scan Docker image
- 提取信息：安全扫描结果、发现的漏洞数量和详情（包括CVE编号、严重程度、影响库）
- 特殊规则：
  - 优先判断漏洞：如果发现漏洞（存在CVE记录、Total > 0、或表格形式的漏洞列表），结果应为 fail，并在detail中列出漏洞详情
  - 如果扫描任务执行成功但未发现漏洞，结果为 pass
  - 如果扫描任务执行失败或超时，结果为 fail 或 timeout
  - 重要：不要直接使用构建任务的状态（exit-code），要以实际漏洞检测结果为准

### unit_test
- 识别关键词：test、TestResults、coverage、passed、failed、Quality Gate UT、Quality Gate Coverage
- 提取信息：测试通过/失败状态、覆盖率分数
- 特殊规则：
  - **判断 pass/fail 的关键依据**：
    - 如果日志中出现 "[SUCCESS]" 或 "policy passed"，结果为 pass
    - 如果日志中出现 "[FAIL]" 或 "policy failed"，结果为 fail
    - 如果日志中出现 "Code Coverage (%): X"，提取覆盖率分数到 score 字段
  - **重要**：warning 不等于 fail！
    - "Warnings policy passed with X warning(s)" 表示测试通过，结果是 pass
    - "warning" 只是警告，不是失败
    - 只有明确出现 "failed"、"error"、"FAIL" 才是失败
  - 如果日志中没有单元测试相关的任务执行记录，结果为 skipped
  - 如果找到覆盖率数据（如 "Code Coverage (%): 31.6167"），提取分数到 extra.score

## 输出格式

请严格按照以下JSON格式输出结果，不要添加任何额外内容：

{"results":[
    {
        "check_type":"compilation",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "chart":"",
            "image":{
                  "amd64":"",
                  "arm64":""
             }
        }
    },
    {
        "check_type":"code_lint",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
              "detail":""
         }
    },
    {
        "check_type":"security_scan",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
              "detail":""
         }
    },
    {
        "check_type":"unit_test",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "score":0
        }
    }
]}

## 输出要求

1. result 字段的取值范围：pass、fail、timeout、cancelled、skipped、running
2. 只输出 JSON 内容，不要添加任何解释性文字
3. 如果某项信息在日志中无法找到，使用空字符串（除score使用0外）
4. 请直接输出JSON结果，不要使用代码块包裹`
}

// GetDeploymentLogPrompt returns the system prompt for deployment log analysis
func (la *LogAnalyzer) GetDeploymentLogPrompt() string {
	return `## 角色

你是一个专业的CI/CD部署日志解析助手，负责从部署日志中提取关键信息并输出结构化的JSON结果。

## 任务

解析给定的部署日志内容，识别部署相关的检查项状态。

## 日志识别规则

### deployment
- 识别关键词：helm install、helm upgrade、kubectl apply、deployment、deploy、Chart、tiller、namespace、pod、service、ingress
- 提取信息：部署是否成功、release名称、namespace、部署的资源、错误信息
- 特殊规则：
  - 如果部署成功完成，结果为 pass
  - 如果部署过程中出现错误（如 helm 失败、pod 创建失败、resource 超时等），结果为 fail，并在 detail 中列出错误详情
  - 如果部署任务超时，结果为 timeout
  - 如果部署任务被取消，结果为 cancelled
  - 如果无法确定部署状态，结果为 unknown

## 输出格式

请严格按照以下JSON格式输出结果，不要添加任何额外内容：

{"results":[
    {
        "check_type":"deployment",
        "result":"pass/fail/timeout/cancelled/skipped/running/unknown",
        "extra":{
            "release":"",
            "namespace":"",
            "detail":""
        }
    }
]}

## 输出要求

1. result 字段的取值范围：pass、fail、timeout、cancelled、skipped、running、unknown
2. 只输出 JSON 内容，不要添加任何解释性文字
3. 如果某项信息在日志中无法找到，使用空字符串
4. 请直接输出JSON结果，不要使用代码块包裹`
}

// GetSpecializedTestsLogPrompt returns the system prompt for specialized tests log analysis
func (la *LogAnalyzer) GetSpecializedTestsLogPrompt() string {
	return `## 角色

你是一个专业的CI/CD专业测试日志解析助手，负责从专业测试日志中提取关键信息并输出结构化的JSON结果。

## 任务

解析给定的专业测试日志内容，识别以下测试类型的检查项状态：

1. api_test（API测试）：检查 API 接口测试、REST、GraphQL 等
2. ui_test（UI测试）：检查前端界面测试、Selenium、Playwright 等
3. e2e_test（端到端测试）：检查端到端业务流程测试
4. performance_test（性能测试）：检查性能测试、压力测试、负载测试、响应时间等

## 日志识别规则

### api_test
- 识别关键词：API test、REST、GraphQL、http、request、response、status code、curl、apifox、postman
- 提取信息：API 测试是否通过、失败的接口、响应状态码
- 特殊规则：
  - 如果所有 API 测试通过，结果为 pass
  - 如果有 API 测试失败，结果为 fail，并在 detail 中列出失败的 API 详情
  - 如果没有执行 API 测试，结果为 skipped

### ui_test
- 识别关键词：UI test、Selenium、Playwright、Cypress、browser、click、element、DOM、screenshot
- 提取信息：UI 测试是否通过、失败的页面/操作
- 特殊规则：
  - 如果所有 UI 测试通过，结果为 pass
  - 如果有 UI 测试失败，结果为 fail，并在 detail 中列出失败的页面或操作
  - 如果没有执行 UI 测试，结果为 skipped

### e2e_test
- 识别关键词：e2e、end-to-end、workflow、scenario、journey、user story
- 提取信息：端到端测试是否通过、失败的流程/场景
- 特殊规则：
  - 如果所有端到端测试通过，结果为 pass
  - 如果有测试失败，结果为 fail，并在 detail 中列出失败的场景
  - 如果没有执行端到端测试，结果为 skipped

### performance_test
- 识别关键词：performance、load test、stress test、QPS、TPS、response time、latency、throughput、concurrency
- 提取信息：性能测试是否通过、QPS/TPS、响应时间、并发数
- 特殊规则：
  - 如果性能指标达标，结果为 pass
  - 如果性能指标不达标（如响应时间过长、QPS 不足），结果为 fail，并在 detail 中列出具体指标
  - 如果没有执行性能测试，结果为 skipped

## 输出格式

请严格按照以下JSON格式输出结果，不要添加任何额外内容：

{"results":[
    {
        "check_type":"api_test",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "detail":""
        }
    },
    {
        "check_type":"ui_test",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "detail":""
        }
    },
    {
        "check_type":"e2e_test",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "detail":""
        }
    },
    {
        "check_type":"performance_test",
        "result":"pass/fail/timeout/cancelled/skipped/running",
        "extra":{
            "qps":0,
            "tps":0,
            "response_time":0,
            "detail":""
        }
    }
]}

## 输出要求

1. result 字段的取值范围：pass、fail、timeout、cancelled、skipped、running
2. 只输出 JSON 内容，不要添加任何解释性文字
3. 如果某项信息在日志中无法找到，使用空字符串（数字类型使用0）
4. 请直接输出JSON结果，不要使用代码块包裹`
}

// GetLogAnalysisPrompt returns the system prompt for log analysis (deprecated, use GetPromptByTaskName)
func (la *LogAnalyzer) GetLogAnalysisPrompt() string {
	return la.GetBuildLogPrompt()
}

// AnalyzeLogs analyzes build logs using AI (deprecated, kept for backward compatibility)
func (la *LogAnalyzer) AnalyzeLogs(logContent string) ([]models.TaskResult, error) {
	return la.analyzeLogsWithPrompt(logContent, la.GetLogAnalysisPrompt())
}

// analyzeLogsWithPrompt analyzes logs using a specific prompt
func (la *LogAnalyzer) analyzeLogsWithPrompt(logContent string, systemPrompt string) ([]models.TaskResult, error) {
	originalSize := len(logContent)
	log.Printf("[LogAnalyzer] Analyzing logs, original content length: %d", originalSize)

	// Truncate log content if it exceeds the maximum size
	truncatedContent := logContent
	if originalSize > MaxSingleLogContentSize {
		truncatedContent = logContent[originalSize-MaxSingleLogContentSize:]
		log.Printf("[LogAnalyzer] Log content truncated from %d to %d characters (kept last portion)",
			originalSize, len(truncatedContent))
	}

	userPrompt := fmt.Sprintf("请解析以下构建日志并输出结果：\n\n%s", truncatedContent)

	chatResp, err := la.aiClient.Chat(&ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call AI for log analysis: %w", err)
	}

	log.Printf("[LogAnalyzer] AI Raw Response: %s", chatResp.Content)

	var analysisResult LogAnalysisResult
	if err := json.Unmarshal([]byte(chatResp.Content), &analysisResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	log.Printf("[LogAnalyzer] Parsed %d results from AI analysis", len(analysisResult.Results))

	return analysisResult.Results, nil
}

// AnalyzeLogDirectory analyzes all log files in a directory concurrently
// Each log file's result is saved as a temporary file, then merged into final result
// taskName is used to select the appropriate prompt for analysis
func (la *LogAnalyzer) AnalyzeLogDirectory(logDir string, taskName string) ([]models.TaskResult, error) {
	log.Printf("[LogAnalyzer] Analyzing log directory: %s for task: %s", logDir, taskName)

	// Get the appropriate prompt based on task name
	prompt := la.GetPromptByTaskName(taskName)
	log.Printf("[LogAnalyzer] Using prompt for task type: %s", taskName)

	// Read all log files from the directory
	logFiles, err := la.readLogFiles(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log files: %w", err)
	}

	if len(logFiles) == 0 {
		return nil, fmt.Errorf("no log files found in directory: %s", logDir)
	}

	log.Printf("[LogAnalyzer] Found %d log files to analyze", len(logFiles))

	// Get AI concurrency setting (default 20, max 50)
	// This is the max concurrent goroutines for this event
	maxConcurrency := 20
	if la.configStorage != nil {
		maxConcurrency, _ = la.configStorage.GetAIConcurrency()
		if maxConcurrency < 1 {
			maxConcurrency = 20
		}
	}
	log.Printf("[LogAnalyzer] Event max concurrency: %d", maxConcurrency)

	// Create a timestamped temp directory for intermediate results inside the log directory
	timestamp := time.Now().Format("20060102_150405")
	tmpDirName := fmt.Sprintf("%s_tmp_%s", filepath.Base(logDir), timestamp)
	tmpDirPath := filepath.Join(logDir, tmpDirName)
	if err := os.MkdirAll(tmpDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	log.Printf("[LogAnalyzer] Created temp directory: %s", tmpDirPath)

	// Get sorted log file names for consistent processing order
	logFileNames := la.getSortedLogFileNames(logFiles)

	// Create a semaphore channel to limit concurrent goroutines
	// This limits concurrency within the event
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Channel to collect results in order
	type analysisResult struct {
		fileName string
		results  []models.TaskResult
		err      error
	}
	resultChan := make(chan analysisResult, len(logFileNames))

	// Phase 1: Process each log file concurrently, send results to channel
	for i, logFileName := range logFileNames {
		wg.Add(1)
		go func(idx int, fileName string) {
			defer wg.Done()

			// Acquire local semaphore (limits concurrency within this event)
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Acquire token from global request pool (limits concurrency across all events)
			// This blocks if the global pool is exhausted
			ctx := context.Background()
			releaseToken, err := la.requestPool.AcquireTokens(ctx, 1)
			if err != nil {
				log.Printf("[LogAnalyzer] Failed to acquire token for %s: %v", fileName, err)
				resultChan <- analysisResult{fileName: fileName, err: err}
				return
			}
			defer releaseToken()

			logContent := logFiles[fileName]
			log.Printf("[LogAnalyzer] [%d/%d] Analyzing: %s (size: %d bytes)",
				idx+1, len(logFileNames), fileName, len(logContent))

			results, err := la.analyzeLogsWithPrompt(logContent, prompt)
			if err != nil {
				log.Printf("[LogAnalyzer] Failed to analyze %s: %v", fileName, err)
				resultChan <- analysisResult{fileName: fileName, err: err}
				return
			}

			// Save individual result to temp file
			tmpResultFileName := fmt.Sprintf("%s_tmp_result.json", fileName)
			tmpResultPath := filepath.Join(tmpDirPath, tmpResultFileName)
			if err := la.saveTmpResult(tmpResultPath, results); err != nil {
				log.Printf("[LogAnalyzer] Failed to save temp result for %s: %v", fileName, err)
			} else {
				log.Printf("[LogAnalyzer] Saved temp result: %s", tmpResultFileName)
			}

			resultChan <- analysisResult{fileName: fileName, results: results}
		}(i, logFileName)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Phase 2: Merge results sequentially in log file order (deterministic)
	allResults := make(map[string]*models.TaskResult)
	for result := range resultChan {
		if result.err != nil {
			continue
		}
		// Merge results in order (this is now sequential, not concurrent)
		for _, r := range result.results {
			la.mergeResult(allResults, r, result.fileName)
		}
	}

	// Convert map to slice
	finalResults := make([]models.TaskResult, 0, len(allResults))
	for _, result := range allResults {
		finalResults = append(finalResults, *result)
	}

	// Sort by check_type for consistent output
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].CheckType < finalResults[j].CheckType
	})

	// Save final merged result
	finalResultPath := filepath.Join(tmpDirPath, "final_result.json")
	if err := la.saveTmpResult(finalResultPath, finalResults); err != nil {
		log.Printf("[LogAnalyzer] Failed to save final result: %v", err)
	} else {
		log.Printf("[LogAnalyzer] Saved final result: final_result.json")
	}

	log.Printf("[LogAnalyzer] Final merged results: %d check types", len(finalResults))
	for _, result := range finalResults {
		log.Printf("[LogAnalyzer]   - %s: %s", result.CheckType, result.Result)
	}

	return finalResults, nil
}

// saveTmpResult saves analysis result to a temporary JSON file
func (la *LogAnalyzer) saveTmpResult(filePath string, results []models.TaskResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// readLogFiles reads all log files from a directory
// Returns a map of filename to content (truncated to MaxSingleLogContentSize if needed)
func (la *LogAnalyzer) readLogFiles(logDir string) (map[string]string, error) {
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
		if !isLogFile(filename) {
			continue
		}

		filePath := filepath.Join(logDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[LogAnalyzer] Failed to read file %s: %v", filename, err)
			continue
		}

		// Truncate if content exceeds MaxSingleLogContentSize
		if len(content) > MaxSingleLogContentSize {
			content = content[len(content)-MaxSingleLogContentSize:]
			log.Printf("[LogAnalyzer] Truncated file %s from %d to %d bytes",
				filename, len(content)+MaxSingleLogContentSize, len(content))
		}

		logFiles[filename] = string(content)
	}

	return logFiles, nil
}

// isLogFile checks if a filename is a valid log file
func isLogFile(filename string) bool {
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
	// Check if the middle part is a number (log_ID.txt format)
	idStr := filename[4 : len(filename)-4]
	_, err := strconv.Atoi(idStr)
	return err == nil
}

// getSortedLogFileNames returns log file names sorted by ID
func (la *LogAnalyzer) getSortedLogFileNames(logFiles map[string]string) []string {
	names := make([]string, 0, len(logFiles))
	for name := range logFiles {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		idI := extractLogID(names[i])
		idJ := extractLogID(names[j])
		return idI < idJ
	})

	return names
}

// extractLogID extracts the numeric ID from a log filename
func extractLogID(filename string) int {
	if len(filename) < 9 {
		return 0
	}
	idStr := filename[4 : len(filename)-4]
	id, _ := strconv.Atoi(idStr)
	return id
}

// mergeResult merges a new result into the results map
// Uses the following merge strategy:
// - If check_type doesn't exist, add it
// - For unit_test: prefer pass results with score over fail results
// - For other types: fail > pass > skipped priority
// - Merge extra fields, with new values taking precedence
func (la *LogAnalyzer) mergeResult(allResults map[string]*models.TaskResult, newResult models.TaskResult, sourceFile string) {
	existing, exists := allResults[newResult.CheckType]

	if !exists {
		// Add new result
		resultCopy := newResult
		allResults[newResult.CheckType] = &resultCopy
		log.Printf("[LogAnalyzer] Added new check_type: %s = %s (from %s)",
			newResult.CheckType, newResult.Result, sourceFile)
		return
	}

	// Special handling for unit_test: prefer pass with score, take the LAST one (by log ID)
	if newResult.CheckType == "unit_test" {
		newScore := la.getScoreFromExtra(newResult.Extra)

		if newResult.Result == "pass" && newScore > 0 {
			// For pass results with score: always use pass and update with this score
			// Since we process sequentially by log ID, the LAST pass with score wins
			existing.Result = "pass"
			if existing.Extra == nil {
				existing.Extra = newResult.Extra
			} else if newResult.Extra != nil {
				existing.Extra["score"] = newScore
			}
			log.Printf("[LogAnalyzer] [unit_test] Updated pass with score %v (from %s)",
				newScore, sourceFile)
			return
		}
		// For pass with score=0, skipped, or fail: use normal priority logic below
	}

	// Normal merge strategy for other types: fail > pass > skipped
	priority := map[string]int{
		"fail":      3,
		"timeout":   2,
		"cancelled": 2,
		"pass":      1,
		"running":   1,
		"skipped":   0,
	}

	// Special handling for code_lint: if fail result has empty detail, treat it as invalid
	// This is because Azure DevOps may report "not available" which AI sometimes misinterprets as fail
	if newResult.CheckType == "code_lint" && newResult.Result == "fail" {
		detail := ""
		if newResult.Extra != nil {
			if d, ok := newResult.Extra["detail"].(string); ok {
				detail = d
			}
		}
		if detail == "" {
			// Fail with empty detail is likely a misjudgment, treat as skipped
			log.Printf("[LogAnalyzer] [code_lint] Skipping fail result with empty detail from %s (likely misjudgment)", sourceFile)
			newResult.Result = "skipped"
		}
	}

	// Track if we're upgrading result priority (e.g., pass -> fail)
	upgradingPriority := priority[newResult.Result] > priority[existing.Result]

	if priority[newResult.Result] > priority[existing.Result] {
		existing.Result = newResult.Result
		log.Printf("[LogAnalyzer] Updated %s result: %s -> %s (from %s)",
			newResult.CheckType, existing.Result, newResult.Result, sourceFile)
	}

	// Merge extra fields - strategy depends on priority change
	if existing.Extra == nil {
		// For unit_test, only set Extra if score > 0 (avoid zero scores)
		if newResult.CheckType != "unit_test" || la.getScoreFromExtra(newResult.Extra) > 0 {
			existing.Extra = newResult.Extra
		} else if newResult.Extra != nil {
			// For unit_test with score=0, create Extra but don't include score
			existing.Extra = make(map[string]interface{})
			for key, value := range newResult.Extra {
				if key != "score" {
					existing.Extra[key] = value
				}
			}
		}
	} else if newResult.Extra != nil {
		// Smart merge: only update with non-empty values
		for key, newValue := range newResult.Extra {
			// Check if existing has this key
			if existingValue, exists := existing.Extra[key]; exists {
				// For nested maps (like image), do a smart merge
				if existingMap, ok := existingValue.(map[string]interface{}); ok {
					if newMap, ok := newValue.(map[string]interface{}); ok {
						// Merge nested map, keeping non-empty values
						for nestedKey, nestedValue := range newMap {
							if nestedStr, ok := nestedValue.(string); ok && nestedStr != "" {
								existingMap[nestedKey] = nestedValue
							}
						}
						continue
					}
				}
				// For strings, only update if new value is non-empty
				if _, ok := existingValue.(string); ok {
					if newStr, ok := newValue.(string); ok {
						if newStr != "" {
							existing.Extra[key] = newValue
						}
					} else if newValue != nil {
						// Non-string new value, replace
						existing.Extra[key] = newValue
					}
					continue
				}
				// For numbers (like score), preserve existing value when:
				// 1. Upgrading priority (e.g., pass -> fail), keep existing score
				// 2. New value is 0, keep existing non-zero score
				// 3. Both values have same priority, keep the higher score
				if _, ok := existingValue.(float64); ok {
					if newFloat, ok := newValue.(float64); ok {
						existingFloat := existingValue.(float64)
						if upgradingPriority {
							// Keep existing score when upgrading priority (pass -> fail)
							continue
						}
						// Same priority: only update if new score is higher (not 0)
						if newFloat > existingFloat {
							existing.Extra[key] = newValue
						}
						// else: keep existing (higher) score instead of 0 or lower score
						continue
					}
				}
				// For other numeric types, use same logic
				if _, ok := existingValue.(int); ok {
					if newInt, ok := newValue.(int); ok {
						existingInt := existingValue.(int)
						if upgradingPriority {
							// Keep existing value when upgrading priority
							continue
						}
						// Same priority: only update if new value is higher
						if newInt > existingInt {
							existing.Extra[key] = newValue
						}
						continue
					}
				}
				// For other types, always update with new value
				existing.Extra[key] = newValue
			} else {
				// Key doesn't exist in existing, add it if non-empty
				if newStr, ok := newValue.(string); ok {
					if newStr != "" {
						existing.Extra[key] = newValue
					}
				} else if newMap, ok := newValue.(map[string]interface{}); ok {
					// For maps, add only if at least one value is non-empty
					hasContent := false
					for _, v := range newMap {
						if str, ok := v.(string); ok && str != "" {
							hasContent = true
							break
						}
					}
					if hasContent {
						existing.Extra[key] = newValue
					}
				} else {
					// Other types (like numbers), add as-is
					existing.Extra[key] = newValue
				}
			}
		}
	}

	// Append output if new result has output
	if newResult.Output != "" {
		if existing.Output != "" {
			existing.Output += "; " + newResult.Output
		} else {
			existing.Output = newResult.Output
		}
	}
}

// getScoreFromExtra extracts the score value from Extra map
// Returns 0 if score doesn't exist or is not a number
func (la *LogAnalyzer) getScoreFromExtra(extra map[string]interface{}) float64 {
	if extra == nil {
		return 0
	}
	if scoreValue, exists := extra["score"]; exists {
		switch v := scoreValue.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case float32:
			return float64(v)
		}
	}
	return 0
}

// AnalyzeLogsFromJSON analyzes logs from JSON request (for API usage)
func (la *LogAnalyzer) AnalyzeLogsFromJSON(jsonRequest []byte) ([]byte, error) {
	var req AnalyzeLogsRequest
	if err := json.Unmarshal(jsonRequest, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	results, err := la.AnalyzeLogs(req.LogContent)
	if err != nil {
		resp := map[string]interface{}{
			"error": err.Error(),
		}
		return json.Marshal(resp)
	}

	resp := LogAnalysisResult{
		Results: results,
	}
	return json.Marshal(resp)
}
