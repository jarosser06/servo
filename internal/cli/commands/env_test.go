package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/internal/project"
)

func TestEnvCommand_SetAndGet(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "servo_env_test_*")
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

	// Initialize a servo project
	servoDir := filepath.Join(tempDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create servo directory: %v", err)
	}

	// Create project file
	projectFile := filepath.Join(servoDir, "project.yaml")
	projectContent := `default_session: default
active_session: default
`
	if err := os.WriteFile(projectFile, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to create project file: %v", err)
	}

	// Create env command
	projectManager := project.NewManager()
	envCmd := NewEnvCommand(projectManager)

	// Test setting environment variables
	err = envCmd.Execute([]string{"set", "TEST_VAR", "test_value"})
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	err = envCmd.Execute([]string{"set", "API_URL", "https://api.example.com"})
	if err != nil {
		t.Fatalf("Failed to set second environment variable: %v", err)
	}

	// Test getting environment variables
	// Capture stdout for get command
	// Note: This test focuses on testing the functionality works without errors
	// In a real scenario, we'd capture stdout to verify the exact output
	err = envCmd.Execute([]string{"get", "TEST_VAR"})
	if err != nil {
		t.Fatalf("Failed to get environment variable: %v", err)
	}

	// Test listing environment variables
	err = envCmd.Execute([]string{"list"})
	if err != nil {
		t.Fatalf("Failed to list environment variables: %v", err)
	}

	// Test deleting environment variable
	err = envCmd.Execute([]string{"delete", "TEST_VAR"})
	if err != nil {
		t.Fatalf("Failed to delete environment variable: %v", err)
	}

	// Verify deleted variable is gone
	err = envCmd.Execute([]string{"get", "TEST_VAR"})
	if err == nil {
		t.Fatalf("Expected error when getting deleted environment variable")
	}
}

func TestEnvCommand_LoadProjectEnvironmentVariables(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "servo_env_load_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create servo directory
	servoDir := filepath.Join(tempDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create servo directory: %v", err)
	}

	// Create env.yaml file directly
	envFile := filepath.Join(servoDir, "env.yaml")
	envContent := `version: "1.0"
env:
  DATABASE_URL: "postgres://localhost:5432/testdb"
  API_KEY: "test-api-key"
  DEBUG_MODE: "true"
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	// Test loading environment variables using the base generator method
	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create project manager and env command
	projectManager := project.NewManager()
	envCmd := NewEnvCommand(projectManager)

	// Load environment variables
	envData, err := envCmd.loadEnvData()
	if err != nil {
		t.Fatalf("Failed to load environment data: %v", err)
	}

	// Verify loaded data
	expectedVars := map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/testdb",
		"API_KEY":      "test-api-key",
		"DEBUG_MODE":   "true",
	}

	if len(envData.Env) != len(expectedVars) {
		t.Fatalf("Expected %d environment variables, got %d", len(expectedVars), len(envData.Env))
	}

	for key, expectedValue := range expectedVars {
		if actualValue, exists := envData.Env[key]; !exists {
			t.Errorf("Expected environment variable %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

func TestEnvCommand_NonExistentProject(t *testing.T) {
	// Create temporary directory that's not a servo project
	tempDir, err := os.MkdirTemp("", "servo_env_non_project_*")
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

	// Create env command
	projectManager := project.NewManager()
	envCmd := NewEnvCommand(projectManager)

	// Test that commands fail when not in a servo project
	err = envCmd.Execute([]string{"set", "TEST_VAR", "test_value"})
	if err == nil {
		t.Fatalf("Expected error when setting env var outside servo project")
	}

	err = envCmd.Execute([]string{"list"})
	if err == nil {
		t.Fatalf("Expected error when listing env vars outside servo project")
	}
}

func TestEnvCommand_Export(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "servo_env_export_test_*")
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

	// Initialize a servo project
	servoDir := filepath.Join(tempDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create servo directory: %v", err)
	}

	// Create project file
	projectFile := filepath.Join(servoDir, "project.yaml")
	projectContent := `default_session: default
active_session: default
`
	if err := os.WriteFile(projectFile, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to create project file: %v", err)
	}

	// Create env command
	projectManager := project.NewManager()
	envCmd := NewEnvCommand(projectManager)

	// Set some environment variables
	envCmd.Execute([]string{"set", "TEST_VAR1", "value1"})
	envCmd.Execute([]string{"set", "TEST_VAR2", "value2"})

	// Test export (should not error)
	err = envCmd.Execute([]string{"export"})
	if err != nil {
		t.Fatalf("Failed to export environment variables: %v", err)
	}
}
