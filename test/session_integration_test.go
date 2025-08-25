package test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestSessionIntegration_DefaultSession validates default session functionality
func TestSessionIntegration_DefaultSession(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_session_default_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running default session integration test in: %s", tempDir)

	// Use absolute path to servo binary
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Step 1: Initialize project without session - should create default session
	cmd := exec.Command(servoPath, "init", "test-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}

	// Verify output mentions default session
	outputStr := string(output)
	if !strings.Contains(outputStr, "Default session: default") {
		t.Errorf("Expected output to show default session, got: %s", outputStr)
	}

	// Step 2: Verify project structure created correctly
	expectedDirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/default",
		".servo/sessions/default/manifests",
		".servo/sessions/default/volumes",
		".servo/sessions/default/config",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	// Step 3: Verify project configuration
	projectYaml, err := ioutil.ReadFile(".servo/project.yaml")
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	if !strings.Contains(string(projectYaml), "default_session: default") {
		t.Errorf("project.yaml should contain default_session: default, got: %s", string(projectYaml))
	}

	if !strings.Contains(string(projectYaml), "active_session: default") {
		t.Errorf("project.yaml should contain active_session: default, got: %s", string(projectYaml))
	}

	t.Logf("✅ Default session integration test passed")
}

// TestSessionIntegration_CustomSession validates custom session functionality
func TestSessionIntegration_CustomSession(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_session_custom_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running custom session integration test in: %s", tempDir)

	// Use absolute path to servo binary
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Step 1: Initialize project with custom session
	cmd := exec.Command(servoPath, "init", "--session", "development", "test-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}

	// Verify output mentions custom session
	outputStr := string(output)
	if !strings.Contains(outputStr, "Default session: development") {
		t.Errorf("Expected output to show default session as 'development', got: %s", outputStr)
	}

	// Step 2: Verify custom session directory structure
	expectedDirs := []string{
		".servo/sessions/development",
		".servo/sessions/development/manifests",
		".servo/sessions/development/volumes",
		".servo/sessions/development/config",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	// Step 3: Verify project configuration uses custom session
	projectYaml, err := ioutil.ReadFile(".servo/project.yaml")
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	if !strings.Contains(string(projectYaml), "default_session: development") {
		t.Errorf("project.yaml should contain default_session: development, got: %s", string(projectYaml))
	}

	if !strings.Contains(string(projectYaml), "active_session: development") {
		t.Errorf("project.yaml should contain active_session: development, got: %s", string(projectYaml))
	}

	t.Logf("✅ Custom session integration test passed")
}

// TestSessionIntegration_SessionSpecificInstallation validates session-specific MCP server installation
func TestSessionIntegration_SessionSpecificInstallation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_session_install_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running session-specific installation test in: %s", tempDir)

	// Use absolute path to servo binary
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Step 1: Initialize project
	cmd := exec.Command(servoPath, "init", "test-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}

	// Step 2: Create a test .servo file
	testServoFile := `servo_version: "1.0"

metadata:
  name: "test-server"
  version: "1.0.0"
  description: "Test MCP server for session testing"

install:
  type: "local"
  method: "local"
  setup_commands:
    - "echo 'Test server setup complete'"

server:
  transport: "stdio"
  command: "echo"
  args: ["Test server running"]`

	err = ioutil.WriteFile("test-server.servo", []byte(testServoFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create test servo file: %v", err)
	}

	// Step 3: Install to default session (should use "default")
	cmd = exec.Command(servoPath, "install", "test-server.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install to default session: %v\nOutput: %s", err, output)
	}

	// Verify output shows correct session
	outputStr := string(output)
	if !strings.Contains(outputStr, "session: default") {
		t.Errorf("Expected install output to show default session, got: %s", outputStr)
	}

	// Step 4: Create testing session
	cmd = exec.Command(servoPath, "session", "create", "testing", "--description", "Test session")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create testing session: %v\nOutput: %s", err, output)
	}

	// Step 5: Install same server to specific session (should work now)
	cmd = exec.Command(servoPath, "install", "--session", "testing", "test-server.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install to testing session: %v\nOutput: %s", err, output)
	}

	// Verify output shows correct session
	outputStr = string(output)
	if !strings.Contains(outputStr, "session: testing") {
		t.Errorf("Expected install output to show testing session, got: %s", outputStr)
	}

	// Step 6: Verify session directories were created
	expectedDirs := []string{
		".servo/sessions/default",
		".servo/sessions/testing",
		".servo/sessions/testing/manifests",
		".servo/sessions/testing/config",
		".servo/sessions/testing/volumes",
		".servo/sessions/testing/logs",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	t.Logf("✅ Session-specific installation test passed")
}

// TestSessionIntegration_MultipleSessionsAndSwitching validates multiple sessions and switching
func TestSessionIntegration_MultipleSessionsAndSwitching(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_session_switching_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running session switching integration test in: %s", tempDir)

	// Use absolute path to servo binary
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Step 1: Initialize project
	cmd := exec.Command(servoPath, "init", "test-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}

	// Step 2: Create different test servers for different sessions
	server1Content := `servo_version: "1.0"
metadata:
  name: "server1"
  version: "1.0.0"
  description: "Server 1 for session testing"
install:
  type: "local"
  method: "local"
  setup_commands:
    - "echo 'Server 1 setup'"
server:
  transport: "stdio"
  command: "echo"
  args: ["Server 1"]`

	server2Content := `servo_version: "1.0"
metadata:
  name: "server2"
  version: "1.0.0"
  description: "Server 2 for session testing"
install:
  type: "local"
  method: "local"
  setup_commands:
    - "echo 'Server 2 setup'"
server:
  transport: "stdio"
  command: "echo"
  args: ["Server 2"]`

	err = ioutil.WriteFile("server1.servo", []byte(server1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create server1.servo: %v", err)
	}

	err = ioutil.WriteFile("server2.servo", []byte(server2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create server2.servo: %v", err)
	}

	// Step 3: Create dev and prod sessions
	cmd = exec.Command(servoPath, "session", "create", "dev", "--description", "Development session")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create dev session: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command(servoPath, "session", "create", "prod", "--description", "Production session")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create prod session: %v\nOutput: %s", err, output)
	}

	// Step 4: Install server1 to session "dev"
	cmd = exec.Command(servoPath, "install", "--session", "dev", "server1.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install server1 to dev session: %v\nOutput: %s", err, output)
	}

	// Step 5: Install server2 to session "prod"
	cmd = exec.Command(servoPath, "install", "--session", "prod", "server2.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install server2 to prod session: %v\nOutput: %s", err, output)
	}

	// Step 6: Verify both sessions exist with their directories
	expectedDirs := []string{
		".servo/sessions/default",
		".servo/sessions/dev",
		".servo/sessions/prod",
		".servo/sessions/dev/manifests",
		".servo/sessions/prod/manifests",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}

	// Step 7: Verify client configurations were created
	expectedFiles := []string{
		".vscode/mcp.json",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected configuration file not created: %s", file)
		}
	}

	// Step 8: Activate dev session and verify VSCode config shows server1
	cmd = exec.Command(servoPath, "session", "activate", "dev")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to activate dev session: %v\nOutput: %s", err, output)
	}

	// Install same server1 to default session explicitly (should work now - same server, different session)
	cmd = exec.Command(servoPath, "install", "--session", "default", "server1.servo")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install prod server: %v\nOutput: %s", err, output)
	}

	// Verify VSCode MCP config contains server1 (from dev session)
	vscodeConfig, err := ioutil.ReadFile(".vscode/mcp.json")
	if err != nil {
		t.Fatalf("Failed to read .vscode/mcp.json: %v", err)
	}

	var mcpConfig map[string]interface{}
	if err := json.Unmarshal(vscodeConfig, &mcpConfig); err != nil {
		t.Fatalf("Failed to parse VSCode MCP config: %v", err)
	}

	servers, ok := mcpConfig["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("VSCode MCP config missing servers section")
	}

	// Should have server1 since dev session is active
	if len(servers) < 1 {
		t.Errorf("Expected at least 1 server in VSCode MCP config, got %d", len(servers))
	}

	if _, exists := servers["server1"]; !exists {
		t.Errorf("server1 not found in VSCode MCP config when dev session is active")
	}

	t.Logf("✅ Session switching integration test passed - %d servers configured in active session", len(servers))
}
