package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

func TestDockerComposeGenerator_EnvironmentVariableIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "servo_docker_compose_env_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create servo directory structure
	servoDir := filepath.Join(tempDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create servo directory: %v", err)
	}

	// Create sessions directory
	sessionsDir := filepath.Join(servoDir, "sessions", "default")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions directory: %v", err)
	}

	// Create project file
	projectFile := filepath.Join(servoDir, "project.yaml")
	projectContent := `default_session: default
active_session: default
mcp_servers:
  - name: test-server
    source: test-server.servo
    sessions:
      - default
`
	if err := os.WriteFile(projectFile, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to create project file: %v", err)
	}

	// Create session file
	sessionFile := filepath.Join(servoDir, "sessions", "default", "session.yaml")
	sessionContent := `name: default
is_default: true
manifests:
  - test-server
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create environment variables file
	envFile := filepath.Join(servoDir, "env.yaml")
	envContent := `version: "1.0"
env:
  PROJECT_DATABASE_URL: "postgres://project-host:5432/project-db"
  PROJECT_API_KEY: "project-api-key"
  GLOBAL_DEBUG: "true"
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	// Create test manifest
	manifestFile := filepath.Join(servoDir, "sessions", "default", "test-server.yaml")
	testManifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "test-server",
		Description:  "Test server for env var integration",
		Version:      "1.0.0",
		Services: map[string]*pkg.ServiceDependency{
			"api": {
				Image: "nginx:latest",
				Ports: []string{"8080:80"},
				Environment: map[string]string{
					"SERVICE_PORT":     "80",
					"PROJECT_API_KEY":  "service-override", // This should override project-level
					"SERVICE_SPECIFIC": "service-value",
				},
			},
		},
	}

	manifestData, err := yaml.Marshal(testManifest)
	if err != nil {
		t.Fatalf("Failed to marshal test manifest: %v", err)
	}

	if err := os.WriteFile(manifestFile, manifestData, 0644); err != nil {
		t.Fatalf("Failed to create manifest file: %v", err)
	}

	// Create devcontainer directory
	devcontainerDir := filepath.Join(tempDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatalf("Failed to create devcontainer directory: %v", err)
	}

	// Create docker-compose generator and generate configuration
	generator := NewDockerComposeGenerator(".servo")
	err = generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate docker-compose configuration: %v", err)
	}

	// Read generated docker-compose.yml
	dockerComposeFile := filepath.Join(".devcontainer", "docker-compose.yml")
	dockerComposeData, err := os.ReadFile(dockerComposeFile)
	if err != nil {
		t.Fatalf("Failed to read generated docker-compose.yml: %v", err)
	}

	// Parse docker-compose configuration
	var dockerCompose map[string]interface{}
	if err := yaml.Unmarshal(dockerComposeData, &dockerCompose); err != nil {
		t.Fatalf("Failed to parse docker-compose.yml: %v", err)
	}

	// Verify services section exists
	services, ok := dockerCompose["services"].(map[string]interface{})
	if !ok {
		t.Fatalf("Services section not found in docker-compose.yml")
	}

	// Find the test-server-api service
	var testServiceConfig map[string]interface{}
	for serviceName, serviceConfig := range services {
		if serviceName == "test-server-api" {
			testServiceConfig = serviceConfig.(map[string]interface{})
			break
		}
	}

	if testServiceConfig == nil {
		t.Fatalf("test-server-api service not found in generated docker-compose.yml")
	}

	// Verify environment variables are present
	environment, ok := testServiceConfig["environment"].([]interface{})
	if !ok {
		t.Fatalf("Environment section not found in test-server-api service")
	}

	// Convert environment slice to map for easier testing
	envMap := make(map[string]string)
	for _, envVar := range environment {
		envStr := envVar.(string)
		parts := splitEnvVar(envStr)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Test that project-level environment variables are included
	if val, exists := envMap["PROJECT_DATABASE_URL"]; !exists {
		t.Errorf("PROJECT_DATABASE_URL not found in service environment")
	} else if val != "postgres://project-host:5432/project-db" {
		t.Errorf("Expected PROJECT_DATABASE_URL=postgres://project-host:5432/project-db, got %s", val)
	}

	if val, exists := envMap["GLOBAL_DEBUG"]; !exists {
		t.Errorf("GLOBAL_DEBUG not found in service environment")
	} else if val != "true" {
		t.Errorf("Expected GLOBAL_DEBUG=true, got %s", val)
	}

	// Test that service-specific environment variables are included
	if val, exists := envMap["SERVICE_PORT"]; !exists {
		t.Errorf("SERVICE_PORT not found in service environment")
	} else if val != "80" {
		t.Errorf("Expected SERVICE_PORT=80, got %s", val)
	}

	if val, exists := envMap["SERVICE_SPECIFIC"]; !exists {
		t.Errorf("SERVICE_SPECIFIC not found in service environment")
	} else if val != "service-value" {
		t.Errorf("Expected SERVICE_SPECIFIC=service-value, got %s", val)
	}

	// Test that service-specific variables override project-level variables
	if val, exists := envMap["PROJECT_API_KEY"]; !exists {
		t.Errorf("PROJECT_API_KEY not found in service environment")
	} else if val != "service-override" {
		t.Errorf("Expected PROJECT_API_KEY=service-override (service should override project), got %s", val)
	}
}

func TestBaseGenerator_LoadProjectEnvironmentVariables(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "servo_base_gen_env_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create servo directory
	servoDir := filepath.Join(tempDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create servo directory: %v", err)
	}

	// Test 1: No env.yaml file (should return empty map without error)
	generator := NewBaseGenerator(".servo")
	
	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	envVars, err := generator.LoadProjectEnvironmentVariables()
	if err != nil {
		t.Fatalf("Expected no error when env.yaml doesn't exist, got: %v", err)
	}

	if len(envVars) != 0 {
		t.Fatalf("Expected empty map when no env.yaml, got %d variables", len(envVars))
	}

	// Test 2: Create env.yaml file and load it
	envFile := filepath.Join(servoDir, "env.yaml")
	envContent := `version: "1.0"
env:
  TEST_VAR1: "value1"
  TEST_VAR2: "value2"
  DATABASE_URL: "postgres://localhost:5432/testdb"
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	envVars, err = generator.LoadProjectEnvironmentVariables()
	if err != nil {
		t.Fatalf("Failed to load environment variables: %v", err)
	}

	expectedVars := map[string]string{
		"TEST_VAR1":    "value1",
		"TEST_VAR2":    "value2",
		"DATABASE_URL": "postgres://localhost:5432/testdb",
	}

	if len(envVars) != len(expectedVars) {
		t.Fatalf("Expected %d environment variables, got %d", len(expectedVars), len(envVars))
	}

	for key, expectedValue := range expectedVars {
		if actualValue, exists := envVars[key]; !exists {
			t.Errorf("Expected environment variable %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

// Helper function to split "KEY=VALUE" into ["KEY", "VALUE"]
func splitEnvVar(envVar string) []string {
	for i, char := range envVar {
		if char == '=' {
			return []string{envVar[:i], envVar[i+1:]}
		}
	}
	return []string{envVar} // No '=' found
}
