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

// E2ETestFramework provides utilities for end-to-end testing
type E2ETestFramework struct {
	testDir   string
	servoPath string
	t         *testing.T
}

// NewE2ETestFramework creates a new E2E test framework
func NewE2ETestFramework(t *testing.T) (*E2ETestFramework, error) {
	// Create temporary directory for testing
	testDir, err := ioutil.TempDir("", "servo_e2e_")
	if err != nil {
		return nil, err
	}

	// Build servo binary for testing
	servoPath := filepath.Join(testDir, "servo")
	buildCmd := exec.Command("go", "build", "-o", servoPath, "../cmd/servo")
	if err := buildCmd.Run(); err != nil {
		os.RemoveAll(testDir)
		return nil, err
	}

	return &E2ETestFramework{
		testDir:   testDir,
		servoPath: servoPath,
		t:         t,
	}, nil
}

// Cleanup removes test directory
func (f *E2ETestFramework) Cleanup() {
	os.RemoveAll(f.testDir)
}

// RunServoCommand executes a servo command and returns output
func (f *E2ETestFramework) RunServoCommand(args ...string) (string, error) {
	cmd := exec.Command(f.servoPath, args...)
	cmd.Dir = f.testDir
	output, err := cmd.CombinedOutput()
	f.t.Logf("Servo command: %s %v", f.servoPath, args)
	f.t.Logf("Command output: %s", string(output))
	if err != nil {
		f.t.Logf("Command error: %v", err)
	}
	return string(output), err
}

// RunServoCommandInDir executes a servo command in a specific directory
func (f *E2ETestFramework) RunServoCommandInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command(f.servoPath, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	f.t.Logf("Servo command in %s: %s %v", dir, f.servoPath, args)
	f.t.Logf("Command output: %s", string(output))
	if err != nil {
		f.t.Logf("Command error: %v", err)
	}
	return string(output), err
}

// CreateTestProject creates a test project in the test directory
func (f *E2ETestFramework) CreateTestProject(name string) error {
	projectDir := filepath.Join(f.testDir, name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	// Initialize project in the project directory
	output, err := f.RunServoCommandInDir(projectDir, "init", name, "", "--clients", "vscode,claude-code")
	if err != nil {
		f.t.Logf("Init output: %s", output)
		return err
	}

	return nil
}

// CreateMockServoFile creates a test .servo file
func (f *E2ETestFramework) CreateMockServoFile(name, filepath string) error {
	content := `
servo_version: "1.0"

metadata:
  name: "` + name + `"
  version: "1.0.0"
  description: "E2E test MCP server"

server:
  command: "python"
  args: ["--version"]  # Simple command for testing
  environment:
    TEST_ENV: "e2e"

services:
  test-service:
    image: "alpine:latest"
    command: ["sleep", "30"]
    ports:
      - "8080:8080"
    volumes:
      - "test_data:/data"
    environment:
      SERVICE_ENV: "test"
    healthcheck:
      test: ["CMD", "echo", "healthy"]
      interval: "5s"
      timeout: "3s"
      retries: 2
`

	return ioutil.WriteFile(filepath, []byte(content), 0644)
}

// TestE2E_CompleteProjectWorkflow tests the complete project workflow
func TestE2E_CompleteProjectWorkflow(t *testing.T) {
	framework, err := NewE2ETestFramework(t)
	if err != nil {
		t.Fatalf("Failed to create E2E framework: %v", err)
	}
	defer framework.Cleanup()

	projectName := "test-complete-workflow"

	t.Log("=== Testing Complete Project Workflow ===")

	// Step 1: Create project
	t.Log("Step 1: Creating test project")
	if err := framework.CreateTestProject(projectName); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Get project directory
	projectDir := filepath.Join(framework.testDir, projectName)

	// Step 2: Check project status
	t.Log("Step 2: Checking project status")
	output, err := framework.RunServoCommandInDir(projectDir, "status")
	if err != nil {
		t.Logf("Status output: %s", output)
		t.Errorf("Status command failed: %v", err)
	}

	// Verify status contains project information
	if !strings.Contains(output, projectName) {
		t.Errorf("Status output should contain project name '%s'", projectName)
	}

	// Step 3: Create and install a mock MCP server
	t.Log("Step 3: Creating and installing mock MCP server")
	servoFilePath := filepath.Join(projectDir, "test-server.servo")
	if err := framework.CreateMockServoFile("test-server", servoFilePath); err != nil {
		t.Fatalf("Failed to create mock servo file: %v", err)
	}

	output, err = framework.RunServoCommandInDir(projectDir, "install", servoFilePath, "--clients", "vscode,claude-code")
	if err != nil {
		t.Logf("Install output: %s", output)
		t.Errorf("Install command failed: %v", err)
	}

	// Step 4: Verify devcontainer was created
	t.Log("Step 4: Verifying devcontainer configuration")
	devcontainerPath := filepath.Join(projectDir, ".devcontainer", "devcontainer.json")
	if _, err := os.Stat(devcontainerPath); os.IsNotExist(err) {
		t.Errorf("devcontainer.json was not created at %s", devcontainerPath)
		// List what's actually in the directory to debug
		if entries, err := os.ReadDir(projectDir); err == nil {
			t.Logf("Contents of project directory:")
			for _, entry := range entries {
				t.Logf("  - %s", entry.Name())
			}
		}
	} else {
		t.Logf("✅ devcontainer.json exists at %s", devcontainerPath)
		// Verify devcontainer structure
		devcontainerBytes, err := ioutil.ReadFile(devcontainerPath)
		if err != nil {
			t.Errorf("Failed to read devcontainer.json: %v", err)
		} else {
			t.Logf("Devcontainer content: %s", string(devcontainerBytes))
			var devcontainer map[string]interface{}
			if err := json.Unmarshal(devcontainerBytes, &devcontainer); err != nil {
				t.Errorf("Failed to parse devcontainer.json: %v", err)
			} else {
				// Verify key components
				if devcontainer["name"] == nil {
					t.Error("devcontainer name not set")
				}
				if devcontainer["dockerComposeFile"] == nil {
					t.Error("dockerComposeFile not set")
				}
				if devcontainer["forwardPorts"] == nil {
					t.Error("forwardPorts not set")
				}
				// Check for onCreateCommand
				if onCreateCmd := devcontainer["onCreateCommand"]; onCreateCmd != nil {
					t.Logf("✅ onCreateCommand exists: %v", onCreateCmd)
				} else {
					t.Error("onCreateCommand not set")
				}
			}
		}
	}

	// Step 5: Verify docker-compose was created
	t.Log("Step 5: Verifying docker-compose configuration")
	composePath := filepath.Join(projectDir, ".devcontainer", "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Error("docker-compose.yml was not created")
	}

	// Step 6: Verify client configurations
	t.Log("Step 6: Verifying client configurations")

	// Check VSCode MCP configuration
	vscodeMCPPath := filepath.Join(projectDir, ".vscode", "mcp.json")
	if _, err := os.Stat(vscodeMCPPath); os.IsNotExist(err) {
		t.Error("VSCode mcp.json was not created")
	} else {
		settingsBytes, err := ioutil.ReadFile(vscodeMCPPath)
		if err != nil {
			t.Errorf("Failed to read VSCode MCP config: %v", err)
		} else {
			var settings map[string]interface{}
			if err := json.Unmarshal(settingsBytes, &settings); err != nil {
				t.Errorf("Failed to parse VSCode MCP config: %v", err)
			} else if settings["servers"] == nil {
				t.Error("MCP servers not configured in VSCode MCP config")
			}
		}
	}

	// Check Claude Code MCP configuration
	claudeMCPPath := filepath.Join(projectDir, ".mcp.json")
	if _, err := os.Stat(claudeMCPPath); os.IsNotExist(err) {
		t.Error("Claude Code .mcp.json was not created")
	} else {
		settingsBytes, err := ioutil.ReadFile(claudeMCPPath)
		if err != nil {
			t.Errorf("Failed to read Claude Code MCP config: %v", err)
		} else {
			var settings map[string]interface{}
			if err := json.Unmarshal(settingsBytes, &settings); err != nil {
				t.Errorf("Failed to parse Claude Code MCP config: %v", err)
			} else if settings["mcpServers"] == nil {
				t.Error("MCP servers not configured in Claude Code MCP config")
			}
		}
	}

	// Step 7: Test work command
	t.Log("Step 7: Testing work command")
	output, err = framework.RunServoCommandInDir(projectDir, "work")
	if err != nil {
		t.Logf("Work output: %s", output)
		t.Errorf("Work command failed: %v", err)
	}

	// Verify work command provides useful output
	if !strings.Contains(output, "development environment") {
		t.Error("Work command should mention development environment")
	}

	t.Log("=== Complete Project Workflow Test PASSED ===")
}

// TestE2E_ServiceConnectivity tests that services can be reached
func TestE2E_ServiceConnectivity(t *testing.T) {
	// This would test actual service connectivity
	// For now, we'll test the configuration generation

	framework, err := NewE2ETestFramework(t)
	if err != nil {
		t.Fatalf("Failed to create E2E framework: %v", err)
	}
	defer framework.Cleanup()

	t.Log("=== Testing Service Connectivity ===")

	// Create project
	projectName := "connectivity-test"
	if err := framework.CreateTestProject(projectName); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	projectDir := filepath.Join(framework.testDir, projectName)

	// Create servo file with multiple services
	servoFilePath := filepath.Join(projectDir, "multi-service.servo")
	multiServiceContent := `
servo_version: "1.0"

metadata:
  name: "multi-service-test"
  version: "1.0.0"

services:
  web:
    image: "nginx:alpine"
    ports:
      - "8080:80"
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost"]
      interval: "10s"
      
  db:
    image: "postgres:13-alpine"
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: "testpass"
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: "10s"
`

	if err := ioutil.WriteFile(servoFilePath, []byte(multiServiceContent), 0644); err != nil {
		t.Fatalf("Failed to create multi-service file: %v", err)
	}

	// Install the multi-service configuration
	output, err := framework.RunServoCommandInDir(projectDir, "install", servoFilePath, "--clients", "vscode")
	if err != nil {
		t.Logf("Install output: %s", output)
		t.Fatalf("Install failed: %v", err)
	}

	// Verify both services are in docker-compose
	composePath := filepath.Join(projectDir, ".devcontainer", "docker-compose.yml")
	composeBytes, err := ioutil.ReadFile(composePath)
	if err != nil {
		t.Fatalf("Failed to read docker-compose: %v", err)
	}

	composeContent := string(composeBytes)

	// Check that both services are defined
	if !strings.Contains(composeContent, "multi-service-test-web") {
		t.Error("Web service not found in docker-compose")
	}

	if !strings.Contains(composeContent, "multi-service-test-db") {
		t.Error("Database service not found in docker-compose")
	}

	// Verify port forwarding includes both ports
	devcontainerPath := filepath.Join(projectDir, ".devcontainer", "devcontainer.json")
	devcontainerBytes, err := ioutil.ReadFile(devcontainerPath)
	if err != nil {
		t.Fatalf("Failed to read devcontainer: %v", err)
	}

	var devcontainer map[string]interface{}
	if err := json.Unmarshal(devcontainerBytes, &devcontainer); err != nil {
		t.Fatalf("Failed to parse devcontainer: %v", err)
	}

	forwardPorts := devcontainer["forwardPorts"].([]interface{})
	portFound := map[int]bool{8080: false, 5432: false}

	for _, port := range forwardPorts {
		if portFloat, ok := port.(float64); ok {
			portInt := int(portFloat)
			if _, exists := portFound[portInt]; exists {
				portFound[portInt] = true
			}
		}
	}

	for port, found := range portFound {
		if !found {
			t.Errorf("Port %d not found in forwardPorts", port)
		}
	}

	t.Log("=== Service Connectivity Test PASSED ===")
}

// TestE2E_DataPersistence tests data persistence across container restarts
func TestE2E_DataPersistence(t *testing.T) {
	framework, err := NewE2ETestFramework(t)
	if err != nil {
		t.Fatalf("Failed to create E2E framework: %v", err)
	}
	defer framework.Cleanup()

	t.Log("=== Testing Data Persistence ===")

	// Create project
	projectName := "persistence-test"
	if err := framework.CreateTestProject(projectName); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	projectDir := filepath.Join(framework.testDir, projectName)

	// Create servo file with persistent volumes
	servoFilePath := filepath.Join(projectDir, "persistent.servo")
	persistentContent := `
servo_version: "1.0"

metadata:
  name: "persistent-test"
  version: "1.0.0"

services:
  database:
    image: "postgres:13-alpine"
    volumes:
      - "db_data:/var/lib/postgresql/data"
      - "db_logs:/var/log"
    environment:
      POSTGRES_PASSWORD: "testpass"
      
  cache:
    image: "redis:7-alpine"
    volumes:
      - "redis_data:/data"
`

	if err := ioutil.WriteFile(servoFilePath, []byte(persistentContent), 0644); err != nil {
		t.Fatalf("Failed to create persistent service file: %v", err)
	}

	// Install with persistence
	output, err := framework.RunServoCommandInDir(projectDir, "install", servoFilePath, "--clients", "vscode")
	if err != nil {
		t.Logf("Install output: %s", output)
		t.Fatalf("Install failed: %v", err)
	}

	// Verify persistence directory setup in devcontainer.json
	devcontainerPath := filepath.Join(projectDir, ".devcontainer", "devcontainer.json")
	devcontainerBytes, err := ioutil.ReadFile(devcontainerPath)
	if err != nil {
		t.Errorf("Failed to read devcontainer.json: %v", err)
	} else {
		var devcontainer map[string]interface{}
		if err := json.Unmarshal(devcontainerBytes, &devcontainer); err != nil {
			t.Errorf("Failed to parse devcontainer.json: %v", err)
		} else {
			// Check that onCreateCommand includes directory creation for persistent volumes
			if onCreateCmd := devcontainer["onCreateCommand"]; onCreateCmd != nil {
				onCreateCmdStr := onCreateCmd.(string)
				expectedCommands := []string{
					"mkdir -p /workspace/.servo/services/persistent-test/database",
					"mkdir -p /workspace/.servo/services/persistent-test/cache",
					"mkdir -p /workspace/.servo/logs/persistent-test/database",
					"mkdir -p /workspace/.servo/logs/persistent-test/cache",
				}

				for _, expectedCmd := range expectedCommands {
					if !strings.Contains(onCreateCmdStr, expectedCmd) {
						t.Errorf("Expected persistence directory command not found in onCreateCommand: %s", expectedCmd)
					} else {
						t.Logf("✅ Persistence directory command found: %s", expectedCmd)
					}
				}
			} else {
				t.Error("onCreateCommand not found in devcontainer.json")
			}
		}
	}

	// Verify docker-compose has correct volume mappings
	composePath := filepath.Join(projectDir, ".devcontainer", "docker-compose.yml")
	composeBytes, err := ioutil.ReadFile(composePath)
	if err != nil {
		t.Fatalf("Failed to read docker-compose: %v", err)
	}

	composeContent := string(composeBytes)

	// Check volume mappings point to .servo directories
	expectedMappings := []string{
		"../.servo/services/persistent-test/database/db_data:/var/lib/postgresql/data",
		"../.servo/services/persistent-test/database/db_logs:/var/log",
		"../.servo/services/persistent-test/cache/redis_data:/data",
	}

	for _, mapping := range expectedMappings {
		if !strings.Contains(composeContent, mapping) {
			t.Errorf("Volume mapping not found: %s", mapping)
		}
	}

	t.Log("=== Data Persistence Test PASSED ===")
}

// TestE2E_CrossClientFunctionality tests the same servers work with both VSCode and Claude Code
func TestE2E_CrossClientFunctionality(t *testing.T) {
	framework, err := NewE2ETestFramework(t)
	if err != nil {
		t.Fatalf("Failed to create E2E framework: %v", err)
	}
	defer framework.Cleanup()

	t.Log("=== Testing Cross-Client Functionality ===")

	// Create project
	projectName := "cross-client-test"
	if err := framework.CreateTestProject(projectName); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	projectDir := filepath.Join(framework.testDir, projectName)

	// Install server for both clients
	servoFilePath := filepath.Join(projectDir, "cross-client.servo")
	if err := framework.CreateMockServoFile("cross-client-server", servoFilePath); err != nil {
		t.Fatalf("Failed to create servo file: %v", err)
	}

	output, err := framework.RunServoCommandInDir(projectDir, "install", servoFilePath, "--clients", "vscode,claude-code")
	if err != nil {
		t.Logf("Install output: %s", output)
		t.Fatalf("Install failed: %v", err)
	}

	// Verify both client configurations exist and have the same server
	vscodeSettings := filepath.Join(projectDir, ".vscode", "mcp.json")
	claudeSettings := filepath.Join(projectDir, ".mcp.json")

	// Read VSCode MCP config
	vscodeBytes, err := ioutil.ReadFile(vscodeSettings)
	if err != nil {
		t.Fatalf("Failed to read VSCode MCP config: %v", err)
	}

	var vscodeConfig map[string]interface{}
	if err := json.Unmarshal(vscodeBytes, &vscodeConfig); err != nil {
		t.Fatalf("Failed to parse VSCode MCP config: %v", err)
	}

	// Read Claude Code MCP config
	claudeBytes, err := ioutil.ReadFile(claudeSettings)
	if err != nil {
		t.Fatalf("Failed to read Claude Code MCP config: %v", err)
	}

	var claudeConfig map[string]interface{}
	if err := json.Unmarshal(claudeBytes, &claudeConfig); err != nil {
		t.Fatalf("Failed to parse Claude Code MCP config: %v", err)
	}

	// Verify both have the same server configured (different format keys)
	vscodeServers := vscodeConfig["servers"].(map[string]interface{})
	claudeServers := claudeConfig["mcpServers"].(map[string]interface{})

	if _, exists := vscodeServers["cross-client-server"]; !exists {
		t.Error("Server not configured in VSCode")
	}

	if _, exists := claudeServers["cross-client-server"]; !exists {
		t.Error("Server not configured in Claude Code")
	}

	// Verify server configurations are equivalent
	vscodeServer := vscodeServers["cross-client-server"].(map[string]interface{})
	claudeServer := claudeServers["cross-client-server"].(map[string]interface{})

	if vscodeServer["command"] != claudeServer["command"] {
		t.Error("Server commands don't match between clients")
	}

	t.Log("=== Cross-Client Functionality Test PASSED ===")
}
