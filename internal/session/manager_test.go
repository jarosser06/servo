package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	tmpDir := t.TempDir()
	return NewManager(tmpDir), tmpDir
}

func TestManager_Create(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []struct {
		name        string
		sessionName string
		description string
		volumePath  string
		expectError bool
	}{
		{
			name:        "valid session creation",
			sessionName: "test-session",
			description: "Test session",
			volumePath:  "",
			expectError: false,
		},
		{
			name:        "session with custom volume path",
			sessionName: "custom-volumes",
			description: "Custom volume test",
			volumePath:  "/custom/path",
			expectError: false,
		},
		{
			name:        "empty session name",
			sessionName: "",
			description: "Invalid test",
			volumePath:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := manager.Create(tt.sessionName, tt.description, tt.volumePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if session.Name != tt.sessionName {
				t.Errorf("expected session name %s, got %s", tt.sessionName, session.Name)
			}

			if session.Description != tt.description {
				t.Errorf("expected description %s, got %s", tt.description, session.Description)
			}

			if tt.volumePath != "" {
				if session.VolumePath != tt.volumePath {
					t.Errorf("expected volume path %s, got %s", tt.volumePath, session.VolumePath)
				}
			} else {
				expectedPath := filepath.Join(manager.getSessionDir(tt.sessionName), "volumes")
				if session.VolumePath != expectedPath {
					t.Errorf("expected default volume path %s, got %s", expectedPath, session.VolumePath)
				}
			}

			// Verify session directories were created
			sessionDir := manager.getSessionDir(tt.sessionName)
			if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
				t.Errorf("session directory was not created")
			}

			manifestsDir := filepath.Join(sessionDir, "manifests")
			if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
				t.Errorf("manifests directory was not created")
			}

			volumesDir := filepath.Join(sessionDir, "volumes")
			if _, err := os.Stat(volumesDir); os.IsNotExist(err) {
				t.Errorf("volumes directory was not created")
			}
		})
	}
}

func TestManager_CreateDuplicate(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create first session
	_, err := manager.Create("duplicate-test", "Test session", "")
	if err != nil {
		t.Fatalf("failed to create first session: %v", err)
	}

	// Try to create duplicate
	_, err = manager.Create("duplicate-test", "Another test session", "")
	if err == nil {
		t.Errorf("expected error when creating duplicate session, but got nil")
	}
}

func TestManager_Get(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a session
	originalSession, err := manager.Create("get-test", "Get test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Retrieve the session
	retrievedSession, err := manager.Get("get-test")
	if err != nil {
		t.Errorf("unexpected error getting session: %v", err)
		return
	}

	if retrievedSession.Name != originalSession.Name {
		t.Errorf("expected name %s, got %s", originalSession.Name, retrievedSession.Name)
	}

	if retrievedSession.Description != originalSession.Description {
		t.Errorf("expected description %s, got %s", originalSession.Description, retrievedSession.Description)
	}

	// Test getting non-existent session
	_, err = manager.Get("non-existent")
	if err == nil {
		t.Errorf("expected error when getting non-existent session, but got nil")
	}
}

func TestManager_List(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Initially should be empty
	sessions, err := manager.List()
	if err != nil {
		t.Errorf("unexpected error listing sessions: %v", err)
		return
	}

	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions initially, got %d", len(sessions))
	}

	// Create multiple sessions
	sessionNames := []string{"session1", "session2", "session3"}
	for _, name := range sessionNames {
		_, err := manager.Create(name, "Test session "+name, "")
		if err != nil {
			t.Fatalf("failed to create session %s: %v", name, err)
		}
	}

	// List sessions
	sessions, err = manager.List()
	if err != nil {
		t.Errorf("unexpected error listing sessions: %v", err)
		return
	}

	if len(sessions) != len(sessionNames) {
		t.Errorf("expected %d sessions, got %d", len(sessionNames), len(sessions))
	}

	// Verify all sessions are present
	sessionMap := make(map[string]*Session)
	for _, session := range sessions {
		sessionMap[session.Name] = session
	}

	for _, expectedName := range sessionNames {
		if _, exists := sessionMap[expectedName]; !exists {
			t.Errorf("expected session %s not found in list", expectedName)
		}
	}
}

func TestManager_Delete(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a session
	_, err := manager.Create("delete-test", "Delete test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify it exists
	exists, err := manager.Exists("delete-test")
	if err != nil {
		t.Fatalf("error checking if session exists: %v", err)
	}
	if !exists {
		t.Fatalf("session should exist before deletion")
	}

	// Delete the session
	err = manager.Delete("delete-test")
	if err != nil {
		t.Errorf("unexpected error deleting session: %v", err)
		return
	}

	// Verify it no longer exists
	exists, err = manager.Exists("delete-test")
	if err != nil {
		t.Errorf("error checking if session exists after deletion: %v", err)
		return
	}
	if exists {
		t.Errorf("session should not exist after deletion")
	}

	// Try to delete non-existent session
	err = manager.Delete("non-existent")
	if err == nil {
		t.Errorf("expected error when deleting non-existent session, but got nil")
	}
}

func TestManager_Activate(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create multiple sessions
	sessions := []string{"session1", "session2", "session3"}
	for _, name := range sessions {
		_, err := manager.Create(name, "Test session", "")
		if err != nil {
			t.Fatalf("failed to create session %s: %v", name, err)
		}
	}

	// Activate first session
	err := manager.Activate("session1")
	if err != nil {
		t.Errorf("unexpected error activating session: %v", err)
		return
	}

	// Check active session
	activeSession, err := manager.GetActive()
	if err != nil {
		t.Errorf("unexpected error getting active session: %v", err)
		return
	}

	if activeSession == nil {
		t.Errorf("expected active session, got nil")
		return
	}

	if activeSession.Name != "session1" {
		t.Errorf("expected active session 'session1', got '%s'", activeSession.Name)
	}

	if !activeSession.Active {
		t.Errorf("expected session to be marked as active")
	}

	// Activate different session
	err = manager.Activate("session2")
	if err != nil {
		t.Errorf("unexpected error activating second session: %v", err)
		return
	}

	// Check new active session
	activeSession, err = manager.GetActive()
	if err != nil {
		t.Errorf("unexpected error getting active session: %v", err)
		return
	}

	if activeSession.Name != "session2" {
		t.Errorf("expected active session 'session2', got '%s'", activeSession.Name)
	}

	// Verify first session is no longer active
	session1, err := manager.Get("session1")
	if err != nil {
		t.Errorf("unexpected error getting session1: %v", err)
		return
	}

	if session1.Active {
		t.Errorf("expected session1 to be inactive after activating session2")
	}
}

func TestManager_ClearActive(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create and activate a session
	_, err := manager.Create("clear-test", "Clear test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = manager.Activate("clear-test")
	if err != nil {
		t.Fatalf("failed to activate session: %v", err)
	}

	// Verify it's active
	activeSession, err := manager.GetActive()
	if err != nil {
		t.Fatalf("unexpected error getting active session: %v", err)
	}
	if activeSession == nil || activeSession.Name != "clear-test" {
		t.Fatalf("expected 'clear-test' to be active")
	}

	// Clear active session
	err = manager.ClearActive()
	if err != nil {
		t.Errorf("unexpected error clearing active session: %v", err)
		return
	}

	// Verify no session is active
	activeSession, err = manager.GetActive()
	if err != nil {
		t.Errorf("unexpected error getting active session: %v", err)
		return
	}
	if activeSession != nil {
		t.Errorf("expected no active session, got %s", activeSession.Name)
	}

	// Verify session is marked as inactive
	session, err := manager.Get("clear-test")
	if err != nil {
		t.Errorf("unexpected error getting session: %v", err)
		return
	}
	if session.Active {
		t.Errorf("expected session to be marked as inactive")
	}
}

func TestManager_Clone(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create source session
	_, err := manager.Create("source-session", "Source session", "")
	if err != nil {
		t.Fatalf("failed to create source session: %v", err)
	}

	// Create a config file in source session
	sourceDir := manager.getSessionDir("source-session")
	configContent := []byte("test: config")
	configPath := filepath.Join(sourceDir, "config.yaml")
	err = os.WriteFile(configPath, configContent, 0644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Clone the session
	clonedSession, err := manager.Clone("source-session", "cloned-session", "Cloned session")
	if err != nil {
		t.Errorf("unexpected error cloning session: %v", err)
		return
	}

	if clonedSession.Name != "cloned-session" {
		t.Errorf("expected cloned session name 'cloned-session', got '%s'", clonedSession.Name)
	}

	if clonedSession.Description != "Cloned session" {
		t.Errorf("expected description 'Cloned session', got '%s'", clonedSession.Description)
	}

	// Verify config was copied
	clonedDir := manager.getSessionDir("cloned-session")
	clonedConfigPath := filepath.Join(clonedDir, "config.yaml")
	if _, err := os.Stat(clonedConfigPath); os.IsNotExist(err) {
		t.Errorf("config file was not copied to cloned session")
		return
	}

	clonedContent, err := os.ReadFile(clonedConfigPath)
	if err != nil {
		t.Errorf("failed to read cloned config file: %v", err)
		return
	}

	if string(clonedContent) != string(configContent) {
		t.Errorf("cloned config content does not match original")
	}

	// Test cloning non-existent session
	_, err = manager.Clone("non-existent", "target", "Test")
	if err == nil {
		t.Errorf("expected error when cloning non-existent session, but got nil")
	}

	// Test cloning to existing target
	_, err = manager.Clone("source-session", "cloned-session", "Another description")
	if err == nil {
		t.Errorf("expected error when cloning to existing target, but got nil")
	}
}

func TestManager_AdoptVolumes(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create source and target sessions
	sourceSession, err := manager.Create("volume-source", "Source session", "/custom/source/volumes")
	if err != nil {
		t.Fatalf("failed to create source session: %v", err)
	}

	targetSession, err := manager.Create("volume-target", "Target session", "")
	if err != nil {
		t.Fatalf("failed to create target session: %v", err)
	}

	originalTargetVolumePath := targetSession.VolumePath

	// Adopt volumes
	err = manager.AdoptVolumes("volume-target", "volume-source")
	if err != nil {
		t.Errorf("unexpected error adopting volumes: %v", err)
		return
	}

	// Verify target session volume path changed
	updatedTarget, err := manager.Get("volume-target")
	if err != nil {
		t.Errorf("unexpected error getting updated target session: %v", err)
		return
	}

	if updatedTarget.VolumePath != sourceSession.VolumePath {
		t.Errorf("expected target volume path to be %s, got %s", sourceSession.VolumePath, updatedTarget.VolumePath)
	}

	if updatedTarget.VolumePath == originalTargetVolumePath {
		t.Errorf("target volume path should have changed from original")
	}
}

func TestManager_SetVolumePath(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create session
	_, err := manager.Create("volume-path-test", "Volume path test", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Set custom volume path
	customPath := "/custom/volume/path"
	err = manager.SetVolumePath("volume-path-test", customPath)
	if err != nil {
		t.Errorf("unexpected error setting volume path: %v", err)
		return
	}

	// Verify path was updated
	session, err := manager.Get("volume-path-test")
	if err != nil {
		t.Errorf("unexpected error getting session: %v", err)
		return
	}

	if session.VolumePath != customPath {
		t.Errorf("expected volume path %s, got %s", customPath, session.VolumePath)
	}

	// Reset to default by passing empty string
	err = manager.SetVolumePath("volume-path-test", "")
	if err != nil {
		t.Errorf("unexpected error resetting volume path: %v", err)
		return
	}

	// Verify path was reset to default
	session, err = manager.Get("volume-path-test")
	if err != nil {
		t.Errorf("unexpected error getting session: %v", err)
		return
	}

	expectedDefault := filepath.Join(manager.getSessionDir("volume-path-test"), "volumes")
	if session.VolumePath != expectedDefault {
		t.Errorf("expected default volume path %s, got %s", expectedDefault, session.VolumePath)
	}
}

func TestManager_Exists(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Test non-existent session
	exists, err := manager.Exists("non-existent")
	if err != nil {
		t.Errorf("unexpected error checking existence: %v", err)
		return
	}
	if exists {
		t.Errorf("expected non-existent session to return false")
	}

	// Create session
	_, err = manager.Create("existence-test", "Test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test existing session
	exists, err = manager.Exists("existence-test")
	if err != nil {
		t.Errorf("unexpected error checking existence: %v", err)
		return
	}
	if !exists {
		t.Errorf("expected existing session to return true")
	}

	// Test empty name
	exists, err = manager.Exists("")
	if err != nil {
		t.Errorf("unexpected error checking existence of empty name: %v", err)
		return
	}
	if exists {
		t.Errorf("expected empty name to return false")
	}
}

func TestManager_SessionPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and session
	manager1 := NewManager(tmpDir)
	originalSession, err := manager1.Create("persistence-test", "Persistence test", "/custom/path")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create new manager with same directory (simulating app restart)
	manager2 := NewManager(tmpDir)

	// Retrieve session
	retrievedSession, err := manager2.Get("persistence-test")
	if err != nil {
		t.Errorf("failed to retrieve persisted session: %v", err)
		return
	}

	// Verify all fields match
	if retrievedSession.Name != originalSession.Name {
		t.Errorf("name mismatch: expected %s, got %s", originalSession.Name, retrievedSession.Name)
	}

	if retrievedSession.Description != originalSession.Description {
		t.Errorf("description mismatch: expected %s, got %s", originalSession.Description, retrievedSession.Description)
	}

	if retrievedSession.VolumePath != originalSession.VolumePath {
		t.Errorf("volume path mismatch: expected %s, got %s", originalSession.VolumePath, retrievedSession.VolumePath)
	}

	// Time comparison (allow small difference due to serialization)
	timeDiff := retrievedSession.CreatedAt.Sub(originalSession.CreatedAt)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("created_at time difference too large: %v", timeDiff)
	}
}

func TestManager_GetSessionDir(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir)

	sessionName := "test-session"
	expectedPath := filepath.Join(tmpDir, "sessions", sessionName)

	actualPath := manager.GetSessionDir(sessionName)
	if actualPath != expectedPath {
		t.Errorf("expected session dir '%s', got '%s'", expectedPath, actualPath)
	}
}

func TestManager_CreateWithOptions(t *testing.T) {
	manager, _ := setupTestManager(t)

	tests := []struct {
		name         string
		sessionName  string
		description  string
		volumePath   string
		devcontainer bool
		clients      []string
		expectError  bool
	}{
		{
			name:         "create with devcontainer enabled",
			sessionName:  "test-devcontainer",
			description:  "Test devcontainer session",
			volumePath:   "",
			devcontainer: true,
			clients:      []string{"vscode", "claude-code"},
			expectError:  false,
		},
		{
			name:         "create with clients only",
			sessionName:  "test-clients",
			description:  "Test clients session",
			volumePath:   "",
			devcontainer: false,
			clients:      []string{"cursor"},
			expectError:  false,
		},
		{
			name:         "create with all options",
			sessionName:  "test-all-options",
			description:  "Test all options",
			volumePath:   "/custom/path",
			devcontainer: true,
			clients:      []string{"vscode", "claude-code", "cursor"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := manager.Create(tt.sessionName, tt.description, tt.volumePath)

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

			if session == nil {
				t.Fatal("session is nil")
			}

			if session.Name != tt.sessionName {
				t.Errorf("expected name %s, got %s", tt.sessionName, session.Name)
			}

			if session.Description != tt.description {
				t.Errorf("expected description %s, got %s", tt.description, session.Description)
			}

		})
	}
}

func TestManager_CreateProjectSession(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Create a temporary directory to simulate project directory
	projectDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	defer func() {
		// Restore original working directory
		os.Chdir(originalWd)
	}()

	session, err := manager.Create("test-project", "Test project session", "")
	if err != nil {
		t.Fatalf("failed to create project session: %v", err)
	}

	// Check that session name is correct
	if session.Name != "test-project" {
		t.Errorf("expected session name test-project, got %s", session.Name)
	}

	// Check the session was saved in the global location
	sessionFile := filepath.Join(manager.servoDir, "sessions", "test-project", "session.yaml")
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Error("expected session.yaml to be created in global sessions directory")
	}
}

func TestManager_ExistsWithOptions(t *testing.T) {
	manager, _ := setupTestManager(t)

	// Test global session
	_, err := manager.Create("global-test", "Global test", "")
	if err != nil {
		t.Fatalf("failed to create global session: %v", err)
	}

	exists, err := manager.Exists("global-test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected global session to exist")
	}

	// Test non-existent global session
	exists, err = manager.Exists("non-existent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected non-existent session to not exist")
	}

	// Test project session
	projectDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	defer func() {
		os.Chdir(originalWd)
	}()

	// Create project session
	_, err = manager.Create("test-project-2", "Test project", "")
	if err != nil {
		t.Fatalf("failed to create project session: %v", err)
	}

	// Check if project session exists
	exists, err = manager.Exists("test-project-2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected project session to exist")
	}
}

func TestManager_Rename(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create test session
	session, err := manager.Create("old-name", "Test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test successful rename
	err = manager.Rename("old-name", "new-name")
	if err != nil {
		t.Errorf("unexpected error renaming session: %v", err)
	}

	// Verify old session no longer exists
	exists, err := manager.Exists("old-name")
	if err != nil {
		t.Errorf("unexpected error checking old session: %v", err)
	}
	if exists {
		t.Error("old session should no longer exist")
	}

	// Verify new session exists
	exists, err = manager.Exists("new-name")
	if err != nil {
		t.Errorf("unexpected error checking new session: %v", err)
	}
	if !exists {
		t.Error("new session should exist")
	}

	// Verify session metadata was updated
	renamedSession, err := manager.Get("new-name")
	if err != nil {
		t.Errorf("failed to get renamed session: %v", err)
	}
	if renamedSession.Name != "new-name" {
		t.Errorf("expected session name to be 'new-name', got '%s'", renamedSession.Name)
	}
	if renamedSession.Description != session.Description {
		t.Errorf("expected description to be preserved")
	}
}

func TestManager_RenameActiveSession(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create and activate test session
	_, err := manager.Create("active-session", "Active test session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = manager.Activate("active-session")
	if err != nil {
		t.Fatalf("failed to activate session: %v", err)
	}

	// Rename the active session
	err = manager.Rename("active-session", "renamed-active")
	if err != nil {
		t.Errorf("unexpected error renaming active session: %v", err)
	}

	// Verify the active session reference was updated
	activeSession, err := manager.GetActive()
	if err != nil {
		t.Errorf("failed to get active session: %v", err)
	}
	if activeSession == nil {
		t.Error("expected active session to exist")
	} else if activeSession.Name != "renamed-active" {
		t.Errorf("expected active session to be 'renamed-active', got '%s'", activeSession.Name)
	}
}

func TestManager_RenameErrors(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create test sessions
	_, err := manager.Create("existing-session", "Existing session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	tests := []struct {
		name        string
		oldName     string
		newName     string
		expectedErr string
	}{
		{
			name:        "empty old name",
			oldName:     "",
			newName:     "new-name",
			expectedErr: "session names cannot be empty",
		},
		{
			name:        "empty new name",
			oldName:     "old-name",
			newName:     "",
			expectedErr: "session names cannot be empty",
		},
		{
			name:        "same names",
			oldName:     "existing-session",
			newName:     "existing-session",
			expectedErr: "old and new session names are the same",
		},
		{
			name:        "non-existent session",
			oldName:     "non-existent",
			newName:     "new-name",
			expectedErr: "session 'non-existent' does not exist",
		},
		{
			name:        "target already exists",
			oldName:     "existing-session",
			newName:     "existing-session", // This would be caught by "same names" first
			expectedErr: "old and new session names are the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Rename(tt.oldName, tt.newName)
			if err == nil {
				t.Error("expected error but got none")
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestManager_RenameWithProjectConfig(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a project directory structure to test project config updates
	projectDir := filepath.Join(tempDir, "test-project")
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Change to project directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		os.Chdir(originalWd)
	}()

	// Create session
	_, err = manager.Create("default", "Default session", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test rename with project config update (should not fail even if no project exists)
	err = manager.Rename("default", "main")
	if err != nil {
		t.Errorf("unexpected error renaming session with project config: %v", err)
	}

	// Verify rename succeeded
	exists, err := manager.Exists("main")
	if err != nil {
		t.Errorf("unexpected error checking renamed session: %v", err)
	}
	if !exists {
		t.Error("renamed session should exist")
	}
}
