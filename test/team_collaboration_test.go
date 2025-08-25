package test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTeamCollaborationWorkflow tests that new team members can see required secrets
// based on installed MCP servers and configure their own secrets independently
func TestTeamCollaborationWorkflow(t *testing.T) {
	// Create temporary directory for original team member
	originalDir, err := ioutil.TempDir("", "servo_team_original_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(originalDir)

	// Create temporary directory for new team member
	teamMemberDir, err := ioutil.TempDir("", "servo_team_member_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(teamMemberDir)

	t.Logf("Testing team collaboration in: %s -> %s", originalDir, teamMemberDir)

	// Change to original team member directory
	if err := os.Chdir(originalDir); err != nil {
		t.Fatalf("Failed to change to original directory: %v", err)
	}

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Original team member: Create project
	t.Logf("📝 Original team member initializes project...")
	cmd := exec.Command(servoPath, "init", "team-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized")

	// Create servo file with multiple required secrets
	t.Logf("📝 Creating MCP server with required secrets...")
	mcpServerContent := `metadata:
  name: "team-auth-server"
  description: "Authentication server for team project"

server:
  command: "team-auth-server"
  args: ["start"]

configuration_schema:
  secrets:
    api_key:
      description: "External API key for authentication"
      required: true
      type: "string"
      env_var: "API_KEY"
    database_url:
      description: "Database connection string"
      required: true
      type: "string"
      env_var: "DATABASE_URL"
    jwt_secret:
      description: "JWT signing secret"
      required: true
      type: "string"
      env_var: "JWT_SECRET"`

	err = ioutil.WriteFile("team-auth-server.servo", []byte(mcpServerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write servo file: %v", err)
	}

	// Original team member: Set up secrets BEFORE installing (proper workflow)
	t.Logf("🔐 Original team member sets secrets...")
	originalSecrets := map[string]string{
		"api_key":      "original-api-key-12345",
		"database_url": "postgres://original:pass@localhost/db",
		"jwt_secret":   "original-jwt-secret-67890",
	}

	// Set up test secrets using base64 encoding
	SetupTestSecrets(t, originalDir, originalSecrets)

	// Original team member: Install MCP server (should now succeed with secrets configured)
	t.Logf("📦 Installing MCP server...")
	cmd = exec.Command(servoPath, "install", "team-auth-server.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install MCP server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ MCP server installed")

	// Verify original team member sees all secrets configured
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get status: %v\nOutput: %s", err, output)
	}

	statusOutput := string(output)
	if !strings.Contains(statusOutput, "🎉 All required secrets are configured!") {
		t.Errorf("Expected original team member to see all secrets configured, got: %s", statusOutput)
	}

	// Simulate git commit and push (copy project files excluding secrets)
	t.Logf("📁 Simulating git clone for new team member...")

	// Copy all project files except secrets.enc
	copyProjectFiles := func(src, dst string) error {
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			// Skip secrets.yaml file (should be gitignored)
			if strings.Contains(relPath, "secrets.yaml") {
				return nil
			}

			dstPath := filepath.Join(dst, relPath)

			if info.IsDir() {
				return os.MkdirAll(dstPath, info.Mode())
			}

			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = dstFile.ReadFrom(srcFile)
			return err
		})
	}

	if err := copyProjectFiles(originalDir, teamMemberDir); err != nil {
		t.Fatalf("Failed to copy project files: %v", err)
	}

	// Change to team member directory
	if err := os.Chdir(teamMemberDir); err != nil {
		t.Fatalf("Failed to change to team member directory: %v", err)
	}

	// New team member: Check status (should see required secrets)
	t.Logf("👥 New team member checks project status...")
	cmd = exec.Command(servoPath, "status")
	// They don't have secrets file yet
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get team member status: %v\nOutput: %s", err, output)
	}

	teamMemberStatus := string(output)
	t.Logf("📋 Team member status: %s", teamMemberStatus)

	// Verify team member sees the required secrets from installed servers
	expectedSecrets := []string{"api_key", "database_url", "jwt_secret"}
	for _, secret := range expectedSecrets {
		if !strings.Contains(teamMemberStatus, secret) {
			t.Errorf("Expected team member to see required secret '%s' in status, got: %s", secret, teamMemberStatus)
		}
	}

	if !strings.Contains(teamMemberStatus, "❌ Missing secrets: 3") {
		t.Errorf("Expected team member to see 3 missing secrets, got: %s", teamMemberStatus)
	}

	// New team member: Configure their own secrets
	t.Logf("🔐 New team member sets their own secrets...")
	teamMemberSecrets := map[string]string{
		"api_key":      "teammember-api-key-99999",
		"database_url": "postgres://teammember:newpass@localhost/teamdb",
		"jwt_secret":   "teammember-jwt-secret-11111",
	}

	for key, value := range teamMemberSecrets {
		cmd := exec.Command(servoPath, "secrets", "set", key, value)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Team member failed to set secret %s: %v\nOutput: %s", key, err, output)
		}
	}
	t.Logf("✅ Team member configured their secrets")

	// Verify team member now sees all secrets configured
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get team member status after config: %v\nOutput: %s", err, output)
	}

	finalStatus := string(output)
	if !strings.Contains(finalStatus, "🎉 All required secrets are configured!") {
		t.Errorf("Expected team member to see all secrets configured after setup, got: %s", finalStatus)
	}

	if !strings.Contains(finalStatus, "✅ Configured secrets: 3") {
		t.Errorf("Expected team member to see 3 configured secrets, got: %s", finalStatus)
	}

	// Verify team member can access their own secret values
	for key, expectedValue := range teamMemberSecrets {
		cmd := exec.Command(servoPath, "secrets", "get", key)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Team member failed to get secret %s: %v\nOutput: %s", key, err, output)
		}

		actualValue := strings.TrimSpace(string(output))
		if actualValue != expectedValue {
			t.Errorf("Expected secret %s to be '%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	// Verify secrets are independent - each team member has their own local secrets file
	t.Logf("🔒 Verifying secrets are stored locally per team member...")
	// Team member has their own secrets.yaml file with base64 encoded values
	secretsPath := filepath.Join(".servo", "secrets.yaml")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		t.Errorf("Expected team member to have local secrets.yaml file")
	}

	t.Logf("✅ Team collaboration workflow validated successfully!")
}
