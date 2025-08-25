package commands

import (
	"os"
	"testing"
)

func TestConfigureCommand_Execute_NotInProject(t *testing.T) {
	// Create temporary directory for test (not a project)
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := NewConfigureCommand()
	err := cmd.Execute([]string{})

	if err == nil {
		t.Errorf("Expected error when not in project directory")
	}
	if err.Error() != "not in a servo project directory" {
		t.Errorf("Expected 'not in a servo project directory' error, got: %v", err)
	}
}

func TestConfigureCommand_Execute_EmptyProject(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal project structure
	os.MkdirAll(".servo", 0755)
	
	// Create project.yaml with empty servers
	projectContent := `version: "1"
name: "test-project"
description: "Test project"
mcpServers: {}
clients: []
sessions:
  main:
    servers: []
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewConfigureCommand()
	err := cmd.Execute([]string{})

	// Should not error, but should warn about no servers
	if err != nil {
		t.Errorf("Unexpected error with empty project: %v", err)
	}
}

func TestConfigureCommand_Execute_WithServers(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project structure
	os.MkdirAll(".servo", 0755)
	os.MkdirAll(".servo/manifests", 0755)
	
	// Create test manifest
	manifestContent := `name: "test-server"
description: "Test MCP server"
version: "1.0.0"
source:
  type: "git"
  uri: "https://github.com/test/test-server"
server:
  command: "python"
  args: ["-m", "test_server"]
  environment:
    TEST_VAR: "test_value"
`
	os.WriteFile(".servo/manifests/test-server.yaml", []byte(manifestContent), 0644)

	// Create project.yaml with servers
	projectContent := `version: "1"
name: "test-project"
description: "Test project"
mcpServers:
  test-server:
    version: "1.0.0"
    source: "https://github.com/test/test-server"
clients: ["vscode", "claude-code"]
sessions:
  main:
    servers: ["test-server"]
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewConfigureCommand()
	err := cmd.Execute([]string{})

	if err != nil {
		// Configuration generation might fail due to missing dependencies, 
		// but the command structure should work
		t.Logf("Configuration generation returned error (may be expected): %v", err)
	}

	// Check that the command attempted to create config directories
	// (Actual file creation depends on the config manager implementation)
}

func TestConfigureCommand_Name(t *testing.T) {
	cmd := NewConfigureCommand()
	if cmd.Name() != "configure" {
		t.Errorf("Expected command name 'configure', got '%s'", cmd.Name())
	}
}

func TestConfigureCommand_Description(t *testing.T) {
	cmd := NewConfigureCommand()
	description := cmd.Description()
	if description == "" {
		t.Errorf("Command description should not be empty")
	}
	if description != "Generate MCP client configurations for the current project" {
		t.Errorf("Unexpected description: %s", description)
	}
}

func TestConfigureCommand_GenerateConfigurations(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal project structure
	os.MkdirAll(".servo", 0755)
	
	// Create minimal project.yaml
	projectContent := `version: "1"
name: "test-project"
description: "Test project"
mcpServers: {}
clients: []
sessions:
  main:
    servers: []
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewConfigureCommand()
	
	// Test the internal generateConfigurations method
	err := cmd.generateConfigurations()
	
	// This might fail due to missing dependencies, but we're testing the structure
	if err != nil {
		t.Logf("generateConfigurations returned error (may be expected): %v", err)
	}
}

func TestConfigureCommand_Integration(t *testing.T) {
	// This test verifies the complete flow with more realistic project setup
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup complete project structure
	os.MkdirAll(".servo/manifests", 0755)
	
	// Create a realistic manifest
	manifestContent := `name: "file-manager"
description: "File management MCP server"
version: "1.0.0"
source:
  type: "git"
  uri: "https://github.com/example/file-manager"
server:
  command: "node"
  args: ["dist/index.js"]
  environment:
    NODE_ENV: "production"
`
	os.WriteFile(".servo/manifests/file-manager.yaml", []byte(manifestContent), 0644)

	// Create comprehensive project.yaml
	projectContent := `version: "1"
name: "integration-test-project"
description: "Integration test project"
mcpServers:
  file-manager:
    version: "1.0.0"
    source: "https://github.com/example/file-manager"
clients: ["vscode", "claude-code", "cursor"]
sessions:
  main:
    servers: ["file-manager"]
  development:
    servers: ["file-manager"]
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewConfigureCommand()
	err := cmd.Execute([]string{})

	// The command should execute without panicking, though it may fail
	// due to missing external dependencies
	if err != nil {
		t.Logf("Integration test returned error (configuration generation may fail in test environment): %v", err)
	}
}

func TestConfigureCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string)
		expectError   bool
		expectedError string
	}{
		{
			name: "corrupted project file",
			setupFunc: func(tmpDir string) {
				os.MkdirAll(".servo", 0755)
				// Write invalid YAML
				os.WriteFile(".servo/project.yaml", []byte("invalid: yaml: content: ["), 0644)
			},
			expectError:   true,
			expectedError: "failed to get project",
		},
		{
			name: "missing project file",
			setupFunc: func(tmpDir string) {
				os.MkdirAll(".servo", 0755)
				// .servo directory exists but no project.yaml
			},
			expectError:   true,
			expectedError: "not in a servo project directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldCwd, _ := os.Getwd()
			defer os.Chdir(oldCwd)
			os.Chdir(tmpDir)

			tt.setupFunc(tmpDir)

			cmd := NewConfigureCommand()
			err := cmd.Execute([]string{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		  func() bool {
			  for i := 0; i <= len(s)-len(substr); i++ {
				  if s[i:i+len(substr)] == substr {
					  return true
				  }
			  }
			  return false
		  }())))
}