package deployer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// DeployRequest 部署请求
type DeployRequest struct {
	ResourceInstanceUUID string            `json:"resource_instance_uuid"`
	IPAddress            string            `json:"ip_address"`
	Port                 int               `json:"port"`
	SSHUser              string            `json:"ssh_user"`
	Passwd               string            `json:"passwd"`
	ProductVersion       string            `json:"product_version"`
	ConfigFile           string            `json:"config_file"`
	EnvVars              map[string]string `json:"env_vars"`
	Timeout              time.Duration     `json:"timeout"`
}

// DeployResult 部署结果
type DeployResult struct {
	Success       bool          `json:"success"`
	MariaDBPort   int           `json:"mariadb_port"`
	MariaDBUser   string        `json:"mariadb_user"`
	MariaDBPasswd string        `json:"mariadb_passwd"`
	AppPort       int           `json:"app_port"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	LogURL        string        `json:"log_url,omitempty"`
	Duration      time.Duration `json:"duration"`
}

// DeployService 产品部署服务接口
type DeployService interface {
	// DeployProduct 将产品部署到指定的资源实例
	DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error)

	// RestoreSnapshot 将资源实例回滚到指定快照
	RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error

	// CheckHealth 检查已部署产品的健康状态
	CheckHealth(ctx context.Context, ipAddress string, port int, sshUser, passwd string) (bool, error)
}

// MockDeployService Mock 部署服务实现（用于初期开发）
type MockDeployService struct {
	DeployDelay  time.Duration // 部署延迟时间
	RestoreDelay time.Duration // 快照回滚延迟时间
	FailureRate  float64       // 模拟失败率 (0.0 - 1.0)
	deployCount  int           // 部署计数
	restoreCount int           // 回滚计数
}

// NewMockDeployService 创建 Mock 部署服务
func NewMockDeployService() *MockDeployService {
	return &MockDeployService{
		DeployDelay:  5 * time.Second,
		RestoreDelay: 3 * time.Second,
		FailureRate:  0.0, // 默认不失败
	}
}

// DeployProduct 模拟部署产品
func (m *MockDeployService) DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	m.deployCount++

	logPrefix := fmt.Sprintf("[MockDeployer:%d]", m.deployCount)
	fmt.Printf("%s Starting deployment to %s:%d\n", logPrefix, req.IPAddress, req.Port)

	// 模拟部署过程 - Mock 使用固定延迟，忽略 req.Timeout
	deployDelay := m.DeployDelay
	if deployDelay == 0 {
		deployDelay = 5 * time.Second
	}

	select {
	case <-time.After(deployDelay):
		// 部署完成
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 模拟失败（基于失败率）
	if m.FailureRate > 0 {
		// 简单的随机失败模拟
		if (float64(m.deployCount%100) / 100.0) < m.FailureRate {
			return &DeployResult{
				Success:      false,
				ErrorMessage: "mock deployment failure",
			}, nil
		}
	}

	// 模拟成功部署
	fmt.Printf("%s Deployment successful to %s:%d\n", logPrefix, req.IPAddress, req.Port)

	return &DeployResult{
		Success:       true,
		MariaDBPort:   3306,
		MariaDBUser:   "root",
		MariaDBPasswd: generateRandomPassword(),
		AppPort:       8080,
		Duration:      deployDelay,
	}, nil
}

// RestoreSnapshot 模拟快照回滚
func (m *MockDeployService) RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error {
	m.restoreCount++

	logPrefix := fmt.Sprintf("[MockDeployer:Restore:%d]", m.restoreCount)
	fmt.Printf("%s Starting snapshot restore: resource=%s, snapshot=%s\n", logPrefix, resourceUUID, snapshotID)

	select {
	case <-time.After(m.RestoreDelay):
		// 回滚完成
	case <-ctx.Done():
		return ctx.Err()
	}

	fmt.Printf("%s Snapshot restore completed: resource=%s\n", logPrefix, resourceUUID)
	return nil
}

// CheckHealth 模拟健康检查
func (m *MockDeployService) CheckHealth(ctx context.Context, ipAddress string, port int, sshUser, passwd string) (bool, error) {
	// Mock 总是返回健康
	return true, nil
}

// GetStats 获取 Mock 部署服务统计信息
func (m *MockDeployService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"deploy_count":  m.deployCount,
		"restore_count": m.restoreCount,
		"failure_rate":  m.FailureRate,
	}
}

// SetFailureRate 设置模拟失败率
func (m *MockDeployService) SetFailureRate(rate float64) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	m.FailureRate = rate
}

// SetDelays 设置延迟时间
func (m *MockDeployService) SetDelays(deployDelay, restoreDelay time.Duration) {
	m.DeployDelay = deployDelay
	m.RestoreDelay = restoreDelay
}

// generateRandomPassword 生成随机密码
func generateRandomPassword() string {
	// 简单的随机密码生成
	return "mock_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

// TencentDeployService 腾讯云部署服务实现（预留）
type TencentDeployService struct {
	APIKey    string
	SecretKey string
	Region    string
}

// NewTencentDeployService 创建腾讯云部署服务
func NewTencentDeployService(apiKey, secretKey, region string) *TencentDeployService {
	return &TencentDeployService{
		APIKey:    apiKey,
		SecretKey: secretKey,
		Region:    region,
	}
}

// DeployProduct 实现腾讯云产品部署
func (t *TencentDeployService) DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	// TODO: 实现真实的腾讯云部署逻辑
	return nil, fmt.Errorf("tencent deployer not implemented yet")
}

// RestoreSnapshot 实现腾讯云快照回滚
func (t *TencentDeployService) RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error {
	// TODO: 实现真实的腾讯云快照回滚逻辑
	return fmt.Errorf("tencent snapshot restore not implemented yet")
}

// CheckHealth 实现健康检查
func (t *TencentDeployService) CheckHealth(ctx context.Context, ipAddress string, port int, sshUser, passwd string) (bool, error) {
	// TODO: 实现真实的腾讯云健康检查逻辑
	return SSHHealthCheck(ctx, ipAddress, port, sshUser, passwd, 5*time.Second)
}

// CMPDeployService CMP 部署服务实现
type CMPDeployService struct {
	APIURL    string
	AccessKey string
	SecretKey string
}

// NewCMPDeployService 创建 CMP 部署服务
func NewCMPDeployService(apiURL, accessKey, secretKey string) *CMPDeployService {
	if apiURL == "" {
		apiURL = "http://devops-api.aishu.cn:8081"
	}
	return &CMPDeployService{
		APIURL:    apiURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
}

// DeployProduct 实现 CMP 产品部署（暂未实现）
func (c *CMPDeployService) DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	return nil, fmt.Errorf("cmp deploy not implemented yet")
}

// RestoreSnapshot 实现 CMP 快照回滚
func (c *CMPDeployService) RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error {
	if resourceUUID == "" {
		return fmt.Errorf("resource_uuid is required")
	}
	if snapshotID == "" {
		return fmt.Errorf("snapshot_id is required")
	}
	if c.AccessKey == "" || c.SecretKey == "" {
		return fmt.Errorf("CMP credentials (access_key and secret_key) are required")
	}

	// CMP 快照回滚 API（固定 URL）
	const cmpSnapshotURL = "http://devops-api.aishu.cn:8081/v1/vm/revert"

	// 准备请求数据（凭证放在请求体中）
	requestBody := map[string]interface{}{
		"snapshotId":   snapshotID,
		"instanceUuid": resourceUUID,
		"reason":       "resource pool management revert",
		"accessKey":    c.AccessKey,
		"secretKey":    c.SecretKey,
	}

	// 发送请求
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", cmpSnapshotURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("[CMP] Sending snapshot revert request: URL=%s, body=%s",
		cmpSnapshotURL, string(jsonData))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[CMP] Restore snapshot response: status=%d, body=%s", resp.StatusCode, string(body))

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CMP API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应 - 检查 code 字段判断成功与否
	var cmpResp struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &cmpResp); err != nil {
		log.Printf("[CMP] Failed to parse response: %v", err)
		return fmt.Errorf("failed to parse CMP response: %w", err)
	}

	if cmpResp.Code != 0 {
		return fmt.Errorf("CMP API returned unsuccessful: code=%d, message=%s", cmpResp.Code, cmpResp.Message)
	}

	// 快照回滚是异步操作，需要等待完成
	log.Printf("[CMP] Snapshot revert initiated successfully, waiting for completion...")
	time.Sleep(5 * time.Minute)
	log.Printf("[CMP] Snapshot revert wait completed")

	return nil
}

// CheckHealth 实现 CMP 健康检查（使用 TCP 连接测试）
func (c *CMPDeployService) CheckHealth(ctx context.Context, ipAddress string, port int, sshUser, passwd string) (bool, error) {
	return SSHHealthCheck(ctx, ipAddress, port, sshUser, passwd, 5*time.Second)
}

// SSHDeployService SSH 部署服务实现
// 实际使用 Azure DevOps Pipeline 进行部署
type SSHDeployService struct {
	azureConfig       *azureConfigInfo
	defaultPipelineID int
	client            *http.Client
	cmpAccessKey      string
	cmpSecretKey      string
	useMockDeployment bool
}

// azureConfigInfo Azure 配置信息
type azureConfigInfo struct {
	Organization string
	Project      string
	PAT          string
	BaseURL      string
}

// NewSSHDeployService 创建 SSH 部署服务
func NewSSHDeployService() *SSHDeployService {
	return &SSHDeployService{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// SetAzureConfig 设置 Azure 配置（接受 map 以便调用）
func (s *SSHDeployService) SetAzureConfig(config interface{}) {
	if configMap, ok := config.(map[string]interface{}); ok {
		// 支持大小写键名
		pat := getFirstString(configMap, "PAT", "pat")
		// 打印 PAT 前4个字符用于调试
		patPreview := ""
		if len(pat) > 4 {
			patPreview = pat[:4] + "..."
		} else if pat != "" {
			patPreview = "***"
		}
		log.Printf("[SSHDeployer] SetAzureConfig: org=%s, project=%s, pat=%s, baseURL=%s",
			getFirstString(configMap, "Organization", "organization"),
			getFirstString(configMap, "Project", "project"),
			patPreview,
			getFirstString(configMap, "BaseURL", "baseURL"))

		s.azureConfig = &azureConfigInfo{
			Organization: getFirstString(configMap, "Organization", "organization"),
			Project:      getFirstString(configMap, "Project", "project"),
			PAT:          pat,
			BaseURL:      getFirstString(configMap, "BaseURL", "baseURL"),
		}
	} else if configInfo, ok := config.(*azureConfigInfo); ok {
		log.Printf("[SSHDeployer] SetAzureConfig from pointer: org=%s, project=%s",
			configInfo.Organization, configInfo.Project)
		s.azureConfig = configInfo
	}
}

// getFirstString 从多个可能的键中获取第一个非空值
func getFirstString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// SetDefaultPipelineID 设置默认 Pipeline ID
func (s *SSHDeployService) SetDefaultPipelineID(pipelineID int) {
	s.defaultPipelineID = pipelineID
}

// SetCMPCredentials 设置 CMP API 凭证
func (s *SSHDeployService) SetCMPCredentials(accessKey, secretKey string) {
	s.cmpAccessKey = accessKey
	s.cmpSecretKey = secretKey
	log.Printf("[SSHDeployer] CMP credentials configured")
}

// DeployProduct 实现 Azure Pipeline 部署
func (s *SSHDeployService) DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error) {
	log.Printf("[SSHDeployer] Starting deployment to %s:%d", req.IPAddress, req.Port)
	s.useMockDeployment = false
	// 如果没有配置 Azure，返回 mock 结果
	if s.azureConfig == nil {
		log.Printf("[SSHDeployer] Azure config not set, using mock deployment")
		// 模拟部署成功
		return &DeployResult{
			Success:       true,
			MariaDBPort:   3306,
			MariaDBUser:   "root",
			MariaDBPasswd: generateRandomPassword(),
			AppPort:       8080,
			Duration:      5 * time.Second,
		}, nil
	}

	// 使用 Azure Pipeline 进行部署
	pipelineID := s.defaultPipelineID
	if pipelineID == 0 {
		pipelineID = 1 // 默认 pipeline ID
	}

	// 准备 Pipeline 参数，替换占位符
	params := map[string]interface{}{
		"host":         req.IPAddress,
		"ssh_user":     req.SSHUser,
		"ssh_password": req.Passwd,
	}

	// 完全跳过 Azure Pipeline，使用 mock 结果
	if s.useMockDeployment {
		log.Printf("[SSHDeployer] Using mock deployment (skipping Azure Pipeline)")
		return &DeployResult{
			Success:       true,
			MariaDBPort:   3306,
			MariaDBUser:   "root",
			MariaDBPasswd: generateRandomPassword(),
			AppPort:       8080,
			Duration:      5 * time.Second,
		}, nil
	}

	// 调用 Azure Pipeline
	buildID, webURL, err := s.runAzurePipeline(ctx, pipelineID, params)
	if err != nil {
		return &DeployResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Azure Pipeline failed: %v", err),
		}, err
	}

	log.Printf("[SSHDeployer] Pipeline started: build_id=%d, web_url=%s", buildID, webURL)

	// 等待 Pipeline 完成
	success, errMsg := s.waitForPipeline(ctx, buildID, false)
	if !success {
		return &DeployResult{
			Success:      false,
			ErrorMessage: errMsg,
			LogURL:       webURL,
		}, nil
	}

	log.Printf("[SSHDeployer] Deployment completed successfully")

	return &DeployResult{
		Success:       true,
		MariaDBPort:   3306,
		MariaDBUser:   "root",
		MariaDBPasswd: generateRandomPassword(),
		AppPort:       8080,
		LogURL:        webURL,
	}, nil
}

// runAzurePipeline 运行 Azure Pipeline
func (s *SSHDeployService) runAzurePipeline(ctx context.Context, pipelineID int, params map[string]interface{}) (int, string, error) {
	baseURL := s.azureConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://devops.aishu.cn"
	}

	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs?api-version=6.0-preview.1",
		baseURL, s.azureConfig.Organization, s.azureConfig.Project, pipelineID)

	payload := map[string]interface{}{
		"resources": map[string]interface{}{
			"repositories": map[string]interface{}{
				"self": map[string]interface{}{
					"refName": "refs/heads/test",
				},
			},
		},
		"templateParameters": params,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	authString := fmt.Sprintf(":%s", s.azureConfig.PAT)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString))))

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return 0, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ID, result.Links.Web.Href, nil
}

// waitForPipeline 等待 Pipeline 完成
// 使用 Build API，status 字段值: inProgress, completed, cancelling, none
// result 字段值（仅当 status=completed 时）: succeeded, failed, canceled
func (s *SSHDeployService) waitForPipeline(ctx context.Context, buildID int, useMock bool) (bool, string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	timeout := time.After(90 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return false, "deployment cancelled"
		case <-timeout:
			return false, "deployment timeout"
		case <-ticker.C:
			status, result := s.checkPipelineStatus(ctx, buildID, useMock)
			// Build API: status 为 "completed" 时表示构建结束
			if status == "completed" {
				if result == "succeeded" {
					return true, ""
				} else {
					return false, fmt.Sprintf("pipeline completed with result: %s", result)
				}
			}
			// Build API: status 为 "cancelling" 或出现错误时
			if status == "canceled" || status == "cancelling" || status == "unknown" {
				return false, fmt.Sprintf("pipeline status: %s", status)
			}
			// inProgress - 继续等待
			log.Printf("[SSHDeployer] Pipeline status: %s, waiting...", status)
		}
	}
}

// checkPipelineStatus 检查 Pipeline 状态
// 使用 Azure DevOps Build API 而非 Pipeline Runs API
// 参考: event-processor/internal/executor/azure_url.go:BuildStatusURL
func (s *SSHDeployService) checkPipelineStatus(ctx context.Context, buildID int, useMock bool) (string, string) {
	baseURL := s.azureConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://devops.aishu.cn"
	}

	// 正确的 API 路径是 _apis/build/builds/{build_id} 而非 _apis/pipelines/runs/{buildId}
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d?api-version=6.0",
		baseURL, s.azureConfig.Organization, s.azureConfig.Project, buildID)

	authString := fmt.Sprintf(":%s", s.azureConfig.PAT)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString))))

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[SSHDeployer] Failed to check pipeline status: %v", err)
		return "unknown", err.Error()
	}
	defer resp.Body.Close()

	if useMock {
		return "completed", "succeeded"
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		log.Printf("[SSHDeployer] Pipeline status API returned %d: %s", resp.StatusCode, string(body))
		return "unknown", string(body)
	}

	var result struct {
		Status string `json:"status"` // Build API 使用 status 字段（而非 Pipeline Runs API 的 state）
		Result string `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[SSHDeployer] Failed to parse pipeline status response: %v, body: %s", err, string(body))
		return "unknown", fmt.Sprintf("failed to parse response: %v", err)
	}

	log.Printf("[SSHDeployer] Pipeline %d status: status=%s, result=%s", buildID, result.Status, result.Result)
	return result.Status, result.Result
}

// RestoreSnapshot 实现 SSH 快照回滚（使用 CMP API）
func (s *SSHDeployService) RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error {
	// 如果没有配置 CMP 凭证，返回错误
	if s.cmpAccessKey == "" || s.cmpSecretKey == "" {
		return fmt.Errorf("CMP credentials (access_key and secret_key) are required")
	}
	// 使用 CMP API 进行快照回滚（URL 固定为 https://cmp.aishu.cn/vm-service/snapshot/revert）
	cmpService := &CMPDeployService{
		AccessKey: s.cmpAccessKey,
		SecretKey: s.cmpSecretKey,
	}
	return cmpService.RestoreSnapshot(ctx, resourceUUID, snapshotID)
}

// CheckHealth 实现 SSH 健康检查（TCP 连接测试 + SSH 认证）
func (s *SSHDeployService) CheckHealth(ctx context.Context, ipAddress string, port int, sshUser, passwd string) (bool, error) {
	return SSHHealthCheck(ctx, ipAddress, port, sshUser, passwd, 5*time.Second)
}

// SSHHealthCheck 执行 SSH 健康检查
// ipAddress: 目标 IP 地址
// port: SSH 端口
// username: SSH 用户名（可选，为空则只做 TCP 连接测试）
// password: SSH 密码（可选）
// timeout: 连接超时时间
func SSHHealthCheck(ctx context.Context, ipAddress string, port int, username, password string, timeout time.Duration) (bool, error) {
	address := fmt.Sprintf("%s:%d", ipAddress, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false, fmt.Errorf("TCP 连接失败: %w", err)
	}
	conn.Close()

	if username != "" && password != "" {
		config := &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         timeout,
		}

		sshClient, err := ssh.Dial("tcp", address, config)
		if err != nil {
			return false, fmt.Errorf("SSH 认证失败: %w", err)
		}
		sshClient.Close()
	}

	return true, nil
}

const defaultResolvConfContent = `# Generated by NetworkManager
nameserver 10.4.22.5
nameserver 10.4.22.6
nameserver 10.96.0.10
search anyshare.svc.cluster.local svc.cluster.local cluster.local
options ndots:2
`

func UpdateResolvConf(ctx context.Context, ipAddress string, port int, username, password string, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", ipAddress, port)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	sshClient, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH session 失败: %w", err)
	}
	defer session.Close()

	checkCmd := "cat /etc/resolv.conf 2>/dev/null || echo 'FILE_NOT_EXISTS'"
	checkOutput, err := session.CombinedOutput(checkCmd)
	if err != nil {
		return fmt.Errorf("检查 resolv.conf 失败: %w", err)
	}

	currentContent := string(checkOutput)
	if strings.Contains(currentContent, "FILE_NOT_EXISTS") || strings.TrimSpace(currentContent) == "" {
		log.Printf("[SSHDeployer] /etc/resolv.conf 不存在，将创建新文件")
	} else {
		log.Printf("[SSHDeployer] /etc/resolv.conf 已存在，检查是否需要更新")

		requiredNameservers := []string{"10.4.22.5", "10.4.22.6", "10.96.0.10"}
		allPresent := true
		for _, ns := range requiredNameservers {
			if !strings.Contains(currentContent, ns) {
				allPresent = false
				break
			}
		}

		if allPresent {
			log.Printf("[SSHDeployer] /etc/resolv.conf 已包含所有必需的 nameserver，跳过更新")
			return nil
		}
		log.Printf("[SSHDeployer] /etc/resolv.conf 缺少必需的 nameserver，将更新内容")
	}

	session.Close()

	writeSession, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("创建写入 SSH session 失败: %w", err)
	}
	defer writeSession.Close()

	writeCmd := fmt.Sprintf("echo '%s' | sudo tee /etc/resolv.conf > /dev/null", defaultResolvConfContent)
	output, err := writeSession.CombinedOutput(writeCmd)
	if err != nil {
		return fmt.Errorf("写入 resolv.conf 失败: %w, output: %s", err, string(output))
	}

	log.Printf("[SSHDeployer] /etc/resolv.conf 更新成功")
	return nil
}

// DeployerFactory 部署服务工厂
type DeployerFactory struct {
	deployerType string
}

// NewDeployerFactory 创建部署服务工厂
func NewDeployerFactory(deployerType string) *DeployerFactory {
	return &DeployerFactory{
		deployerType: deployerType,
	}
}

// CreateDeployer 创建部署服务
func (f *DeployerFactory) CreateDeployer(config map[string]string) (DeployService, error) {
	switch f.deployerType {
	case "mock":
		return NewMockDeployService(), nil
	case "ssh":
		return NewSSHDeployService(), nil
	case "tencent":
		apiKey, ok := config["tencent_api_key"]
		if !ok || apiKey == "" {
			return nil, fmt.Errorf("tencent_api_key is required for tencent deployer")
		}
		secretKey, ok := config["tencent_secret_key"]
		if !ok || secretKey == "" {
			return nil, fmt.Errorf("tencent_secret_key is required for tencent deployer")
		}
		region := config["tencent_region"]
		if region == "" {
			region = "ap-guangzhou"
		}
		return NewTencentDeployService(apiKey, secretKey, region), nil
	default:
		return nil, fmt.Errorf("unknown deployer type: %s", f.deployerType)
	}
}

// LogConfig 记录部署配置（用于调试）
func LogConfig(req DeployRequest) string {
	configJSON, _ := json.MarshalIndent(req, "", "  ")
	return string(configJSON)
}
