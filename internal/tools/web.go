package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// FetchWebpageTool implements fetch_webpage
type FetchWebpageTool struct {
	client *http.Client
}

// NewFetchWebpageTool creates a new fetch_webpage tool
func NewFetchWebpageTool() *FetchWebpageTool {
	return &FetchWebpageTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *FetchWebpageTool) Name() string {
	return "fetch_webpage"
}

func (t *FetchWebpageTool) Description() string {
	return "Fetch content from a web page (must be http or https)"
}

func (t *FetchWebpageTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to fetch (must be http or https)",
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Request timeout in seconds (default: 10)",
				"default":     10,
			},
		},
		"required": []string{"url"},
	}
}

func (t *FetchWebpageTool) Validate(params map[string]interface{}, workingDir string) error {
	// Validate URL
	urlStr, ok := params["url"].(string)
	if !ok || strings.TrimSpace(urlStr) == "" {
		return fmt.Errorf("url parameter is required and must be a non-empty string")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	// Validate host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Validate timeout if provided
	if timeout, ok := params["timeout_seconds"].(float64); ok {
		if timeout < 1 || timeout > 60 {
			return fmt.Errorf("timeout_seconds must be between 1 and 60")
		}
	}

	return nil
}

func (t *FetchWebpageTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	urlStr := strings.TrimSpace(params["url"].(string))
	timeoutSeconds := 10
	if timeout, ok := params["timeout_seconds"].(float64); ok {
		timeoutSeconds = int(timeout)
	}

	logging.Debug("fetch_webpage: url=%s timeout=%d", urlStr, timeoutSeconds)

	// Parse URL for robots.txt check
	parsedURL, _ := url.Parse(urlStr)

	// Check robots.txt
	if !t.checkRobotsTxt(parsedURL) {
		logging.Warn("Robots.txt disallows access", "url", urlStr)
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Access to %s is disallowed by robots.txt", urlStr),
			Error:           "robots.txt disallows access",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("robots.txt disallows access")
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", urlStr, nil)
	if err != nil {
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Failed to create request: %v", err),
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	// Set user agent
	req.Header.Set("User-Agent", "wink-cli/1.0 (AI coding assistant)")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		if reqCtx.Err() == context.DeadlineExceeded {
			return &types.ToolResult{
				Success:         false,
				Output:          fmt.Sprintf("Request timed out after %d seconds", timeoutSeconds),
				Error:           "timeout",
				ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			}, fmt.Errorf("request timed out")
		}
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Failed to fetch: %v", err),
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status),
			Error:           fmt.Sprintf("status code %d", resp.StatusCode),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			Metadata: map[string]interface{}{
				"status_code": resp.StatusCode,
				"url":         urlStr,
			},
		}, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	// Read content with size limit (1MB)
	const maxContentSize = 1024 * 1024 // 1MB
	limitedReader := io.LimitReader(resp.Body, maxContentSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Failed to read response: %v", err),
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	// Check if content was truncated
	truncated := false
	if len(content) > maxContentSize {
		content = content[:maxContentSize]
		truncated = true
	}

	executionTime := time.Since(startTime).Milliseconds()

	// Format output
	contentStr := string(content)
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Fetched content from %s\n", urlStr))
	output.WriteString(fmt.Sprintf("Status: %d %s\n", resp.StatusCode, resp.Status))
	output.WriteString(fmt.Sprintf("Content-Type: %s\n", resp.Header.Get("Content-Type")))
	output.WriteString(fmt.Sprintf("Size: %.2f KB\n", float64(len(content))/1024))
	if truncated {
		output.WriteString("⚠️  Content truncated to 1MB limit\n")
	}
	output.WriteString("\nContent:\n")
	output.WriteString(contentStr)

	logging.Debug("fetch_webpage: status=%d size=%d time=%dms",
		resp.StatusCode, len(content), executionTime)

	return &types.ToolResult{
		Success:         true,
		Output:          output.String(),
		ExecutionTimeMs: executionTime,
		Metadata: map[string]interface{}{
			"url":            urlStr,
			"status_code":    resp.StatusCode,
			"content_length": len(content),
			"content_type":   resp.Header.Get("Content-Type"),
			"truncated":      truncated,
		},
	}, nil
}

func (t *FetchWebpageTool) RequiresApproval() bool {
	return true
}

func (t *FetchWebpageTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelDangerous
}

// checkRobotsTxt checks if the URL is allowed by robots.txt
func (t *FetchWebpageTool) checkRobotsTxt(parsedURL *url.URL) bool {
	// Build robots.txt URL
	robotsURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   "/robots.txt",
	}

	// Try to fetch robots.txt with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL.String(), nil)
	if err != nil {
		// If we can't check, allow access
		return true
	}

	req.Header.Set("User-Agent", "wink-cli/1.0 (AI coding assistant)")

	resp, err := t.client.Do(req)
	if err != nil {
		// If robots.txt doesn't exist or can't be fetched, allow access
		return true
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No robots.txt, allow access
		return true
	}

	if resp.StatusCode != 200 {
		// Can't read robots.txt, allow access
		return true
	}

	// Read robots.txt (limit to 100KB)
	robotsContent, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		// Error reading, allow access
		return true
	}

	// Simple robots.txt parsing (check for Disallow rules)
	// This is a basic implementation - a full parser would be more complex
	lines := strings.Split(string(robotsContent), "\n")
	userAgentMatch := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check User-agent
		if strings.HasPrefix(strings.ToLower(line), "user-agent:") {
			agent := strings.TrimSpace(line[11:])
			if agent == "*" || strings.Contains(strings.ToLower(agent), "wink") {
				userAgentMatch = true
			} else {
				userAgentMatch = false
			}
			continue
		}

		// Check Disallow rules
		if userAgentMatch && strings.HasPrefix(strings.ToLower(line), "disallow:") {
			disallowPath := strings.TrimSpace(line[9:])
			if disallowPath == "/" {
				// Disallow all
				return false
			}
			// Check if current path starts with disallowed path
			if disallowPath != "" && strings.HasPrefix(parsedURL.Path, disallowPath) {
				return false
			}
		}
	}

	// Default: allow
	return true
}
