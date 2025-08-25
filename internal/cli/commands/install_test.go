package commands

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/servo/servo/internal/mcp"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/session"
	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

func TestInstallCommand_ManifestStorage(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "servo_install_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := setupInstallTestProject(t); err != nil {
		t.Fatalf("Failed to setup project: %v", err)
	}

	// Create test .servo file
	mockServoContent := `servo_version: "1.0"
name: "test-server"
version: "1.0.0"
description: "Test MCP server"

server:
  command: "python"
  args: ["-m", "test_server.main"]
  environment:
    TEST_VAR: "test_value"

services:
  database:
    image: "postgres:15"
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: "testdb"
      POSTGRES_USER: "user"
      POSTGRES_PASSWORD: "password"`

	if err := os.WriteFile("test-server.servo", []byte(mockServoContent), 0644); err != nil {
		t.Fatalf("Failed to create servo file: %v", err)
	}

	// Create command
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewInstallCommand(parser, validator)

	// Run install
	args := []string{"test-server.servo", "--clients", "vscode"}
	if err := cmd.Execute(args); err != nil {
		t.Fatalf("Install command failed: %v", err)
	}

	// Verify manifest was stored in session
	manifestPath := ".servo/sessions/default/manifests/test-server.servo"
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("Manifest not stored in session: %s", manifestPath)
	}

	// Verify manifest content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read stored manifest: %v", err)
	}

	var storedManifest pkg.ServoDefinition
	if err := yaml.Unmarshal(data, &storedManifest); err != nil {
		t.Fatalf("Failed to parse stored manifest: %v", err)
	}

	if storedManifest.Name != "test-server" {
		t.Errorf("Expected manifest name 'test-server', got %s", storedManifest.Name)
	}

	// Verify VSCode config was generated
	vscodeConfigPath := ".vscode/mcp.json"
	if _, err := os.Stat(vscodeConfigPath); os.IsNotExist(err) {
		t.Errorf("VSCode config not generated: %s", vscodeConfigPath)
	}

	// Verify devcontainer was generated
	devcontainerPath := ".devcontainer/devcontainer.json"
	if _, err := os.Stat(devcontainerPath); os.IsNotExist(err) {
		t.Errorf("Devcontainer config not generated: %s", devcontainerPath)
	}

	// Verify docker-compose was generated
	dockerComposePath := ".devcontainer/docker-compose.yml"
	if _, err := os.Stat(dockerComposePath); os.IsNotExist(err) {
		t.Errorf("Docker compose config not generated: %s", dockerComposePath)
	}
}

func TestInstallCommand_ConfigGeneration(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "servo_install_config_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := setupInstallTestProject(t); err != nil {
		t.Fatalf("Failed to setup project: %v", err)
	}

	// Create test .servo file with services
	mockServoContent := `servo_version: "1.0"
name: "graphiti-server"
version: "1.0.0"

server:
  command: "python"
  args: ["-m", "graphiti.main"]
  environment:
    NEO4J_URI: "neo4j://localhost:7687"

services:
  neo4j:
    image: "neo4j:5"
    ports:
      - "7474:7474"
      - "7687:7687"
    environment:
      NEO4J_AUTH: "neo4j/password"
    volumes:
      - "neo4j_data:/data"`

	if err := os.WriteFile("graphiti-server.servo", []byte(mockServoContent), 0644); err != nil {
		t.Fatalf("Failed to create servo file: %v", err)
	}

	// Create command and run install
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewInstallCommand(parser, validator)

	args := []string{"graphiti-server.servo", "--clients", "vscode,claude-code"}
	if err := cmd.Execute(args); err != nil {
		t.Fatalf("Install command failed: %v", err)
	}

	// Test VSCode MCP config
	vscodeConfig, err := os.ReadFile(".vscode/mcp.json")
	if err != nil {
		t.Fatalf("Failed to read VSCode config: %v", err)
	}

	var vscodeData map[string]interface{}
	if err := json.Unmarshal(vscodeConfig, &vscodeData); err != nil {
		t.Fatalf("Failed to parse VSCode config: %v", err)
	}

	servers, ok := vscodeData["servers"].(map[string]interface{})
	if !ok || len(servers) == 0 {
		t.Error("VSCode config missing servers")
	}

	if _, exists := servers["graphiti-server"]; !exists {
		t.Error("graphiti-server not found in VSCode config")
	}

	// Test Claude Code MCP config
	claudeConfig, err := os.ReadFile(".mcp.json")
	if err != nil {
		t.Fatalf("Failed to read Claude Code config: %v", err)
	}

	var claudeData map[string]interface{}
	if err := json.Unmarshal(claudeConfig, &claudeData); err != nil {
		t.Fatalf("Failed to parse Claude Code config: %v", err)
	}

	claudeServers, ok := claudeData["mcpServers"].(map[string]interface{})
	if !ok || len(claudeServers) == 0 {
		t.Error("Claude Code config missing mcpServers")
	}

	if _, exists := claudeServers["graphiti-server"]; !exists {
		t.Error("graphiti-server not found in Claude Code config")
	}

	// Test docker-compose generation
	dockerCompose, err := os.ReadFile(".devcontainer/docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to read docker-compose: %v", err)
	}

	var composeData map[string]interface{}
	if err := yaml.Unmarshal(dockerCompose, &composeData); err != nil {
		t.Fatalf("Failed to parse docker-compose: %v", err)
	}

	services, ok := composeData["services"].(map[string]interface{})
	if !ok {
		t.Fatal("docker-compose missing services section")
	}

	// Check that Neo4j service was added with correct prefix
	if _, exists := services["graphiti-server-neo4j"]; !exists {
		t.Error("graphiti-server-neo4j service not found in docker-compose")
	}
}

func TestInstallCommand_SessionSpecific(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "servo_session_install_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := setupInstallTestProject(t); err != nil {
		t.Fatalf("Failed to setup project: %v", err)
	}

	// Create test .servo file
	mockServoContent := `servo_version: "1.0"
name: "dev-server"
version: "1.0.0"

server:
  command: "python"
  args: ["-m", "dev_server.main"]`

	if err := os.WriteFile("dev-server.servo", []byte(mockServoContent), 0644); err != nil {
		t.Fatalf("Failed to create servo file: %v", err)
	}

	// Create the development session first
	sessionManager := session.NewManager(".servo")
	_, sessionErr := sessionManager.Create("development", "Development session", "")
	if sessionErr != nil {
		t.Fatalf("Failed to create development session: %v", sessionErr)
	}

	// Create command
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewInstallCommand(parser, validator)

	// Install to specific session
	args := []string{"dev-server.servo"}
	clients := []string{"vscode"}
	sessionName := "development"
	if err := cmd.ExecuteWithOptions(args, clients, sessionName, false); err != nil {
		t.Fatalf("Install command failed: %v", err)
	}

	// Verify manifest was stored in development session
	manifestPath := ".servo/sessions/development/manifests/dev-server.servo"
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("Manifest not stored in development session: %s", manifestPath)
	}

	// Verify session directory structure was created
	expectedDirs := []string{
		".servo/sessions/development",
		".servo/sessions/development/manifests",
		".servo/sessions/development/config",
		".servo/sessions/development/volumes",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	// Verify configs were generated
	expectedFiles := []string{
		".vscode/mcp.json",
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected config file not generated: %s", file)
		}
	}
}

func TestInstallCommand_GitRepo(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "servo_git_install_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := setupInstallTestProject(t); err != nil {
		t.Fatalf("Failed to setup project: %v", err)
	}

	// Create command
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewInstallCommand(parser, validator)

	// Test installing from GitHub repo (this should work with public repos)
	args := []string{"github.com/servo/test-mcp-server", "--clients", "vscode"}

	err = cmd.Execute(args)

	// We expect this to either succeed or fail with a network/auth error, not a parsing error
	if err != nil && !strings.Contains(err.Error(), "failed to clone") && !strings.Contains(err.Error(), "authentication") && !strings.Contains(err.Error(), "network") {
		t.Errorf("Unexpected error type (should be network/auth related): %v", err)
	}
}

func TestInstallCommand_InvalidYAML(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "servo_validate_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Initialize project structure
	if err := setupInstallTestProject(t); err != nil {
		t.Fatalf("Failed to setup project: %v", err)
	}

	// Create invalid YAML file (malformed syntax)
	invalidServoContent := `servo_version: "1.0"
name: "invalid-server"
invalid_yaml: [
  this is not valid yaml
  missing closing bracket`

	if err := os.WriteFile("invalid-server.servo", []byte(invalidServoContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid servo file: %v", err)
	}

	// Create command
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewInstallCommand(parser, validator)

	// Try to install invalid manifest - should fail during parsing
	args := []string{"invalid-server.servo", "--clients", "vscode"}
	err = cmd.Execute(args)

	if err == nil {
		t.Error("Expected install to fail with invalid YAML, but it succeeded")
		return
	}

	if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "yaml") {
		t.Errorf("Expected parsing error, got: %v", err)
	}
}

// Helper function to setup test project structure
func setupInstallTestProject(t *testing.T) error {
	// Create .servo directory structure
	dirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/default",
		".servo/sessions/default/manifests",
		".servo/sessions/default/config",
		".servo/sessions/default/docker",
		".servo/sessions/default/devcontainer",
		".servo/sessions/default/volumes",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create project.yaml
	projectData := &project.Project{
		Clients:        []string{"vscode"},
		DefaultSession: "default",
		ActiveSession:  "default",
	}

	data, err := yaml.Marshal(projectData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".servo/project.yaml", data, 0644); err != nil {
		return err
	}

	// Create active session file
	if err := os.WriteFile(".servo/active_session", []byte("default"), 0644); err != nil {
		return err
	}

	// Create default session file
	sessionInfo := map[string]interface{}{
		"name":        "default",
		"description": "Default session",
		"active":      true,
		"created_at":  "2024-01-01T00:00:00Z",
	}

	sessionData, err := yaml.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/default/session.yaml", sessionData, 0644)
}
