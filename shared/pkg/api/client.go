package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
)

// ClientConfig holds the client configuration.
type ClientConfig struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	MaxRetries int
}

// Client is an HTTP client for service-to-service communication.
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
	logger     *logger.Logger
}

// NewClient creates a new HTTP client.
func NewClient(cfg *ClientConfig) *Client {
	if cfg == nil {
		cfg = &ClientConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger.New(logger.Config{Level: logger.InfoLevel}),
	}
}

// SetLogger sets the client's logger.
func (c *Client) SetLogger(log *logger.Logger) {
	c.logger = log
}

// Do executes an HTTP request.
func (c *Client) Do(req *http.Request) (*Response, error) {
	// Add authentication header
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Execute request with retry
	var resp *http.Response
	var err error

	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			c.logger.Debug("Retrying request",
				logger.String("url", req.URL.String()),
				logger.Int("attempt", i+1),
			)
			time.Sleep(time.Duration(i) * time.Second)
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// Parse JSON response
	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// If not our standard response format, return raw body
		return &Response{
			Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
			Data:    string(body),
		}, nil
	}

	c.logger.Debug("HTTP response",
		logger.String("url", req.URL.String()),
		logger.Int("status", resp.StatusCode),
		logger.String("success", fmt.Sprintf("%v", apiResp.Success)),
	)

	return &apiResp, nil
}

// DoWithContext executes an HTTP request with context.
func (c *Client) DoWithContext(ctx context.Context, req *http.Request) (*Response, error) {
	req = req.WithContext(ctx)
	return c.Do(req)
}

// Get executes a GET request.
func (c *Client) Get(path string) (*Response, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.Do(req)
}

// GetWithContext executes a GET request with context.
func (c *Client) GetWithContext(ctx context.Context, path string) (*Response, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.Do(req)
}

// Post executes a POST request.
func (c *Client) Post(path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// PostWithContext executes a POST request with context.
func (c *Client) PostWithContext(ctx context.Context, path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// Put executes a PUT request.
func (c *Client) Put(path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// PutWithContext executes a PUT request with context.
func (c *Client) PutWithContext(ctx context.Context, path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// Delete executes a DELETE request.
func (c *Client) Delete(path string) (*Response, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.Do(req)
}

// DeleteWithContext executes a DELETE request with context.
func (c *Client) DeleteWithContext(ctx context.Context, path string) (*Response, error) {
	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.Do(req)
}

// Patch executes a PATCH request.
func (c *Client) Patch(path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// PatchWithContext executes a PATCH request with context.
func (c *Client) PatchWithContext(ctx context.Context, path string, data interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// GetData executes a GET request and unmarshals the response data into dest.
func (c *Client) GetData(path string, dest interface{}) error {
	resp, err := c.Get(path)
	if err != nil {
		return err
	}
	if !resp.Success {
		if resp.Error != nil {
			return fmt.Errorf("request failed: %s", resp.Error.Message)
		}
		return fmt.Errorf("request failed")
	}
	return jsonMarshal(resp.Data, dest)
}

// PostData executes a POST request and unmarshals the response data into dest.
func (c *Client) PostData(path string, reqData, respData interface{}) error {
	resp, err := c.Post(path, reqData)
	if err != nil {
		return err
	}
	if !resp.Success {
		if resp.Error != nil {
			return fmt.Errorf("request failed: %s", resp.Error.Message)
		}
		return fmt.Errorf("request failed")
	}
	return jsonMarshal(resp.Data, respData)
}

// jsonMarshal is a helper to marshal data into dest.
func jsonMarshal(data, dest interface{}) error {
	// Convert data to JSON then unmarshal into dest
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}
	return json.Unmarshal(jsonBytes, dest)
}

// SetBaseURL updates the client's base URL.
func (c *Client) SetBaseURL(baseURL string) {
	c.config.BaseURL = baseURL
}

// SetToken updates the client's authentication token.
func (c *Client) SetToken(token string) {
	c.config.Token = token
}

// SetTimeout updates the client's timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.httpClient.Timeout = timeout
}
