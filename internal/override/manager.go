package override

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration overrides with precedence: session > project > defaults
type Manager struct {
	sessionDir string
	projectDir string
}

// NewManager creates a new override manager
func NewManager(sessionDir, projectDir string) *Manager {
	return &Manager{
		sessionDir: sessionDir,
		projectDir: projectDir,
	}
}

// DockerComposeOverride represents docker-compose configuration overrides
type DockerComposeOverride struct {
	Version  string                     `yaml:"version,omitempty"`
	Services map[string]ServiceOverride `yaml:"services,omitempty"`
	Networks map[string]interface{}     `yaml:"networks,omitempty"`
	Volumes  map[string]interface{}     `yaml:"volumes,omitempty"`
	Secrets  map[string]interface{}     `yaml:"secrets,omitempty"`
}

// ServiceOverride represents service-specific overrides
type ServiceOverride struct {
	Image       string                 `yaml:"image,omitempty"`
	Environment map[string]string      `yaml:"environment,omitempty"`
	Ports       []string               `yaml:"ports,omitempty"`
	Volumes     []string               `yaml:"volumes,omitempty"`
	Command     []string               `yaml:"command,omitempty"`
	DependsOn   []string               `yaml:"depends_on,omitempty"`
	Networks    []string               `yaml:"networks,omitempty"`
	Labels      map[string]string      `yaml:"labels,omitempty"`
	Extra       map[string]interface{} `yaml:",inline"` // For any additional fields
}

// DevcontainerOverride represents devcontainer configuration overrides
type DevcontainerOverride struct {
	Name              string                 `json:"name,omitempty"`
	Image             string                 `json:"image,omitempty"`
	Features          map[string]interface{} `json:"features,omitempty"`
	Customizations    map[string]interface{} `json:"customizations,omitempty"`
	ForwardPorts      []int                  `json:"forwardPorts,omitempty"`
	PostCreateCommand string                 `json:"postCreateCommand,omitempty"`
	PostStartCommand  string                 `json:"postStartCommand,omitempty"`
	RemoteUser        string                 `json:"remoteUser,omitempty"`
	WorkspaceFolder   string                 `json:"workspaceFolder,omitempty"`
	Mounts            []string               `json:"mounts,omitempty"`
	Extra             map[string]interface{} `json:"-"` // Handle with custom marshal/unmarshal
}

// GetDockerComposeOverrides retrieves docker-compose overrides with precedence
func (m *Manager) GetDockerComposeOverrides() (*DockerComposeOverride, error) {
	// Load in precedence order: defaults < project < session
	merged := &DockerComposeOverride{
		Services: make(map[string]ServiceOverride),
		Networks: make(map[string]interface{}),
		Volumes:  make(map[string]interface{}),
		Secrets:  make(map[string]interface{}),
	}

	// 1. Load default overrides (none currently, but structure is ready)

	// 2. Load project-level overrides
	if m.projectDir != "" {
		projectOverrides, err := m.loadDockerComposeOverride(filepath.Join(m.projectDir, ".servo", "config", "docker-compose.yml"))
		if err == nil {
			merged = m.mergeDockerComposeOverrides(merged, projectOverrides)
		}
	}

	// 3. Load session-level overrides (highest precedence)
	if m.sessionDir != "" {
		sessionOverrides, err := m.loadDockerComposeOverride(filepath.Join(m.sessionDir, "config", "docker-compose.yml"))
		if err == nil {
			merged = m.mergeDockerComposeOverrides(merged, sessionOverrides)
		}
	}

	return merged, nil
}

// GetDevcontainerOverrides retrieves devcontainer overrides with precedence
func (m *Manager) GetDevcontainerOverrides() (*DevcontainerOverride, error) {
	// Load in precedence order: defaults < project < session
	merged := &DevcontainerOverride{
		Features:       make(map[string]interface{}),
		Customizations: make(map[string]interface{}),
		Extra:          make(map[string]interface{}),
	}

	// 1. Load default overrides (none currently)

	// 2. Load project-level overrides
	if m.projectDir != "" {
		projectOverrides, err := m.loadDevcontainerOverride(filepath.Join(m.projectDir, ".servo", "config", "devcontainer.json"))
		if err == nil {
			merged = m.mergeDevcontainerOverrides(merged, projectOverrides)
		}
	}

	// 3. Load session-level overrides (highest precedence)
	if m.sessionDir != "" {
		sessionOverrides, err := m.loadDevcontainerOverride(filepath.Join(m.sessionDir, "config", "devcontainer.json"))
		if err == nil {
			merged = m.mergeDevcontainerOverrides(merged, sessionOverrides)
		}
	}

	return merged, nil
}

// SaveDockerComposeOverride saves docker-compose overrides to the specified level
func (m *Manager) SaveDockerComposeOverride(level string, override *DockerComposeOverride) error {
	var filePath string

	switch level {
	case "project":
		if m.projectDir == "" {
			return fmt.Errorf("project directory not set")
		}
		overrideDir := filepath.Join(m.projectDir, ".servo", "config")
		if err := os.MkdirAll(overrideDir, 0755); err != nil {
			return fmt.Errorf("failed to create project override directory: %w", err)
		}
		filePath = filepath.Join(overrideDir, "docker-compose.yml")

	case "session":
		if m.sessionDir == "" {
			return fmt.Errorf("session directory not set")
		}
		overrideDir := filepath.Join(m.sessionDir, "config")
		if err := os.MkdirAll(overrideDir, 0755); err != nil {
			return fmt.Errorf("failed to create session override directory: %w", err)
		}
		filePath = filepath.Join(overrideDir, "docker-compose.yml")

	default:
		return fmt.Errorf("invalid override level: %s (must be 'project' or 'session')", level)
	}

	data, err := yaml.Marshal(override)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose override: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// SaveDevcontainerOverride saves devcontainer overrides to the specified level
func (m *Manager) SaveDevcontainerOverride(level string, override *DevcontainerOverride) error {
	var filePath string

	switch level {
	case "project":
		if m.projectDir == "" {
			return fmt.Errorf("project directory not set")
		}
		overrideDir := filepath.Join(m.projectDir, ".servo", "config")
		if err := os.MkdirAll(overrideDir, 0755); err != nil {
			return fmt.Errorf("failed to create project override directory: %w", err)
		}
		filePath = filepath.Join(overrideDir, "devcontainer.json")

	case "session":
		if m.sessionDir == "" {
			return fmt.Errorf("session directory not set")
		}
		overrideDir := filepath.Join(m.sessionDir, "config")
		if err := os.MkdirAll(overrideDir, 0755); err != nil {
			return fmt.Errorf("failed to create session override directory: %w", err)
		}
		filePath = filepath.Join(overrideDir, "devcontainer.json")

	default:
		return fmt.Errorf("invalid override level: %s (must be 'project' or 'session')", level)
	}

	data, err := json.MarshalIndent(override, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal devcontainer override: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// loadDockerComposeOverride loads docker-compose override from file
func (m *Manager) loadDockerComposeOverride(filePath string) (*DockerComposeOverride, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &DockerComposeOverride{
			Services: make(map[string]ServiceOverride),
			Networks: make(map[string]interface{}),
			Volumes:  make(map[string]interface{}),
			Secrets:  make(map[string]interface{}),
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read override file: %w", err)
	}

	var override DockerComposeOverride
	if err := yaml.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker-compose override: %w", err)
	}

	// Initialize maps if they're nil
	if override.Services == nil {
		override.Services = make(map[string]ServiceOverride)
	}
	if override.Networks == nil {
		override.Networks = make(map[string]interface{})
	}
	if override.Volumes == nil {
		override.Volumes = make(map[string]interface{})
	}
	if override.Secrets == nil {
		override.Secrets = make(map[string]interface{})
	}

	return &override, nil
}

// loadDevcontainerOverride loads devcontainer override from file
func (m *Manager) loadDevcontainerOverride(filePath string) (*DevcontainerOverride, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &DevcontainerOverride{
			Features:       make(map[string]interface{}),
			Customizations: make(map[string]interface{}),
			Extra:          make(map[string]interface{}),
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read override file: %w", err)
	}

	var override DevcontainerOverride
	if err := json.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devcontainer override: %w", err)
	}

	// Initialize maps if they're nil
	if override.Features == nil {
		override.Features = make(map[string]interface{})
	}
	if override.Customizations == nil {
		override.Customizations = make(map[string]interface{})
	}
	if override.Extra == nil {
		override.Extra = make(map[string]interface{})
	}

	return &override, nil
}

// mergeDockerComposeOverrides merges two docker-compose overrides with second taking precedence
func (m *Manager) mergeDockerComposeOverrides(base, override *DockerComposeOverride) *DockerComposeOverride {
	result := &DockerComposeOverride{
		Version:  base.Version,
		Services: make(map[string]ServiceOverride),
		Networks: make(map[string]interface{}),
		Volumes:  make(map[string]interface{}),
		Secrets:  make(map[string]interface{}),
	}

	// Override version if specified
	if override.Version != "" {
		result.Version = override.Version
	}

	// Merge services
	for name, service := range base.Services {
		result.Services[name] = service
	}
	for name, service := range override.Services {
		if existing, exists := result.Services[name]; exists {
			result.Services[name] = m.mergeServiceOverrides(existing, service)
		} else {
			result.Services[name] = service
		}
	}

	// Merge networks
	for name, network := range base.Networks {
		result.Networks[name] = network
	}
	for name, network := range override.Networks {
		result.Networks[name] = network // Override takes precedence
	}

	// Merge volumes
	for name, volume := range base.Volumes {
		result.Volumes[name] = volume
	}
	for name, volume := range override.Volumes {
		result.Volumes[name] = volume // Override takes precedence
	}

	// Merge secrets
	for name, secret := range base.Secrets {
		result.Secrets[name] = secret
	}
	for name, secret := range override.Secrets {
		result.Secrets[name] = secret // Override takes precedence
	}

	return result
}

// mergeServiceOverrides merges two service overrides with second taking precedence
func (m *Manager) mergeServiceOverrides(base, override ServiceOverride) ServiceOverride {
	result := base

	// Override individual fields if specified in override
	if override.Image != "" {
		result.Image = override.Image
	}
	if override.Command != nil {
		result.Command = override.Command
	}
	if override.Networks != nil {
		result.Networks = override.Networks
	}
	if override.DependsOn != nil {
		result.DependsOn = override.DependsOn
	}

	// Merge environment variables
	if result.Environment == nil {
		result.Environment = make(map[string]string)
	}
	for key, value := range override.Environment {
		result.Environment[key] = value
	}

	// Merge labels
	if result.Labels == nil {
		result.Labels = make(map[string]string)
	}
	for key, value := range override.Labels {
		result.Labels[key] = value
	}

	// Append ports and volumes (allowing duplicates to be filtered later)
	result.Ports = append(result.Ports, override.Ports...)
	result.Volumes = append(result.Volumes, override.Volumes...)

	// Merge extra fields
	if result.Extra == nil {
		result.Extra = make(map[string]interface{})
	}
	for key, value := range override.Extra {
		result.Extra[key] = value
	}

	return result
}

// mergeDevcontainerOverrides merges two devcontainer overrides with second taking precedence
func (m *Manager) mergeDevcontainerOverrides(base, override *DevcontainerOverride) *DevcontainerOverride {
	result := &DevcontainerOverride{
		Features:       make(map[string]interface{}),
		Customizations: make(map[string]interface{}),
		Extra:          make(map[string]interface{}),
	}

	// Copy base fields
	result.Name = base.Name
	result.Image = base.Image
	result.PostCreateCommand = base.PostCreateCommand
	result.PostStartCommand = base.PostStartCommand
	result.RemoteUser = base.RemoteUser
	result.WorkspaceFolder = base.WorkspaceFolder
	result.ForwardPorts = append([]int{}, base.ForwardPorts...)
	result.Mounts = append([]string{}, base.Mounts...)

	// Override individual fields if specified
	if override.Name != "" {
		result.Name = override.Name
	}
	if override.Image != "" {
		result.Image = override.Image
	}
	if override.PostCreateCommand != "" {
		result.PostCreateCommand = override.PostCreateCommand
	}
	if override.PostStartCommand != "" {
		result.PostStartCommand = override.PostStartCommand
	}
	if override.RemoteUser != "" {
		result.RemoteUser = override.RemoteUser
	}
	if override.WorkspaceFolder != "" {
		result.WorkspaceFolder = override.WorkspaceFolder
	}

	// Append arrays (allowing duplicates to be filtered later)
	result.ForwardPorts = append(result.ForwardPorts, override.ForwardPorts...)
	result.Mounts = append(result.Mounts, override.Mounts...)

	// Merge features
	for key, value := range base.Features {
		result.Features[key] = value
	}
	for key, value := range override.Features {
		result.Features[key] = value // Override takes precedence
	}

	// Merge customizations recursively
	for key, value := range base.Customizations {
		result.Customizations[key] = value
	}
	for key, value := range override.Customizations {
		if existing, exists := result.Customizations[key]; exists {
			// If both are maps, merge them recursively
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if overrideMap, ok := value.(map[string]interface{}); ok {
					merged := make(map[string]interface{})
					for k, v := range existingMap {
						merged[k] = v
					}
					for k, v := range overrideMap {
						merged[k] = v
					}
					result.Customizations[key] = merged
					continue
				}
			}
		}
		result.Customizations[key] = value // Override takes precedence
	}

	// Merge extra fields
	for key, value := range base.Extra {
		result.Extra[key] = value
	}
	for key, value := range override.Extra {
		result.Extra[key] = value
	}

	return result
}
