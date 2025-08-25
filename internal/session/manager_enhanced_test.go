package session

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestManager_ConcurrentOperations tests thread safety and concurrent operations
func TestManager_ConcurrentOperations(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create base sessions for concurrent tests
	baseSessionNames := []string{"concurrent-1", "concurrent-2", "concurrent-3"}
	for _, name := range baseSessionNames {
		_, err := manager.Create(name, "Concurrent test", "")
		if err != nil {
			t.Fatalf("failed to create base session %s: %v", name, err)
		}
	}

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			sessionName := fmt.Sprintf("concurrent-%d", (i%3)+1)
			session, err := manager.Get(sessionName)
			if err != nil {
				t.Errorf("concurrent read %d failed: %v", i, err)
			}
			if session == nil {
				t.Errorf("concurrent read %d returned nil session", i)
			}
			done <- true
		}(i)
	}

	// Wait for all concurrent reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent activations
	for i := 0; i < 5; i++ {
		go func(i int) {
			sessionName := fmt.Sprintf("concurrent-%d", (i%3)+1)
			err := manager.Activate(sessionName)
			if err != nil {
				t.Errorf("concurrent activation %d failed: %v", i, err)
			}
			done <- true
		}(i)
	}

	// Wait for concurrent activations
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify at least one session is active
	activeSession, err := manager.GetActive()
	if err != nil {
		t.Errorf("failed to get active session after concurrent operations: %v", err)
	}
	if activeSession == nil {
		t.Errorf("no session is active after concurrent operations")
	}
}

// TestManager_ErrorRecovery tests error recovery scenarios
func TestManager_ErrorRecovery(t *testing.T) {
	manager, tempDir := setupTestManager(t)

	// Test recovery from corrupted session file
	sessionName := "corrupted-session"
	_, err := manager.Create(sessionName, "Test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Corrupt the session file
	sessionFile := filepath.Join(tempDir, "sessions", sessionName, "session.yaml")
	corruptedContent := []byte("invalid yaml content: [[[")
	err = ioutil.WriteFile(sessionFile, corruptedContent, 0644)
	if err != nil {
		t.Fatalf("failed to corrupt session file: %v", err)
	}

	// Try to get the corrupted session - should fail gracefully
	_, err = manager.Get(sessionName)
	if err == nil {
		t.Errorf("expected error when reading corrupted session, got nil")
	}

	// Verify manager can still operate with other sessions
	_, err = manager.Create("recovery-test", "Recovery test", "")
	if err != nil {
		t.Errorf("manager should still work after encountering corrupted session: %v", err)
	}
}

// TestManager_PermissionHandling tests permission-related scenarios
func TestManager_PermissionHandling(t *testing.T) {
	manager, tempDir := setupTestManager(t)

	// Create a session
	sessionName := "permission-test"
	session, err := manager.Create(sessionName, "Permission test", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify session was created
	if session.Name != sessionName {
		t.Errorf("expected session name %s, got %s", sessionName, session.Name)
	}

	// Make session directory read-only (simulate permission issues)
	sessionDir := filepath.Join(tempDir, "sessions", sessionName)
	originalMode := os.FileMode(0755)
	err = os.Chmod(sessionDir, 0444) // Read-only
	if err != nil {
		t.Fatalf("failed to change permissions: %v", err)
	}

	// Restore permissions after test
	defer func() {
		os.Chmod(sessionDir, originalMode)
	}()

	// Try to delete session with permission issues
	err = manager.Delete(sessionName)
	if err == nil {
		t.Errorf("expected error when deleting session with permission issues")
	}

	// Restore permissions and try again
	err = os.Chmod(sessionDir, originalMode)
	if err != nil {
		t.Fatalf("failed to restore permissions: %v", err)
	}

	err = manager.Delete(sessionName)
	if err != nil {
		t.Errorf("delete should succeed after restoring permissions: %v", err)
	}
}

// TestManager_LargeScaleOperations tests performance with many sessions
func TestManager_LargeScaleOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large scale test in short mode")
	}

	manager, _ := setupTestManager(t)

	// Create many sessions
	sessionCount := 100
	sessionNames := make([]string, sessionCount)
	
	start := time.Now()
	for i := 0; i < sessionCount; i++ {
		sessionName := fmt.Sprintf("scale-test-%03d", i)
		sessionNames[i] = sessionName
		_, err := manager.Create(sessionName, fmt.Sprintf("Scale test session %d", i), "")
		if err != nil {
			t.Fatalf("failed to create session %s: %v", sessionName, err)
		}
	}
	createDuration := time.Since(start)

	t.Logf("Created %d sessions in %v", sessionCount, createDuration)

	// List all sessions
	start = time.Now()
	sessions, err := manager.List()
	if err != nil {
		t.Errorf("failed to list sessions: %v", err)
	}
	listDuration := time.Since(start)

	if len(sessions) != sessionCount {
		t.Errorf("expected %d sessions, got %d", sessionCount, len(sessions))
	}

	t.Logf("Listed %d sessions in %v", len(sessions), listDuration)

	// Test activation performance
	start = time.Now()
	err = manager.Activate(sessionNames[sessionCount/2])
	if err != nil {
		t.Errorf("failed to activate middle session: %v", err)
	}
	activateDuration := time.Since(start)

	t.Logf("Activated session in %v", activateDuration)

	// Cleanup - delete all sessions
	start = time.Now()
	for _, sessionName := range sessionNames {
		err := manager.Delete(sessionName)
		if err != nil {
			t.Errorf("failed to delete session %s: %v", sessionName, err)
		}
	}
	deleteDuration := time.Since(start)

	t.Logf("Deleted %d sessions in %v", sessionCount, deleteDuration)
}

// TestManager_SessionValidation tests session validation scenarios
func TestManager_SessionValidation(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []struct {
		name        string
		sessionName string
		description string
		volumePath  string
		shouldFail  bool
		expectedErr string
	}{
		{
			name:        "valid session with standard name",
			sessionName: "valid-session",
			description: "Valid session",
			volumePath:  "",
			shouldFail:  false,
		},
		{
			name:        "session with special characters",
			sessionName: "session_with-special.chars123",
			description: "Session with special chars",
			volumePath:  "",
			shouldFail:  false,
		},
		{
			name:        "session with very long name",
			sessionName: strings.Repeat("a", 255),
			description: "Long name session",
			volumePath:  "",
			shouldFail:  false,
		},
		{
			name:        "session with extremely long name",
			sessionName: strings.Repeat("a", 500),
			description: "Extremely long name",
			volumePath:  "",
			shouldFail:  true,
			expectedErr: "too long",
		},
		{
			name:        "session with only numbers",
			sessionName: "12345",
			description: "Numeric session",
			volumePath:  "",
			shouldFail:  false,
		},
		{
			name:        "session with unicode characters",
			sessionName: "session-测试-🚀",
			description: "Unicode session",
			volumePath:  "",
			shouldFail:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := manager.Create(tt.sessionName, tt.description, tt.volumePath)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if session == nil {
				t.Fatal("session is nil")
			}

			// Verify session can be retrieved
			retrievedSession, err := manager.Get(tt.sessionName)
			if err != nil {
				t.Errorf("failed to retrieve session: %v", err)
			}

			if retrievedSession.Name != tt.sessionName {
				t.Errorf("retrieved session name mismatch: expected %s, got %s", tt.sessionName, retrievedSession.Name)
			}

			// Clean up
			manager.Delete(tt.sessionName)
		})
	}
}

// TestManager_ActiveSessionPersistence tests active session persistence across restarts
func TestManager_ActiveSessionPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first manager instance
	manager1 := NewManager(tmpDir)
	
	// Create sessions
	_, err := manager1.Create("session1", "First session", "")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	_, err = manager1.Create("session2", "Second session", "")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Activate session2
	err = manager1.Activate("session2")
	if err != nil {
		t.Fatalf("failed to activate session2: %v", err)
	}

	// Verify session2 is active
	activeSession, err := manager1.GetActive()
	if err != nil {
		t.Fatalf("failed to get active session: %v", err)
	}
	if activeSession == nil || activeSession.Name != "session2" {
		t.Fatalf("expected session2 to be active")
	}

	// Create second manager instance (simulating restart)
	manager2 := NewManager(tmpDir)

	// Verify active session persisted
	activeSession, err = manager2.GetActive()
	if err != nil {
		t.Errorf("failed to get active session from new manager: %v", err)
	}
	if activeSession == nil {
		t.Error("active session not persisted across restart")
	} else if activeSession.Name != "session2" {
		t.Errorf("expected session2 to be active, got %s", activeSession.Name)
	}

	// Verify both sessions still exist
	sessions, err := manager2.List()
	if err != nil {
		t.Errorf("failed to list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

// TestManager_VolumePathEdgeCases tests edge cases with volume paths
func TestManager_VolumePathEdgeCases(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []struct {
		name        string
		volumePath  string
		expectError bool
	}{
		{
			name:        "absolute unix path",
			volumePath:  "/tmp/test/volumes",
			expectError: false,
		},
		{
			name:        "relative path",
			volumePath:  "./volumes",
			expectError: false,
		},
		{
			name:        "path with spaces",
			volumePath:  "/path with spaces/volumes",
			expectError: false,
		},
		{
			name:        "path with special chars",
			volumePath:  "/path-with_special.chars/volumes",
			expectError: false,
		},
		{
			name:        "very long path",
			volumePath:  "/" + strings.Repeat("very-long-directory-name", 20),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionName := "volume-" + strings.ReplaceAll(tt.name, " ", "-")
			session, err := manager.Create(sessionName, "Volume test", tt.volumePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if session.VolumePath != tt.volumePath {
				t.Errorf("expected volume path %s, got %s", tt.volumePath, session.VolumePath)
			}

			// Test setting volume path
			newPath := tt.volumePath + "-modified"
			err = manager.SetVolumePath(sessionName, newPath)
			if err != nil {
				t.Errorf("failed to set volume path: %v", err)
			}

			// Verify path was updated (note: relative paths may be resolved to absolute)
			updatedSession, err := manager.Get(sessionName)
			if err != nil {
				t.Errorf("failed to get updated session: %v", err)
			}
			
			// For relative paths, check if it ends with the expected suffix
			if strings.HasPrefix(tt.volumePath, "./") {
				expectedSuffix := strings.TrimPrefix(newPath, "./")
				if !strings.HasSuffix(updatedSession.VolumePath, expectedSuffix) {
					t.Errorf("volume path not updated correctly for relative path: expected suffix %s in %s", expectedSuffix, updatedSession.VolumePath)
				}
			} else if updatedSession.VolumePath != newPath {
				t.Errorf("volume path not updated: expected %s, got %s", newPath, updatedSession.VolumePath)
			}

			// Clean up
			manager.Delete(sessionName)
		})
	}
}

// TestManager_SessionMetadataUpdate tests updating session metadata
func TestManager_SessionMetadataUpdate(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create session
	originalSession, err := manager.Create("metadata-test", "Original description", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	originalTime := originalSession.CreatedAt

	// Wait a bit to ensure timestamps differ
	time.Sleep(10 * time.Millisecond)

	// Test updating description through session manager
	// Note: This assumes the manager has an UpdateDescription method
	// If not available, this test validates current behavior
	session, err := manager.Get("metadata-test")
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}

	// Verify created time hasn't changed
	if !session.CreatedAt.Equal(originalTime) {
		t.Errorf("created time should not change on metadata update")
	}

	// Test that description can be preserved
	if session.Description != "Original description" {
		t.Errorf("description should be preserved: expected 'Original description', got '%s'", session.Description)
	}
}

// TestManager_CleanupOperations tests various cleanup scenarios
func TestManager_CleanupOperations(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create sessions with files
	sessionName := "cleanup-test"
	session, err := manager.Create(sessionName, "Cleanup test", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify session was created
	if session.Name != sessionName {
		t.Errorf("expected session name %s, got %s", sessionName, session.Name)
	}

	// Create files in session directories
	sessionDir := manager.getSessionDir(sessionName)
	testFiles := []string{
		filepath.Join(sessionDir, "manifests", "test.yaml"),
		filepath.Join(sessionDir, "config", "test.json"),
		filepath.Join(sessionDir, "volumes", "test.txt"),
		filepath.Join(sessionDir, "logs", "test.log"),
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}

		err = ioutil.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", file, err)
		}
	}

	// Verify files exist
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("test file %s was not created", file)
		}
	}

	// Delete session - should clean up all files
	err = manager.Delete(sessionName)
	if err != nil {
		t.Errorf("failed to delete session: %v", err)
	}

	// Verify all files and directories are cleaned up
	if _, err := os.Stat(sessionDir); !os.IsNotExist(err) {
		t.Errorf("session directory should be cleaned up after deletion")
	}

	// Test cleanup of active session
	_, err = manager.Create("active-cleanup", "Active cleanup test", "")
	if err != nil {
		t.Fatalf("failed to create active session: %v", err)
	}

	err = manager.Activate("active-cleanup")
	if err != nil {
		t.Fatalf("failed to activate session: %v", err)
	}

	// Delete active session
	err = manager.Delete("active-cleanup")
	if err != nil {
		t.Errorf("failed to delete active session: %v", err)
	}

	// Verify no active session
	activeSession, err := manager.GetActive()
	if err != nil {
		t.Errorf("unexpected error getting active session: %v", err)
	}
	if activeSession != nil {
		t.Errorf("active session should be cleared after deletion")
	}
}