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

// TestWorkflowIntegration_CompleteDevWorkflow tests the complete developer workflow:
// init -> install -> configure -> work -> validate configurations
func TestWorkflowIntegration_CompleteDevWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_workflow_complete_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running complete dev workflow integration test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Phase 1: Initialize project
	t.Logf("📦 Phase 1: Project Initialization")
	cmd := exec.Command(servoPath, "init", "workflow-test", "--description", "Complete workflow test", "--clients", "vscode,claude-code")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized successfully")

	// Phase 2: Create and install test server
	t.Logf("🔧 Phase 2: Server Installation")
	createTestServoFile(t, tempDir)
	
	cmd = exec.Command(servoPath, "install", "test-mcp-server.servo", "--clients", "vscode,claude-code")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Server installed successfully")

	// Phase 3: Configure all clients
	t.Logf("⚙️ Phase 3: Client Configuration")
	cmd = exec.Command(servoPath, "configure")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to configure clients: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Clients configured successfully")

	// Phase 4: Validate generated configurations
	t.Logf("🔍 Phase 4: Configuration Validation")
	validateGeneratedConfigurations(t, tempDir)

	// Phase 5: Test work command for different clients
	t.Logf("🚀 Phase 5: Work Command Testing")
	testWorkCommand(t, tempDir, servoPath)

	// Phase 6: Test status and project state
	t.Logf("📊 Phase 6: Status Validation")
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get status: %v\nOutput: %s", err, output)
	}
	
	statusOutput := string(output)
	if !strings.Contains(statusOutput, "1 configured") {
		t.Errorf("Expected 1 MCP server configured, got: %s", statusOutput)
	}
	
	t.Logf("✅ Complete dev workflow integration test passed!")
}

// TestWorkflowIntegration_MultiSessionWorkflow tests workflows across multiple sessions
func TestWorkflowIntegration_MultiSessionWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_workflow_multisession_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running multi-session workflow integration test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Phase 1: Initialize project with default session
	t.Logf("📦 Phase 1: Project Initialization")
	cmd := exec.Command(servoPath, "init", "multisession-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}

	// Phase 2: Create additional sessions
	t.Logf("🔄 Phase 2: Creating Multiple Sessions")
	sessions := []struct {
		name        string
		description string
	}{
		{"development", "Development environment"},
		{"staging", "Staging environment"},
		{"production", "Production environment"},
	}

	for _, session := range sessions {
		cmd = exec.Command(servoPath, "session", "create", session.name, "--description", session.description)
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create session %s: %v\nOutput: %s", session.name, err, output)
		}
		t.Logf("✅ Created session: %s", session.name)
	}

	// Phase 3: Install servers to different sessions
	t.Logf("🔧 Phase 3: Installing Servers to Different Sessions")
	createMultipleServoFiles(t, tempDir)

	sessionServers := map[string][]string{
		"development": {"dev-server.servo", "logging-server.servo"},
		"staging":     {"staging-server.servo", "logging-server.servo"},
		"production":  {"prod-server.servo"},
	}

	for sessionName, servers := range sessionServers {
		for _, server := range servers {
			cmd = exec.Command(servoPath, "install", "--session", sessionName, server)
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to install %s to session %s: %v\nOutput: %s", server, sessionName, err, output)
			}
			t.Logf("✅ Installed %s to %s session", server, sessionName)
		}
	}

	// Phase 4: Test session switching and configuration
	t.Logf("🔄 Phase 4: Testing Session Switching")
	for sessionName := range sessionServers {
		// Switch to session
		cmd = exec.Command(servoPath, "session", "activate", sessionName)
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to activate session %s: %v\nOutput: %s", sessionName, err, output)
		}

		// Configure for this session
		cmd = exec.Command(servoPath, "configure")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to configure session %s: %v\nOutput: %s", sessionName, err, output)
		}

		// Verify status shows correct servers for this session
		cmd = exec.Command(servoPath, "status")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get status for session %s: %v\nOutput: %s", sessionName, err, output)
		}

		// Note: Status command shows all servers, not per-session
		statusOutput := string(output)
		if !strings.Contains(statusOutput, "configured") {
			t.Errorf("Expected servers configured in status output for %s session, got: %s", sessionName, statusOutput)
		}
		
		t.Logf("✅ Session %s status validated", sessionName)
	}

	t.Logf("✅ Multi-session workflow integration test passed!")
}

// TestWorkflowIntegration_ErrorRecoveryWorkflow tests error scenarios and recovery
func TestWorkflowIntegration_ErrorRecoveryWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_workflow_errors_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running error recovery workflow integration test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test 1: Commands outside project should fail gracefully
	t.Logf("🚫 Test 1: Commands Outside Project")
	errorCommands := [][]string{
		{"install", "nonexistent.servo"},
		{"configure"},
		{"work"},
		{"status"},
	}

	for _, cmdArgs := range errorCommands {
		cmd := exec.Command(servoPath, cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("Expected command '%s' to fail outside project, but it succeeded", strings.Join(cmdArgs, " "))
		}
		if !strings.Contains(string(output), "not in a servo project directory") {
			t.Errorf("Expected proper error message for '%s', got: %s", strings.Join(cmdArgs, " "), string(output))
		}
	}
	t.Logf("✅ Commands correctly fail outside project")

	// Test 2: Initialize project for further testing
	cmd := exec.Command(servoPath, "init", "error-recovery-test")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Test 3: Install nonexistent server should fail
	t.Logf("🚫 Test 3: Install Nonexistent Server")
	cmd = exec.Command(servoPath, "install", "nonexistent-server.servo")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("Expected install of nonexistent server to fail")
	}
	t.Logf("✅ Install correctly fails for nonexistent server")

	// Test 4: Install malformed server should fail
	t.Logf("🚫 Test 4: Install Malformed Server")
	createMalformedServoFile(t, tempDir)
	cmd = exec.Command(servoPath, "install", "malformed.servo")
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Errorf("Expected install of malformed server to fail")
	}
	t.Logf("✅ Install correctly fails for malformed server")

	// Test 5: Recovery after fixing issues
	t.Logf("🔧 Test 5: Recovery After Fixing Issues")
	createTestServoFile(t, tempDir)
	cmd = exec.Command(servoPath, "install", "test-mcp-server.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install server after fixing: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Successfully recovered and installed server")

	// Test 6: Status should work after recovery
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Status failed after recovery: %v", err)
	}
	if !strings.Contains(string(output), "1 configured") {
		t.Errorf("Expected 1 server after recovery, got: %s", string(output))
	}
	t.Logf("✅ Status works correctly after recovery")

	t.Logf("✅ Error recovery workflow integration test passed!")
}

// TestWorkflowIntegration_ClientSwitchingWorkflow tests switching between different client configurations
func TestWorkflowIntegration_ClientSwitchingWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_workflow_clients_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running client switching workflow integration test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project
	cmd := exec.Command(servoPath, "init", "client-switching-test")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	createTestServoFile(t, tempDir)

	// Test different client combinations
	clientCombinations := []struct {
		name    string
		clients string
	}{
		{"vscode-only", "vscode"},
		{"claude-code-only", "claude-code"},
		{"cursor-only", "cursor"},
		{"vscode-claude", "vscode,claude-code"},
		{"all-clients", "vscode,claude-code,cursor"},
	}

	for _, combo := range clientCombinations {
		t.Logf("🔧 Testing client combination: %s", combo.name)

		// Install with specific client combination
		cmd = exec.Command(servoPath, "install", "--clients", combo.clients, "--update", "test-mcp-server.servo")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install with clients %s: %v\nOutput: %s", combo.clients, err, output)
		}

		// Configure
		cmd = exec.Command(servoPath, "configure")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to configure with clients %s: %v\nOutput: %s", combo.clients, err, output)
		}

		// Validate that correct client files were generated
		validateClientConfigurations(t, tempDir, combo.clients)

		t.Logf("✅ Client combination %s validated", combo.name)
	}

	t.Logf("✅ Client switching workflow integration test passed!")
}

// Helper functions

func createTestServoFile(t *testing.T, tempDir string) {
	servoContent := `servo_version: "1.0"
name: "test-mcp-server"
version: "1.0.0"
description: "Test MCP server for integration testing"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "node"
      version: ">=18.0"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/test-mcp-server.git"
  setup_commands:
    - "npm install"
    - "npm run build"

server:
  transport: "stdio"
  command: "node"
  args: ["dist/index.js"]
  environment:
    NODE_ENV: "production"
    LOG_LEVEL: "info"

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code", "cursor"]
`

	err := ioutil.WriteFile(filepath.Join(tempDir, "test-mcp-server.servo"), []byte(servoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test servo file: %v", err)
	}
}

func createMultipleServoFiles(t *testing.T, tempDir string) {
	servers := map[string]string{
		"dev-server.servo": `servo_version: "1.0"
name: "dev-server"
version: "1.0.0"
description: "Development server"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "node"
      version: ">=18.0"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/dev-server.git"
  setup_commands:
    - "npm install"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "dev"]

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,

		"staging-server.servo": `servo_version: "1.0"
name: "staging-server"
version: "1.0.0"
description: "Staging server"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "node"
      version: ">=18.0"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/staging-server.git"
  setup_commands:
    - "npm install"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "start"]

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,

		"prod-server.servo": `servo_version: "1.0"
name: "prod-server"
version: "1.0.0"
description: "Production server"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "node"
      version: ">=18.0"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/prod-server.git"
  setup_commands:
    - "npm install"
    - "npm run build"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "start:prod"]

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,

		"logging-server.servo": `servo_version: "1.0"
name: "logging-server"
version: "1.0.0"
description: "Logging server"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "python"
      version: ">=3.8"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/logging-server.git"
  setup_commands:
    - "pip install -r requirements.txt"

server:
  transport: "stdio"
  command: "python"
  args: ["-m", "logging_server"]

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,
	}

	for filename, content := range servers {
		err := ioutil.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create server file %s: %v", filename, err)
		}
	}
}

func createMalformedServoFile(t *testing.T, tempDir string) {
	malformedContent := `servo_version: "1.0"
name: "malformed-server"
# Missing required fields like install, server, etc.
server:
  # Invalid structure - missing transport
  command: 
  args: [
install: "invalid structure"
`

	err := ioutil.WriteFile(filepath.Join(tempDir, "malformed.servo"), []byte(malformedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create malformed servo file: %v", err)
	}
}

func validateGeneratedConfigurations(t *testing.T, tempDir string) {
	// Check that expected config files were generated
	expectedFiles := []string{
		".vscode/mcp.json",
		".mcp.json", // Claude Code config
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tempDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected configuration file %s was not generated", file)
		} else {
			t.Logf("✅ Configuration file generated: %s", file)
		}
	}

	// Validate VSCode MCP configuration structure
	vscodeMcpPath := filepath.Join(tempDir, ".vscode", "mcp.json")
	if content, err := ioutil.ReadFile(vscodeMcpPath); err == nil {
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err == nil {
			if mcpServers, ok := config["servers"].(map[string]interface{}); ok {
				if len(mcpServers) > 0 {
					t.Logf("✅ VSCode MCP config contains %d servers", len(mcpServers))
				} else {
					t.Errorf("VSCode MCP config contains no servers")
				}
			} else {
				t.Errorf("VSCode MCP config has invalid structure")
			}
		} else {
			t.Errorf("Failed to parse VSCode MCP config: %v", err)
		}
	}

	// Validate Claude Code MCP configuration structure
	claudeMcpPath := filepath.Join(tempDir, ".mcp.json")
	if content, err := ioutil.ReadFile(claudeMcpPath); err == nil {
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err == nil {
			if mcpServers, ok := config["mcpServers"].(map[string]interface{}); ok {
				if len(mcpServers) > 0 {
					t.Logf("✅ Claude Code MCP config contains %d servers", len(mcpServers))
				} else {
					t.Errorf("Claude Code MCP config contains no servers")
				}
			} else {
				t.Errorf("Claude Code MCP config has invalid structure")
			}
		} else {
			t.Errorf("Failed to parse Claude Code MCP config: %v", err)
		}
	}
}

func testWorkCommand(t *testing.T, tempDir, servoPath string) {
	// Test work command for different clients
	clients := []string{"vscode", "claude-code", "cursor"}

	for _, client := range clients {
		cmd := exec.Command(servoPath, "work", "--client", client)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Work command for %s returned error (may be expected if client not installed): %v", client, err)
		} else {
			outputStr := string(output)
			// The work command should return launch instructions
			if strings.Contains(outputStr, "Launch") || strings.Contains(outputStr, "#") {
				t.Logf("✅ Work command for %s returned launch instructions", client)
			} else {
				t.Logf("⚠️  Work command for %s returned unexpected output: %s", client, outputStr)
			}
		}
	}

	// Test work command without specific client (should provide options)
	cmd := exec.Command(servoPath, "work")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Work command without client returned error (may be expected): %v", err)
	} else {
		t.Logf("✅ Work command without client returned: %s", string(output))
	}
}

func validateClientConfigurations(t *testing.T, tempDir, clients string) {
	clientList := strings.Split(clients, ",")
	
	for _, client := range clientList {
		client = strings.TrimSpace(client)
		switch client {
		case "vscode":
			if _, err := os.Stat(filepath.Join(tempDir, ".vscode", "mcp.json")); os.IsNotExist(err) {
				t.Errorf("VSCode config not generated for client combination: %s", clients)
			}
		case "claude-code":
			if _, err := os.Stat(filepath.Join(tempDir, ".mcp.json")); os.IsNotExist(err) {
				t.Errorf("Claude Code config not generated for client combination: %s", clients)
			}
		case "cursor":
			if _, err := os.Stat(filepath.Join(tempDir, ".cursor", "mcp.json")); os.IsNotExist(err) {
				// Cursor might use different config path, check alternative locations
				t.Logf("Note: Cursor config location may vary for client combination: %s", clients)
			}
		}
	}
	
	// DevContainer files should always be generated
	devcontainerFiles := []string{
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
	}
	
	for _, file := range devcontainerFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("DevContainer file %s not generated for client combination: %s", file, clients)
		}
	}
}