package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/servo/servo/internal/utils"
	"gopkg.in/yaml.v3"
)

// Session represents a named global workflow session
type Session struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	VolumePath  string    `yaml:"volume_path" json:"volume_path"`
	Active      bool      `yaml:"active" json:"active"`
}

// Manager handles session operations
type Manager struct {
	servoDir string
}

// NewManager creates a new session manager
func NewManager(servoDir string) *Manager {
	return &Manager{
		servoDir: servoDir,
	}
}

// Create creates a new named global session
func (m *Manager) Create(name, description, volumePath string) (*Session, error) {
	if name == "" {
		return nil, fmt.Errorf("session name cannot be empty")
	}

	if exists, err := m.Exists(name); err != nil {
		return nil, fmt.Errorf("failed to check if session exists: %w", err)
	} else if exists {
		return nil, fmt.Errorf("session '%s' already exists", name)
	}

	if volumePath == "" {
		volumePath = filepath.Join(m.getSessionDir(name), "volumes")
	}

	session := &Session{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		VolumePath:  volumePath,
		Active:      false,
	}

	if err := m.createSessionDirectories(name); err != nil {
		return nil, fmt.Errorf("failed to create session directories: %w", err)
	}

	if err := m.saveSession(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// List returns all available sessions
func (m *Manager) List() ([]*Session, error) {
	sessionsDir := filepath.Join(m.servoDir, "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []*Session{}, nil
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionName := entry.Name()
		session, err := m.Get(sessionName)
		if err != nil {
			// Skip invalid sessions but continue
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Get retrieves a session by name
func (m *Manager) Get(name string) (*Session, error) {
	if name == "" {
		return nil, fmt.Errorf("session name cannot be empty")
	}

	// Sessions are project-specific, stored in .servo/sessions/{name}/session.yaml
	sessionFile := filepath.Join(m.getSessionDir(name), "session.yaml")

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("session '%s' does not exist", name)
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := yaml.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &session, nil
}

// Delete removes a session
func (m *Manager) Delete(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	if exists, err := m.Exists(name); err != nil {
		return fmt.Errorf("failed to check if session exists: %w", err)
	} else if !exists {
		return fmt.Errorf("session '%s' does not exist", name)
	}

	activeSession, err := m.GetActive()
	if err == nil && activeSession != nil && activeSession.Name == name {
		if err := m.ClearActive(); err != nil {
			return fmt.Errorf("failed to clear active session: %w", err)
		}
	}

	// Remove session directory
	sessionDir := m.getSessionDir(name)
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}

	return nil
}

// Activate makes a session the active one
func (m *Manager) Activate(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	// Check if session exists
	session, err := m.Get(name)
	if err != nil {
		return fmt.Errorf("session '%s' does not exist: %w", name, err)
	}

	// Update active session file
	activeFile := filepath.Join(m.servoDir, "active_session")
	if err := os.WriteFile(activeFile, []byte(name), 0644); err != nil {
		return fmt.Errorf("failed to write active session: %w", err)
	}

	// Mark session as active and save
	session.Active = true
	if err := m.saveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Mark all other sessions as inactive
	sessions, err := m.List()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	for _, s := range sessions {
		if s.Name != name && s.Active {
			s.Active = false
			if err := m.saveSession(s); err != nil {
				// Log error but don't fail activation
				fmt.Printf("Warning: failed to mark session '%s' as inactive: %v\n", s.Name, err)
			}
		}
	}

	return nil
}

// GetActive returns the currently active session
func (m *Manager) GetActive() (*Session, error) {
	activeFile := filepath.Join(m.servoDir, "active_session")
	data, err := os.ReadFile(activeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No active session
		}
		return nil, fmt.Errorf("failed to read active session: %w", err)
	}

	sessionName := string(data)
	if sessionName == "" {
		return nil, nil
	}

	return m.Get(sessionName)
}

// ClearActive clears the active session
func (m *Manager) ClearActive() error {
	activeFile := filepath.Join(m.servoDir, "active_session")
	if err := os.Remove(activeFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear active session: %w", err)
	}

	// Mark all sessions as inactive
	sessions, err := m.List()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	for _, session := range sessions {
		if session.Active {
			session.Active = false
			if err := m.saveSession(session); err != nil {
				fmt.Printf("Warning: failed to mark session '%s' as inactive: %v\n", session.Name, err)
			}
		}
	}

	return nil
}

// Exists checks if a session exists
func (m *Manager) Exists(name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	sessionDir := m.getSessionDir(name)
	sessionFile := filepath.Join(sessionDir, "session.yaml")
	_, err := os.Stat(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Clone creates a new session by copying configuration from an existing session.
//
// This operation performs a selective copy of session data, including:
// - Configuration files (config.yaml)
// - Manifest definitions
// - Override settings
//
// For security reasons, secrets and volume data are NOT copied during cloning.
// This ensures that sensitive data remains isolated between sessions while allowing
// developers to quickly set up new environments based on existing configurations.
//
// The cloning process creates a completely independent session that can be modified
// without affecting the source session.
func (m *Manager) Clone(sourceName, targetName, description string) (*Session, error) {
	if sourceName == "" || targetName == "" {
		return nil, fmt.Errorf("source and target session names cannot be empty")
	}

	// Check if source exists
	_, err := m.Get(sourceName)
	if err != nil {
		return nil, fmt.Errorf("source session '%s' does not exist: %w", sourceName, err)
	}

	// Check if target already exists
	if exists, err := m.Exists(targetName); err != nil {
		return nil, fmt.Errorf("failed to check if target session exists: %w", err)
	} else if exists {
		return nil, fmt.Errorf("target session '%s' already exists", targetName)
	}

	// Create new session
	targetSession, err := m.Create(targetName, description, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create target session: %w", err)
	}

	// Copy configuration and other files from source
	sourceDir := m.getSessionDir(sourceName)
	targetDir := m.getSessionDir(targetName)

	// Copy config if it exists
	sourceConfig := filepath.Join(sourceDir, "config.yaml")
	if _, err := os.Stat(sourceConfig); err == nil {
		targetConfig := filepath.Join(targetDir, "config.yaml")
		if err := copyFile(sourceConfig, targetConfig); err != nil {
			fmt.Printf("Warning: failed to copy config: %v\n", err)
		}
	}

	// Copy manifests directory if it exists
	sourceManifests := filepath.Join(sourceDir, "manifests")
	if info, err := os.Stat(sourceManifests); err == nil && info.IsDir() {
		targetManifests := filepath.Join(targetDir, "manifests")
		if err := copyDir(sourceManifests, targetManifests); err != nil {
			fmt.Printf("Warning: failed to copy manifests: %v\n", err)
		}
	}

	// Secrets and volumes are not copied for security reasons

	return targetSession, nil
}

// AdoptVolumes configures a session to use volumes from another session
func (m *Manager) AdoptVolumes(sessionName, sourceSessionName string) error {
	if sessionName == "" || sourceSessionName == "" {
		return fmt.Errorf("session names cannot be empty")
	}

	// Get both sessions
	session, err := m.Get(sessionName)
	if err != nil {
		return fmt.Errorf("session '%s' does not exist: %w", sessionName, err)
	}

	sourceSession, err := m.Get(sourceSessionName)
	if err != nil {
		return fmt.Errorf("source session '%s' does not exist: %w", sourceSessionName, err)
	}

	// Update volume path to point to source session's volumes
	session.VolumePath = sourceSession.VolumePath

	// Save updated session
	if err := m.saveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// SetVolumePath sets a custom volume path for a session
func (m *Manager) SetVolumePath(sessionName, volumePath string) error {
	if sessionName == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	session, err := m.Get(sessionName)
	if err != nil {
		return fmt.Errorf("session '%s' does not exist: %w", sessionName, err)
	}

	// Expand path if it's relative or contains ~
	if volumePath != "" {
		if filepath.IsAbs(volumePath) {
			session.VolumePath = volumePath
		} else {
			absPath, err := filepath.Abs(volumePath)
			if err != nil {
				return fmt.Errorf("failed to resolve volume path: %w", err)
			}
			session.VolumePath = absPath
		}
	} else {
		// Reset to default
		session.VolumePath = filepath.Join(m.getSessionDir(sessionName), "volumes")
	}

	// Save updated session
	if err := m.saveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// GetSessionDir returns the directory path for a session
func (m *Manager) GetSessionDir(name string) string {
	return m.getSessionDir(name)
}

// Helper methods

func (m *Manager) getSessionDir(name string) string {
	return filepath.Join(m.servoDir, "sessions", name)
}

func (m *Manager) createSessionDirectories(name string) error {
	sessionDir := m.getSessionDir(name)

	// Create all required directories
	dirs := []string{
		sessionDir,
		filepath.Join(sessionDir, "manifests"), // Server manifests (.servo files)
		filepath.Join(sessionDir, "config"),    // User configuration overrides
		filepath.Join(sessionDir, "volumes"),   // Default Docker volumes
		filepath.Join(sessionDir, "logs"),      // Session-specific logs
	}
	
	return utils.EnsureDirectoryStructure(dirs)
}

// SaveSession saves a session to its appropriate file location
func (m *Manager) SaveSession(session *Session) error {
	return m.saveSession(session)
}

func (m *Manager) saveSession(session *Session) error {
	// Global session - save to ~/.servo/sessions/name/session.yaml
	sessionDir := m.getSessionDir(session.Name)
	sessionFile := filepath.Join(sessionDir, "session.yaml")

	if err := utils.WriteYAMLFile(sessionFile, session); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Rename renames a session and updates all references
func (m *Manager) Rename(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("session names cannot be empty")
	}

	if oldName == newName {
		return fmt.Errorf("old and new session names are the same")
	}

	// Check if old session exists
	oldSession, err := m.Get(oldName)
	if err != nil {
		return fmt.Errorf("session '%s' does not exist: %w", oldName, err)
	}

	// Check if new name already exists
	if exists, err := m.Exists(newName); err != nil {
		return fmt.Errorf("failed to check if new session name exists: %w", err)
	} else if exists {
		return fmt.Errorf("session '%s' already exists", newName)
	}

	// Get old and new session directories
	oldSessionDir := m.getSessionDir(oldName)
	newSessionDir := m.getSessionDir(newName)

	// Create new session directory
	if err := os.MkdirAll(newSessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create new session directory: %w", err)
	}

	// Copy all contents to new directory
	if err := copyDir(oldSessionDir, newSessionDir); err != nil {
		// Clean up on failure
		os.RemoveAll(newSessionDir)
		return fmt.Errorf("failed to copy session data: %w", err)
	}

	// Update session metadata
	oldSession.Name = newName
	newSessionFile := filepath.Join(newSessionDir, "session.yaml")

	if err := utils.WriteYAMLFile(newSessionFile, oldSession); err != nil {
		os.RemoveAll(newSessionDir)
		return fmt.Errorf("failed to write updated session file: %w", err)
	}

	// Update active session reference if this was the active session
	isActive := oldSession.Active
	if isActive {
		activeFile := filepath.Join(m.servoDir, "active_session")
		if err := os.WriteFile(activeFile, []byte(newName), 0644); err != nil {
			// Try to rollback
			os.RemoveAll(newSessionDir)
			return fmt.Errorf("failed to update active session reference: %w", err)
		}
	}

	// Update project configuration if this session is referenced
	if err := m.updateProjectConfigForRename(oldName, newName); err != nil {
		// Log warning but don't fail the rename
		fmt.Printf("Warning: failed to update project configuration: %v\n", err)
	}

	// Remove old session directory only after everything else succeeds
	if err := os.RemoveAll(oldSessionDir); err != nil {
		fmt.Printf("Warning: failed to remove old session directory '%s': %v\n", oldSessionDir, err)
	}

	return nil
}

// updateProjectConfigForRename updates project configuration when a session is renamed
func (m *Manager) updateProjectConfigForRename(oldName, newName string) error {
	// This is a simplified integration - in a full implementation this would
	// need to import the project package, but that could create circular imports.
	// For now, we'll directly manipulate the project.yaml file

	projectFile := filepath.Join(m.servoDir, "project.yaml")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		return nil // No project file, nothing to update
	}

	// Read project file
	data, err := os.ReadFile(projectFile)
	if err != nil {
		return fmt.Errorf("failed to read project file: %w", err)
	}

	// Parse as generic map to avoid circular imports
	var projectData map[string]interface{}
	if err := yaml.Unmarshal(data, &projectData); err != nil {
		return fmt.Errorf("failed to parse project file: %w", err)
	}

	// Update references to the old session name
	updated := false

	if defaultSession, ok := projectData["default_session"].(string); ok && defaultSession == oldName {
		projectData["default_session"] = newName
		updated = true
	}

	if activeSession, ok := projectData["active_session"].(string); ok && activeSession == oldName {
		projectData["active_session"] = newName
		updated = true
	}

	// Write back if we made changes
	if updated {
		newData, err := yaml.Marshal(projectData)
		if err != nil {
			return fmt.Errorf("failed to marshal updated project: %w", err)
		}

		if err := os.WriteFile(projectFile, newData, 0644); err != nil {
			return fmt.Errorf("failed to write updated project file: %w", err)
		}
	}

	return nil
}

// Utility functions for file operations
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}
