package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
)

// TestSearchWorkflow tests complete search workflows
func TestSearchWorkflow(t *testing.T) {
	// Create temp directory with realistic project structure
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create realistic project structure
	projectFiles := map[string]string{
		"main.go":                  "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		"util.go":                  "package main\n\n// TODO: optimize this\nfunc helper() string {\n\treturn \"data\"\n}",
		"src/app.go":               "package app\n\n// FIXME: handle errors\nfunc Run() error {\n\treturn nil\n}",
		"src/config.go":            "package app\n\ntype Config struct {\n\tName string\n}",
		"docs/README.md":           "# Project\n\nTODO: write documentation",
		"docs/API.md":              "# API Reference\n\n## Endpoints",
		"test/unit_test.go":        "package test\n\n// TODO: add more tests\nfunc TestSample() {}",
		"test/integration_test.go": "package test\n\nfunc TestIntegration() {}",
		".gitignore":               "*.log\nbin/",
		"README.md":                "# Main README\n\nSee docs/ for more info",
	}

	for path, content := range projectFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Register tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewFileSearchTool()); err != nil {
		t.Fatalf("Failed to register file_search: %v", err)
	}
	if err := registry.Register(tools.NewGrepSearchTool()); err != nil {
		t.Fatalf("Failed to register grep_search: %v", err)
	}

	ctx := context.Background()

	t.Run("Find and analyze TODO comments", func(t *testing.T) {
		// Workflow: Find all Go files, then search for TODOs in them

		// Step 1: Find all Go files
		fileSearchResult, err := registry.Execute(ctx, "file_search",
			map[string]interface{}{
				"pattern": "**/*.go",
			}, tmpDir)
		if err != nil {
			t.Fatalf("file_search failed: %v", err)
		}

		if !fileSearchResult.Success {
			t.Fatalf("file_search unsuccessful: %s", fileSearchResult.Output)
		}

		matches, ok := fileSearchResult.Metadata["matches"].(int)
		if !ok || matches < 4 {
			t.Errorf("Expected at least 4 Go files, got %d", matches)
		}

		// Step 2: Search for TODOs in Go files
		grepResult, err := registry.Execute(ctx, "grep_search",
			map[string]interface{}{
				"pattern":      "TODO",
				"file_pattern": "*.go", // Match all .go files (will search recursively)
			}, tmpDir)
		if err != nil {
			t.Fatalf("grep_search failed: %v", err)
		}

		if !grepResult.Success {
			t.Fatalf("grep_search unsuccessful: %s", grepResult.Output)
		}

		todoMatches, ok := grepResult.Metadata["total_matches"].(int)
		if !ok || todoMatches < 2 {
			t.Errorf("Expected at least 2 TODO comments, got %d. Output: %s",
				todoMatches, grepResult.Output)
		}

		// Note: file_pattern "*.go" with WalkDir will match .go files at any level
		// Verify we found TODOs
		if !strings.Contains(grepResult.Output, "TODO") {
			t.Errorf("Expected TODO in search results")
		}
	})

	t.Run("Find markdown documentation", func(t *testing.T) {
		// Workflow: Find all markdown files (use pattern that matches root too)

		result, err := registry.Execute(ctx, "file_search",
			map[string]interface{}{
				"pattern": "*.md", // This will match all .md recursively via WalkDir
			}, tmpDir)
		if err != nil {
			t.Fatalf("file_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("file_search unsuccessful: %s", result.Output)
		}

		matches, ok := result.Metadata["matches"].(int)
		if !ok || matches < 2 {
			t.Errorf("Expected at least 2 markdown files, got %d. Output: %s",
				matches, result.Output)
		}
	})

	t.Run("Search for error handling patterns", func(t *testing.T) {
		// Workflow: Use regex to find error handling

		result, err := registry.Execute(ctx, "grep_search",
			map[string]interface{}{
				"pattern":      "FIXME|TODO|BUG",
				"is_regex":     true,
				"file_pattern": "*.go",
			}, tmpDir)
		if err != nil {
			t.Fatalf("grep_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("grep_search unsuccessful: %s", result.Output)
		}

		matches, ok := result.Metadata["total_matches"].(int)
		if !ok || matches < 2 {
			t.Errorf("Expected at least 2 matches for TODO|FIXME, got %d", matches)
		}

		// Should find both TODO and FIXME
		if !strings.Contains(result.Output, "TODO") && !strings.Contains(result.Output, "FIXME") {
			t.Errorf("Expected TODO or FIXME in results: %s", result.Output)
		}
	})

	t.Run("Limit search results", func(t *testing.T) {
		// Workflow: Test max_results limiting

		result, err := registry.Execute(ctx, "grep_search",
			map[string]interface{}{
				"pattern":     "package|import|func",
				"is_regex":    true,
				"max_results": 3.0,
			}, tmpDir)
		if err != nil {
			t.Fatalf("grep_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("grep_search unsuccessful: %s", result.Output)
		}

		matches, ok := result.Metadata["total_matches"].(int)
		if !ok {
			t.Error("Missing total_matches metadata")
		}

		// Should be limited to 3
		if matches > 3 {
			t.Errorf("Expected max 3 matches due to limit, got %d", matches)
		}
	})

	t.Run("Search specific subdirectory", func(t *testing.T) {
		// Workflow: Search only in docs/ directory

		result, err := registry.Execute(ctx, "file_search",
			map[string]interface{}{
				"pattern":   "*.md",
				"base_path": "docs",
			}, tmpDir)
		if err != nil {
			t.Fatalf("file_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("file_search unsuccessful: %s", result.Output)
		}

		matches, ok := result.Metadata["matches"].(int)
		if !ok || matches != 2 {
			t.Errorf("Expected exactly 2 markdown files in docs/, got %d. Output: %s",
				matches, result.Output)
		}

		// Should NOT include root README.md
		if strings.Contains(result.Output, "README.md") && !strings.Contains(result.Output, "docs") {
			t.Errorf("Should not include root README.md, only docs/*.md")
		}
	})
}

// TestSearchPerformance tests search with many files
func TestSearchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create temp directory with many files
	tmpDir, err := os.MkdirTemp("", "wink-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create 100 files
	for i := 0; i < 100; i++ {
		path := filepath.Join(tmpDir, "src", fmt.Sprintf("file%03d.go", i))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		content := fmt.Sprintf("package src\n\n// TODO: implement feature %d\nfunc Func%d() {}", i, i)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Register tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewFileSearchTool()); err != nil {
		t.Fatalf("Failed to register file_search: %v", err)
	}
	if err := registry.Register(tools.NewGrepSearchTool()); err != nil {
		t.Fatalf("Failed to register grep_search: %v", err)
	}

	ctx := context.Background()

	t.Run("file_search with many files", func(t *testing.T) {
		result, err := registry.Execute(ctx, "file_search",
			map[string]interface{}{
				"pattern": "**/*.go",
			}, tmpDir)
		if err != nil {
			t.Fatalf("file_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("file_search unsuccessful: %s", result.Output)
		}

		// Should complete in reasonable time
		if result.ExecutionTimeMs > 5000 {
			t.Errorf("Search took too long: %dms", result.ExecutionTimeMs)
		}

		matches, ok := result.Metadata["matches"].(int)
		if !ok || matches != 100 {
			t.Errorf("Expected 100 matches, got %d", matches)
		}
	})

	t.Run("grep_search with many files", func(t *testing.T) {
		result, err := registry.Execute(ctx, "grep_search",
			map[string]interface{}{
				"pattern":     "TODO",
				"max_results": 100.0,
			}, tmpDir)
		if err != nil {
			t.Fatalf("grep_search failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("grep_search unsuccessful: %s", result.Output)
		}

		// Should complete in reasonable time
		if result.ExecutionTimeMs > 5000 {
			t.Errorf("Search took too long: %dms", result.ExecutionTimeMs)
		}

		matches, ok := result.Metadata["total_matches"].(int)
		if !ok || matches != 100 {
			t.Errorf("Expected 100 TODO matches, got %d", matches)
		}
	})
}
