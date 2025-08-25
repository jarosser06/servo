package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

func TestDevcontainerGeneration_BasicManifest(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupDevcontainerTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createBasicDevcontainerManifest(); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyBasicDevcontainer(t)
}

func TestDevcontainerGeneration_WithMCPServers(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupDevcontainerTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createMCPServerManifest(); err != nil {
		t.Fatalf("Failed to create MCP server manifest: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyDevcontainerWithMCPServers(t)
}

func TestDevcontainerGeneration_WithPortForwarding(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-ports-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupDevcontainerTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createPortForwardingManifest(); err != nil {
		t.Fatalf("Failed to create port forwarding manifest: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyDevcontainerWithPorts(t)
}

func TestDevcontainerGeneration_WithFeatures(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-features-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupDevcontainerTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createFeaturesManifest(); err != nil {
		t.Fatalf("Failed to create features manifest: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyDevcontainerWithFeatures(t)
}

func TestDevcontainerGeneration_EmptyManifests(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupDevcontainerTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyMinimalDevcontainer(t)
}

// Helper functions

func setupDevcontainerTestProject() error {
	dirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/test",
		".servo/sessions/test/manifests",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	project := &project.Project{
		Clients:        []string{"vscode"},
		DefaultSession: "test",
		ActiveSession:  "test",
	}

	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".servo/project.yaml", data, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(".servo/active_session", []byte("test"), 0644); err != nil {
		return err
	}

	sessionInfo := map[string]interface{}{
		"name":        "test",
		"description": "Test session",
		"active":      true,
		"created_at":  "2024-01-01T00:00:00Z",
	}

	sessionData, err := yaml.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/session.yaml", sessionData, 0644)
}

func createBasicDevcontainerManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "basic-app",
		Description:  "Basic application for testing",
		Services: map[string]*pkg.ServiceDependency{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80"},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/basic-app.servo", data, 0644)
}

func createMCPServerManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "mcp-server",
		Server: pkg.Server{
			Transport: "stdio",
			Command:   "python",
			Args:      []string{"-m", "myserver", "--port", "8080"},
			Environment: map[string]string{
				"API_KEY": "test-key",
			},
		},
		Requirements: &pkg.Requirements{
			Runtimes: []pkg.RuntimeRequirement{
				{
					Name:    "python",
					Version: "3.11",
				},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/mcp-server.servo", data, 0644)
}

func createPortForwardingManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "port-app",
		Services: map[string]*pkg.ServiceDependency{
			"api": {
				Image: "node:16",
				Ports: []string{"3000:3000", "3001:3001"},
			},
			"db": {
				Image: "postgres:13",
				Ports: []string{"5432:5432"},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/port-app.servo", data, 0644)
}

func createFeaturesManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "features-app",
		Requirements: &pkg.Requirements{
			Runtimes: []pkg.RuntimeRequirement{
				{
					Name:    "node",
					Version: "18",
				},
				{
					Name:    "docker",
					Version: "latest",
				},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/features-app.servo", data, 0644)
}

func verifyBasicDevcontainer(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	if config["name"] != "Servo Development Environment" {
		t.Errorf("Expected name 'Servo Development Environment', got %v", config["name"])
	}

	dockerComposeFile, ok := config["dockerComposeFile"].([]interface{})
	if !ok || len(dockerComposeFile) != 1 || dockerComposeFile[0] != "docker-compose.yml" {
		t.Errorf("Expected dockerComposeFile ['docker-compose.yml'], got %v", config["dockerComposeFile"])
	}

	if config["service"] != "workspace" {
		t.Errorf("Expected service 'workspace', got %v", config["service"])
	}

	if config["workspaceFolder"] != "/workspace" {
		t.Errorf("Expected workspaceFolder '/workspace', got %v", config["workspaceFolder"])
	}
}

func verifyDevcontainerWithMCPServers(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	// Check features were added
	features, ok := config["features"].(map[string]interface{})
	if !ok {
		t.Fatal("Features section not found")
	}

	if _, exists := features["ghcr.io/devcontainers/features/python:1"]; !exists {
		t.Error("Python feature not found in devcontainer features")
	}

	// Devcontainer should NOT have client-specific customizations
	// Those are handled by individual clients, not the devcontainer generator
	if customizations, exists := config["customizations"]; exists {
		t.Errorf("Devcontainer should not include client customizations, found: %v", customizations)
	}
}

func verifyDevcontainerWithPorts(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	// Check port forwarding
	forwardPorts, ok := config["forwardPorts"].([]interface{})
	if !ok {
		t.Fatal("forwardPorts not found")
	}

	expectedPorts := []int{3000, 3001, 5432}
	if len(forwardPorts) != len(expectedPorts) {
		t.Errorf("Expected %d forwarded ports, got %d", len(expectedPorts), len(forwardPorts))
	}

	portMap := make(map[int]bool)
	for _, port := range forwardPorts {
		if portFloat, ok := port.(float64); ok {
			portMap[int(portFloat)] = true
		}
	}

	for _, expectedPort := range expectedPorts {
		if !portMap[expectedPort] {
			t.Errorf("Expected port %d not found in forwardPorts", expectedPort)
		}
	}
}

func verifyDevcontainerWithFeatures(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	// Check features
	features, ok := config["features"].(map[string]interface{})
	if !ok {
		t.Fatal("Features section not found")
	}

	expectedFeatures := []string{
		"ghcr.io/devcontainers/features/node:1",
		"ghcr.io/devcontainers/features/docker-outside-of-docker:1",
	}

	for _, featureName := range expectedFeatures {
		if _, exists := features[featureName]; !exists {
			t.Errorf("Feature %s not found", featureName)
		}
	}

	// Verify feature configuration
	nodeFeature, ok := features["ghcr.io/devcontainers/features/node:1"].(map[string]interface{})
	if !ok {
		t.Error("Node feature configuration not found")
	} else if nodeFeature["version"] != "18" {
		t.Errorf("Expected Node version '18', got %v", nodeFeature["version"])
	}

	dockerFeature, ok := features["ghcr.io/devcontainers/features/docker-outside-of-docker:1"].(map[string]interface{})
	if !ok {
		t.Error("Docker feature configuration not found")
	} else if dockerFeature["version"] != "latest" {
		t.Errorf("Expected Docker version 'latest', got %v", dockerFeature["version"])
	}
}

func verifyMinimalDevcontainer(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	// Should have basic structure
	if config["name"] != "Servo Development Environment" {
		t.Errorf("Expected name 'Servo Development Environment', got %v", config["name"])
	}

	// Should have minimal required fields
	requiredFields := []string{"dockerComposeFile", "service", "workspaceFolder"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			t.Errorf("Required field %s not found", field)
		}
	}

	// Features should be empty or minimal
	features, ok := config["features"].(map[string]interface{})
	if ok && len(features) > 1 { // May have base features
		t.Logf("Features found in minimal config: %v", features)
	}
}
