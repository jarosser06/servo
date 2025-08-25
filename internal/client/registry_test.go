package client

import (
	"testing"

	"github.com/servo/servo/pkg"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	name            string
	description     string
	platforms       []string
	installed       bool
	supportedScopes []pkg.ClientScope
}

func (m *MockClient) Name() string                          { return m.name }
func (m *MockClient) Description() string                   { return m.description }
func (m *MockClient) SupportedPlatforms() []string          { return m.platforms }
func (m *MockClient) IsInstalled() bool                     { return m.installed }
func (m *MockClient) GetVersion() (string, error)           { return "1.0.0", nil }
func (m *MockClient) GetSupportedScopes() []pkg.ClientScope { return m.supportedScopes }
func (m *MockClient) ValidateScope(scope string) error {
	for _, s := range m.supportedScopes {
		if string(s) == scope {
			return nil
		}
	}
	return pkg.ErrInvalidScope
}
func (m *MockClient) GetCurrentConfig(scope string) (*pkg.MCPConfig, error) {
	return &pkg.MCPConfig{}, nil
}
func (m *MockClient) ValidateConfig(scope string) error                  { return nil }
func (m *MockClient) RequiresRestart() bool                              { return false }
func (m *MockClient) TriggerReload() error                               { return nil }
func (m *MockClient) GetLaunchCommand(projectPath string) string         { return "test-command" }
func (m *MockClient) SupportsDevcontainers() bool                        { return true }
func (m *MockClient) IsPlatformSupported() bool                          { return true }
func (m *MockClient) RemoveServer(scope string, serverName string) error { return nil }
func (m *MockClient) ListServers(scope string) ([]string, error)         { return []string{}, nil }
func (m *MockClient) GenerateConfig(manifests []pkg.ServoDefinition, secretsProvider func(string) (string, error)) error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	client := &MockClient{
		name:        "test-client",
		description: "Test client",
		platforms:   []string{"linux", "darwin"},
	}

	err := registry.Register(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(client)
	if err == nil {
		t.Error("Expected error when registering duplicate client")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	client := &MockClient{
		name:        "test-client",
		description: "Test client",
	}

	registry.Register(client)

	retrieved, err := registry.Get("test-client")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if retrieved.Name() != "test-client" {
		t.Errorf("Expected client name 'test-client', got '%s'", retrieved.Name())
	}

	// Test non-existent client
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent client")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	client1 := &MockClient{name: "client1"}
	client2 := &MockClient{name: "client2"}

	registry.Register(client1)
	registry.Register(client2)

	clients := registry.List()
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
}

func TestRegistry_GetSupportingScope(t *testing.T) {
	registry := NewRegistry()

	localOnlyClient := &MockClient{
		name: "local-only",
		supportedScopes: []pkg.ClientScope{
			pkg.LocalScope,
		},
	}

	localScopeClient := &MockClient{
		name: "local-scopes",
		supportedScopes: []pkg.ClientScope{
			pkg.LocalScope,
		},
	}

	registry.Register(localOnlyClient)
	registry.Register(localScopeClient)

	localSupporting := registry.GetSupportingScope("local")
	if len(localSupporting) != 2 {
		t.Errorf("Expected 2 clients supporting local scope, got %d", len(localSupporting))
	}

	globalSupporting := registry.GetSupportingScope("global")
	if len(globalSupporting) != 0 {
		t.Errorf("Expected 0 clients supporting global scope, got %d", len(globalSupporting))
	}

	// Check that both expected clients are present (order is not guaranteed)
	foundLocalScopes := false
	foundLocalOnly := false
	for _, client := range localSupporting {
		if client.Name() == "local-scopes" {
			foundLocalScopes = true
		}
		if client.Name() == "local-only" {
			foundLocalOnly = true
		}
	}
	if !foundLocalScopes {
		t.Errorf("Expected 'local-scopes' client to be in supporting clients")
	}
	if !foundLocalOnly {
		t.Errorf("Expected 'local-only' client to be in supporting clients")
	}
}

func TestRegistry_GetInstalledClients(t *testing.T) {
	registry := NewRegistry()

	installedClient := &MockClient{
		name:      "installed",
		installed: true,
	}

	notInstalledClient := &MockClient{
		name:      "not-installed",
		installed: false,
	}

	registry.Register(installedClient)
	registry.Register(notInstalledClient)

	installed := registry.GetInstalledClients()
	if len(installed) != 1 {
		t.Errorf("Expected 1 installed client, got %d", len(installed))
	}

	if installed[0].Name() != "installed" {
		t.Errorf("Expected 'installed' client, got '%s'", installed[0].Name())
	}
}
