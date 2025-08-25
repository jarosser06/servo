package config

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

func TestDockerComposeOverrideGeneration_BasicMerging(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-override-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupOverrideTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createBasicDockerComposeManifest(); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	if err := createDockerComposeOverride(); err != nil {
		t.Fatalf("Failed to create docker-compose override: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDockerCompose(); err != nil {
		t.Fatalf("Failed to generate docker-compose: %v", err)
	}

	verifyDockerComposeOverrideMerging(t)
}

func TestDevcontainerOverrideGeneration_BasicMerging(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-devcontainer-override-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupOverrideTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createBasicDevcontainerOverrideManifest(); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	if err := createDevcontainerOverride(); err != nil {
		t.Fatalf("Failed to create devcontainer override: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDevcontainer(); err != nil {
		t.Fatalf("Failed to generate devcontainer: %v", err)
	}

	verifyDevcontainerOverrideMerging(t)
}

func TestOverrideGeneration_PrecedenceOrder(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-precedence-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupOverrideTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createBasicDockerComposeManifest(); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	if err := createMultipleOverrides(); err != nil {
		t.Fatalf("Failed to create multiple overrides: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDockerCompose(); err != nil {
		t.Fatalf("Failed to generate docker-compose: %v", err)
	}

	verifyPrecedenceOrder(t)
}

func TestOverrideGeneration_ComplexServiceMerging(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-complex-override-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupOverrideTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createComplexDockerComposeManifest(); err != nil {
		t.Fatalf("Failed to create complex manifest: %v", err)
	}

	if err := createComplexServiceOverride(); err != nil {
		t.Fatalf("Failed to create complex service override: %v", err)
	}

	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDockerCompose(); err != nil {
		t.Fatalf("Failed to generate docker-compose: %v", err)
	}

	verifyComplexServiceMerging(t)
}

// Helper functions

func setupOverrideTestProject() error {
	dirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/test",
		".servo/sessions/test/manifests",
		".servo/sessions/test/config",
		".servo/config",
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

func createBasicDockerComposeManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "basic-app",
		Services: map[string]*pkg.ServiceDependency{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80"},
				Environment: map[string]string{
					"ENV": "production",
				},
			},
			"api": {
				Image: "node:16",
				Ports: []string{"3000:3000"},
				Environment: map[string]string{
					"NODE_ENV": "production",
					"PORT":     "3000",
				},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/basic-app.servo", data, 0644)
}

func createBasicDevcontainerOverrideManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "dev-app",
		Requirements: &pkg.Requirements{
			Runtimes: []pkg.RuntimeRequirement{
				{
					Name:    "node",
					Version: "16",
				},
			},
		},
		Services: map[string]*pkg.ServiceDependency{
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

	return os.WriteFile(".servo/sessions/test/manifests/dev-app.servo", data, 0644)
}

func createDockerComposeOverride() error {
	override := map[string]interface{}{
		"services": map[string]interface{}{
			"basic-app-web": map[string]interface{}{
				"environment": map[string]string{
					"DEBUG": "true",
					"ENV":   "development", // Override production value
				},
				"ports": []string{"8080:80"}, // Override default port
			},
			"basic-app-api": map[string]interface{}{
				"volumes": []string{
					"./src:/app/src:ro",
				},
			},
			"cache": map[string]interface{}{
				"image": "redis:alpine",
				"ports": []string{"6379:6379"},
			},
		},
		"networks": map[string]interface{}{
			"override-network": map[string]interface{}{
				"driver": "bridge",
			},
		},
	}

	data, err := yaml.Marshal(override)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/config/docker-compose.yml", data, 0644)
}

func createDevcontainerOverride() error {
	override := map[string]interface{}{
		"features": map[string]interface{}{
			"ghcr.io/devcontainers/features/docker-outside-of-docker:1": map[string]interface{}{
				"version": "latest",
			},
		},
		"customizations": map[string]interface{}{
			"vscode": map[string]interface{}{
				"settings": map[string]interface{}{
					"terminal.integrated.shell.linux": "/bin/bash",
				},
				"extensions": []string{
					"ms-vscode.vscode-json",
				},
			},
		},
		"forwardPorts": []int{8080, 9000},
	}

	data, err := json.Marshal(override)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/config/devcontainer.json", data, 0644)
}

func createComplexDockerComposeManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "complex-app",
		Services: map[string]*pkg.ServiceDependency{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80", "443:443"},
				Environment: map[string]string{
					"ENV":      "production",
					"SSL_MODE": "enabled",
				},
				Volumes: []string{
					"./nginx.conf:/etc/nginx/nginx.conf:ro",
				},
			},
			"api": {
				Image: "node:16",
				Ports: []string{"3000:3000"},
				Environment: map[string]string{
					"NODE_ENV": "production",
					"PORT":     "3000",
				},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/complex-app.servo", data, 0644)
}

func createMultipleOverrides() error {
	// Project-level override
	projectOverride := map[string]interface{}{
		"services": map[string]interface{}{
			"basic-app-web": map[string]interface{}{
				"environment": map[string]string{
					"ENV":   "staging",
					"DEBUG": "false",
				},
			},
		},
	}

	projectData, err := yaml.Marshal(projectOverride)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".servo/config/docker-compose.yml", projectData, 0644); err != nil {
		return err
	}

	// Session-level override (higher precedence)
	sessionOverride := map[string]interface{}{
		"services": map[string]interface{}{
			"basic-app-web": map[string]interface{}{
				"environment": map[string]string{
					"ENV": "development", // Should override both manifest and project
				},
			},
		},
	}

	sessionData, err := yaml.Marshal(sessionOverride)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/config/docker-compose.yml", sessionData, 0644)
}

func createComplexServiceOverride() error {
	override := map[string]interface{}{
		"services": map[string]interface{}{
			"complex-app-web": map[string]interface{}{
				"environment": map[string]string{
					"ENV":       "development",
					"LOG_LEVEL": "debug",
				},
				"volumes": []string{
					"./logs:/var/log/nginx",
					"./static:/usr/share/nginx/html:ro",
				},
				"command":    []string{"nginx", "-g", "daemon off;"},
				"depends_on": []string{"complex-app-api"},
			},
			"complex-app-api": map[string]interface{}{
				"build": map[string]interface{}{
					"context":    ".",
					"dockerfile": "Dockerfile.dev",
				},
				"volumes": []string{
					"./src:/app/src",
					"./node_modules:/app/node_modules",
				},
				"command": []string{"npm", "run", "dev"},
			},
			"db": map[string]interface{}{
				"image": "postgres:14",
				"environment": map[string]string{
					"POSTGRES_DB":       "devdb",
					"POSTGRES_USER":     "dev",
					"POSTGRES_PASSWORD": "devpass",
				},
				"volumes": []string{
					"db_data:/var/lib/postgresql/data",
				},
				"ports": []string{"5432:5432"},
			},
		},
		"volumes": map[string]interface{}{
			"db_data": nil,
		},
		"networks": map[string]interface{}{
			"dev-network": map[string]interface{}{
				"driver": "bridge",
			},
		},
	}

	data, err := yaml.Marshal(override)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/config/docker-compose.yml", data, 0644)
}

// Verification functions

func verifyDockerComposeOverrideMerging(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse docker-compose.yml: %v", err)
	}

	services := config["services"].(map[string]interface{})

	// Check web service override
	webService := services["basic-app-web"].(map[string]interface{})
	webEnv := webService["environment"].([]interface{})

	// Parse environment variables from string array format
	envMap := make(map[string]string)
	for _, envVar := range webEnv {
		envStr := envVar.(string)
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG=true from override, got %v", envMap["DEBUG"])
	}

	if envMap["ENV"] != "development" {
		t.Errorf("Expected ENV=development from override, got %v", envMap["ENV"])
	}

	webPorts := webService["ports"].([]interface{})
	if len(webPorts) != 1 || webPorts[0] != "8080:80" {
		t.Errorf("Expected overridden port 8080:80, got %v", webPorts)
	}

	// Check api service override
	apiService := services["basic-app-api"].(map[string]interface{})
	if apiVolumes, ok := apiService["volumes"]; !ok {
		t.Error("Expected volumes to be added to api service")
	} else {
		volumes := apiVolumes.([]interface{})
		if len(volumes) != 1 || volumes[0] != "./src:/app/src:ro" {
			t.Errorf("Expected volume ./src:/app/src:ro, got %v", volumes)
		}
	}

	// Check new cache service
	if _, ok := services["cache"]; !ok {
		t.Error("Expected cache service to be added from override")
	}

	// Check networks
	if networks, ok := config["networks"]; !ok {
		t.Error("Expected networks section from override")
	} else {
		networksMap := networks.(map[string]interface{})
		if _, ok := networksMap["override-network"]; !ok {
			t.Error("Expected override-network to be added")
		}
	}
}

func verifyDevcontainerOverrideMerging(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/devcontainer.json")
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse devcontainer.json: %v", err)
	}

	// Check features from override
	features := config["features"].(map[string]interface{})
	if _, ok := features["ghcr.io/devcontainers/features/node:1"]; !ok {
		t.Error("Expected node feature from manifest")
	}

	if _, ok := features["ghcr.io/devcontainers/features/docker-outside-of-docker:1"]; !ok {
		t.Error("Expected docker feature from override")
	}

	// Check customizations
	customizations := config["customizations"].(map[string]interface{})
	vscode := customizations["vscode"].(map[string]interface{})
	settings := vscode["settings"].(map[string]interface{})

	if settings["terminal.integrated.shell.linux"] != "/bin/bash" {
		t.Error("Expected terminal shell setting from override")
	}

	extensions := vscode["extensions"].([]interface{})
	if len(extensions) == 0 || extensions[0] != "ms-vscode.vscode-json" {
		t.Error("Expected extension from override")
	}

	// Check forward ports
	forwardPorts := config["forwardPorts"].([]interface{})
	expectedPorts := map[float64]bool{5432: true, 8080: true, 9000: true}

	for _, port := range forwardPorts {
		portFloat := port.(float64)
		if !expectedPorts[portFloat] {
			t.Errorf("Unexpected port %v in forwardPorts", portFloat)
		}
		delete(expectedPorts, portFloat)
	}

	if len(expectedPorts) > 0 {
		t.Errorf("Missing ports in forwardPorts: %v", expectedPorts)
	}
}

func verifyPrecedenceOrder(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse docker-compose.yml: %v", err)
	}

	services := config["services"].(map[string]interface{})
	webService := services["basic-app-web"].(map[string]interface{})
	webEnv := webService["environment"].([]interface{})

	// Parse environment variables from string array format
	envMap := make(map[string]string)
	for _, envVar := range webEnv {
		envStr := envVar.(string)
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Session override should have highest precedence
	if envMap["ENV"] != "development" {
		t.Errorf("Expected ENV=development from session override (highest precedence), got %v", envMap["ENV"])
	}

	// DEBUG should come from project override since session doesn't override it
	if envMap["DEBUG"] != "false" {
		t.Errorf("Expected DEBUG=false from project override, got %v", envMap["DEBUG"])
	}
}

func verifyComplexServiceMerging(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse docker-compose.yml: %v", err)
	}

	services := config["services"].(map[string]interface{})

	// Check web service complex merging
	webService := services["complex-app-web"].(map[string]interface{})
	webEnv := webService["environment"].([]interface{})

	// Parse environment variables from string array format
	envMap := make(map[string]string)
	for _, envVar := range webEnv {
		envStr := envVar.(string)
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Original + override environment variables
	if envMap["ENV"] != "development" {
		t.Errorf("Expected ENV=development from override, got %v", envMap["ENV"])
	}

	if envMap["SSL_MODE"] != "enabled" {
		t.Errorf("Expected SSL_MODE=enabled from manifest, got %v", envMap["SSL_MODE"])
	}

	if envMap["LOG_LEVEL"] != "debug" {
		t.Errorf("Expected LOG_LEVEL=debug from override, got %v", envMap["LOG_LEVEL"])
	}

	// Check volumes merging
	webVolumes := webService["volumes"].([]interface{})
	expectedVolumes := []string{
		"./nginx.conf:/etc/nginx/nginx.conf:ro", // from manifest
		"./logs:/var/log/nginx",                 // from override
		"./static:/usr/share/nginx/html:ro",     // from override
	}

	if len(webVolumes) != len(expectedVolumes) {
		t.Errorf("Expected %d volumes, got %d", len(expectedVolumes), len(webVolumes))
	}

	// Check api service build override
	apiService := services["complex-app-api"].(map[string]interface{})
	if build, ok := apiService["build"]; !ok {
		t.Error("Expected build configuration from override")
	} else {
		buildMap := build.(map[string]interface{})
		if buildMap["dockerfile"] != "Dockerfile.dev" {
			t.Error("Expected Dockerfile.dev from override")
		}
	}

	// Check new db service
	if dbService, ok := services["db"]; !ok {
		t.Error("Expected db service from override")
	} else {
		db := dbService.(map[string]interface{})
		dbEnv := db["environment"].([]interface{})

		// Parse environment variables from string array format
		dbEnvMap := make(map[string]string)
		for _, envVar := range dbEnv {
			envStr := envVar.(string)
			parts := strings.SplitN(envStr, "=", 2)
			if len(parts) == 2 {
				dbEnvMap[parts[0]] = parts[1]
			}
		}

		if dbEnvMap["POSTGRES_DB"] != "devdb" {
			t.Error("Expected POSTGRES_DB=devdb from override")
		}
	}

	// Check volumes section
	if volumes, ok := config["volumes"]; !ok {
		t.Error("Expected volumes section from override")
	} else {
		volumesMap := volumes.(map[string]interface{})
		if _, ok := volumesMap["db_data"]; !ok {
			t.Error("Expected db_data volume from override")
		}
	}

	// Check networks section
	if networks, ok := config["networks"]; !ok {
		t.Error("Expected networks section from override")
	} else {
		networksMap := networks.(map[string]interface{})
		if _, ok := networksMap["dev-network"]; !ok {
			t.Error("Expected dev-network from override")
		}
	}
}
