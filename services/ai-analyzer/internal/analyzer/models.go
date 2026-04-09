package analyzer

// LogAnalysisResult represents the AI response for log analysis
type LogAnalysisResult struct {
	Results []CheckResult `json:"results"`
}

// CheckResult represents a single check result from log analysis
type CheckResult struct {
	CheckType string                 `json:"check_type"`
	Result    string                 `json:"result"` // pass/fail/timeout/cancelled/skipped/running
	Extra     map[string]interface{} `json:"extra,omitempty"`
	Output    string                 `json:"output,omitempty"`
}

// AnalyzeRequest represents the request for log analysis
type AnalyzeRequest struct {
	LogContent string `json:"log_content"`
	TaskName   string `json:"task_name,omitempty"`
}

// BatchAnalyzeRequest represents the request for batch log analysis
type BatchAnalyzeRequest struct {
	LogContents []string `json:"log_contents"`
	TaskName    string   `json:"task_name,omitempty"`
}

// PoolSizeRequest represents the request to set pool size
type PoolSizeRequest struct {
	Size int `json:"size"`
}
