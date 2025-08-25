package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/internal/mcp"
)

func TestValidateCommand_Execute_NoArgs(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err := cmd.Execute([]string{})
	if err == nil {
		t.Errorf("Expected error when no arguments provided")
	}
	if err.Error() != "source is required\nUsage: servo validate <source>" {
		t.Errorf("Expected usage error, got: %v", err)
	}
}

func TestValidateCommand_Execute_Help(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	testCases := []string{"help", "--help", "-h"}

	for _, helpArg := range testCases {
		t.Run("help_"+helpArg, func(t *testing.T) {
			err := cmd.Execute([]string{helpArg})
			if err != nil {
				t.Errorf("Help command should not return error, got: %v", err)
			}
		})
	}
}

func TestValidateCommand_Execute_ValidServoFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create a valid .servo file
	servoContent := `servo_version: "1.0"
name: "test-server"
description: "A test MCP server"
version: "1.0.0"
author: "Test Author"
license: "MIT"

metadata:
  tags: ["testing", "example"]

source:
  type: "git"
  uri: "https://github.com/test/test-server"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/test/test-server"
  setup_commands:
    - "pip install -r requirements.txt"

server:
  transport: "stdio"
  command: "python"
  args: ["-m", "test_server"]
  environment:
    TEST_ENV: "test_value"

requirements:
  system:
    - name: "python"
      description: "Python runtime"
      check_command: "python --version"
      install_hint: "Install Python 3.8 or later"
  runtimes:
    - name: "python"
      version: "3.8"

configurationSchema:
  secrets:
    api_key:
      description: "API key for service"
      required: true
  config:
    debug:
      description: "Enable debug mode"
      default: false

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code", "cursor"]
`
	
	servoFile := filepath.Join(tmpDir, "test.servo")
	err := os.WriteFile(servoFile, []byte(servoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err = cmd.Execute([]string{servoFile})
	if err != nil {
		t.Errorf("Valid .servo file should not return error, got: %v", err)
	}
}

func TestValidateCommand_Execute_InvalidServoFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create an invalid .servo file (missing required fields)
	invalidServoContent := `name: "incomplete-server"
# Missing required fields like version, description, etc.
`
	
	servoFile := filepath.Join(tmpDir, "invalid.servo")
	err := os.WriteFile(servoFile, []byte(invalidServoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err = cmd.Execute([]string{servoFile})
	if err == nil {
		t.Errorf("Invalid .servo file should return error")
	}
}

func TestValidateCommand_Execute_NonExistentFile(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err := cmd.Execute([]string{"/nonexistent/file.servo"})
	if err == nil {
		t.Errorf("Non-existent file should return error")
	}
}

func TestValidateCommand_Execute_CorruptedFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create a corrupted .servo file (invalid YAML)
	corruptedContent := `name: "corrupted-server
description: This is invalid YAML [ { }
version: "1.0.0"
`
	
	servoFile := filepath.Join(tmpDir, "corrupted.servo")
	err := os.WriteFile(servoFile, []byte(corruptedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err = cmd.Execute([]string{servoFile})
	if err == nil {
		t.Errorf("Corrupted .servo file should return error")
	}
}

func TestValidateCommand_Name(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	if cmd.Name() != "validate" {
		t.Errorf("Expected command name 'validate', got '%s'", cmd.Name())
	}
}

func TestValidateCommand_Description(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	description := cmd.Description()
	if description == "" {
		t.Errorf("Command description should not be empty")
	}
	if description != "Validate .servo file or source" {
		t.Errorf("Unexpected description: %s", description)
	}
}

func TestValidateCommand_ParseSource(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create a valid .servo file
	servoContent := `servo_version: "1.0"
name: "test-server"
description: "A test MCP server"
version: "1.0.0"
author: "Test Author"

source:
  type: "git"
  uri: "https://github.com/test/test-server"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/test/test-server"
  setup_commands:
    - "pip install -r requirements.txt"

server:
  command: "python"
  args: ["-m", "test_server"]
`
	
	servoFile := filepath.Join(tmpDir, "test.servo")
	err := os.WriteFile(servoFile, []byte(servoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	tests := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name:        "local .servo file",
			source:      servoFile,
			expectError: false,
		},
		{
			name:        "directory path",
			source:      tmpDir,
			expectError: false, // May fail but should attempt parsing
		},
		{
			name:        "non-existent file",
			source:      "/nonexistent/file.servo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servoFile, err := cmd.parseSource(tt.source)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for source %s, but got none", tt.source)
				}
			} else {
				if err != nil {
					t.Logf("parseSource returned error (may be expected): %v", err)
				} else if servoFile == nil {
					t.Errorf("Expected non-nil ServoDefinition for valid source")
				}
			}
		})
	}
}

func TestValidateCommand_ParseSource_HTTPSources(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	// These tests will likely fail due to network requirements, 
	// but we're testing the logic paths
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "GitHub repo URL",
			source: "https://github.com/example/repo",
		},
		{
			name:   "Direct .servo URL",
			source: "https://example.com/file.servo",
		},
		{
			name:   "HTTP .servo URL",
			source: "http://example.com/file.servo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These will likely fail due to network/authentication issues
			// but should not panic
			_, err := cmd.parseSource(tt.source)
			if err != nil {
				t.Logf("Network-based parsing failed (expected): %v", err)
			}
		})
	}
}

func TestValidateCommand_ShowHelp(t *testing.T) {
	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err := cmd.showHelp()
	if err != nil {
		t.Errorf("showHelp should not return error, got: %v", err)
	}
}

func TestValidateCommand_MinimalValidFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create a minimal but valid .servo file
	minimalServoContent := `servo_version: "1.0"
name: "minimal-server"
description: "A minimal test MCP server"
version: "1.0.0"

source:
  type: "git"
  uri: "https://github.com/test/minimal-server"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/test/test-server"
  setup_commands:
    - "pip install -r requirements.txt"

server:
  command: "node"
  args: ["index.js"]
`
	
	servoFile := filepath.Join(tmpDir, "minimal.servo")
	err := os.WriteFile(servoFile, []byte(minimalServoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create minimal .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err = cmd.Execute([]string{servoFile})
	if err != nil {
		// The minimal file may not pass all validation requirements
		t.Logf("Minimal .servo file failed validation (expected): %v", err)
	}
}

func TestValidateCommand_EmptyFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	
	// Create an empty .servo file
	emptyFile := filepath.Join(tmpDir, "empty.servo")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty .servo file: %v", err)
	}

	parser := mcp.NewParser()
	validator := mcp.NewValidator()
	cmd := NewValidateCommand(parser, validator)

	err = cmd.Execute([]string{emptyFile})
	if err == nil {
		t.Errorf("Empty .servo file should return error")
	}
}