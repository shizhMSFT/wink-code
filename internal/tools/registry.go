// Package tools provides tool registration and dispatch
package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/shizhMSFT/wink-code/pkg/types"
)

// Registry manages available tools
type Registry struct {
	tools map[string]types.Tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]types.Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool types.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (types.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	return tool, nil
}

// GetAll returns all registered tools
func (r *Registry) GetAll() []types.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]types.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Execute executes a tool by name with given parameters
func (r *Registry) Execute(ctx context.Context, toolName string, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	// Get tool
	tool, err := r.Get(toolName)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := tool.Validate(params, workingDir); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Execute tool
	result, err := tool.Execute(ctx, params, workingDir)
	if err != nil {
		// Return result even on error (it may contain useful metadata)
		return result, fmt.Errorf("execution failed: %w", err)
	}

	return result, nil
}

// List returns names of all registered tools
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}
