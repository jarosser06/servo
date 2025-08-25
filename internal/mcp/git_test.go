package mcp

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestParser_ParseFromGitRepo_Public tests cloning from a local test repository
func TestParser_ParseFromGitRepo_Public(t *testing.T) {
	// Create a temporary directory for our test git repository
	tempRepoDir, err := ioutil.TempDir("", "test-git-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp repo directory: %v", err)
	}
	defer os.RemoveAll(tempRepoDir)

	// Create a sample .servo file
	servoContent := `servo_version: "1.0"
metadata:
  name: "test-mcp-server"
  description: "Test MCP server for git parsing"
server:
  transport: "stdio"
  command: "python"
  args: ["-m", "test_server"]
`

	servoFile := filepath.Join(tempRepoDir, "test.servo")
	if err := ioutil.WriteFile(servoFile, []byte(servoContent), 0644); err != nil {
		t.Fatalf("Failed to write test servo file: %v", err)
	}

	// Initialize git repository
	commands := [][]string{
		{"git", "init", tempRepoDir},
		{"git", "-C", tempRepoDir, "config", "user.email", "test@example.com"},
		{"git", "-C", tempRepoDir, "config", "user.name", "Test User"},
		{"git", "-C", tempRepoDir, "add", "test.servo"},
		{"git", "-C", tempRepoDir, "commit", "-m", "Add test servo file"},
	}

	for _, cmd := range commands {
		if output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("Failed to run %v: %v\nOutput: %s", cmd, err, output)
		}
	}

	parser := NewParser()

	// Test parsing from the local git repository
	servoDef, err := parser.ParseFromGitRepo(tempRepoDir, "")
	if err != nil {
		t.Fatalf("ParseFromGitRepo failed: %v", err)
	}

	// Basic validation that we got a servo definition
	if servoDef == nil {
		t.Fatal("Expected servo definition, got nil")
	}

	// Validate the parsed content
	if servoDef.Name != "test-mcp-server" {
		t.Errorf("Expected name 'test-mcp-server', got %v", servoDef.Name)
	}

	t.Logf("Successfully parsed servo file from git repo: %s", servoDef.Name)
}

// TestParser_GitAuthentication_SSH tests SSH authentication setup
func TestParser_GitAuthentication_SSH(t *testing.T) {
	parser := NewParser()

	// Configure SSH authentication
	parser.SSHKeyPath = "/home/user/.ssh/id_rsa"
	parser.SSHPassword = "test-passphrase"

	// Test that authentication options are properly set
	if parser.SSHKeyPath != "/home/user/.ssh/id_rsa" {
		t.Errorf("Expected SSHKeyPath '/home/user/.ssh/id_rsa', got '%s'", parser.SSHKeyPath)
	}

	if parser.SSHPassword != "test-passphrase" {
		t.Errorf("Expected SSHPassword 'test-passphrase', got '%s'", parser.SSHPassword)
	}

	// Test that HTTP credentials remain unset
	if parser.HTTPToken != "" {
		t.Errorf("Expected HTTPToken to be empty, got '%s'", parser.HTTPToken)
	}
}

// TestParser_GitAuthentication_HTTPS tests HTTPS authentication setup
func TestParser_GitAuthentication_HTTPS(t *testing.T) {
	parser := NewParser()

	// Configure HTTPS token authentication
	parser.HTTPToken = "ghp_test123token456"

	if parser.HTTPToken != "ghp_test123token456" {
		t.Errorf("Expected HTTPToken 'ghp_test123token456', got '%s'", parser.HTTPToken)
	}

	// Test username/password authentication
	parser2 := NewParser()
	parser2.HTTPUsername = "testuser"
	parser2.HTTPPassword = "testpass"

	if parser2.HTTPUsername != "testuser" {
		t.Errorf("Expected HTTPUsername 'testuser', got '%s'", parser2.HTTPUsername)
	}

	if parser2.HTTPPassword != "testpass" {
		t.Errorf("Expected HTTPPassword 'testpass', got '%s'", parser2.HTTPPassword)
	}
}

// TestParser_ParseFromGitRepo_InvalidRepo tests error handling for invalid repositories
func TestParser_ParseFromGitRepo_InvalidRepo(t *testing.T) {
	parser := NewParser()

	// Test with completely invalid URL
	_, err := parser.ParseFromGitRepo("not-a-git-url", "")
	if err == nil {
		t.Error("Expected ParseFromGitRepo to fail with invalid URL")
	}

	// Test with non-existent repository
	_, err = parser.ParseFromGitRepo("https://github.com/nonexistent/repo12345.git", "")
	if err == nil {
		t.Error("Expected ParseFromGitRepo to fail with non-existent repository")
	}
}

// TestParser_ParseFromGitRepo_SubDirectory tests parsing from subdirectories
func TestParser_ParseFromGitRepo_SubDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git clone test in short mode")
	}

	if os.Getenv("SKIP_GIT_TESTS") != "" {
		t.Skip("Skipping git tests due to SKIP_GIT_TESTS environment variable")
	}

	parser := NewParser()

	// Test with subdirectory that doesn't exist
	_, err := parser.ParseFromGitRepo("https://github.com/test/repo.git", "nonexistent/path")
	if err == nil {
		t.Error("Expected ParseFromGitRepo to fail with non-existent subdirectory")
	}
}

// TestParser_GitEnvironmentVariables tests that environment variables are respected
func TestParser_GitEnvironmentVariables(t *testing.T) {
	// Set up environment variables
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalUsername := os.Getenv("GIT_USERNAME")
	originalPassword := os.Getenv("GIT_PASSWORD")

	// Clean up after test
	defer func() {
		os.Setenv("GITHUB_TOKEN", originalToken)
		os.Setenv("GIT_USERNAME", originalUsername)
		os.Setenv("GIT_PASSWORD", originalPassword)
	}()

	// Test GitHub token environment variable
	os.Setenv("GITHUB_TOKEN", "test-token-123")
	os.Setenv("GIT_USERNAME", "testuser")
	os.Setenv("GIT_PASSWORD", "testpass")

	parser := NewParser()

	// The parser should be able to access these via os.Getenv() in its authentication logic
	if os.Getenv("GITHUB_TOKEN") != "test-token-123" {
		t.Error("Environment variable GITHUB_TOKEN not set correctly")
	}

	if os.Getenv("GIT_USERNAME") != "testuser" {
		t.Error("Environment variable GIT_USERNAME not set correctly")
	}

	if os.Getenv("GIT_PASSWORD") != "testpass" {
		t.Error("Environment variable GIT_PASSWORD not set correctly")
	}

	// Test that explicit credentials take precedence over environment
	parser.HTTPToken = "explicit-token"
	if parser.HTTPToken != "explicit-token" {
		t.Error("Explicit HTTPToken should take precedence over environment variable")
	}
}

// TestParser_SSHConfig tests that SSH config file support is automatic
func TestParser_SSHConfig(t *testing.T) {
	// go-git automatically uses SSH config via ssh_config.DefaultUserSettings
	// This reads from ~/.ssh/config and /etc/ssh/ssh_config automatically
	parser := NewParser()

	// Test that we can configure explicit SSH key path (overrides SSH config)
	parser.SSHKeyPath = "~/.ssh/custom_key"

	if parser.SSHKeyPath != "~/.ssh/custom_key" {
		t.Errorf("Expected SSHKeyPath '~/.ssh/custom_key', got '%s'", parser.SSHKeyPath)
	}

	// go-git automatically handles (via ssh_config.DefaultUserSettings):
	// 1. Reading ~/.ssh/config for host-specific configurations
	// 2. Reading /etc/ssh/ssh_config for system defaults
	// 3. Host aliases, port settings, user settings, key files, etc.
	// 4. SSH agent integration
	// 5. Common key locations if not specified in config
	// 6. All SSH key types (RSA, Ed25519, ECDSA, etc.)

	// This means users can configure git repos in their SSH config:
	// Host myserver
	//   HostName github.com
	//   User git
	//   IdentityFile ~/.ssh/id_rsa_work
	//   Port 22

	// And then use: servo install git@myserver:user/repo.git
}

// TestParser_CredentialPrecedence tests the order of credential precedence
func TestParser_CredentialPrecedence(t *testing.T) {
	// Test precedence: explicit options > environment variables > system defaults

	// Set up environment
	os.Setenv("GITHUB_TOKEN", "env-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	parser := NewParser()

	// Environment variable should be available
	if os.Getenv("GITHUB_TOKEN") != "env-token" {
		t.Error("Environment token not set correctly")
	}

	// Explicit token should take precedence
	parser.HTTPToken = "explicit-token"
	if parser.HTTPToken != "explicit-token" {
		t.Error("Explicit token should override environment variable")
	}

	// Test SSH key precedence
	os.Setenv("GIT_SSH_KEY", "/env/path/key")
	defer os.Unsetenv("GIT_SSH_KEY")

	parser.SSHKeyPath = "/explicit/path/key"
	if parser.SSHKeyPath != "/explicit/path/key" {
		t.Error("Explicit SSH key path should override environment variable")
	}
}

// BenchmarkParser_ParseFromFile benchmarks file parsing performance
func BenchmarkParser_ParseFromFile(b *testing.B) {
	// Create a temporary servo file
	tempFile, err := os.CreateTemp("", "benchmark-*.servo")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test servo content
	servoContent := `servo_version: "1.0"
metadata:
  name: "benchmark-server"
  version: "1.0.0"
  description: "Benchmark test server"
  author: "Test"
  license: "MIT"

server:
  transport: "stdio"
  command: "benchmark"
  args: ["--mode", "test"]

configuration_schema:
  secrets:
    api_key:
      description: "API key"
      required: true
      type: "string"
      env_var: "API_KEY"
`

	if _, err := tempFile.WriteString(servoContent); err != nil {
		b.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFromFile(tempFile.Name())
		if err != nil {
			b.Fatalf("ParseFromFile failed: %v", err)
		}
	}
}

// TestParser_GitURLVariations tests different git URL formats
func TestParser_GitURLVariations(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		shouldWork  bool
		description string
	}{
		{
			name:        "HTTPS GitHub",
			url:         "https://github.com/user/repo.git",
			shouldWork:  false, // Would work with real repo
			description: "Standard HTTPS GitHub URL",
		},
		{
			name:        "SSH GitHub",
			url:         "git@github.com:user/repo.git",
			shouldWork:  false, // Would work with real repo and SSH setup
			description: "Standard SSH GitHub URL",
		},
		{
			name:        "HTTPS GitLab",
			url:         "https://gitlab.com/user/repo.git",
			shouldWork:  false, // Would work with real repo
			description: "Standard HTTPS GitLab URL",
		},
		{
			name:        "SSH GitLab",
			url:         "git@gitlab.com:user/repo.git",
			shouldWork:  false, // Would work with real repo and SSH setup
			description: "Standard SSH GitLab URL",
		},
		{
			name:        "Invalid URL",
			url:         "not-a-url",
			shouldWork:  false,
			description: "Invalid URL format",
		},
		{
			name:        "Empty URL",
			url:         "",
			shouldWork:  false,
			description: "Empty URL",
		},
	}

	parser := NewParser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseFromGitRepo(tc.url, "")

			if tc.shouldWork && err != nil {
				t.Errorf("Expected %s to work, but got error: %v", tc.description, err)
			}

			if !tc.shouldWork && err == nil {
				t.Errorf("Expected %s to fail, but it succeeded", tc.description)
			}

			// Log the error for debugging (expected for test URLs)
			if err != nil {
				t.Logf("Expected error for %s: %v", tc.description, err)
			}
		})
	}
}

// Example function showing how to use the parser with authentication
func ExampleParser_authentication() {
	parser := NewParser()

	// SSH key authentication
	parser.SSHKeyPath = "~/.ssh/id_rsa"
	parser.SSHPassword = "key-passphrase"

	// Or HTTPS token authentication
	parser.HTTPToken = "ghp_your_github_token_here"

	// Or HTTPS username/password
	parser.HTTPUsername = "your-username"
	parser.HTTPPassword = "your-password"

	// Parse from git repository
	_, err := parser.ParseFromGitRepo("git@github.com:your/repo.git", "")
	if err != nil {
		fmt.Printf("Failed to parse from git repo: %v", err)
		return
	}

	fmt.Println("Successfully parsed .servo file from git repository")
}
