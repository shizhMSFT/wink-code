//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shizhMSFT/wink-code/internal/agent"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/internal/tools"
)

// BenchmarkStartupTime measures CLI startup time (Constitution: ≤500ms)
func BenchmarkStartupTime(b *testing.B) {
	logging.InitLogger(false) // No debug for benchmarks

	ollamaURL := os.Getenv("WINK_OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Create agent (simulate CLI startup)
		agentInstance, err := agent.NewAgent(ollamaURL, "qwen3:8b", 30)
		if err != nil {
			b.Fatalf("Failed to create agent: %v", err)
		}

		// Register tools
		createFile := tools.NewCreateFileTool()
		_ = agentInstance.RegisterTool(createFile)

		elapsed := time.Since(start)

		// Verify startup time meets constitution requirement
		if elapsed > 500*time.Millisecond {
			b.Logf("Warning: Startup time %v exceeds 500ms target", elapsed)
		}
	}
}

// BenchmarkToolExecutionOverhead measures tool execution overhead (Constitution: <100ms)
func BenchmarkToolExecutionOverhead(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "wink-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createFile := tools.NewCreateFileTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    "bench.txt",
		"content": "benchmark test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clean up previous file
		testFile := tempDir + "/bench.txt"
		os.Remove(testFile)

		start := time.Now()

		// Execute tool (excluding LLM time)
		_, err := createFile.Execute(ctx, params, tempDir)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}

		elapsed := time.Since(start)

		// Verify execution overhead meets constitution requirement
		if elapsed > 100*time.Millisecond {
			b.Logf("Warning: Tool overhead %v exceeds 100ms target", elapsed)
		}
	}
}

// TestStartupPerformance validates startup time requirement
func TestStartupPerformance(t *testing.T) {
	logging.InitLogger(false)

	ollamaURL := os.Getenv("WINK_OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	start := time.Now()

	agentInstance, err := agent.NewAgent(ollamaURL, "qwen3:8b", 30)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	createFile := tools.NewCreateFileTool()
	_ = agentInstance.RegisterTool(createFile)

	elapsed := time.Since(start)

	t.Logf("Startup time: %v", elapsed)

	// Constitution requirement: ≤500ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("Startup time %v exceeds constitution requirement of 500ms", elapsed)
	}
}

// TestToolExecutionPerformance validates tool overhead requirement
func TestToolExecutionPerformance(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createFile := tools.NewCreateFileTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"path":    "perf-test.txt",
		"content": "performance test content",
	}

	start := time.Now()
	_, err = createFile.Execute(ctx, params, tempDir)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	elapsed := time.Since(start)

	t.Logf("Tool execution time: %v", elapsed)

	// Constitution requirement: <100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Tool execution time %v exceeds constitution requirement of 100ms", elapsed)
	}
}
