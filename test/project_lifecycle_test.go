package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestProjectLifecycle_CompleteFlow tests the complete project lifecycle:
// init -> secrets -> install -> configure -> work -> validate -> clean workflows
func TestProjectLifecycle_CompleteFlow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_lifecycle_complete_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running complete project lifecycle test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Phase 1: Project Initialization
	t.Logf("🏗️  Phase 1: Project Initialization")
	cmd := exec.Command(servoPath, "init", "lifecycle-test", "--description", "Complete lifecycle test project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized successfully")

	// Verify project structure was created
	verifyProjectStructure(t, tempDir)

	// Phase 2: Secrets Configuration (for servers that need them)
	t.Logf("🔐 Phase 2: Secrets Configuration")
	createServoFileWithSecrets(t, tempDir)
	
	// Set up required secrets
	secrets := map[string]string{
		"api_key":      "test-api-key-12345",
		"database_url": "postgresql://user:pass@localhost:5432/testdb",
		"auth_token":   "test-auth-token-67890",
	}

	for key, value := range secrets {
		cmd = exec.Command(servoPath, "secrets", "set", key, value)
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to set secret %s: %v\nOutput: %s", key, err, output)
		}
		t.Logf("✅ Secret %s configured", key)
	}

	// Verify secrets can be listed
	cmd = exec.Command(servoPath, "secrets", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list secrets: %v", err)
	}
	
	secretsList := string(output)
	for key := range secrets {
		if !strings.Contains(secretsList, key) {
			t.Errorf("Secret %s not found in secrets list: %s", key, secretsList)
		}
	}
	t.Logf("✅ All secrets configured and verified")

	// Phase 3: Server Installation
	t.Logf("📦 Phase 3: Server Installation")
	cmd = exec.Command(servoPath, "install", "secure-server.servo", "--clients", "vscode,claude-code")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Server with secrets installed successfully")

	// Phase 4: Configuration Generation
	t.Logf("⚙️  Phase 4: Configuration Generation")
	cmd = exec.Command(servoPath, "configure")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to configure clients: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Client configurations generated")

	// Validate that configurations contain secret references but not actual values
	validateSecretHandling(t, tempDir)

	// Phase 5: Status Verification
	t.Logf("📊 Phase 5: Status Verification")
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get project status: %v", err)
	}
	
	statusOutput := string(output)
	if !strings.Contains(statusOutput, "1 configured") {
		t.Errorf("Expected 1 server configured, got: %s", statusOutput)
	}
	if !strings.Contains(statusOutput, "Configured secrets: 3") {
		t.Errorf("Expected 3 secrets configured, got: %s", statusOutput)
	}
	t.Logf("✅ Status verified: 1 server, 3 secrets")

	// Phase 6: Work Command Testing
	t.Logf("🚀 Phase 6: Work Command Testing")
	testWorkCommandWithSecrets(t, servoPath)

	// Phase 7: Session Management
	t.Logf("🔄 Phase 7: Session Management")
	testSessionManagementWorkflow(t, servoPath)

	// Phase 8: Project Validation
	t.Logf("🔍 Phase 8: Project Validation")
	validateCompleteProject(t, tempDir)

	t.Logf("✅ Complete project lifecycle test passed!")
}

// TestProjectLifecycle_TeamCollaboration tests collaboration scenarios
func TestProjectLifecycle_TeamCollaboration(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_lifecycle_team_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running team collaboration lifecycle test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Simulate original team member setup
	t.Logf("👤 Original Team Member: Project Setup")
	setupOriginalMemberProject(t, servoPath, tempDir)

	// Simulate new team member onboarding
	t.Logf("👥 New Team Member: Project Onboarding") 
	simulateNewMemberOnboarding(t, servoPath, tempDir)

	// Test that both members can work with the project
	t.Logf("🤝 Team Collaboration: Verify Both Members Can Work")
	verifyTeamCollaboration(t, servoPath, tempDir)

	t.Logf("✅ Team collaboration lifecycle test passed!")
}

// TestProjectLifecycle_EnvironmentVariations tests different environment scenarios
func TestProjectLifecycle_EnvironmentVariations(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_lifecycle_env_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running environment variations lifecycle test in: %s", tempDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test different environment configurations
	environments := []struct {
		name    string
		session string
		clients string
	}{
		{"development", "dev", "vscode,claude-code"},
		{"staging", "staging", "vscode"},
		{"production", "prod", "claude-code"},
	}

	// Initialize project
	cmd := exec.Command(servoPath, "init", "env-variations-test")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	createEnvironmentServoFiles(t, tempDir)

	for _, env := range environments {
		t.Logf("🌍 Testing %s environment", env.name)
		
		// Create session for environment
		cmd = exec.Command(servoPath, "session", "create", env.session, "--description", fmt.Sprintf("%s environment", env.name))
		_, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create %s session: %v", env.name, err)
		}

		// Install environment-specific server
		cmd = exec.Command(servoPath, "install", "--session", env.session, fmt.Sprintf("%s-server.servo", env.name), "--clients", env.clients)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install %s server: %v\nOutput: %s", env.name, err, output)
		}

		// Switch to this environment
		cmd = exec.Command(servoPath, "session", "activate", env.session)
		_, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to activate %s session: %v", env.name, err)
		}

		// Configure for this environment
		cmd = exec.Command(servoPath, "configure")
		_, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to configure %s environment: %v", env.name, err)
		}

		// Verify configuration
		validateEnvironmentConfiguration(t, tempDir, env.name, env.clients)
		
		t.Logf("✅ %s environment configured and validated", env.name)
	}

	t.Logf("✅ Environment variations lifecycle test passed!")
}

// Helper functions

func verifyProjectStructure(t *testing.T, tempDir string) {
	expectedDirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/default",
		".servo/sessions/default/manifests",
		".servo/sessions/default/config",
		".servo/sessions/default/volumes",
	}

	expectedFiles := []string{
		".servo/project.yaml",
		".servo/.gitignore",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(filepath.Join(tempDir, dir)); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", dir)
		}
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

func createServoFileWithSecrets(t *testing.T, tempDir string) {
	servoContent := `servo_version: "1.0"
name: "secure-server"
version: "1.0.0"
description: "Server requiring secrets"
author: "Test Author"
license: "MIT"

requirements:
  runtimes:
    - name: "node"
      version: ">=18.0"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/secure-server.git"
  setup_commands:
    - "npm install"
    - "npm run build"

server:
  transport: "stdio"
  command: "node"
  args: ["secure-server.js"]
  environment:
    API_KEY: "{{SERVO_SECRET:api_key}}"
    DATABASE_URL: "{{SERVO_SECRET:database_url}}"
    AUTH_TOKEN: "{{SERVO_SECRET:auth_token}}"
    NODE_ENV: "production"

configuration_schema:
  secrets:
    api_key:
      description: "API key for external service"
      required: true
    database_url:
      description: "Database connection URL"
      required: true
    auth_token:
      description: "Authentication token"
      required: true

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code", "cursor"]
`

	err := ioutil.WriteFile(filepath.Join(tempDir, "secure-server.servo"), []byte(servoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create servo file with secrets: %v", err)
	}
}

func validateSecretHandling(t *testing.T, tempDir string) {
	configFiles := []string{
		".vscode/mcp.json",
		".mcp.json",
		".devcontainer/docker-compose.yml",
	}

	for _, configFile := range configFiles {
		fullPath := filepath.Join(tempDir, configFile)
		if content, err := ioutil.ReadFile(fullPath); err == nil {
			contentStr := string(content)
			
			// Should NOT contain actual secret values
			secretValues := []string{"test-api-key-12345", "test-auth-token-67890", "postgresql://user:pass@localhost"}
			for _, secret := range secretValues {
				if strings.Contains(contentStr, secret) {
					t.Errorf("Configuration file %s contains actual secret value: %s", configFile, secret)
				}
			}

			// Should contain secret placeholders or environment variable references
			if strings.Contains(contentStr, "SERVO_SECRET") || strings.Contains(contentStr, "API_KEY") {
				t.Logf("✅ Configuration %s properly handles secrets", configFile)
			}
		}
	}
}

func testWorkCommandWithSecrets(t *testing.T, servoPath string) {
	// Test that work command provides launch instructions without exposing secrets
	cmd := exec.Command(servoPath, "work", "--client", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Work command returned error (may be expected if VSCode not installed): %v", err)
	}
	
	outputStr := string(output)
	
	// Should not contain actual secret values in output
	secretValues := []string{"test-api-key-12345", "test-auth-token-67890"}
	for _, secret := range secretValues {
		if strings.Contains(outputStr, secret) {
			t.Errorf("Work command output contains secret value: %s", secret)
		}
	}
	
	t.Logf("✅ Work command properly handles secrets")
}

func testSessionManagementWorkflow(t *testing.T, servoPath string) {
	// Create a test session
	cmd := exec.Command(servoPath, "session", "create", "test-session", "--description", "Test session for lifecycle")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// List sessions
	cmd = exec.Command(servoPath, "session", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}
	
	if !strings.Contains(string(output), "test-session") {
		t.Errorf("Test session not found in session list: %s", string(output))
	}

	// Switch to test session
	cmd = exec.Command(servoPath, "session", "activate", "test-session")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to activate test session: %v", err)
	}

	// Verify active session
	cmd = exec.Command(servoPath, "session", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list sessions after activation: %v", err)
	}
	
	outputStr := string(output)
	if !strings.Contains(outputStr, "test-session") || !strings.Contains(outputStr, "active") {
		t.Errorf("Test session not shown as active: %s", outputStr)
	}

	t.Logf("✅ Session management workflow validated")
}

func validateCompleteProject(t *testing.T, tempDir string) {
	// Verify all expected files exist
	expectedFiles := []string{
		".servo/project.yaml",
		".servo/secrets.yaml",
		".vscode/mcp.json",
		".mcp.json",
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("Expected project file %s was not created", file)
		} else {
			t.Logf("✅ Project file exists: %s", file)
		}
	}

	// Verify project configuration structure
	projectFile := filepath.Join(tempDir, ".servo", "project.yaml")
	if content, err := ioutil.ReadFile(projectFile); err == nil {
		if strings.Contains(string(content), "secure-server") {
			t.Logf("✅ Project configuration references installed server")
		} else {
			t.Errorf("Project configuration does not reference installed server")
		}
	}
}

func setupOriginalMemberProject(t *testing.T, servoPath, tempDir string) {
	// Original member initializes project
	cmd := exec.Command(servoPath, "init", "team-project", "--description", "Team collaboration test")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Original member failed to init project: %v", err)
	}

	// Create team servo file
	createTeamServoFile(t, tempDir)
	
	// Install server
	cmd = exec.Command(servoPath, "install", "team-server.servo")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Original member failed to install server: %v", err)
	}

	t.Logf("✅ Original member set up project")
}

func simulateNewMemberOnboarding(t *testing.T, servoPath, tempDir string) {
	// New member should be able to configure existing project
	cmd := exec.Command(servoPath, "configure")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("New member failed to configure project: %v\nOutput: %s", err, output)
	}

	// New member should see project status
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("New member failed to get status: %v", err)
	}
	
	if !strings.Contains(string(output), "1 configured") {
		t.Errorf("New member doesn't see configured server: %s", string(output))
	}

	t.Logf("✅ New member successfully onboarded")
}

func verifyTeamCollaboration(t *testing.T, servoPath, tempDir string) {
	// Both members should be able to work with the project
	expectedFiles := []string{
		".vscode/mcp.json",
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("Team collaboration file %s missing", file)
		}
	}

	// Both should see the same status
	cmd := exec.Command(servoPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Team collaboration status check failed: %v", err)
	}

	if !strings.Contains(string(output), "team-server") {
		t.Errorf("Team server not visible in collaborative project: %s", string(output))
	}

	t.Logf("✅ Team collaboration verified")
}

func createEnvironmentServoFiles(t *testing.T, tempDir string) {
	environments := map[string]string{
		"development-server.servo": `servo_version: "1.0"
name: "development-server"
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
  repository: "https://github.com/example/development-server.git"
  setup_commands:
    - "npm install"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "dev"]
  environment:
    NODE_ENV: "development"
    DEBUG: "true"

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
  environment:
    NODE_ENV: "staging"
    DEBUG: "false"

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,

		"production-server.servo": `servo_version: "1.0"
name: "production-server"
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
  repository: "https://github.com/example/production-server.git"
  setup_commands:
    - "npm install"
    - "npm run build"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "start:prod"]
  environment:
    NODE_ENV: "production"
    DEBUG: "false"

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]`,
	}

	for filename, content := range environments {
		err := ioutil.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create environment servo file %s: %v", filename, err)
		}
	}
}

func createTeamServoFile(t *testing.T, tempDir string) {
	servoContent := `servo_version: "1.0"
name: "team-server"
version: "1.0.0"
description: "Team collaboration server"
author: "Team Lead"
license: "MIT"

requirements:
  runtimes:
    - name: "python"
      version: ">=3.8"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/example/team-server.git"
  setup_commands:
    - "pip install -r requirements.txt"

server:
  transport: "stdio"
  command: "python"
  args: ["-m", "team_server"]
  environment:
    TEAM_MODE: "true"
    LOG_LEVEL: "info"

clients:
  recommended: ["vscode", "claude-code"]
  tested: ["vscode", "claude-code"]
`

	err := ioutil.WriteFile(filepath.Join(tempDir, "team-server.servo"), []byte(servoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create team servo file: %v", err)
	}
}

func validateEnvironmentConfiguration(t *testing.T, tempDir, envName, clients string) {
	clientList := strings.Split(clients, ",")
	
	// Check that appropriate client configs exist
	for _, client := range clientList {
		client = strings.TrimSpace(client)
		var configFile string
		
		switch client {
		case "vscode":
			configFile = ".vscode/mcp.json"
		case "claude-code":
			configFile = ".mcp.json"
		case "cursor":
			configFile = ".cursor/mcp.json"
		}
		
		if configFile != "" {
			fullPath := filepath.Join(tempDir, configFile)
			if content, err := ioutil.ReadFile(fullPath); err == nil {
				if strings.Contains(string(content), fmt.Sprintf("%s-server", envName)) {
					t.Logf("✅ %s config contains %s environment server", client, envName)
				} else {
					t.Errorf("%s config missing %s environment server", client, envName)
				}
			}
		}
	}
}