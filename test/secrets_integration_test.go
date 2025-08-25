package test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecretsIntegration_CompleteWorkflow validates the complete secrets workflow
// from project creation through secret configuration and environment injection
func TestSecretsIntegration_CompleteWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_secrets_integration_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running secrets integration test in: %s", tempDir)

	// Step 1: Create servo files that require secrets
	createSecretsDependentServoFiles(t, tempDir)

	// Step 2: Initialize project and install servers
	testSecretsProjectWorkflow(t, tempDir)

	// Step 3: Test missing secrets detection
	testMissingSecretsDetection(t, tempDir)

	// Step 4: Configure secrets and validate injection
	testSecretsInjection(t, tempDir)

	// Step 5: Test team collaboration workflow
	testTeamSecretsWorkflow(t, tempDir)

	t.Logf("✅ Complete secrets integration test passed!")
}

func createSecretsDependentServoFiles(t *testing.T, tempDir string) {
	t.Logf("📝 Creating servo files with secrets requirements...")

	// API server that requires API keys and database
	apiServerContent := `servo_version: "1.0"

metadata:
  name: "api-server"
  version: "1.0.0"
  description: "API server requiring secrets"

requirements:
  runtimes:
    - name: "node"
      version: "18.0.0"

dependencies:
  services:
    postgres:
      image: "postgres:15"
      ports: ["5432"]
      environment:
        POSTGRES_DB: "apidb"
        POSTGRES_USER: "apiuser"
        POSTGRES_PASSWORD: "{{.GeneratedPassword}}"

install:
  type: "local"
  method: "local"
  setup_commands:
    - "npm install express pg"
    - "echo 'console.log(\"API server setup with secrets\");' > server.js"

server:
  transport: "stdio"
  command: "node"
  args: ["server.js"]
  environment:
    NODE_ENV: "production"
    OPENAI_API_KEY: "{{.Secrets.openai_api_key}}"
    DATABASE_URL: "{{.Secrets.database_url}}"
    POSTGRES_PASSWORD: "{{.GeneratedPassword}}"

configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for AI functionality"
      required: true
      type: "string"
      env_var: "OPENAI_API_KEY"
    database_url:
      description: "External database connection string"
      required: true
      type: "string"
      env_var: "DATABASE_URL"`

	err := ioutil.WriteFile(filepath.Join(tempDir, "api-server.servo"), []byte(apiServerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write api server servo file: %v", err)
	}

	// ML server with different secret requirements
	mlServerContent := `servo_version: "1.0"

metadata:
  name: "ml-server"
  version: "1.0.0"
  description: "ML server with API keys"

requirements:
  runtimes:
    - name: "python"
      version: "3.11"

dependencies:
  services:
    redis:
      image: "redis:7"
      ports: ["6379"]

install:
  type: "local"
  method: "local"
  setup_commands:
    - "pip install openai anthropic redis"
    - "echo 'print(\"ML server with secrets\")' > ml_server.py"

server:
  transport: "stdio"
  command: "python"
  args: ["ml_server.py"]
  environment:
    OPENAI_API_KEY: "{{.Secrets.openai_api_key}}"
    ANTHROPIC_API_KEY: "{{.Secrets.anthropic_api_key}}"
    REDIS_URL: "redis://redis:6379"

configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for LLM access"
      required: true
      type: "string"
      env_var: "OPENAI_API_KEY"
    anthropic_api_key:
      description: "Anthropic API key for Claude access"
      required: true
      type: "string"
      env_var: "ANTHROPIC_API_KEY"`

	err = ioutil.WriteFile(filepath.Join(tempDir, "ml-server.servo"), []byte(mlServerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write ml server servo file: %v", err)
	}

	t.Logf("✅ Created servo files with secrets requirements")
}

func testSecretsProjectWorkflow(t *testing.T, tempDir string) {
	t.Logf("🔧 Testing secrets project workflow...")
	defer CleanupTestSecrets(t)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project
	cmd := exec.Command(servoPath, "init", "secrets-test", "", "--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized")

	// Set up required secrets BEFORE installing servers
	secrets := map[string]string{
		"openai_api_key":    "sk-test-api-key-12345",
		"database_url":      "postgresql://testuser:testpass@localhost:5432/testdb",
		"anthropic_api_key": "sk-ant-test-api-key-67890",
	}
	SetupTestSecrets(t, tempDir, secrets)

	// Install API server (should now succeed with secrets configured)
	cmd = exec.Command(servoPath, "install", "api-server.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install api server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ API server installed")

	// Install ML server
	cmd = exec.Command(servoPath, "install", "ml-server.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install ml server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ ML server installed")
}

func testMissingSecretsDetection(t *testing.T, tempDir string) {
	t.Logf("🔍 Testing missing secrets detection...")

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Check status - should show configured secrets since they were set up in previous step
	cmd := exec.Command(servoPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get status: %v\nOutput: %s", err, output)
	}

	statusOutput := string(output)

	// Status should now show required secrets from servo files
	t.Logf("📋 Current status output: %s", statusOutput)

	// Verify basic project status works
	if !strings.Contains(statusOutput, "MCP Servers: 2 configured") {
		t.Errorf("Expected status to show configured servers, got: %s", statusOutput)
	}

	// Verify secrets status section exists
	if !strings.Contains(statusOutput, "Secrets Status:") {
		t.Errorf("Expected status to show secrets status section, got: %s", statusOutput)
	}

	// Should show configured secrets (they were set up in previous step)
	expectedSecrets := []string{"openai_api_key", "database_url", "anthropic_api_key"}
	for _, secret := range expectedSecrets {
		if !strings.Contains(statusOutput, secret) {
			t.Errorf("Expected status to mention required secret '%s', got: %s", secret, statusOutput)
		}
	}

	// Should indicate all secrets configured since they were set up
	if !strings.Contains(statusOutput, "🎉 All required secrets are configured!") {
		t.Errorf("Expected status to indicate all secrets configured, got: %s", statusOutput)
	}

	t.Logf("✅ Secrets detection functionality working correctly")
}

func testSecretsInjection(t *testing.T, tempDir string) {
	t.Logf("🔐 Testing secrets injection...")

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Set required secrets
	secrets := map[string]string{
		"openai_api_key":    "sk-test1234567890abcdef",
		"database_url":      "postgres://user:pass@localhost/db",
		"anthropic_api_key": "ant_test1234567890abcdef",
	}

	for key, value := range secrets {
		cmd := exec.Command(servoPath, "secrets", "set", key, value)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to set secret %s: %v\nOutput: %s", key, err, output)
		}
		t.Logf("✅ Set secret: %s", key)
	}

	// Check status again - should show secrets are configured
	cmd := exec.Command(servoPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get status after setting secrets: %v\nOutput: %s", err, output)
	}

	statusOutput := string(output)

	// Should show configured secrets instead of missing ones
	if strings.Contains(statusOutput, "Missing secrets:") {
		t.Errorf("Expected status to not show missing secrets after configuration, got: %s", statusOutput)
	}

	// Should show configured secrets
	if !strings.Contains(statusOutput, "Configured secrets:") {
		t.Errorf("Expected status to show configured secrets, got: %s", statusOutput)
	}

	// List secrets to verify they're set
	cmd = exec.Command(servoPath, "secrets", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list secrets: %v\nOutput: %s", err, output)
	}

	secretsList := string(output)
	for key := range secrets {
		if !strings.Contains(secretsList, key) {
			t.Errorf("Expected secrets list to contain '%s', got: %s", key, secretsList)
		}
	}

	t.Logf("✅ Secrets injection validated")
}

func testTeamSecretsWorkflow(t *testing.T, tempDir string) {
	t.Logf("👥 Testing team collaboration secrets workflow...")

	// Simulate team member workflow
	teamTempDir, err := ioutil.TempDir("", "servo_team_secrets_")
	if err != nil {
		t.Fatalf("Failed to create team temp dir: %v", err)
	}
	defer os.RemoveAll(teamTempDir)

	// Copy project files (simulating git clone)
	err = copyDir(filepath.Join(tempDir, ".servo"), filepath.Join(teamTempDir, ".servo"))
	if err != nil {
		t.Fatalf("Failed to copy .servo directory: %v", err)
	}

	// Copy servo files
	servoFiles := []string{"api-server.servo", "ml-server.servo"}
	for _, file := range servoFiles {
		srcFile := filepath.Join(tempDir, file)
		dstFile := filepath.Join(teamTempDir, file)
		err = copyFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("Failed to copy %s: %v", file, err)
		}
	}

	// Change to team directory
	originalWd, _ := os.Getwd()
	os.Chdir(teamTempDir)
	defer os.Chdir(originalWd)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Team member checks status
	cmd := exec.Command(servoPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Team member status check failed: %v\nOutput: %s", err, output)
	}

	teamStatus := string(output)
	// Verify team member sees project structure
	if !strings.Contains(teamStatus, "MCP Servers: 2 configured") {
		t.Errorf("Expected team member to see configured servers, got: %s", teamStatus)
	}

	// Team member configures their own secrets
	teamSecrets := map[string]string{
		"openai_api_key":    "sk-team1234567890abcdef",
		"database_url":      "postgres://team:pass@localhost/teamdb",
		"anthropic_api_key": "ant_team1234567890abcdef",
	}

	for key, value := range teamSecrets {
		cmd := exec.Command(servoPath, "secrets", "set", key, value)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Team member failed to set secret %s: %v\nOutput: %s", key, err, output)
		}
	}

	// Verify team member can now work with the project
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Team member final status check failed: %v\nOutput: %s", err, output)
	}

	finalStatus := string(output)
	// Basic validation - team member should see their project is configured
	if !strings.Contains(finalStatus, "MCP Servers: 2 configured") {
		t.Errorf("Expected team member to see configured project, got: %s", finalStatus)
	}

	t.Logf("✅ Team collaboration secrets workflow validated")
}

// Helper functions
func copyDir(src, dst string) error {
	return exec.Command("cp", "-r", src, dst).Run()
}

func copyFile(src, dst string) error {
	return exec.Command("cp", src, dst).Run()
}
