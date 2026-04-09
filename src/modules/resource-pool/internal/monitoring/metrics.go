package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AlertSeverity 告警严重程度
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertType 告警类型
type AlertType string

const (
	AlertTypeTestbedUnhealthy      AlertType = "testbed_unhealthy"
	AlertTypeTestbedExhausted      AlertType = "testbed_exhausted"
	AlertTypeResourceLow           AlertType = "resource_low"
	AlertTypeAllocationFailed      AlertType = "allocation_failed"
	AlertTypeDeploymentFailed      AlertType = "deployment_failed"
	AlertTypeHealthCheckFailed     AlertType = "health_check_failed"
	AlertTypeReplenishmentFailed   AlertType = "replenishment_failed"
)

// Alert 告警信息
type Alert struct {
	ID        string      `json:"id"`
	Type      AlertType   `json:"type"`
	Severity  AlertSeverity `json:"severity"`
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Metadata  interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	Acked     bool        `json:"acked"`
	AckedAt   *time.Time  `json:"acked_at,omitempty"`
}

// AlertHandler 告警处理器接口
type AlertHandler interface {
	HandleAlert(alert *Alert) error
}

// Monitor 监控器接口
type Monitor interface {
	Start() error
	Stop() error
	Name() string
}

// MetricsService 监控指标服务
type MetricsService struct {
	alertHandlers  []AlertHandler
	monitors       []Monitor
	alertHistory   []*Alert
	mu             sync.RWMutex
	stopChan       chan struct{}
	interval       time.Duration
}

// NewMetricsService 创建监控服务
func NewMetricsService() *MetricsService {
	return &MetricsService{
		alertHandlers: make([]AlertHandler, 0),
		monitors:      make([]Monitor, 0),
		alertHistory:  make([]*Alert, 0, 1000),
		stopChan:      make(chan struct{}),
		interval:      1 * time.Minute,
	}
}

// RegisterAlertHandler 注册告警处理器
func (s *MetricsService) RegisterAlertHandler(handler AlertHandler) {
	s.alertHandlers = append(s.alertHandlers, handler)
}

// RegisterMonitor 注册监控器
func (s *MetricsService) RegisterMonitor(monitor Monitor) {
	s.monitors = append(s.monitors, monitor)
}

// Start 启动监控服务
func (s *MetricsService) Start() error {
	log.Printf("[MetricsService] Starting monitoring service")

	// 启动所有监控器
	for _, monitor := range s.monitors {
		if err := monitor.Start(); err != nil {
			log.Printf("[MetricsService] Failed to start monitor %s: %v", monitor.Name(), err)
		}
	}

	// 启动告警清理协程
	go s.cleanupOldAlerts()

	return nil
}

// Stop 停止监控服务
func (s *MetricsService) Stop() {
	log.Printf("[MetricsService] Stopping monitoring service")

	// 停止所有监控器
	for _, monitor := range s.monitors {
		_ = monitor.Stop()
	}

	close(s.stopChan)
}

// cleanupOldAlerts 清理旧的告警记录
func (s *MetricsService) cleanupOldAlerts() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			cutoff := time.Now().Add(-24 * time.Hour) // 保留24小时内的告警
			var newHistory []*Alert
			for _, alert := range s.alertHistory {
				if alert.CreatedAt.After(cutoff) {
					newHistory = append(newHistory, alert)
				}
			}
			s.alertHistory = newHistory
			s.mu.Unlock()
		case <-s.stopChan:
			return
		}
	}
}

// SendAlert 发送告警
func (s *MetricsService) SendAlert(alertType AlertType, severity AlertSeverity, title, message string, metadata interface{}) {
	alert := &Alert{
		ID:        generateAlertID(),
		Type:      alertType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		Acked:     false,
	}

	// 记录告警历史
	s.mu.Lock()
	if len(s.alertHistory) >= cap(s.alertHistory) {
		// 移除最旧的告警
		s.alertHistory = s.alertHistory[1:]
	}
	s.alertHistory = append(s.alertHistory, alert)
	s.mu.Unlock()

	// 记录日志
	log.Printf("[Alert] %s: %s - %s", severity, title, message)

	// 通知所有处理器
	for _, handler := range s.alertHandlers {
		if err := handler.HandleAlert(alert); err != nil {
			log.Printf("[Alert] Handler failed: %v", err)
		}
	}
}

// GetRecentAlerts 获取最近的告警
func (s *MetricsService) GetRecentAlerts(limit int) []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.alertHistory) {
		limit = len(s.alertHistory)
	}

	// 返回最近的告警（倒序）
	result := make([]*Alert, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.alertHistory[len(s.alertHistory)-1-i]
	}
	return result
}

// GetAlertsByType 根据类型获取告警
func (s *MetricsService) GetAlertsByType(alertType AlertType, limit int) []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Alert
	for i := len(s.alertHistory) - 1; i >= 0 && len(result) < limit; i-- {
		if s.alertHistory[i].Type == alertType {
			result = append(result, s.alertHistory[i])
		}
	}
	return result
}

// AcknowledgeAlert 确认告警
func (s *MetricsService) AcknowledgeAlert(alertID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, alert := range s.alertHistory {
		if alert.ID == alertID {
			now := time.Now()
			alert.Acked = true
			alert.AckedAt = &now
			return nil
		}
	}
	return ErrAlertNotFound
}

// GetStats 获取告警统计
func (s *MetricsService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total"] = len(s.alertHistory)
	stats["acked"] = 0
	stats["unacked"] = 0

	typeCount := make(map[AlertType]int)
	severityCount := make(map[AlertSeverity]int)

	ackedCount := 0
	unackedCount := 0

	for _, alert := range s.alertHistory {
		if alert.Acked {
			ackedCount++
		} else {
			unackedCount++
		}
		typeCount[alert.Type]++
		severityCount[alert.Severity]++
	}

	stats["acked"] = ackedCount
	stats["unacked"] = unackedCount

	// 转换为普通 map
	statsByType := make(map[string]int)
	for k, v := range typeCount {
		statsByType[string(k)] = v
	}
	stats["by_type"] = statsByType

	statsBySeverity := make(map[string]int)
	for k, v := range severityCount {
		statsBySeverity[string(k)] = v
	}
	stats["by_severity"] = statsBySeverity

	return stats
}

// generateAlertID 生成告警ID
func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

// LogAlertHandler 日志告警处理器
type LogAlertHandler struct{}

// NewLogAlertHandler 创建日志告警处理器
func NewLogAlertHandler() *LogAlertHandler {
	return &LogAlertHandler{}
}

// HandleAlert 处理告警，输出到日志
func (h *LogAlertHandler) HandleAlert(alert *Alert) error {
	log.Printf("[AlertHandler] Type=%s, Severity=%s, Title=%s, Message=%s",
		alert.Type, alert.Severity, alert.Title, alert.Message)
	return nil
}

// EmailAlertHandler 邮件告警处理器（预留）
type EmailAlertHandler struct {
	recipients []string
	smtpHost   string
	smtpPort   int
	username   string
	password   string
	from       string
}

// NewEmailAlertHandler 创建邮件告警处理器
func NewEmailAlertHandler(recipients []string, smtpHost string, smtpPort int, username, password, from string) *EmailAlertHandler {
	return &EmailAlertHandler{
		recipients: recipients,
		smtpHost:   smtpHost,
		smtpPort:   smtpPort,
		username:   username,
		password:   password,
		from:       from,
	}
}

// HandleAlert 发送邮件告警
func (h *EmailAlertHandler) HandleAlert(alert *Alert) error {
	// TODO: 实现邮件发送逻辑
	log.Printf("[EmailAlertHandler] Would send email alert: %s", alert.Title)
	return nil
}

// WebhookAlertHandler Webhook 告警处理器
type WebhookAlertHandler struct {
	url     string
	headers map[string]string
}

// NewWebhookAlertHandler 创建 Webhook 告警处理器
func NewWebhookAlertHandler(url string, headers map[string]string) *WebhookAlertHandler {
	return &WebhookAlertHandler{
		url:     url,
		headers: headers,
	}
}

// HandleAlert 发送 Webhook 告警
func (h *WebhookAlertHandler) HandleAlert(alert *Alert) error {
	// TODO: 实现 HTTP POST 到 webhook URL
	log.Printf("[WebhookAlertHandler] Would send webhook to %s: %s", h.url, alert.Title)
	return nil
}

// TestbedMonitor Testbed 监控器
type TestbedMonitor struct {
	testbedStorage TestbedStorage
	quotaStorage   QuotaPolicyStorage
	interval       time.Duration
	stopChan       chan struct{}
	metricsService *MetricsService
}

// TestbedStorage Testbed 存储接口（简化版）
type TestbedStorage interface {
	CountAvailableTestbedsByCategory(categoryUUID string) (int, error)
}

// QuotaPolicyStorage 配额策略存储接口（简化版）
type QuotaPolicyStorage interface {
	ListPoliciesByPriority() ([]*QuotaPolicy, error)
}

// QuotaPolicy 配额策略（简化版）
type QuotaPolicy struct {
	CategoryUUID        string
	ReplenishThreshold  int
	AutoReplenish        bool
}

// NewTestbedMonitor 创建 Testbed 监控器
func NewTestbedMonitor(testbedStorage TestbedStorage, quotaStorage QuotaPolicyStorage, metricsService *MetricsService) *TestbedMonitor {
	return &TestbedMonitor{
		testbedStorage: testbedStorage,
		quotaStorage:   quotaStorage,
		interval:       2 * time.Minute,
		stopChan:       make(chan struct{}),
		metricsService: metricsService,
	}
}

// Start 启动监控
func (m *TestbedMonitor) Start() error {
	log.Printf("[TestbedMonitor] Starting testbed monitor")

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.check()
			case <-m.stopChan:
				log.Printf("[TestbedMonitor] Stopped")
				return
			}
		}
	}()

	return nil
}

// Stop 停止监控
func (m *TestbedMonitor) Stop() error {
	close(m.stopChan)
	return nil
}

// Name 监控器名称
func (m *TestbedMonitor) Name() string {
	return "TestbedMonitor"
}

// check 执行检查
func (m *TestbedMonitor) check() {
	// 检查所有类别的可用 Testbed 数量
	policies, err := m.quotaStorage.ListPoliciesByPriority()
	if err != nil {
		log.Printf("[TestbedMonitor] Failed to list policies: %v", err)
		return
	}

	for _, policy := range policies {
		if !policy.AutoReplenish {
			continue
		}

		availableCount, err := m.testbedStorage.CountAvailableTestbedsByCategory(policy.CategoryUUID)
		if err != nil {
			continue
		}

		// 检查是否接近阈值
		if availableCount <= 2 {
			m.metricsService.SendAlert(
				AlertTypeTestbedExhausted,
				AlertSeverityWarning,
				"Testbed 数量不足",
				fmt.Sprintf("类别 %s 可用 Testbed 仅剩 %d 个", policy.CategoryUUID, availableCount),
				map[string]interface{}{
					"category_uuid":   policy.CategoryUUID,
					"available_count":  availableCount,
					"threshold":        policy.ReplenishThreshold,
				},
			)
		}
	}
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	resourceStorage ResourceInstanceStorage
	interval         time.Duration
	stopChan         chan struct{}
	metricsService   *MetricsService
}

// ResourceInstanceStorage 资源实例存储接口（简化版）
type ResourceInstanceStorage interface {
	CountAvailableInstances() (int, error)
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(resourceStorage ResourceInstanceStorage, metricsService *MetricsService) *ResourceMonitor {
	return &ResourceMonitor{
		resourceStorage: resourceStorage,
		interval:         3 * time.Minute,
		stopChan:         make(chan struct{}),
		metricsService:   metricsService,
	}
}

// Start 启动监控
func (m *ResourceMonitor) Start() error {
	log.Printf("[ResourceMonitor] Starting resource monitor")

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.check()
			case <-m.stopChan:
				log.Printf("[ResourceMonitor] Stopped")
				return
			}
		}
	}()

	return nil
}

// Stop 停止监控
func (m *ResourceMonitor) Stop() error {
	close(m.stopChan)
	return nil
}

// Name 监控器名称
func (m *ResourceMonitor) Name() string {
	return "ResourceMonitor"
}

// check 执行检查
func (m *ResourceMonitor) check() {
	availableCount, err := m.resourceStorage.CountAvailableInstances()
	if err != nil {
		return
	}

	// 可用资源少于 5 个时发出警告
	if availableCount < 5 {
		m.metricsService.SendAlert(
			AlertTypeResourceLow,
			AlertSeverityWarning,
			"可用资源实例不足",
			fmt.Sprintf("可用资源实例仅剩 %d 个", availableCount),
			map[string]interface{}{
				"available_count": availableCount,
			},
		)
	}
}

// Errors
var (
	ErrAlertNotFound = fmt.Errorf("alert not found")
)
