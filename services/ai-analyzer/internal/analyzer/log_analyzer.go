package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/quality-gateway/ai-analyzer/internal/client"
	"github.com/quality-gateway/ai-analyzer/internal/pool"
	"github.com/quality-gateway/ai-analyzer/internal/types"
)

const (
	// MaxSingleLogContentSize is the maximum size for a single log file (10KB)
	MaxSingleLogContentSize = 10240
)

// LogAnalyzer handles AI-based log analysis
type LogAnalyzer struct {
	aiClient    *client.AIClient
	requestPool *pool.AIRequestPool
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer(aiClient *client.AIClient, requestPool *pool.AIRequestPool) *LogAnalyzer {
	return &LogAnalyzer{
		aiClient:    aiClient,
		requestPool: requestPool,
	}
}

// GetPromptByTaskName returns the appropriate prompt based on task name
func (la *LogAnalyzer) GetPromptByTaskName(taskName string) string {
	switch taskName {
	case "deployment_deployment", "deployment":
		return la.GetDeploymentLogPrompt()
	case "specialized_tests", "specialized_tests_api_test", "specialized_tests_module_e2e", "specialized_tests_agent_e2e", "specialized_tests_ai_e2e":
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

// AnalyzeLog analyzes a single log content using AI
func (la *LogAnalyzer) AnalyzeLog(logContent string, taskName string) ([]CheckResult, error) {
	prompt := la.GetPromptByTaskName(taskName)
	return la.analyzeLogWithPrompt(logContent, prompt)
}

// analyzeLogWithPrompt analyzes logs using a specific prompt
func (la *LogAnalyzer) analyzeLogWithPrompt(logContent string, systemPrompt string) ([]CheckResult, error) {
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

	chatResp, err := la.aiClient.Chat(&types.ChatRequest{
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

// AnalyzeBatch analyzes multiple logs concurrently
func (la *LogAnalyzer) AnalyzeBatch(ctx context.Context, logContents []string, taskName string) ([]CheckResult, error) {
	if len(logContents) == 0 {
		return nil, fmt.Errorf("no log contents provided")
	}

	prompt := la.GetPromptByTaskName(taskName)
	log.Printf("[LogAnalyzer] Batch analyzing %d logs for task: %s", len(logContents), taskName)

	// Create a semaphore channel to limit concurrent goroutines
	maxConcurrency := 20
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Channel to collect results
	type analysisResult struct {
		index   int
		results []CheckResult
		err     error
	}
	resultChan := make(chan analysisResult, len(logContents))

	// Process each log content concurrently
	for i, logContent := range logContents {
		wg.Add(1)
		go func(idx int, content string) {
			defer wg.Done()

			// Acquire token from global request pool
			releaseToken, err := la.requestPool.AcquireTokens(ctx, 1)
			if err != nil {
				log.Printf("[LogAnalyzer] Failed to acquire token for log %d: %v", idx, err)
				resultChan <- analysisResult{index: idx, err: err}
				return
			}
			defer releaseToken()

			// Acquire local semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("[LogAnalyzer] [%d/%d] Analyzing log (size: %d bytes)",
				idx+1, len(logContents), len(content))

			results, err := la.analyzeLogWithPrompt(content, prompt)
			if err != nil {
				log.Printf("[LogAnalyzer] Failed to analyze log %d: %v", idx, err)
				resultChan <- analysisResult{index: idx, err: err}
				return
			}

			resultChan <- analysisResult{index: idx, results: results}
		}(i, logContent)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Merge results from all logs
	allResults := make(map[string]*CheckResult)
	for result := range resultChan {
		if result.err != nil {
			continue
		}
		for _, r := range result.results {
			la.mergeResult(allResults, r)
		}
	}

	// Convert map to slice
	finalResults := make([]CheckResult, 0, len(allResults))
	for _, result := range allResults {
		finalResults = append(finalResults, *result)
	}

	log.Printf("[LogAnalyzer] Batch analysis complete: %d check types", len(finalResults))

	return finalResults, nil
}

// mergeResult merges a new result into the results map
func (la *LogAnalyzer) mergeResult(allResults map[string]*CheckResult, newResult CheckResult) {
	existing, exists := allResults[newResult.CheckType]

	if !exists {
		// Add new result
		resultCopy := newResult
		allResults[newResult.CheckType] = &resultCopy
		return
	}

	// Merge strategy: fail > pass > skipped
	priority := map[string]int{
		"fail":      3,
		"timeout":   2,
		"cancelled": 2,
		"pass":      1,
		"running":   1,
		"skipped":   0,
	}

	if priority[newResult.Result] > priority[existing.Result] {
		existing.Result = newResult.Result
	}

	// Merge extra fields
	if existing.Extra == nil {
		existing.Extra = newResult.Extra
	} else if newResult.Extra != nil {
		for key, value := range newResult.Extra {
			existing.Extra[key] = value
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
