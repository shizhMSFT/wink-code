// Package tools provides security utilities for tool operations
package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath ensures a path is within the working directory (security jail)
func ValidatePath(workingDir, requestedPath string) error {
	// Resolve to absolute paths
	absRequested, err := filepath.Abs(requestedPath)
	if err != nil {
		return fmt.Errorf("invalid path '%s': %w", requestedPath, err)
	}

	absWorking, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("invalid working directory '%s': %w", workingDir, err)
	}

	// Clean paths to resolve . and ..
	absRequested = filepath.Clean(absRequested)
	absWorking = filepath.Clean(absWorking)

	// Check if requested path is within working directory
	relPath, err := filepath.Rel(absWorking, absRequested)
	if err != nil {
		return fmt.Errorf("path '%s' is outside working directory", requestedPath)
	}

	// Check for path traversal attempts (..)
	if strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, string(filepath.Separator)) {
		return fmt.Errorf("path '%s' is outside working directory (resolved to '%s')",
			requestedPath, absRequested)
	}

	return nil
}

// SanitizePathForDisplay converts an absolute path to a relative path for display
// This prevents leaking full system paths in logs
func SanitizePathForDisplay(workingDir, absPath string) string {
	relPath, err := filepath.Rel(workingDir, absPath)
	if err != nil {
		// If we can't make it relative, just show the base name
		return filepath.Base(absPath)
	}
	return relPath
}

// ResolvePath resolves a path relative to working directory
func ResolvePath(workingDir, requestedPath string) (string, error) {
	// If already absolute, validate it
	if filepath.IsAbs(requestedPath) {
		if err := ValidatePath(workingDir, requestedPath); err != nil {
			return "", err
		}
		return filepath.Clean(requestedPath), nil
	}

	// Join with working directory
	absPath := filepath.Join(workingDir, requestedPath)

	// Validate
	if err := ValidatePath(workingDir, absPath); err != nil {
		return "", err
	}

	return filepath.Clean(absPath), nil
}
