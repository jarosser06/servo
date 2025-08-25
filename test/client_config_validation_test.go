package test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVSCodeMCPConfigValidation validates that servo generates correct VSCode MCP configurations
func TestVSCodeMCPConfigValidation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_vscode_config_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Testing VSCode MCP configuration in: %s", tempDir)

	// Create diverse MCP servers
	createVSCodeTestServoFiles(t, tempDir)

	// Run servo workflow for VSCode
	runVSCodeServoWorkflow(t, tempDir)

	// Validate VSCode MCP configuration
	validateVSCodeMCPSettings(t, tempDir)

	t.Logf("✅ VSCode MCP configuration validation passed!")
}

// TestClaudeCodeMCPConfigValidation validates that servo generates correct Claude Code configurations
func TestClaudeCodeMCPConfigValidation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_claude_config_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Testing Claude Code MCP configuration in: %s", tempDir)

	// Create diverse MCP servers
	createClaudeCodeTestServoFiles(t, tempDir)

	// Run servo workflow for Claude Code
	runClaudeCodeServoWorkflow(t, tempDir)

	// Validate Claude Code MCP configuration
	validateClaudeCodeMCPSettings(t, tempDir)

	t.Logf("✅ Claude Code MCP configuration validation passed!")
}

// TestDevcontainerValidation validates devcontainer config without building
func TestDevcontainerValidation(t *testing.T) {
	if !isDevcontainerCLIAvailable() {
		t.Skip("Devcontainer CLI not available")
	}

	tempDir, err := ioutil.TempDir("", "servo_devcontainer_validation_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Testing devcontainer validation in: %s", tempDir)

	// Create complex servo files
	createDevcontainerTestServoFiles(t, tempDir)

	// Run servo workflow
	runDevcontainerServoWorkflow(t, tempDir)

	// Validate with devcontainer CLI (read-only)
	validateDevcontainerWithCLI(t, tempDir)

	t.Logf("✅ Devcontainer validation passed!")
}

func createVSCodeTestServoFiles(t *testing.T, tempDir string) {
	// Node.js MCP server with complex environment variables
	nodeServer := `servo_version: "1.0"
metadata:
  name: "vscode-node-server"
  version: "1.2.3"
  description: "Node.js MCP server for VSCode testing"

requirements:
  runtimes:
    - name: "node"
      version: "18.0.0"

dependencies:
  services:
    redis:
      image: "redis:alpine"
      ports: ["6379"]

server:
  transport: "stdio"
  command: "node"
  args: ["index.js"]
  environment:
    NODE_ENV: "development"
    REDIS_URL: "redis://redis:6379"
    API_KEY: "${API_KEY}"
    DEBUG: "mcp:*"`

	err := ioutil.WriteFile(filepath.Join(tempDir, "vscode-node.servo"), []byte(nodeServer), 0644)
	if err != nil {
		t.Fatalf("Failed to write VSCode node servo: %v", err)
	}

	// Python MCP server with working directory
	pythonServer := `servo_version: "1.0"
metadata:
  name: "vscode-python-server"
  version: "2.0.0"
  description: "Python MCP server for VSCode testing"

requirements:
  runtimes:
    - name: "python"
      version: "3.11"

server:
  transport: "stdio" 
  command: "python"
  args: ["-m", "server.main"]
  working_directory: "/workspace/python-server"
  environment:
    PYTHONPATH: "/workspace/python-server"
    LOG_LEVEL: "DEBUG"`

	err = ioutil.WriteFile(filepath.Join(tempDir, "vscode-python.servo"), []byte(pythonServer), 0644)
	if err != nil {
		t.Fatalf("Failed to write VSCode python servo: %v", err)
	}

	t.Logf("✅ Created VSCode test servo files")
}

func createClaudeCodeTestServoFiles(t *testing.T, tempDir string) {
	// Go MCP server
	goServer := `servo_version: "1.0"
metadata:
  name: "claude-go-server"
  version: "1.0.0"
  description: "Go MCP server for Claude Code testing"

requirements:
  runtimes:
    - name: "go"
      version: "1.21"

server:
  transport: "stdio"
  command: "go"
  args: ["run", "main.go"]
  environment:
    CGO_ENABLED: "0"
    GOOS: "linux"`

	err := ioutil.WriteFile(filepath.Join(tempDir, "claude-go.servo"), []byte(goServer), 0644)
	if err != nil {
		t.Fatalf("Failed to write Claude Code go servo: %v", err)
	}

	// Rust MCP server (testing different runtime)
	rustServer := `servo_version: "1.0"
metadata:
  name: "claude-rust-server"
  version: "0.5.0"
  description: "Rust MCP server for Claude Code testing"

requirements:
  system:
    - name: "build-essential"
  runtimes:
    - name: "rust"
      version: "1.70.0"

server:
  transport: "stdio"
  command: "cargo"
  args: ["run", "--release"]
  working_directory: "/workspace/rust-server"`

	err = ioutil.WriteFile(filepath.Join(tempDir, "claude-rust.servo"), []byte(rustServer), 0644)
	if err != nil {
		t.Fatalf("Failed to write Claude Code rust servo: %v", err)
	}

	t.Logf("✅ Created Claude Code test servo files")
}

func createDevcontainerTestServoFiles(t *testing.T, tempDir string) {
	// Complex multi-service setup
	complexServer := `servo_version: "1.0"
metadata:
  name: "complex-service"
  version: "1.0.0"
  description: "Complex service for devcontainer validation"

requirements:
  runtimes:
    - name: "node"
      version: "20.0.0"
    - name: "python"
      version: "3.11"

dependencies:
  services:
    redis:
      image: "redis:7"
      ports: ["6379"]
    postgres:
      image: "postgres:15"
      ports: ["5432"]
      environment:
        POSTGRES_DB: "testdb"
        POSTGRES_USER: "testuser"
        POSTGRES_PASSWORD: "testpass"

server:
  transport: "stdio"
  command: "node"
  args: ["server.js"]`

	err := ioutil.WriteFile(filepath.Join(tempDir, "complex.servo"), []byte(complexServer), 0644)
	if err != nil {
		t.Fatalf("Failed to write complex servo: %v", err)
	}

	t.Logf("✅ Created devcontainer test servo file")
}

func runVSCodeServoWorkflow(t *testing.T, tempDir string) {
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project for VSCode
	cmd := exec.Command(servoPath, "init", "vscode-test", "", "--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize VSCode project: %v\nOutput: %s", err, output)
	}

	// Install servers
	servoFiles := []string{"vscode-node.servo", "vscode-python.servo"}
	for _, file := range servoFiles {
		cmd = exec.Command(servoPath, "install", file, "--clients", "vscode")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install %s for VSCode: %v\nOutput: %s", file, err, output)
		}
	}

	t.Logf("✅ VSCode servo workflow completed")
}

func runClaudeCodeServoWorkflow(t *testing.T, tempDir string) {
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project for Claude Code
	cmd := exec.Command(servoPath, "init", "claude-test", "", "--clients", "claude-code")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize Claude Code project: %v\nOutput: %s", err, output)
	}

	// Install servers
	servoFiles := []string{"claude-go.servo", "claude-rust.servo"}
	for _, file := range servoFiles {
		cmd = exec.Command(servoPath, "install", file, "--clients", "claude-code")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install %s for Claude Code: %v\nOutput: %s", file, err, output)
		}
	}

	t.Logf("✅ Claude Code servo workflow completed")
}

func runDevcontainerServoWorkflow(t *testing.T, tempDir string) {
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project for devcontainer
	cmd := exec.Command(servoPath, "init", "devcontainer-test", "", "--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize devcontainer project: %v\nOutput: %s", err, output)
	}

	// Install complex server
	cmd = exec.Command(servoPath, "install", "complex.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install complex servo: %v\nOutput: %s", err, output)
	}

	t.Logf("✅ Devcontainer servo workflow completed")
}

func validateVSCodeMCPSettings(t *testing.T, tempDir string) {
	t.Logf("🔍 Validating VSCode MCP settings structure...")

	settingsPath := filepath.Join(tempDir, ".vscode", "mcp.json")
	content, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read VSCode MCP config: %v", err)
	}

	var settings map[string]interface{}
	err = json.Unmarshal(content, &settings)
	if err != nil {
		t.Fatalf("Invalid JSON in VSCode MCP config: %v", err)
	}

	// Validate MCP configuration structure (VSCode uses servers section)
	servers, exists := settings["servers"].(map[string]interface{})
	if !exists {
		t.Fatal("Missing 'servers' configuration in VSCode MCP config")
	}

	// Validate specific server configurations
	validateVSCodeServer(t, servers, "vscode-node-server", map[string]interface{}{
		"command": "node",
		"args":    []interface{}{"index.js"},
		"env": map[string]interface{}{
			"NODE_ENV":  "development",
			"REDIS_URL": "redis://redis:6379",
			"API_KEY":   "${API_KEY}",
			"DEBUG":     "mcp:*",
		},
	})

	validateVSCodeServer(t, servers, "vscode-python-server", map[string]interface{}{
		"command": "python",
		"args":    []interface{}{"-m", "server.main"},
		"env": map[string]interface{}{
			"PYTHONPATH": "/workspace/python-server",
			"LOG_LEVEL":  "DEBUG",
		},
	})

	t.Logf("✅ VSCode MCP settings validation passed!")
}

func validateClaudeCodeMCPSettings(t *testing.T, tempDir string) {
	t.Logf("🔍 Validating Claude Code MCP settings structure...")

	settingsPath := filepath.Join(tempDir, ".mcp.json")
	content, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Code MCP config: %v", err)
	}

	var settings map[string]interface{}
	err = json.Unmarshal(content, &settings)
	if err != nil {
		t.Fatalf("Invalid JSON in Claude Code MCP config: %v", err)
	}

	// Validate MCP configuration structure for Claude Code
	mcp, exists := settings["mcpServers"].(map[string]interface{})
	if !exists {
		t.Fatal("Missing 'mcpServers' configuration in Claude Code MCP config")
	}

	// Validate specific server configurations
	validateClaudeCodeServer(t, mcp, "claude-go-server", map[string]interface{}{
		"command": "go",
		"args":    []interface{}{"run", "main.go"},
		"env": map[string]interface{}{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
		},
	})

	validateClaudeCodeServer(t, mcp, "claude-rust-server", map[string]interface{}{
		"command": "cargo",
		"args":    []interface{}{"run", "--release"},
	})

	t.Logf("✅ Claude Code MCP settings validation passed!")
}

func validateDevcontainerWithCLI(t *testing.T, tempDir string) {
	t.Logf("🔍 Validating devcontainer with CLI...")

	// Use devcontainer CLI to validate configuration
	cmd := exec.Command("npx", "@devcontainers/cli", "read-configuration", "--workspace-folder", tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Devcontainer CLI validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check for critical components
	if !strings.Contains(outputStr, "features") {
		t.Error("Devcontainer CLI didn't recognize features")
	}

	if !strings.Contains(outputStr, "dockerComposeFile") {
		t.Error("Devcontainer CLI didn't recognize docker-compose configuration")
	}

	// Parse output for feature validation
	if strings.Contains(outputStr, "node") {
		t.Logf("✅ Node.js feature recognized by devcontainer CLI")
	}

	if strings.Contains(outputStr, "python") {
		t.Logf("✅ Python feature recognized by devcontainer CLI")
	}

	t.Logf("✅ Devcontainer CLI validation passed!")
}

func validateVSCodeServer(t *testing.T, servers map[string]interface{}, serverName string, expected map[string]interface{}) {
	server, exists := servers[serverName].(map[string]interface{})
	if !exists {
		t.Fatalf("Missing server '%s' in VSCode MCP settings", serverName)
	}

	// Check command
	if expectedCmd, ok := expected["command"]; ok {
		if server["command"] != expectedCmd {
			t.Errorf("Server '%s' command mismatch: expected '%v', got '%v'", serverName, expectedCmd, server["command"])
		}
	}

	// Check args
	if expectedArgs, ok := expected["args"]; ok {
		serverArgs, argsExist := server["args"].([]interface{})
		if !argsExist {
			t.Errorf("Server '%s' missing args", serverName)
		} else {
			expectedArgsSlice := expectedArgs.([]interface{})
			if len(serverArgs) != len(expectedArgsSlice) {
				t.Errorf("Server '%s' args length mismatch", serverName)
			}
		}
	}

	// Check environment variables
	if expectedEnv, ok := expected["env"]; ok {
		serverEnv, envExists := server["env"].(map[string]interface{})
		if !envExists {
			t.Errorf("Server '%s' missing environment variables", serverName)
		} else {
			expectedEnvMap := expectedEnv.(map[string]interface{})
			for key, expectedValue := range expectedEnvMap {
				if serverEnv[key] != expectedValue {
					t.Errorf("Server '%s' env '%s' mismatch: expected '%v', got '%v'", serverName, key, expectedValue, serverEnv[key])
				}
			}
		}
	}

	// Check working directory
	if expectedCwd, ok := expected["cwd"]; ok {
		if server["cwd"] != expectedCwd {
			t.Errorf("Server '%s' cwd mismatch: expected '%v', got '%v'", serverName, expectedCwd, server["cwd"])
		}
	}

	t.Logf("✅ VSCode server '%s' configuration validated", serverName)
}

func validateClaudeCodeServer(t *testing.T, servers map[string]interface{}, serverName string, expected map[string]interface{}) {
	server, exists := servers[serverName].(map[string]interface{})
	if !exists {
		t.Fatalf("Missing server '%s' in Claude Code MCP settings", serverName)
	}

	// Check command
	if expectedCmd, ok := expected["command"]; ok {
		if server["command"] != expectedCmd {
			t.Errorf("Server '%s' command mismatch: expected '%v', got '%v'", serverName, expectedCmd, server["command"])
		}
	}

	// Check args
	if expectedArgs, ok := expected["args"]; ok {
		serverArgs, argsExist := server["args"].([]interface{})
		if !argsExist {
			t.Errorf("Server '%s' missing args", serverName)
		} else {
			expectedArgsSlice := expectedArgs.([]interface{})
			if len(serverArgs) != len(expectedArgsSlice) {
				t.Errorf("Server '%s' args length mismatch", serverName)
			}
		}
	}

	// Check environment variables
	if expectedEnv, ok := expected["env"]; ok {
		serverEnv, envExists := server["env"].(map[string]interface{})
		if !envExists {
			t.Errorf("Server '%s' missing environment variables", serverName)
		} else {
			expectedEnvMap := expectedEnv.(map[string]interface{})
			for key, expectedValue := range expectedEnvMap {
				if serverEnv[key] != expectedValue {
					t.Errorf("Server '%s' env '%s' mismatch: expected '%v', got '%v'", serverName, key, expectedValue, serverEnv[key])
				}
			}
		}
	}

	// Check working directory
	if expectedCwd, ok := expected["cwd"]; ok {
		if server["cwd"] != expectedCwd {
			t.Errorf("Server '%s' cwd mismatch: expected '%v', got '%v'", serverName, expectedCwd, server["cwd"])
		}
	}

	t.Logf("✅ Claude Code server '%s' configuration validated", serverName)
}
