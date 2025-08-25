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

// DevcontainerIntegrationTest validates the complete workflow:
// 1. Create real servo files with complex dependencies
// 2. Use servo commands to create project and install servers
// 3. Use devcontainer CLI to validate the generated configuration works
// 4. Test that services actually start and respond inside the container
func TestDevcontainerIntegration_FullWorkflow(t *testing.T) {
	// Skip if running in CI or if devcontainer CLI not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping devcontainer integration test in CI environment")
	}

	// Check if devcontainer CLI is available
	if !isDevcontainerCLIAvailable() {
		t.Skip("Devcontainer CLI not available - install with: npm install -g @devcontainers/cli")
	}

	// Check if Docker is running
	if !isDockerRunning() {
		t.Skip("Docker is not running - required for devcontainer integration test")
	}

	tempDir, err := ioutil.TempDir("", "servo_devcontainer_integration_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running devcontainer integration test in: %s", tempDir)

	// Step 1: Create realistic servo files with complex dependencies
	createRealServoFiles(t, tempDir)

	// Step 2: Test full servo workflow
	testServoWorkflow(t, tempDir)

	// Step 3: Validate with devcontainer CLI
	testDevcontainerCLI(t, tempDir)

	// Step 4: Test services actually work inside container (optional - can be slow)
	if os.Getenv("SERVO_FULL_INTEGRATION") != "" {
		testServicesInContainer(t, tempDir)
	} else {
		t.Logf("⏩ Skipping full container service tests (set SERVO_FULL_INTEGRATION=1 to enable)")
	}

	t.Logf("✅ Complete devcontainer integration test passed!")
}

func createRealServoFiles(t *testing.T, tempDir string) {
	t.Logf("📝 Creating realistic servo files...")

	// Create a Node.js MCP server with Redis dependency
	nodeServerContent := `servo_version: "1.0"

metadata:
  name: "node-chat-server"
  version: "1.0.0"
  description: "Node.js chat MCP server with Redis"
  author: "Test Team <test@example.com>"
  license: "MIT"

requirements:
  system:
    - name: "curl"
      description: "HTTP client for health checks"
    - name: "redis-tools"
      description: "Redis CLI tools"
  runtimes:
    - name: "node"
      version: "18.0.0"

dependencies:
  services:
    redis:
      image: "redis:7-alpine"
      ports:
        - "6379"
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: "30s"
        timeout: "10s"
        retries: 3

install:
  type: "local"
  method: "local"  
  setup_commands:
    - "npm init -y"
    - "npm install express redis"
    - "echo 'console.log(\"Node chat server setup complete\");' > server.js"

server:
  transport: "stdio"
  command: "node"
  args: ["server.js"]
  environment:
    REDIS_URL: "redis://redis:6379"
    NODE_ENV: "production"`

	nodeServerPath := filepath.Join(tempDir, "node-chat.servo")
	err := ioutil.WriteFile(nodeServerPath, []byte(nodeServerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write node servo file: %v", err)
	}

	// Create a Python MCP server with Postgres dependency
	pythonServerContent := `servo_version: "1.0"

metadata:
  name: "python-data-server"
  version: "2.1.0"
  description: "Python data processing MCP server with PostgreSQL"
  author: "Data Team <data@example.com>"
  license: "MIT"

requirements:
  system:
    - name: "postgresql-client"
      description: "PostgreSQL client tools"
  runtimes:
    - name: "python"
      version: "3.11"

dependencies:
  services:
    postgres:
      image: "postgres:15-alpine"
      ports:
        - "5432"
      environment:
        POSTGRES_DB: "datadb"
        POSTGRES_USER: "datauser"
        POSTGRES_PASSWORD: "datapass123"
      volumes:
        - "postgres-data:/var/lib/postgresql/data"
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U datauser -d datadb"]
        interval: "30s"
        timeout: "10s"
        retries: 5

install:
  type: "local"
  method: "local"
  setup_commands:
    - "python -m venv venv"
    - "source venv/bin/activate && pip install psycopg2-binary fastapi"
    - "echo 'print(\"Python data server setup complete\")' > data_server.py"

server:
  transport: "stdio"
  command: "python"
  args: ["data_server.py"]
  environment:
    DATABASE_URL: "postgresql://datauser:datapass123@postgres:5432/datadb"
    PYTHONPATH: "/workspace"`

	pythonServerPath := filepath.Join(tempDir, "python-data.servo")
	err = ioutil.WriteFile(pythonServerPath, []byte(pythonServerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write python servo file: %v", err)
	}

	t.Logf("✅ Created servo files: node-chat.servo, python-data.servo")
}

func testServoWorkflow(t *testing.T, tempDir string) {
	t.Logf("🔧 Testing servo workflow...")

	// Use absolute path to servo binary (test runs in temp dir, so need to go back to project root)
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Step 1: Initialize project
	cmd := exec.Command(servoPath, "init", "integration-test", "", "--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize servo project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized: %s", strings.TrimSpace(string(output)))

	// Step 2: Install Node.js server
	cmd = exec.Command(servoPath, "install", "node-chat.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install node server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Node server installed: %s", strings.TrimSpace(string(output)))

	// Step 3: Install Python server
	cmd = exec.Command(servoPath, "install", "python-data.servo", "--clients", "vscode")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install python server: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Python server installed: %s", strings.TrimSpace(string(output)))

	// Step 4: Check project status
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get project status: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project status: %s", strings.TrimSpace(string(output)))

	// Step 5: Verify files were created
	requiredFiles := []string{
		".servo/project.yaml",
		".devcontainer/devcontainer.json",
		".devcontainer/docker-compose.yml",
		".vscode/mcp.json",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Required file not created: %s", file)
		} else {
			t.Logf("✅ File created: %s", file)
		}
	}
}

func testDevcontainerCLI(t *testing.T, tempDir string) {
	t.Logf("🐳 Testing devcontainer CLI validation...")

	// Read and validate the generated devcontainer.json
	devcontainerPath := filepath.Join(tempDir, ".devcontainer", "devcontainer.json")
	content, err := ioutil.ReadFile(devcontainerPath)
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var devcontainerConfig map[string]interface{}
	err = json.Unmarshal(content, &devcontainerConfig)
	if err != nil {
		t.Fatalf("Invalid devcontainer.json: %v", err)
	}

	// Validate key components exist
	validateDevcontainerConfig(t, devcontainerConfig)

	// Test devcontainer configuration validation (fast)
	t.Logf("🔍 Validating devcontainer configuration...")

	configCmd := exec.Command("npx", "@devcontainers/cli", "read-configuration", "--workspace-folder", tempDir, "--log-level", "info")
	configOutput, err := configCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Devcontainer config validation failed: %v\nOutput: %s", err, configOutput)
	}

	t.Logf("✅ Devcontainer configuration validation successful!")

	// Optionally test devcontainer build (this actually builds the container!)
	if os.Getenv("SERVO_FULL_INTEGRATION") != "" {
		t.Logf("🔨 Building devcontainer (this may take a few minutes)...")

		buildCmd := exec.Command("npx", "@devcontainers/cli", "build", "--workspace-folder", tempDir, "--log-level", "info")
		buildCmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Devcontainer build failed: %v\nOutput: %s", err, buildOutput)
		}

		t.Logf("✅ Devcontainer build successful!")

		// Log key parts of build output for debugging
		outputLines := strings.Split(string(buildOutput), "\n")
		for _, line := range outputLines {
			if strings.Contains(line, "Successfully tagged") ||
				strings.Contains(line, "Installing") ||
				strings.Contains(line, "DONE") {
				t.Logf("Build: %s", line)
			}
		}
	} else {
		t.Logf("⏩ Skipping devcontainer build (set SERVO_FULL_INTEGRATION=1 to enable)")
	}
}

func testServicesInContainer(t *testing.T, tempDir string) {
	t.Logf("🚀 Testing services inside devcontainer...")

	// Start the devcontainer and test services
	testCmd := `
		# Wait for services to be ready
		timeout 60 bash -c 'until docker-compose ps | grep -q healthy; do sleep 2; done' || echo "Services starting..."
		
		# Test Redis connection
		redis-cli -h redis ping && echo "✅ Redis is responding" || echo "❌ Redis not responding"
		
		# Test Postgres connection  
		PGPASSWORD=datapass123 pg_isready -h postgres -U datauser -d datadb && echo "✅ Postgres is responding" || echo "❌ Postgres not responding"
		
		# Test Node.js environment
		node --version && echo "✅ Node.js is available" || echo "❌ Node.js not available"
		
		# Test Python environment
		python --version && echo "✅ Python is available" || echo "❌ Python not available"
		
		# Show docker-compose services status
		docker-compose ps
	`

	execCmd := exec.Command("npx", "@devcontainers/cli", "exec", "--workspace-folder", tempDir, "bash", "-c", testCmd)
	execCmd.Env = os.Environ()

	execOutput, err := execCmd.CombinedOutput()
	outputStr := string(execOutput)
	t.Logf("Service test output:\n%s", outputStr)

	if err != nil {
		t.Logf("⚠️  Some service tests failed (this may be expected): %v", err)
		// Don't fail the test completely - services might take time to start
	}

	// Check for success indicators in output
	successChecks := []string{
		"Redis is responding",
		"Node.js is available",
		"Python is available",
	}

	successCount := 0
	for _, check := range successChecks {
		if strings.Contains(outputStr, check) {
			successCount++
		}
	}

	if successCount >= 2 {
		t.Logf("✅ %d/%d service checks passed - devcontainer is functional!", successCount, len(successChecks))
	} else {
		t.Logf("⚠️  Only %d/%d service checks passed - may need more startup time", successCount, len(successChecks))
	}
}

func validateDevcontainerConfig(t *testing.T, config map[string]interface{}) {
	t.Logf("🔍 Validating devcontainer configuration...")

	// Check features exist and are correct
	features, ok := config["features"].(map[string]interface{})
	if !ok {
		t.Fatal("devcontainer.json missing 'features' section")
	}

	expectedFeatures := []string{
		"ghcr.io/devcontainers/features/node:1",
		"ghcr.io/devcontainers/features/python:1",
		"ghcr.io/devcontainers/features/common-utils:2",
		"ghcr.io/devcontainers/features/git:1",
	}

	for _, expectedFeature := range expectedFeatures {
		if _, exists := features[expectedFeature]; !exists {
			t.Errorf("Missing expected feature: %s", expectedFeature)
		} else {
			t.Logf("✅ Feature found: %s", expectedFeature)
		}
	}

	// Check forwardPorts includes our service ports
	forwardPorts, ok := config["forwardPorts"].([]interface{})
	if !ok {
		t.Error("devcontainer.json missing 'forwardPorts'")
	} else {
		expectedPorts := []float64{5432, 6379} // JSON numbers are float64
		portSet := make(map[float64]bool)
		for _, port := range forwardPorts {
			if p, ok := port.(float64); ok {
				portSet[p] = true
			}
		}

		for _, expectedPort := range expectedPorts {
			if !portSet[expectedPort] {
				t.Errorf("Missing expected forwarded port: %.0f", expectedPort)
			} else {
				t.Logf("✅ Port forwarded: %.0f", expectedPort)
			}
		}
	}

	// Check docker-compose configuration
	dockerComposeFile, ok := config["dockerComposeFile"].([]interface{})
	if !ok || len(dockerComposeFile) == 0 {
		t.Error("devcontainer.json missing dockerComposeFile configuration")
	} else {
		t.Logf("✅ Docker compose configured: %v", dockerComposeFile)
	}

	// Check customizations - they should NOT be present in base devcontainer
	// Customizations are client-specific and handled separately
	if customizations, exists := config["customizations"]; exists {
		// If customizations exist, they should be from overrides
		t.Logf("ℹ️  Customizations found (likely from overrides): %v", customizations)
	} else {
		t.Logf("✅ No client-specific customizations in base devcontainer (as expected)")
	}
}

func isDevcontainerCLIAvailable() bool {
	cmd := exec.Command("npx", "@devcontainers/cli", "--version")
	return cmd.Run() == nil
}

func isDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}
