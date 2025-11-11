// Package agent handles session persistence
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

const (
	sessionsDir = ".wink/sessions"
)

// SessionManager handles session persistence
type SessionManager struct {
	sessionsPath string
}

// NewSessionManager creates a new session manager
func NewSessionManager() (*SessionManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionsPath := filepath.Join(homeDir, sessionsDir)

	// Create sessions directory if it doesn't exist
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &SessionManager{
		sessionsPath: sessionsPath,
	}, nil
}

// Create creates a new session
func (sm *SessionManager) Create(workingDir, model string) (*types.Session, error) {
	session := &types.Session{
		ID:          uuid.New().String(),
		WorkingDir:  workingDir,
		Model:       model,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages:    []types.Message{},
		ToolResults: []types.ToolResult{},
		Status:      types.SessionStatusActive,
	}

	// Save to disk
	if err := sm.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

// Save persists a session to disk
func (sm *SessionManager) Save(session *types.Session) error {
	session.UpdatedAt = time.Now()

	filePath := filepath.Join(sm.sessionsPath, session.ID+".json")

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load loads a session from disk
func (sm *SessionManager) Load(sessionID string) (*types.Session, error) {
	filePath := filepath.Join(sm.sessionsPath, sessionID+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &session, nil
}

// GetLatest returns the most recently updated session
func (sm *SessionManager) GetLatest() (*types.Session, error) {
	entries, err := os.ReadDir(sm.sessionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Find most recent file
	var latestFile string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = entry.Name()
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no valid session files found")
	}

	// Extract session ID from filename
	sessionID := latestFile[:len(latestFile)-5] // Remove .json extension

	return sm.Load(sessionID)
}

// Delete removes a session file
func (sm *SessionManager) Delete(sessionID string) error {
	filePath := filepath.Join(sm.sessionsPath, sessionID+".json")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// List returns all session IDs
func (sm *SessionManager) List() ([]string, error) {
	entries, err := os.ReadDir(sm.sessionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	sessionIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-5]
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs, nil
}
