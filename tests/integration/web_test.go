package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
)

// TestWebWorkflow tests complete web fetching workflows
func TestWebWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test-page":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h1>Test Page</h1><p>This is a test page content.</p></body></html>")

		case "/json-api":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status": "success", "data": {"message": "Hello World"}}`)

		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not Found")

		case "/large-content":
			w.WriteHeader(http.StatusOK)
			// Send 2MB of data (should be truncated)
			data := strings.Repeat("A", 2*1024*1024)
			fmt.Fprint(w, data)

		case "/robots.txt":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "User-agent: *\nDisallow: /admin\n")

		case "/admin":
			// This should be blocked by robots.txt
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Admin page")

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Register tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewFetchWebpageTool()); err != nil {
		t.Fatalf("Failed to register fetch_webpage: %v", err)
	}

	ctx := context.Background()

	t.Run("Fetch HTML page successfully", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/test-page",
			}, tmpDir)
		if err != nil {
			t.Fatalf("fetch_webpage failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Request failed: %s", result.Output)
		}

		// Check output contains expected content
		if !strings.Contains(result.Output, "Test Page") {
			t.Errorf("Expected 'Test Page' in output: %s", result.Output)
		}

		// Check metadata
		statusCode, ok := result.Metadata["status_code"].(int)
		if !ok || statusCode != 200 {
			t.Errorf("Expected status code 200, got %v", statusCode)
		}

		contentType, ok := result.Metadata["content_type"].(string)
		if !ok || !strings.Contains(contentType, "text/html") {
			t.Errorf("Expected content type text/html, got %v", contentType)
		}
	})

	t.Run("Fetch JSON API successfully", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/json-api",
			}, tmpDir)
		if err != nil {
			t.Fatalf("fetch_webpage failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Request failed: %s", result.Output)
		}

		// Check output contains JSON content
		if !strings.Contains(result.Output, "Hello World") {
			t.Errorf("Expected JSON content in output: %s", result.Output)
		}

		// Note: httptest sets default content type, so we skip strict content-type check
	})

	t.Run("Handle 404 error", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/not-found",
			}, tmpDir)

		// Should return error for 404
		if err == nil {
			t.Error("Expected error for 404")
		}

		if result == nil {
			t.Fatal("Result should not be nil even on error")
		}

		if result.Success {
			t.Error("Expected failure for 404")
		}

		// Check status code in metadata
		if result.Metadata != nil {
			statusCode, ok := result.Metadata["status_code"].(int)
			if !ok || statusCode != 404 {
				t.Errorf("Expected status code 404, got %v", statusCode)
			}
		}
	})

	t.Run("Handle large content (truncation)", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/large-content",
			}, tmpDir)
		if err != nil {
			t.Fatalf("fetch_webpage failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Request failed: %s", result.Output)
		}

		// Check truncation flag
		truncated, ok := result.Metadata["truncated"].(bool)
		if !ok || !truncated {
			t.Error("Expected content to be truncated")
		}

		// Content should be limited to 1MB
		contentLength, ok := result.Metadata["content_length"].(int)
		if !ok {
			t.Error("Missing content_length in metadata")
		}
		if contentLength > 1024*1024 {
			t.Errorf("Content should be truncated to 1MB, got %d bytes", contentLength)
		}

		// Output should mention truncation
		if !strings.Contains(result.Output, "truncated") {
			t.Error("Output should mention content was truncated")
		}
	})

	t.Run("Robots.txt allows access", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/test-page",
			}, tmpDir)
		if err != nil {
			t.Fatalf("fetch_webpage failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Request should succeed for allowed path: %s", result.Output)
		}
	})

	t.Run("Robots.txt blocks access to admin", func(t *testing.T) {
		result, err := registry.Execute(ctx, "fetch_webpage",
			map[string]interface{}{
				"url": mockServer.URL + "/admin",
			}, tmpDir)

		// Should be blocked by robots.txt
		if err == nil || result.Success {
			t.Error("Expected error due to robots.txt disallow")
		}

		if !strings.Contains(result.Output, "robots.txt") && !strings.Contains(result.Error, "robots.txt") {
			t.Errorf("Expected robots.txt message in output or error: %s / %s",
				result.Output, result.Error)
		}
	})
}

// TestWebTimeout tests timeout handling
func TestWebTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create slow mock server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than timeout
		<-r.Context().Done()
		w.WriteHeader(http.StatusRequestTimeout)
	}))
	defer slowServer.Close()

	tool := tools.NewFetchWebpageTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url":             slowServer.URL + "/slow",
		"timeout_seconds": 1.0,
	}, tmpDir)

	// Should timeout
	if err == nil {
		t.Error("Expected timeout error")
	}

	if result.Success {
		t.Error("Expected failure due to timeout")
	}

	if !strings.Contains(result.Output, "timeout") && !strings.Contains(result.Error, "timeout") {
		t.Errorf("Expected timeout message: %s / %s", result.Output, result.Error)
	}
}

// TestWebSafety tests safety features
func TestWebSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewFetchWebpageTool()

	t.Run("Requires approval", func(t *testing.T) {
		if !tool.RequiresApproval() {
			t.Error("fetch_webpage should require approval")
		}

		if tool.RiskLevel() != "dangerous" {
			t.Errorf("Expected dangerous risk level, got %s", tool.RiskLevel())
		}
	})

	t.Run("Rejects file:// URLs", func(t *testing.T) {
		err := tool.Validate(map[string]interface{}{
			"url": "file:///etc/passwd",
		}, tmpDir)

		if err == nil {
			t.Error("Expected validation error for file:// URL")
		}
	})

	t.Run("Rejects javascript: URLs", func(t *testing.T) {
		err := tool.Validate(map[string]interface{}{
			"url": "javascript:alert(1)",
		}, tmpDir)

		if err == nil {
			t.Error("Expected validation error for javascript: URL")
		}
	})

	t.Run("Rejects FTP URLs", func(t *testing.T) {
		err := tool.Validate(map[string]interface{}{
			"url": "ftp://example.com/file.txt",
		}, tmpDir)

		if err == nil {
			t.Error("Expected validation error for ftp:// URL")
		}
	})
}

// TestWebMetadata tests metadata consistency
func TestWebMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Test content")
	}))
	defer mockServer.Close()

	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewFetchWebpageTool()); err != nil {
		t.Fatalf("Failed to register fetch_webpage: %v", err)
	}

	ctx := context.Background()

	result, err := registry.Execute(ctx, "fetch_webpage",
		map[string]interface{}{
			"url": mockServer.URL + "/test",
		}, tmpDir)
	if err != nil {
		t.Fatalf("fetch_webpage failed: %v", err)
	}

	// Check all expected metadata fields
	expectedFields := []string{"url", "status_code", "content_length", "content_type", "truncated"}
	for _, field := range expectedFields {
		if _, ok := result.Metadata[field]; !ok {
			t.Errorf("Missing metadata field: %s", field)
		}
	}

	// Verify metadata values
	url, _ := result.Metadata["url"].(string)
	if !strings.Contains(url, mockServer.URL) {
		t.Errorf("URL in metadata doesn't match request: %s", url)
	}

	statusCode, _ := result.Metadata["status_code"].(int)
	if statusCode != 200 {
		t.Errorf("Expected status code 200, got %d", statusCode)
	}

	contentLength, _ := result.Metadata["content_length"].(int)
	if contentLength <= 0 {
		t.Errorf("Expected positive content length, got %d", contentLength)
	}
}
