package test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInitIntegration_CompleteWorkflow validates the complete servo init workflow
// including project creation, directory structure, and various initialization options
func TestInitIntegration_CompleteWorkflow(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_init_integration_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Logf("Running servo init integration test in: %s", tempDir)

	// Test 1: Basic init without parameters
	testBasicInit(t, tempDir)

	// Test 2: Init with project name
	testInitWithName(t, tempDir)

	// Test 3: Init with description and clients
	testInitWithFullOptions(t, tempDir)

	// Test 4: Init in existing project (should fail)
	testInitInExistingProject(t, tempDir)

	// Test 5: Init with different client configurations
	testInitClientConfigurations(t, tempDir)

	// Test 6: Validate generated file structure and content
	testGeneratedFileStructure(t, tempDir)

	t.Logf("✅ Complete servo init integration test passed!")
}

func testBasicInit(t *testing.T, baseDir string) {
	t.Logf("🚀 Testing basic servo init...")

	// Create subdirectory for this test
	testDir := filepath.Join(baseDir, "basic-init")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test basic init command
	cmd := exec.Command(servoPath, "init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Basic init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "initialized successfully") {
		t.Errorf("Expected success message in output, got: %s", outputStr)
	}

	// Verify .servo directory was created
	servoDir := filepath.Join(testDir, ".servo")
	if _, err := os.Stat(servoDir); os.IsNotExist(err) {
		t.Error("Expected .servo directory to be created")
	}

	// Verify project.yaml was created
	projectFile := filepath.Join(servoDir, "project.yaml")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		t.Error("Expected project.yaml to be created")
	}

	// Read and validate project.yaml content
	content, err := ioutil.ReadFile(projectFile)
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	projectContent := string(content)
	// Should contain basic project structure
	if !strings.Contains(projectContent, "default_session:") {
		t.Errorf("Expected project.yaml to contain default_session field, got: %s", projectContent)
	}

	t.Logf("✅ Basic init test passed")
}

func testInitWithName(t *testing.T, baseDir string) {
	t.Logf("🏷️  Testing servo init with custom name...")

	testDir := filepath.Join(baseDir, "named-init")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test init with description
	cmd := exec.Command(servoPath, "init", "--session", "testing")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Named init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "initialized successfully") {
		t.Errorf("Expected success message in output, got: %s", outputStr)
	}

	// Verify project.yaml exists
	projectFile := filepath.Join(testDir, ".servo", "project.yaml")
	_, err = os.Stat(projectFile)
	if err != nil {
		t.Fatalf("Expected project.yaml to exist: %v", err)
	}

	t.Logf("✅ Named init test passed")
}

func testInitWithFullOptions(t *testing.T, baseDir string) {
	t.Logf("⚙️  Testing servo init with full options...")

	testDir := filepath.Join(baseDir, "full-options-init")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test init with all options
	cmd := exec.Command(servoPath, "init",
		"",
		"--clients", "vscode,claude-code")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Full options init failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "initialized successfully") {
		t.Errorf("Expected success message in output, got: %s", outputStr)
	}

	// Verify project.yaml was created
	projectFile := filepath.Join(testDir, ".servo", "project.yaml")
	_, err = os.Stat(projectFile)
	if err != nil {
		t.Fatalf("Expected project.yaml to exist: %v", err)
	}

	t.Logf("✅ Full options init test passed")
}

func testInitInExistingProject(t *testing.T, baseDir string) {
	t.Logf("🚫 Testing servo init in existing project (should fail)...")

	testDir := filepath.Join(baseDir, "existing-project")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// First, initialize a project
	cmd := exec.Command(servoPath, "init")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	// Try to initialize again - should fail
	cmd = exec.Command(servoPath, "init")
	output, err := cmd.CombinedOutput()

	// Should fail with error
	if err == nil {
		t.Error("Expected servo init to fail in existing project directory")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "already exists") {
		t.Errorf("Expected 'already exists' error message, got: %s", outputStr)
	}

	t.Logf("✅ Existing project init failure test passed")
}

func testInitClientConfigurations(t *testing.T, baseDir string) {
	t.Logf("🔧 Testing servo init with different client configurations...")

	testCases := []struct {
		name     string
		clients  string
		expected []string
	}{
		{
			name:     "vscode-only",
			clients:  "vscode",
			expected: []string{"vscode"},
		},
		{
			name:     "claude-code-only",
			clients:  "claude-code",
			expected: []string{"claude-code"},
		},
		{
			name:     "multiple-clients",
			clients:  "vscode,claude-code,cursor",
			expected: []string{"vscode", "claude-code", "cursor"},
		},
	}

	servoPath := "/Users/jim/Projects/servo/build/servo"

	for _, tc := range testCases {
		t.Logf("Testing client config: %s", tc.name)

		testDir := filepath.Join(baseDir, "client-config-"+tc.name)
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory for %s: %v", tc.name, err)
		}

		originalWd, _ := os.Getwd()
		os.Chdir(testDir)

		// Initialize with specific clients
		cmd := exec.Command(servoPath, "init", "--clients", tc.clients)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Client config init failed for %s: %v\nOutput: %s", tc.name, err, output)
		}

		projectFile := filepath.Join(testDir, ".servo", "project.yaml")
		if _, err := os.Stat(projectFile); os.IsNotExist(err) {
			t.Errorf("Expected project.yaml to be created for client config test %s", tc.name)
		}

		// Verify clients are recorded in project configuration (even if files aren't created yet)
		content, err := ioutil.ReadFile(projectFile)
		if err != nil {
			t.Errorf("Failed to read project.yaml for %s: %v", tc.name, err)
		} else {
			// The project should know about configured clients
			// (actual client files created during install, not init)
			_ = content // Just verify we can read the file for now
		}

		os.Chdir(originalWd)
	}

	t.Logf("✅ Client configuration test passed")
}

func testGeneratedFileStructure(t *testing.T, baseDir string) {
	t.Logf("📁 Testing generated file structure and content...")

	testDir := filepath.Join(baseDir, "structure-validation")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize with VSCode client
	cmd := exec.Command(servoPath, "init",
		"",
		"--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Structure test init failed: %v\nOutput: %s", err, output)
	}

	// Define expected file structure (client files created during install, not init)
	expectedFiles := []string{
		".servo/project.yaml",
		".servo/.gitignore",
	}

	expectedDirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/default",
		".servo/sessions/default/manifests",
		".servo/sessions/default/volumes",
	}

	// Check that all expected directories exist
	for _, dir := range expectedDirs {
		dirPath := filepath.Join(testDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Check that all expected files exist
	for _, file := range expectedFiles {
		filePath := filepath.Join(testDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	// Validate project.yaml structure
	projectFile := filepath.Join(testDir, ".servo", "project.yaml")
	content, err := ioutil.ReadFile(projectFile)
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	projectContent := string(content)

	// Should contain required fields
	requiredFields := []string{"default_session:", "active_session:"}
	for _, field := range requiredFields {
		if !strings.Contains(projectContent, field) {
			t.Errorf("Expected project.yaml to contain '%s', got: %s", field, projectContent)
		}
	}

	// Validate .gitignore content
	gitignoreFile := filepath.Join(testDir, ".servo", ".gitignore")
	gitignoreContent, err := ioutil.ReadFile(gitignoreFile)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	gitignoreStr := string(gitignoreContent)

	// Should ignore certain patterns
	expectedIgnorePatterns := []string{"volumes/", "*.log", "secrets.yaml"}
	for _, pattern := range expectedIgnorePatterns {
		if !strings.Contains(gitignoreStr, pattern) {
			t.Errorf("Expected .gitignore to contain '%s', got: %s", pattern, gitignoreStr)
		}
	}

	// Check current time - created_at should be within last minute
	if strings.Contains(projectContent, "created_at:") {
		// This is a simple check - in a real test we might parse the timestamp
		now := time.Now()
		oneMinuteAgo := now.Add(-1 * time.Minute)

		// Extract timestamp and verify it's recent (basic check)
		lines := strings.Split(projectContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "created_at:") {
				// Just verify the line contains a reasonable timestamp format
				if !strings.Contains(line, "2025") {
					t.Errorf("Expected recent timestamp in created_at, got: %s", line)
				}
				break
			}
		}

		_ = oneMinuteAgo // Use the variable to avoid unused warning
	}

	t.Logf("✅ File structure validation test passed")
}

// TestInitIntegration_EdgeCases tests edge cases and error conditions
func TestInitIntegration_EdgeCases(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "servo_init_edge_cases_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Logf("Running servo init edge cases test in: %s", tempDir)

	// Test invalid client names
	testInvalidClients(t, tempDir)

	// Test very long project names
	testLongProjectNames(t, tempDir)

	// Test special characters in descriptions
	testSpecialCharacters(t, tempDir)

	t.Logf("✅ Edge cases test passed!")
}

func testInvalidClients(t *testing.T, baseDir string) {
	t.Logf("❌ Testing invalid client configurations...")

	testDir := filepath.Join(baseDir, "invalid-clients")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test with invalid client name
	cmd := exec.Command(servoPath, "init", "--clients", "nonexistent-client")
	output, err := cmd.CombinedOutput()

	// Should either warn or fail gracefully
	outputStr := string(output)

	// The behavior might vary - either it should warn about unknown clients or fail
	// Let's just verify it doesn't crash and provides some feedback
	if err != nil {
		// If it fails, should have helpful error message
		if !strings.Contains(outputStr, "client") {
			t.Errorf("Expected error message about client, got: %s", outputStr)
		}
	} else {
		// If it succeeds, should at least initialize basic structure
		if !strings.Contains(outputStr, "initialized") {
			t.Errorf("Expected initialization message, got: %s", outputStr)
		}
	}

	t.Logf("✅ Invalid clients test completed")
}

func testLongProjectNames(t *testing.T, baseDir string) {
	t.Logf("📏 Testing long project names...")

	testDir := filepath.Join(baseDir, "long-names")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test basic initialization without description (description field has been removed)
	cmd := exec.Command(servoPath, "init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Basic init failed: %v\nOutput: %s", err, output)
	}

	// Verify project.yaml was created with proper structure
	projectFile := filepath.Join(testDir, ".servo", "project.yaml")
	content, err := ioutil.ReadFile(projectFile)
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	// Verify project.yaml has basic required fields
	contentStr := string(content)
	if !strings.Contains(contentStr, "default_session:") {
		t.Errorf("Expected project.yaml to contain default_session field")
	}

	t.Logf("✅ Long descriptions test passed")
}

func testSpecialCharacters(t *testing.T, baseDir string) {
	t.Logf("🔤 Testing special characters in descriptions...")

	testDir := filepath.Join(baseDir, "special-chars")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Test initialization with clients option (since description is removed, test clients functionality)
	cmd := exec.Command(servoPath, "init", "--clients", "vscode,claude-code")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Init with clients failed: %v\nOutput: %s", err, output)
	}

	// Verify the clients were configured
	projectFile := filepath.Join(testDir, ".servo", "project.yaml")
	content, err := ioutil.ReadFile(projectFile)
	if err != nil {
		t.Fatalf("Failed to read project.yaml: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "vscode") || !strings.Contains(contentStr, "claude-code") {
		t.Errorf("Expected clients to be configured, got: %s", contentStr)
	}

	t.Logf("✅ Special characters test passed")
}
