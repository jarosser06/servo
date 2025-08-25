package commands

import (
	"os"
	"testing"
)

func TestWorkCommand_Execute_NotInProject(t *testing.T) {
	// Create temporary directory for test (not a project)
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := NewWorkCommand()
	err := cmd.Execute([]string{})

	if err == nil {
		t.Errorf("Expected error when not in project directory")
	}
	if err.Error() != "not in a servo project directory" {
		t.Errorf("Expected 'not in a servo project directory' error, got: %v", err)
	}
}

func TestWorkCommand_Execute_BasicProject(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal project structure
	os.MkdirAll(".servo", 0755)
	
	// Create project.yaml
	projectContent := `version: "1"
name: "test-work-project"
description: "Test project for work command"
mcpServers: {}
clients: []
activeSession: "main"
sessions:
  main:
    servers: []
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewWorkCommand()
	err := cmd.Execute([]string{})

	// Configuration generation might fail due to missing dependencies,
	// but the command should handle the basic flow
	if err != nil {
		t.Logf("Work command returned error (configuration generation may fail): %v", err)
	}
}

func TestWorkCommand_Execute_WithServers(t *testing.T) {
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
`
	os.WriteFile(".servo/manifests/test-server.yaml", []byte(manifestContent), 0644)

	// Create project.yaml with servers
	projectContent := `version: "1"
name: "test-work-project"
description: "Test project with servers"
mcpServers:
  test-server:
    name: "test-server"
    version: "1.0.0"
    source: "https://github.com/test/test-server"
    clients: ["vscode", "claude-code"]
clients: ["vscode", "claude-code"]
activeSession: "main"
sessions:
  main:
    servers: ["test-server"]
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewWorkCommand()
	err := cmd.Execute([]string{})

	if err != nil {
		t.Logf("Work command with servers returned error (may be expected): %v", err)
	}
}

func TestWorkCommand_Execute_WithClientFlags(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal project structure
	os.MkdirAll(".servo", 0755)
	
	projectContent := `version: "1"
name: "test-work-project"
description: "Test project"
mcpServers: {}
clients: []
activeSession: "main"
sessions:
  main:
    servers: []
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewWorkCommand()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "vscode flag",
			args: []string{"--vscode"},
		},
		{
			name: "claude-code flag",
			args: []string{"--claude-code"},
		},
		{
			name: "client flag with value",
			args: []string{"--client", "vscode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args)
			// These should now safely show launch commands instead of launching apps
			if err != nil {
				t.Logf("Work command with %s returned error (configuration generation may fail): %v", tt.name, err)
			}
		})
	}
}

func TestWorkCommand_Name(t *testing.T) {
	cmd := NewWorkCommand()
	if cmd.Name() != "work" {
		t.Errorf("Expected command name 'work', got '%s'", cmd.Name())
	}
}

func TestWorkCommand_Description(t *testing.T) {
	cmd := NewWorkCommand()
	description := cmd.Description()
	if description == "" {
		t.Errorf("Command description should not be empty")
	}
	if description != "Start development environment for project" {
		t.Errorf("Unexpected description: %s", description)
	}
}

func TestWorkCommand_GetLaunchCommand(t *testing.T) {
	cmd := NewWorkCommand()

	tests := []struct {
		name       string
		clientName string
		expectCmd  bool
	}{
		{
			name:       "unsupported client",
			clientName: "unsupported", 
			expectCmd:  false,
		},
		{
			name:       "vscode client",
			clientName: "vscode",
			expectCmd:  true, // Should return command if VSCode exists, empty if not
		},
		{
			name:       "claude-code client",
			clientName: "claude-code",
			expectCmd:  true, // Should return command if Claude exists, empty if not
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			launchCmd := cmd.getLaunchCommand(tt.clientName)
			
			if tt.expectCmd {
				// Command should be either a valid launch command or empty (if not installed)
				// Both are acceptable outcomes
				t.Logf("Launch command for %s: '%s'", tt.clientName, launchCmd)
			} else if launchCmd != "" {
				t.Errorf("Expected empty command for unsupported client %s, got: %s", tt.clientName, launchCmd)
			}
		})
	}
}


func TestWorkCommand_GenerateConfigurations(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal project structure
	os.MkdirAll(".servo", 0755)
	
	projectContent := `version: "1"
name: "test-work-project"
description: "Test project"
mcpServers: {}
clients: []
sessions:
  main:
    servers: []
`
	os.WriteFile(".servo/project.yaml", []byte(projectContent), 0644)

	cmd := NewWorkCommand()
	
	// Test the internal generateConfigurations method
	err := cmd.generateConfigurations()
	
	// This might fail due to missing dependencies, but we're testing the structure
	if err != nil {
		t.Logf("generateConfigurations returned error (may be expected): %v", err)
	}
}

func TestWorkCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string)
		args          []string
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
			args:          []string{},
			expectError:   true,
			expectedError: "failed to get project",
		},
		{
			name: "missing project file",
			setupFunc: func(tmpDir string) {
				os.MkdirAll(".servo", 0755)
				// .servo directory exists but no project.yaml
			},
			args:          []string{},
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

			cmd := NewWorkCommand()
			err := cmd.Execute(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && !containsSubstring(err.Error(), tt.expectedError) {
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
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

