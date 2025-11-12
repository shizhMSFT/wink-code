package tools

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// FileSearchTool implements file_search using glob patterns
type FileSearchTool struct{}

// NewFileSearchTool creates a new file_search tool
func NewFileSearchTool() *FileSearchTool {
	return &FileSearchTool{}
}

func (t *FileSearchTool) Name() string {
	return "file_search"
}

func (t *FileSearchTool) Description() string {
	return "Search for files matching a glob pattern (e.g., '**/*.py', 'src/**/*.go')"
}

func (t *FileSearchTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern (e.g., '**/*.py', 'src/**/*.go')",
			},
			"base_path": map[string]interface{}{
				"type":        "string",
				"description": "Base directory to search from (default: current directory)",
				"default":     ".",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *FileSearchTool) Validate(params map[string]interface{}, workingDir string) error {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return fmt.Errorf("pattern parameter is required and must be a non-empty string")
	}

	// Validate glob pattern syntax
	// Check for invalid characters like unclosed brackets
	if strings.Contains(pattern, "[") && !strings.Contains(pattern, "]") {
		return fmt.Errorf("invalid glob pattern '%s': unclosed bracket", pattern)
	}
	// For patterns without **, test with filepath.Match
	if !strings.Contains(pattern, "**") {
		_, err := filepath.Match(pattern, "")
		if err != nil {
			return fmt.Errorf("invalid glob pattern '%s': %v", pattern, err)
		}
	}

	// Validate base_path if provided
	if basePath, ok := params["base_path"].(string); ok && basePath != "" {
		absBase := filepath.Join(workingDir, basePath)
		if err := ValidatePath(workingDir, absBase); err != nil {
			return fmt.Errorf("base_path '%s' is outside working directory", basePath)
		}
	}

	return nil
}

func (t *FileSearchTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	pattern := params["pattern"].(string)
	basePath := "."
	if bp, ok := params["base_path"].(string); ok && bp != "" {
		basePath = bp
	}

	absBase := filepath.Join(workingDir, basePath)
	logging.Debug("file_search: pattern=%s base=%s", pattern, absBase)

	var matches []string
	const maxResults = 1000
	const maxDepth = 20

	err := filepath.WalkDir(absBase, func(path string, d fs.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // Skip errors
		}

		// Check depth
		relPath, _ := filepath.Rel(absBase, path)
		depth := strings.Count(relPath, string(filepath.Separator))
		if depth > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Skip directories for matching
		if d.IsDir() {
			return nil
		}

		// Get relative path for matching
		relToWorking, err := filepath.Rel(workingDir, path)
		if err != nil {
			return nil
		}

		// Try to match the pattern against the relative path
		matched, err := matchGlob(pattern, relToWorking)
		if err != nil {
			logging.Debug("file_search: pattern match error: %v", err)
			return nil
		}

		if matched {
			matches = append(matches, relToWorking)
			if len(matches) >= maxResults {
				return fs.SkipAll
			}
		}

		return nil
	})

	executionTime := time.Since(startTime).Milliseconds()

	if err != nil && err != fs.SkipAll {
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Error during file search: %v", err),
			ExecutionTimeMs: executionTime,
		}, err
	}

	// Format output
	var output strings.Builder
	if len(matches) == 0 {
		output.WriteString(fmt.Sprintf("No files found matching pattern '%s'", pattern))
	} else {
		output.WriteString(fmt.Sprintf("Found %d file(s) matching '%s':\n", len(matches), pattern))
		for _, match := range matches {
			output.WriteString(fmt.Sprintf("  %s\n", match))
		}
		if len(matches) >= maxResults {
			output.WriteString(fmt.Sprintf("\nWarning: Found %d+ matches, showing first %d", maxResults, maxResults))
		}
	}

	logging.Debug("file_search: found %d matches in %dms", len(matches), executionTime)

	return &types.ToolResult{
		Success:         true,
		Output:          output.String(),
		ExecutionTimeMs: executionTime,
		Metadata: map[string]interface{}{
			"matches": len(matches),
			"pattern": pattern,
		},
	}, nil
}

func (t *FileSearchTool) RequiresApproval() bool {
	return true
}

func (t *FileSearchTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelReadOnly
}

// matchGlob handles glob pattern matching with ** support
func matchGlob(pattern, path string) (bool, error) {
	// Normalize paths
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ** (match any depth)
	if strings.Contains(pattern, "**") {
		// Convert ** glob to regex
		regexPattern := "^" + regexp.QuoteMeta(pattern)
		regexPattern = strings.ReplaceAll(regexPattern, "\\*\\*\\/", "(.*/)?")
		regexPattern = strings.ReplaceAll(regexPattern, "\\*\\*", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "\\*", "[^/]*")
		regexPattern = strings.ReplaceAll(regexPattern, "\\?", "[^/]")
		regexPattern += "$"

		matched, err := regexp.MatchString(regexPattern, path)
		return matched, err
	}

	// Standard glob matching
	return filepath.Match(pattern, path)
}

// GrepSearchTool implements grep_search for content searching
type GrepSearchTool struct{}

// NewGrepSearchTool creates a new grep_search tool
func NewGrepSearchTool() *GrepSearchTool {
	return &GrepSearchTool{}
}

func (t *GrepSearchTool) Name() string {
	return "grep_search"
}

func (t *GrepSearchTool) Description() string {
	return "Search file contents for text or regex pattern"
}

func (t *GrepSearchTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Text or regex pattern to search for",
			},
			"is_regex": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether pattern is regex (default: false)",
				"default":     false,
			},
			"file_pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to limit files searched (default: all files)",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 100)",
				"default":     100,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepSearchTool) Validate(params map[string]interface{}, workingDir string) error {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return fmt.Errorf("pattern parameter is required and must be a non-empty string")
	}

	// If regex, validate it
	if isRegex, ok := params["is_regex"].(bool); ok && isRegex {
		_, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	// Validate max_results
	if maxResults, ok := params["max_results"].(float64); ok {
		if maxResults < 1 || maxResults > 1000 {
			return fmt.Errorf("max_results must be between 1 and 1000")
		}
	}

	return nil
}

func (t *GrepSearchTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	pattern := params["pattern"].(string)
	isRegex := false
	if ir, ok := params["is_regex"].(bool); ok {
		isRegex = ir
	}

	filePattern := ""
	if fp, ok := params["file_pattern"].(string); ok {
		filePattern = fp
	}

	maxResults := 100
	if mr, ok := params["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	logging.Debug("grep_search: pattern=%s is_regex=%v file_pattern=%s max=%d", pattern, isRegex, filePattern, maxResults)

	// Compile regex if needed
	var re *regexp.Regexp
	var err error
	if isRegex {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return &types.ToolResult{
				Success: false,
				Output:  fmt.Sprintf("Invalid regex pattern: %v", err),
			}, err
		}
	}

	type match struct {
		file string
		line int
		text string
	}

	var matches []match
	filesSearched := 0

	err = filepath.WalkDir(workingDir, func(path string, d fs.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // Skip errors
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(workingDir, path)
		if err != nil {
			return nil
		}

		// Filter by file pattern if provided
		if filePattern != "" {
			matched, err := matchGlob(filePattern, relPath)
			if err != nil || !matched {
				return nil
			}
		}

		// Skip binary files
		if isBinaryFile(path) {
			return nil
		}

		filesSearched++

		// Search file contents
		file, err := os.Open(path)
		if err != nil {
			return nil // Skip files we can't open
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			matched := false
			if isRegex {
				matched = re.MatchString(line)
			} else {
				matched = strings.Contains(line, pattern)
			}

			if matched {
				matches = append(matches, match{
					file: relPath,
					line: lineNum,
					text: strings.TrimSpace(line),
				})

				if len(matches) >= maxResults {
					return fs.SkipAll
				}
			}
		}

		return nil
	})

	executionTime := time.Since(startTime).Milliseconds()

	if err != nil && err != fs.SkipAll {
		return &types.ToolResult{
			Success:         false,
			Output:          fmt.Sprintf("Error during grep search: %v", err),
			ExecutionTimeMs: executionTime,
		}, err
	}

	// Format output
	var output strings.Builder
	if len(matches) == 0 {
		output.WriteString(fmt.Sprintf("No matches found for '%s'", pattern))
	} else {
		output.WriteString(fmt.Sprintf("Found %d match(es) for '%s':\n", len(matches), pattern))
		for _, m := range matches {
			output.WriteString(fmt.Sprintf("  %s:%d: %s\n", m.file, m.line, m.text))
		}
		if len(matches) >= maxResults {
			output.WriteString(fmt.Sprintf("\nWarning: Reached limit of %d results", maxResults))
		}
	}

	logging.Debug("grep_search: found %d matches in %d files (%dms)", len(matches), filesSearched, executionTime)

	return &types.ToolResult{
		Success:         true,
		Output:          output.String(),
		ExecutionTimeMs: executionTime,
		Metadata: map[string]interface{}{
			"total_matches":  len(matches),
			"files_searched": filesSearched,
			"pattern":        pattern,
		},
	}, nil
}

func (t *GrepSearchTool) RequiresApproval() bool {
	return true
}

func (t *GrepSearchTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelReadOnly
}

// isBinaryFile checks if a file is binary by reading first 512 bytes
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	// Check for null bytes (binary indicator)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}
