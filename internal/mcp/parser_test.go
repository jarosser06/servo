package mcp

import (
	"os"
	"testing"
)

func TestParser_ParseFromFile(t *testing.T) {
	// Create a temporary servo file for testing
	tempFile, err := os.CreateTemp("", "test-*.servo")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test servo content
	servoContent := `servo_version: "1.0"
name: "test-server"
version: "1.0.0"
description: "Test MCP server"
author: "Test Author"
license: "MIT"

server:
  transport: "stdio"
  command: "test-server"
  args: ["--port", "8080"]

configuration_schema:
  secrets:
    api_key:
      description: "API key for external service"
      required: true
      type: "string"
      env_var: "API_KEY"
    
  config:
    debug_mode:
      description: "Enable debug logging"
      type: "boolean"
      default: false
      env_var: "DEBUG_MODE"
`

	if _, err := tempFile.WriteString(servoContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Test parsing
	parser := NewParser()
	servoDef, err := parser.ParseFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("ParseFromFile failed: %v", err)
	}

	// Validate parsed content
	if servoDef.ServoVersion != "1.0" {
		t.Errorf("Expected servo_version '1.0', got '%s'", servoDef.ServoVersion)
	}

	if servoDef.Name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", servoDef.Name)
	}

	if servoDef.Server.Command != "test-server" {
		t.Errorf("Expected server.command 'test-server', got '%s'", servoDef.Server.Command)
	}

	if len(servoDef.Server.Args) != 2 {
		t.Errorf("Expected 2 server args, got %d", len(servoDef.Server.Args))
	}

	// Test configuration schema
	if servoDef.ConfigurationSchema == nil {
		t.Fatal("Expected configuration_schema to be present")
	}

	if _, exists := servoDef.ConfigurationSchema.Secrets["api_key"]; !exists {
		t.Error("Expected api_key secret to be present")
	}

	if _, exists := servoDef.ConfigurationSchema.Config["debug_mode"]; !exists {
		t.Error("Expected debug_mode config to be present")
	}
}

func TestParser_ParseFromDirectory(t *testing.T) {
	// Create temporary directory with a servo file
	tempDir, err := os.MkdirTemp("", "servo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create servo file in directory
	servoFile := tempDir + "/test.servo"
	servoContent := `servo_version: "1.0"
name: "directory-test"
version: "1.0.0"
description: "Test from directory"
author: "Test"
license: "MIT"

server:
  transport: "stdio"
  command: "test"
  args: ["start"]
`

	if err := os.WriteFile(servoFile, []byte(servoContent), 0644); err != nil {
		t.Fatalf("Failed to write servo file: %v", err)
	}

	// Test parsing from directory
	parser := NewParser()
	servoDef, err := parser.ParseFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("ParseFromDirectory failed: %v", err)
	}

	if servoDef.Name != "directory-test" {
		t.Errorf("Expected name 'directory-test', got '%s'", servoDef.Name)
	}
}

func TestParser_ParseFromDirectory_MultipleFiles(t *testing.T) {
	// Create temporary directory with multiple servo files
	tempDir, err := os.MkdirTemp("", "servo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create first servo file
	servoContent := `servo_version: "1.0"
name: "test1"
version: "1.0.0"
description: "Test 1"
author: "Test"
license: "MIT"

server:
  transport: "stdio"
  command: "test1"
  args: ["start"]
`

	if err := os.WriteFile(tempDir+"/test1.servo", []byte(servoContent), 0644); err != nil {
		t.Fatalf("Failed to write first servo file: %v", err)
	}

	// Create second servo file
	if err := os.WriteFile(tempDir+"/test2.servo", []byte(servoContent), 0644); err != nil {
		t.Fatalf("Failed to write second servo file: %v", err)
	}

	// Test parsing - should fail with multiple files
	parser := NewParser()
	_, err = parser.ParseFromDirectory(tempDir)
	if err == nil {
		t.Error("Expected ParseFromDirectory to fail with multiple .servo files")
	}
}

func TestParser_ParseFromDirectory_NoFiles(t *testing.T) {
	// Create empty temporary directory
	tempDir, err := os.MkdirTemp("", "servo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test parsing from empty directory
	parser := NewParser()
	_, err = parser.ParseFromDirectory(tempDir)
	if err == nil {
		t.Error("Expected ParseFromDirectory to fail with no .servo files")
	}
}

func TestParser_AuthenticationConfiguration(t *testing.T) {
	parser := NewParser()

	// Test setting SSH authentication
	parser.SSHKeyPath = "/path/to/key"
	parser.SSHPassword = "passphrase"

	if parser.SSHKeyPath != "/path/to/key" {
		t.Errorf("Expected SSHKeyPath '/path/to/key', got '%s'", parser.SSHKeyPath)
	}

	if parser.SSHPassword != "passphrase" {
		t.Errorf("Expected SSHPassword 'passphrase', got '%s'", parser.SSHPassword)
	}

	// Test setting HTTP authentication
	parser.HTTPToken = "token123"
	parser.HTTPUsername = "user"
	parser.HTTPPassword = "pass"

	if parser.HTTPToken != "token123" {
		t.Errorf("Expected HTTPToken 'token123', got '%s'", parser.HTTPToken)
	}

	if parser.HTTPUsername != "user" {
		t.Errorf("Expected HTTPUsername 'user', got '%s'", parser.HTTPUsername)
	}

	if parser.HTTPPassword != "pass" {
		t.Errorf("Expected HTTPPassword 'pass', got '%s'", parser.HTTPPassword)
	}
}

func TestParser_ParseFromURL_InvalidURL(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFromURL("not-a-valid-url")
	if err == nil {
		t.Error("Expected ParseFromURL to fail with invalid URL")
	}
}

func TestParser_ParseFromFile_NonExistent(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFromFile("/non/existent/file.servo")
	if err == nil {
		t.Error("Expected ParseFromFile to fail with non-existent file")
	}
}

func TestParser_ParseFromFile_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempFile, err := os.CreateTemp("", "invalid-*.servo")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write invalid YAML content (malformed syntax)
	invalidYAML := `servo_version: "1.0"
name: "test"
invalid_yaml_syntax: [
  this is not proper yaml syntax
  missing closing bracket and quotes
`

	if _, err := tempFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Test parsing - should handle invalid YAML gracefully
	parser := NewParser()
	_, err = parser.ParseFromFile(tempFile.Name())
	if err == nil {
		t.Error("Expected ParseFromFile to fail with invalid YAML")
	}
}
