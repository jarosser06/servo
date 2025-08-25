package project

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servo/servo/internal/utils"
	"gopkg.in/yaml.v3"
)

// RequiredSecret represents a secret that the project needs
type RequiredSecret struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

// MCPServer represents an MCP server to be installed in the devcontainer
type MCPServer struct {
	Name     string   `yaml:"name" json:"name"`
	Source   string   `yaml:"source" json:"source"`
	Clients  []string `yaml:"clients,omitempty" json:"clients,omitempty"`
	Sessions []string `yaml:"sessions,omitempty" json:"sessions,omitempty"` // Sessions where this server is installed
}

// ServerAlreadyExistsError indicates a server already exists and no update was requested
type ServerAlreadyExistsError struct {
	ServerName  string
	SessionName string
}

func (e *ServerAlreadyExistsError) Error() string {
	return fmt.Sprintf("MCP server %s already exists in session %s", e.ServerName, e.SessionName)
}

// Project represents a servo project configuration
type Project struct {
	Clients         []string         `yaml:"clients,omitempty" json:"clients,omitempty"`
	DefaultSession  string           `yaml:"default_session" json:"default_session"`                   // Default session name
	ActiveSession   string           `yaml:"active_session,omitempty" json:"active_session,omitempty"` // Currently active session
	MCPServers      []MCPServer      `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	RequiredSecrets []RequiredSecret `yaml:"required_secrets,omitempty" json:"required_secrets,omitempty"`
}

// Manager handles project operations in the current directory
type Manager struct {
	// Project manager operates on current working directory only
}

// NewManager creates a new project manager
func NewManager() *Manager {
	return &Manager{}
}

// projectFileExists checks if project.yaml exists
func (m *Manager) projectFileExists() bool {
	projectFile := filepath.Join(m.GetServoDir(), "project.yaml")
	_, err := os.Stat(projectFile)
	return err == nil
}

// createProjectDirectories creates the basic project directory structure
func (m *Manager) createProjectDirectories() error {
	servoDir := m.GetServoDir()

	// Create all required directories
	dirs := []string{
		servoDir,
		filepath.Join(servoDir, "config"), // Project-level config overrides
		filepath.Join(servoDir, "sessions"),
		filepath.Join(servoDir, "sessions", "default"),
		filepath.Join(servoDir, "sessions", "default", "manifests"),
		filepath.Join(servoDir, "sessions", "default", "config"), // Session-level config overrides
		filepath.Join(servoDir, "sessions", "default", "volumes"),
	}

	if err := utils.EnsureDirectoryStructure(dirs); err != nil {
		return fmt.Errorf("failed to create project directories: %w", err)
	}

	// Create .gitignore file
	gitignorePath := filepath.Join(servoDir, ".gitignore")
	gitignoreContent := `# Servo project files
secrets.yaml
*.log

# Legacy structure (for backwards compatibility)
volumes/

# Session-specific files
sessions/*/config/
sessions/*/logs/

# OS files
.DS_Store
Thumbs.db
`
	if err := utils.WriteFileWithDir(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

// saveProject saves the project configuration to disk
func (m *Manager) saveProject(project *Project) error {
	projectFile := filepath.Join(m.GetServoDir(), "project.yaml")

	if err := utils.WriteYAMLFile(projectFile, project); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// IsProject checks if current directory contains a servo project
func (m *Manager) IsProject() bool {
	return m.projectFileExists()
}

// Init initializes a new servo project in the current directory
func (m *Manager) Init(sessionName string, clients []string) (*Project, error) {
	if m.IsProject() {
		return nil, fmt.Errorf("servo project already exists in current directory")
	}

	// Use "default" as session name if none provided
	if sessionName == "" {
		sessionName = "default"
	}

	// Create project structure
	if err := m.createProjectDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create project directories: %w", err)
	}

	// Create project configuration
	project := &Project{
		Clients:        clients,
		DefaultSession: sessionName,
		ActiveSession:  sessionName,
	}

	if err := m.saveProject(project); err != nil {
		return nil, fmt.Errorf("failed to save project configuration: %w", err)
	}

	// Note: Session directories are created on-demand by install command

	return project, nil
}

// Get returns the current project configuration
func (m *Manager) Get() (*Project, error) {
	if !m.IsProject() {
		return nil, fmt.Errorf("not in a servo project directory")
	}

	projectFile := filepath.Join(".servo", "project.yaml")
	data, err := os.ReadFile(projectFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project file: %w", err)
	}

	return &project, nil
}

// Save saves the project configuration
func (m *Manager) Save(project *Project) error {
	return m.saveProject(project)
}

// Delete removes the servo project from current directory
func (m *Manager) Delete() error {
	if !m.IsProject() {
		return fmt.Errorf("not in a servo project directory")
	}

	// Remove entire .servo directory
	if err := os.RemoveAll(".servo"); err != nil {
		return fmt.Errorf("failed to remove project directory: %w", err)
	}

	return nil
}

// AddClient adds a client to the project configuration
func (m *Manager) AddClient(clientName string) error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	// Check if client already exists
	for _, client := range project.Clients {
		if client == clientName {
			return fmt.Errorf("client '%s' already configured", clientName)
		}
	}

	project.Clients = append(project.Clients, clientName)
	return m.Save(project)
}

// RemoveClient removes a client from the project configuration
func (m *Manager) RemoveClient(clientName string) error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	var newClients []string
	found := false
	for _, client := range project.Clients {
		if client == clientName {
			found = true
		} else {
			newClients = append(newClients, client)
		}
	}

	if !found {
		return fmt.Errorf("client '%s' not found in project", clientName)
	}

	project.Clients = newClients
	return m.Save(project)
}

// SetActiveSession sets the active session for this project
func (m *Manager) SetActiveSession(sessionName string) error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	project.ActiveSession = sessionName
	return m.Save(project)
}

// ClearActiveSession removes the active session reference
func (m *Manager) ClearActiveSession() error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	project.ActiveSession = ""
	return m.Save(project)
}

// GetProjectPath returns the absolute path of the current project
func (m *Manager) GetProjectPath() (string, error) {
	if !m.IsProject() {
		return "", fmt.Errorf("not in a servo project directory")
	}

	return os.Getwd()
}

// GetServoDir returns the .servo directory path for the current project
func (m *Manager) GetServoDir() string {
	return ".servo"
}

// GetProjectName returns the project name based on the current directory
func (m *Manager) GetProjectName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return filepath.Base(cwd), nil
}

// AddRequiredSecret adds a required secret to the project
func (m *Manager) AddRequiredSecret(name, description string) error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	// Check if secret already exists
	for _, secret := range project.RequiredSecrets {
		if secret.Name == name {
			return fmt.Errorf("required secret '%s' already exists", name)
		}
	}

	// Add the new secret
	project.RequiredSecrets = append(project.RequiredSecrets, RequiredSecret{
		Name:        name,
		Description: description,
	})

	return m.Save(project)
}

// RemoveRequiredSecret removes a required secret from the project
func (m *Manager) RemoveRequiredSecret(name string) error {
	project, err := m.Get()
	if err != nil {
		return err
	}

	// Find and remove the secret
	for i, secret := range project.RequiredSecrets {
		if secret.Name == name {
			project.RequiredSecrets = append(project.RequiredSecrets[:i], project.RequiredSecrets[i+1:]...)
			return m.Save(project)
		}
	}

	return fmt.Errorf("required secret '%s' not found", name)
}

// GetMissingSecrets checks which required secrets are not set and returns them
func (m *Manager) GetMissingSecrets() ([]RequiredSecret, error) {
	project, err := m.Get()
	if err != nil {
		return nil, err
	}

	// Get configured secrets by trying to decrypt the secrets file
	configuredSecrets, err := m.getConfiguredSecrets()
	if err != nil {
		// If we can't read secrets (file missing, wrong password, etc.), assume all are missing
		return project.RequiredSecrets, nil
	}

	// Build set of configured secret names
	configuredSet := make(map[string]bool)
	for secretName := range configuredSecrets {
		configuredSet[secretName] = true
	}

	// Return only missing secrets
	var missing []RequiredSecret
	for _, required := range project.RequiredSecrets {
		if !configuredSet[required.Name] {
			missing = append(missing, required)
		}
	}

	return missing, nil
}

// getConfiguredSecrets returns a map of configured secret names to their values
func (m *Manager) getConfiguredSecrets() (map[string]string, error) {
	secretsPath := filepath.Join(m.GetServoDir(), "secrets.yaml")

	// Check if secrets file exists
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return make(map[string]string), nil // No secrets file = no configured secrets
	}

	// Read and parse the secrets file
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	var fileData struct {
		Version string            `yaml:"version"`
		Secrets map[string]string `yaml:"secrets"`
	}

	if err := yaml.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	// Decode base64 secrets
	decodedSecrets := make(map[string]string)
	for key, encodedValue := range fileData.Secrets {
		decodedValue, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			// If decoding fails, assume it's plain text (for migration)
			decodedSecrets[key] = encodedValue
		} else {
			decodedSecrets[key] = string(decodedValue)
		}
	}

	return decodedSecrets, nil
}

// GetConfiguredSecrets returns a map of configured secret names (for external use)
func (m *Manager) GetConfiguredSecrets() (map[string]bool, error) {
	secrets, err := m.getConfiguredSecrets()
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for name := range secrets {
		result[name] = true
	}

	return result, nil
}

// AddMCPServer adds an MCP server to the project, session-aware
func (m *Manager) AddMCPServer(serverName string, source string, clients []string) error {
	return m.AddMCPServerToSession(serverName, source, clients, "", false)
}

// AddMCPServerWithUpdate adds an MCP server to the project with update option
func (m *Manager) AddMCPServerWithUpdate(serverName string, source string, clients []string, forceUpdate bool) error {
	return m.AddMCPServerToSession(serverName, source, clients, "", forceUpdate)
}

// AddMCPServerToSession adds an MCP server to a specific session
func (m *Manager) AddMCPServerToSession(serverName string, source string, clients []string, sessionName string, forceUpdate bool) error {
	project, err := m.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// If no session specified, use default
	if sessionName == "" {
		sessionName = project.DefaultSession
	}

	// Check if server already exists in this session
	for i, server := range project.MCPServers {
		if server.Name == serverName {
			// Server exists - check if it's already in this session
			for _, existingSession := range server.Sessions {
				if existingSession == sessionName {
					if !forceUpdate {
						// Return a special error type to indicate "already exists, nothing to do"
						return &ServerAlreadyExistsError{ServerName: serverName, SessionName: sessionName}
					}
					// Force update: update the existing server configuration
					project.MCPServers[i].Source = source
					project.MCPServers[i].Clients = clients
					return m.saveProject(project)
				}
			}
			// Server exists but not in this session - add session to existing server
			project.MCPServers[i].Sessions = append(project.MCPServers[i].Sessions, sessionName)
			return m.saveProject(project)
		}
	}

	// Add new server with session
	newServer := MCPServer{
		Name:     serverName,
		Source:   source,
		Clients:  clients,
		Sessions: []string{sessionName},
	}

	project.MCPServers = append(project.MCPServers, newServer)

	return m.saveProject(project)
}

// RemoveMCPServer removes an MCP server from the project
func (m *Manager) RemoveMCPServer(serverName string) error {
	project, err := m.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Find and remove the server
	for i, server := range project.MCPServers {
		if server.Name == serverName {
			project.MCPServers = append(project.MCPServers[:i], project.MCPServers[i+1:]...)
			return m.saveProject(project)
		}
	}

	return fmt.Errorf("MCP server %s not found", serverName)
}
