package pkg

import (
	"strings"
	"testing"
)

func TestClientScope_String(t *testing.T) {
	local := LocalScope

	if string(local) != "local" {
		t.Errorf("expected local scope to be 'local', got '%s'", string(local))
	}
}

func TestErrConstants(t *testing.T) {
	if ErrInvalidScope == nil {
		t.Errorf("ErrInvalidScope should not be nil")
	}

	if ErrConfigNotFound == nil {
		t.Errorf("ErrConfigNotFound should not be nil")
	}

	if !strings.Contains(ErrInvalidScope.Error(), "invalid scope") {
		t.Errorf("ErrInvalidScope should contain 'invalid scope', got '%s'", ErrInvalidScope.Error())
	}

	if !strings.Contains(ErrConfigNotFound.Error(), "configuration not found") {
		t.Errorf("ErrConfigNotFound should contain 'configuration not found', got '%s'", ErrConfigNotFound.Error())
	}
}

func TestServoDefinition_Validation(t *testing.T) {
	// Test that ServoDefinition struct can be created with required fields
	servo := &ServoDefinition{
		ServoVersion: "1.0",
		Name:         "test-server",
		Version:      "1.0.0",
		Description:  "Test server",
		Author:       "Test Author",
		License:      "MIT",
		Install: Install{
			Type:   "local",
			Method: "local",
		},
		Server: Server{
			Transport: "stdio",
			Command:   "test",
			Args:      []string{"--stdio"},
		},
	}

	if servo.ServoVersion != "1.0" {
		t.Errorf("expected servo version '1.0', got '%s'", servo.ServoVersion)
	}

	if servo.Name != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", servo.Name)
	}

	if servo.Install.Type != "local" {
		t.Errorf("expected install type 'local', got '%s'", servo.Install.Type)
	}

	if servo.Server.Transport != "stdio" {
		t.Errorf("expected server transport 'stdio', got '%s'", servo.Server.Transport)
	}
}

func TestConfiguration_Defaults(t *testing.T) {
	config := &Configuration{
		Version: "1.0",
		Scope:   "local",
		Settings: Settings{
			LogRetentionDays:    30,
			AutoStartServices:   false,
			InheritGlobal:       true,
			UpdateCheckInterval: "daily",
			SecretsEncryption:   true,
		},
		Clients: make(ClientConfigs),
		Servers: make(ServerConfigs),
	}

	if config.Settings.LogRetentionDays != 30 {
		t.Errorf("expected log retention days 30, got %d", config.Settings.LogRetentionDays)
	}

	if config.Settings.AutoStartServices != false {
		t.Errorf("expected auto start services false, got %v", config.Settings.AutoStartServices)
	}

	if config.Settings.InheritGlobal != true {
		t.Errorf("expected inherit global true, got %v", config.Settings.InheritGlobal)
	}
}

func TestMCPServerConfig_Fields(t *testing.T) {
	config := MCPServerConfig{
		Command:     "test-command",
		Args:        []string{"--arg1", "--arg2"},
		Environment: map[string]string{"ENV1": "value1", "ENV2": "value2"},
	}

	if config.Command != "test-command" {
		t.Errorf("expected command 'test-command', got '%s'", config.Command)
	}

	if len(config.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(config.Args))
	}

	if config.Args[0] != "--arg1" {
		t.Errorf("expected first arg '--arg1', got '%s'", config.Args[0])
	}

	if len(config.Environment) != 2 {
		t.Errorf("expected 2 environment variables, got %d", len(config.Environment))
	}

	if config.Environment["ENV1"] != "value1" {
		t.Errorf("expected ENV1='value1', got '%s'", config.Environment["ENV1"])
	}
}

func TestServerConfig_Fields(t *testing.T) {
	config := ServerConfig{
		Version:         "1.0.0",
		Clients:         []string{"vscode", "cursor"},
		Config:          map[string]interface{}{"debug": true, "port": 8080},
		SecretsRequired: []string{"api_key", "database_url"},
	}

	if config.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", config.Version)
	}

	if len(config.Clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(config.Clients))
	}

	if config.Clients[0] != "vscode" {
		t.Errorf("expected first client 'vscode', got '%s'", config.Clients[0])
	}

	if len(config.SecretsRequired) != 2 {
		t.Errorf("expected 2 required secrets, got %d", len(config.SecretsRequired))
	}

	if config.Config["debug"] != true {
		t.Errorf("expected debug=true, got '%v'", config.Config["debug"])
	}

	if config.Config["port"] != 8080 {
		t.Errorf("expected port=8080, got '%v'", config.Config["port"])
	}
}

func TestClientConfig_Fields(t *testing.T) {
	config := ClientConfig{
		Enabled:       true,
		AutoConfigure: false,
		ConfigPath:    "/path/to/config",
		PluginPath:    "/path/to/plugin",
		Extensions:    []string{"ext1", "ext2"},
	}

	if !config.Enabled {
		t.Error("expected client to be enabled")
	}

	if config.AutoConfigure {
		t.Error("expected auto configure to be false")
	}

	if config.ConfigPath != "/path/to/config" {
		t.Errorf("expected config path '/path/to/config', got '%s'", config.ConfigPath)
	}

	if len(config.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(config.Extensions))
	}
}

func TestInstall_Fields(t *testing.T) {
	install := Install{
		Type:          "docker",
		Method:        "compose",
		Repository:    "https://github.com/user/repo",
		Subdirectory:  "src/server",
		SetupCommands: []string{"npm install", "npm run build"},
		BuildCommands: []string{"docker build", "docker tag"},
		TestCommands:  []string{"npm test", "docker run --rm"},
	}

	if install.Type != "docker" {
		t.Errorf("expected type 'docker', got '%s'", install.Type)
	}

	if install.Method != "compose" {
		t.Errorf("expected method 'compose', got '%s'", install.Method)
	}

	if install.Repository != "https://github.com/user/repo" {
		t.Errorf("expected repository 'https://github.com/user/repo', got '%s'", install.Repository)
	}

	if install.Subdirectory != "src/server" {
		t.Errorf("expected subdirectory 'src/server', got '%s'", install.Subdirectory)
	}

	if len(install.SetupCommands) != 2 {
		t.Errorf("expected 2 setup commands, got %d", len(install.SetupCommands))
	}

	if len(install.BuildCommands) != 2 {
		t.Errorf("expected 2 build commands, got %d", len(install.BuildCommands))
	}

	if len(install.TestCommands) != 2 {
		t.Errorf("expected 2 test commands, got %d", len(install.TestCommands))
	}
}
