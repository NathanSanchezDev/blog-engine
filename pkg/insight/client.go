package insight

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Go-Insight client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// LogRequest represents a log entry for Go-Insight
type LogRequest struct {
	ServiceName string                 `json:"service_name"`
	LogLevel    string                 `json:"log_level"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MetricRequest represents a metric for Go-Insight
type MetricRequest struct {
	ServiceName string       `json:"service_name"`
	Path        string       `json:"path"`
	Method      string       `json:"method"`
	StatusCode  int          `json:"status_code"`
	Duration    float64      `json:"duration_ms"`
	Source      MetricSource `json:"source"`
	Environment string       `json:"environment,omitempty"`
	RequestID   string       `json:"request_id,omitempty"`
}

type MetricSource struct {
	Language  string `json:"language"`
	Framework string `json:"framework"`
	Version   string `json:"version"`
}

// SendLog sends a log entry to Go-Insight
func (c *Client) SendLog(serviceName, level, message string, metadata map[string]interface{}) error {
	logReq := LogRequest{
		ServiceName: serviceName,
		LogLevel:    level,
		Message:     message,
		Metadata:    metadata,
	}

	return c.sendRequest("POST", "/logs", logReq)
}

// SendLogWithTrace sends a log entry with trace correlation
func (c *Client) SendLogWithTrace(serviceName, level, message, traceID, spanID string, metadata map[string]interface{}) error {
	logReq := LogRequest{
		ServiceName: serviceName,
		LogLevel:    level,
		Message:     message,
		TraceID:     traceID,
		SpanID:      spanID,
		Metadata:    metadata,
	}

	return c.sendRequest("POST", "/logs", logReq)
}

// SendMetric sends a performance metric to Go-Insight
func (c *Client) SendMetric(serviceName, path, method string, statusCode int, duration float64) error {
	metricReq := MetricRequest{
		ServiceName: serviceName,
		Path:        path,
		Method:      method,
		StatusCode:  statusCode,
		Duration:    duration,
		Source: MetricSource{
			Language:  "go",
			Framework: "mux",
			Version:   "1.0.0",
		},
		Environment: "development", // Could come from env var
	}

	return c.sendRequest("POST", "/metrics", metricReq)
}

// sendRequest is the internal method that handles HTTP requests to Go-Insight
func (c *Client) sendRequest(method, endpoint string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return nil
}

// Health checks if Go-Insight is reachable
func (c *Client) Health() error {
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
