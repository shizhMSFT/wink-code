package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shizhMSFT/wink-code/internal/agent"
	"github.com/shizhMSFT/wink-code/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionPersistence validates that sessions are properly saved to disk
func TestSessionPersistence(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".wink", "sessions")
	err := os.MkdirAll(sessionsDir, 0755)
	require.NoError(t, err)

	// Create a new session
	session := &types.Session{
		ID:         "test-session-123",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
		Messages: []types.Message{
			{
				Role:      "user",
				Content:   "Create a test file",
				Timestamp: time.Now(),
			},
			{
				Role:      "assistant",
				Content:   "I'll create the test file for you.",
				Timestamp: time.Now(),
			},
		},
		ToolResults: []types.ToolResult{
			{
				ToolCallID:      "call-1",
				Success:         true,
				Output:          "Created file: test.txt",
				ExecutionTimeMs: 5,
				FilesAffected:   []string{"test.txt"},
			},
		},
	}

	// Save session
	sessionPath := filepath.Join(sessionsDir, session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(sessionPath, data, 0644)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, sessionPath)

	// Load session back
	loadedData, err := os.ReadFile(sessionPath)
	require.NoError(t, err)

	var loadedSession types.Session
	err = json.Unmarshal(loadedData, &loadedSession)
	require.NoError(t, err)

	// Verify session data
	assert.Equal(t, session.ID, loadedSession.ID)
	assert.Equal(t, session.WorkingDir, loadedSession.WorkingDir)
	assert.Equal(t, session.Model, loadedSession.Model)
	assert.Equal(t, len(session.Messages), len(loadedSession.Messages))
	assert.Equal(t, len(session.ToolResults), len(loadedSession.ToolResults))
	assert.Equal(t, types.SessionStatusActive, loadedSession.Status)
}

// TestSessionContinuation validates the full workflow:
// 1. Create a session with some messages
// 2. Save it
// 3. Load it back
// 4. Add more messages
// 5. Verify context is preserved
func TestSessionContinuation(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".wink", "sessions")
	err := os.MkdirAll(sessionsDir, 0755)
	require.NoError(t, err)

	// Original WINK_HOME
	originalHome := os.Getenv("WINK_HOME")
	defer os.Setenv("WINK_HOME", originalHome)
	os.Setenv("WINK_HOME", tempDir)

	// Phase 1: Create initial session
	t.Run("Create initial session", func(t *testing.T) {
		session := &types.Session{
			ID:         "continuation-test-session",
			WorkingDir: tempDir,
			Model:      "qwen3:8b",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Status:     types.SessionStatusActive,
			Messages: []types.Message{
				{
					Role:      "user",
					Content:   "What files are in this directory?",
					Timestamp: time.Now(),
				},
				{
					Role:      "assistant",
					Content:   "Let me check the directory for you.",
					Timestamp: time.Now(),
				},
			},
		}

		// Save session using direct file I/O since we're testing persistence
		sessionPath := filepath.Join(sessionsDir, session.ID+".json")
		data, err := json.MarshalIndent(session, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(sessionPath, data, 0644)
		require.NoError(t, err)

		// Verify session file exists
		assert.FileExists(t, sessionPath)
	})

	// Phase 2: Load and continue session
	t.Run("Load and continue session", func(t *testing.T) {
		// Load the session
		sessionPath := filepath.Join(sessionsDir, "continuation-test-session.json")
		data, err := os.ReadFile(sessionPath)
		require.NoError(t, err)

		var session types.Session
		err = json.Unmarshal(data, &session)
		require.NoError(t, err)

		// Verify loaded session has original messages
		assert.Equal(t, "continuation-test-session", session.ID)
		assert.Equal(t, 2, len(session.Messages))
		assert.Equal(t, "What files are in this directory?", session.Messages[0].Content)

		// Add new messages to continued session
		session.Messages = append(session.Messages, types.Message{
			Role:      "user",
			Content:   "Now create a README.md file",
			Timestamp: time.Now(),
		})

		session.Messages = append(session.Messages, types.Message{
			Role:      "assistant",
			Content:   "I'll create the README.md file for you.",
			Timestamp: time.Now(),
		})

		session.UpdatedAt = time.Now()

		// Save updated session
		data, err = json.MarshalIndent(&session, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(sessionPath, data, 0644)
		require.NoError(t, err)
	})

	// Phase 3: Verify continued session persists correctly
	t.Run("Verify continued session persistence", func(t *testing.T) {
		// Load session again
		sessionPath := filepath.Join(sessionsDir, "continuation-test-session.json")
		data, err := os.ReadFile(sessionPath)
		require.NoError(t, err)

		var session types.Session
		err = json.Unmarshal(data, &session)
		require.NoError(t, err)

		// Verify all messages are present (original + continued)
		assert.Equal(t, 4, len(session.Messages))
		assert.Equal(t, "What files are in this directory?", session.Messages[0].Content)
		assert.Equal(t, "Now create a README.md file", session.Messages[2].Content)
	})
}

// TestSessionContextPruning validates that sessions maintain the last N messages
func TestSessionContextPruning(t *testing.T) {
	tempDir := t.TempDir()

	session := &types.Session{
		ID:         "pruning-test-session",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
		Messages:   []types.Message{},
	}

	// Add 110 messages (exceeding the 100 message limit)
	for i := 0; i < 110; i++ {
		msg := types.Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		session.Messages = append(session.Messages, msg)
	}

	// Create context manager
	contextMgr := agent.NewContextManager(100) // 100 message limit

	// Prune messages
	contextMgr.PruneMessages(session)

	// Verify only last 100 messages remain
	assert.Equal(t, 100, len(session.Messages))

	// Verify the most recent messages are kept
	// Last message should be "Message 109"
	assert.Equal(t, "Message 109", session.Messages[len(session.Messages)-1].Content)
	// First message after pruning should be "Message 10"
	assert.Equal(t, "Message 10", session.Messages[0].Content)
}

// TestLatestSessionRetrieval validates finding the most recent session
func TestLatestSessionRetrieval(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".wink", "sessions")
	err := os.MkdirAll(sessionsDir, 0755)
	require.NoError(t, err)

	// Set HOME to temp dir so SessionManager uses our test directory
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		}
	}()

	// Windows uses USERPROFILE, Unix uses HOME
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)

	// Create multiple sessions with different timestamps
	// Note: file modification time will determine order, not UpdatedAt field
	time.Sleep(10 * time.Millisecond) // Ensure different mod times

	session1 := &types.Session{
		ID:         "session-1",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
	}
	data, err := json.MarshalIndent(session1, "", "  ")
	require.NoError(t, err)
	sessionPath1 := filepath.Join(sessionsDir, "session-1.json")
	err = os.WriteFile(sessionPath1, data, 0644)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	session2 := &types.Session{
		ID:         "session-2",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
	}
	data, err = json.MarshalIndent(session2, "", "  ")
	require.NoError(t, err)
	sessionPath2 := filepath.Join(sessionsDir, "session-2.json")
	err = os.WriteFile(sessionPath2, data, 0644)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	session3 := &types.Session{
		ID:         "session-3",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
	}
	data, err = json.MarshalIndent(session3, "", "  ")
	require.NoError(t, err)
	sessionPath3 := filepath.Join(sessionsDir, "session-3.json")
	err = os.WriteFile(sessionPath3, data, 0644)
	require.NoError(t, err)

	// Create session manager (will use HOME/.wink/sessions)
	sessionMgr, err := agent.NewSessionManager()
	require.NoError(t, err)

	// Find latest session (should be session-3 as it was created last)
	latestSession, err := sessionMgr.GetLatest()
	require.NoError(t, err)
	require.NotNil(t, latestSession)

	// Verify latest session is session-3 (most recent file modification time)
	assert.Equal(t, "session-3", latestSession.ID)
}

// TestSessionStatusTransitions validates state changes
func TestSessionStatusTransitions(t *testing.T) {
	tempDir := t.TempDir()
	sessionsDir := filepath.Join(tempDir, ".wink", "sessions")
	err := os.MkdirAll(sessionsDir, 0755)
	require.NoError(t, err)

	// Create session
	session := &types.Session{
		ID:         "status-test-session",
		WorkingDir: tempDir,
		Model:      "qwen3:8b",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     types.SessionStatusActive,
	}

	sessionPath := filepath.Join(sessionsDir, session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(sessionPath, data, 0644)
	require.NoError(t, err)

	// Test status transitions
	testCases := []struct {
		name      string
		newStatus types.SessionStatus
	}{
		{"Active to Paused", types.SessionStatusPaused},
		{"Paused to Active", types.SessionStatusActive},
		{"Active to Completed", types.SessionStatusCompleted},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load session
			loadedData, err := os.ReadFile(sessionPath)
			require.NoError(t, err)

			var loadedSession types.Session
			err = json.Unmarshal(loadedData, &loadedSession)
			require.NoError(t, err)

			// Update status
			loadedSession.Status = tc.newStatus
			loadedSession.UpdatedAt = time.Now()

			// Save
			data, err := json.MarshalIndent(&loadedSession, "", "  ")
			require.NoError(t, err)
			err = os.WriteFile(sessionPath, data, 0644)
			require.NoError(t, err)

			// Reload and verify
			reloadedData, err := os.ReadFile(sessionPath)
			require.NoError(t, err)

			var reloadedSession types.Session
			err = json.Unmarshal(reloadedData, &reloadedSession)
			require.NoError(t, err)
			assert.Equal(t, tc.newStatus, reloadedSession.Status)
		})
	}
}
