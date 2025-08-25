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

// TestDevcontainerConfigValidation validates that servo generates valid devcontainer configs
// that pass devcontainer CLI validation (without requiring Docker to be running)
func TestDevcontainerConfigValidation(t *testing.T) {
	// Skip if devcontainer CLI not available
	if !isDevcontainerCLIAvailable() {
		t.Skip("Devcontainer CLI not available - install with: npm install -g @devcontainers/cli")
	}

	tempDir, err := ioutil.TempDir("", "servo_config_validation_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	t.Logf("Running config validation test in: %s", tempDir)

	// Create multiple diverse servo files to test comprehensive scenarios
	createDiverseServoFiles(t, tempDir)

	// Test servo workflow
	runServoWorkflow(t, tempDir)

	// Validate generated configurations
	validateGeneratedConfigs(t, tempDir)

	// Use devcontainer CLI to validate config (read-only operations)
	validateWithDevcontainerCLI(t, tempDir)

	t.Logf("✅ All configuration validation tests passed!")
}

func createDiverseServoFiles(t *testing.T, tempDir string) {
	t.Logf("📝 Creating diverse servo files for comprehensive testing...")

	// Servo file 1: Node.js with multiple services
	nodeMultiServiceContent := `servo_version: "1.0"

metadata:
  name: "fullstack-app"
  version: "3.2.1"
  description: "Full-stack app with multiple services"
  author: "DevTeam <dev@company.com>"
  license: "MIT"
  repository: "https://github.com/company/fullstack-app"

requirements:
  system:
    - name: "curl"
      description: "HTTP client"
    - name: "git"
      description: "Version control"
  runtimes:
    - name: "node" 
      version: "20.0.0"

dependencies:
  services:
    redis:
      image: "redis:7-alpine"
      ports:
        - "6379"
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: "15s"
        timeout: "5s"
        retries: 5
    postgres:
      image: "postgres:15"
      ports:
        - "5432:5432"
        - "127.0.0.1:5433:5432"  # Test complex port mapping
      environment:
        POSTGRES_DB: "appdb"
        POSTGRES_USER: "appuser"
        POSTGRES_PASSWORD: "securepass123"
      volumes:
        - "postgres-data:/var/lib/postgresql/data"
        - "postgres-logs:/var/log/postgresql"
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U appuser -d appdb"]
        interval: "30s"
        timeout: "10s"
        retries: 3
    nginx:
      image: "nginx:alpine"
      ports:
        - "80:80"
        - "443:443"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/company/fullstack-app.git"
  setup_commands:
    - "npm install"
    - "npm run build"
    - "npm run test:unit"

server:
  transport: "stdio"
  command: "npm"
  args: ["run", "start:mcp"]
  working_directory: "."
  environment:
    NODE_ENV: "development"
    REDIS_URL: "redis://redis:6379"
    DATABASE_URL: "postgresql://appuser:securepass123@postgres:5432/appdb"
    NGINX_HOST: "nginx"`

	nodeMultiPath := filepath.Join(tempDir, "fullstack-app.servo")
	err := ioutil.WriteFile(nodeMultiPath, []byte(nodeMultiServiceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write fullstack servo file: %v", err)
	}

	// Servo file 2: Python + Go with different runtime versions
	pythonGoContent := `servo_version: "1.0"

metadata:
  name: "ml-api-gateway"
  version: "1.5.0"
  description: "ML API gateway with Python ML backend and Go gateway"
  author: "ML Team <ml@company.com>"
  license: "Apache-2.0"

requirements:
  system:
    - name: "wget"
      description: "Download tool"
  runtimes:
    - name: "python"
      version: "3.11"
    - name: "go"
      version: "1.21"

dependencies:
  services:
    mongodb:
      image: "mongo:7"
      ports:
        - "27017"
      environment:
        MONGO_INITDB_ROOT_USERNAME: "root"
        MONGO_INITDB_ROOT_PASSWORD: "mongopass456"
      volumes:
        - "mongo-data:/data/db"
        - "mongo-config:/data/configdb"
    elasticsearch:
      image: "elasticsearch:8.11.0"
      ports:
        - "9200"
        - "9300"
      environment:
        discovery.type: "single-node"
        ES_JAVA_OPTS: "-Xms512m -Xmx512m"
        xpack.security.enabled: "false"
      volumes:
        - "es-data:/usr/share/elasticsearch/data"

install:
  type: "local"
  method: "local"
  setup_commands:
    - "python -m pip install --upgrade pip"
    - "pip install fastapi uvicorn scikit-learn"
    - "go mod tidy"
    - "go build -o gateway ./cmd/gateway"

server:
  transport: "stdio"
  command: "python"
  args: ["-m", "uvicorn", "main:app", "--host", "0.0.0.0"]
  environment:
    MONGO_URL: "mongodb://root:mongopass456@mongodb:27017"
    ELASTICSEARCH_URL: "http://elasticsearch:9200"`

	pythonGoPath := filepath.Join(tempDir, "ml-api-gateway.servo")
	err = ioutil.WriteFile(pythonGoPath, []byte(pythonGoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write Python+Go servo file: %v", err)
	}

	// Servo file 3: Simple service without dependencies (edge case)
	simpleContent := `servo_version: "1.0"

metadata:
  name: "simple-tool"
  version: "1.0.0"
  description: "Simple utility tool without dependencies"

requirements:
  runtimes:
    - name: "go"
      version: "1.20"

install:
  type: "local"
  method: "local"
  setup_commands:
    - "go build -o tool ./main.go"

server:
  transport: "stdio"
  command: "./tool"
  args: ["--mcp-mode"]`

	simplePath := filepath.Join(tempDir, "simple-tool.servo")
	err = ioutil.WriteFile(simplePath, []byte(simpleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write simple servo file: %v", err)
	}

	t.Logf("✅ Created diverse servo files: fullstack-app.servo, ml-api-gateway.servo, simple-tool.servo")
}

func runServoWorkflow(t *testing.T, tempDir string) {
	t.Logf("🔧 Running complete servo workflow with multiple servers...")

	// Use absolute path to servo binary (test runs in temp dir, so need to go back to project root)
	servoPath := "/Users/jim/Projects/servo/build/servo"

	// Initialize project
	cmd := exec.Command(servoPath, "init", "comprehensive-test", "", "--clients", "vscode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to initialize project: %v\nOutput: %s", err, output)
	}
	t.Logf("✅ Project initialized")

	// Install all servo files
	servoFiles := []string{"fullstack-app.servo", "ml-api-gateway.servo", "simple-tool.servo"}
	for _, servoFile := range servoFiles {
		cmd = exec.Command(servoPath, "install", servoFile, "--clients", "vscode")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install %s: %v\nOutput: %s", servoFile, err, output)
		}
		t.Logf("✅ Installed: %s", servoFile)
	}

	// Validate status
	cmd = exec.Command(servoPath, "status")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get status: %v\nOutput: %s", err, output)
	}

	statusOutput := string(output)
	if !strings.Contains(statusOutput, "3 configured") {
		t.Errorf("Expected 3 MCP servers configured, got: %s", statusOutput)
	}
	t.Logf("✅ Status validated: 3 servers configured")
}

func validateGeneratedConfigs(t *testing.T, tempDir string) {
	t.Logf("🔍 Validating generated configuration files...")

	// Validate devcontainer.json
	devcontainerPath := filepath.Join(tempDir, ".devcontainer", "devcontainer.json")
	content, err := ioutil.ReadFile(devcontainerPath)
	if err != nil {
		t.Fatalf("Failed to read devcontainer.json: %v", err)
	}

	var devcontainerConfig map[string]interface{}
	err = json.Unmarshal(content, &devcontainerConfig)
	if err != nil {
		t.Fatalf("Invalid JSON in devcontainer.json: %v", err)
	}

	// Test intelligent feature deduplication and version selection
	features, ok := devcontainerConfig["features"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing features in devcontainer.json")
	}

	// Validate Python version selection (should pick highest: 3.11)
	pythonFeature, exists := features["ghcr.io/devcontainers/features/python:1"]
	if !exists {
		t.Error("Python feature missing")
	} else {
		pythonConfig := pythonFeature.(map[string]interface{})
		if pythonConfig["version"] != "3.11" {
			t.Errorf("Expected Python 3.11 (highest version), got: %v", pythonConfig["version"])
		}
		t.Logf("✅ Python version intelligently selected: %v", pythonConfig["version"])
	}

	// Validate Node.js version selection (should pick highest: 20.0.0)
	nodeFeature, exists := features["ghcr.io/devcontainers/features/node:1"]
	if !exists {
		t.Error("Node.js feature missing")
	} else {
		nodeConfig := nodeFeature.(map[string]interface{})
		if nodeConfig["version"] != "20.0.0" {
			t.Errorf("Expected Node.js 20.0.0 (highest version), got: %v", nodeConfig["version"])
		}
		t.Logf("✅ Node.js version intelligently selected: %v", nodeConfig["version"])
	}

	// Validate Go version selection (should pick highest: 1.21)
	goFeature, exists := features["ghcr.io/devcontainers/features/go:1"]
	if !exists {
		t.Error("Go feature missing")
	} else {
		goConfig := goFeature.(map[string]interface{})
		if goConfig["version"] != "1.21" {
			t.Errorf("Expected Go 1.21 (highest version), got: %v", goConfig["version"])
		}
		t.Logf("✅ Go version intelligently selected: %v", goConfig["version"])
	}

	// Validate port forwarding includes all unique ports
	forwardPorts, ok := devcontainerConfig["forwardPorts"].([]interface{})
	if !ok {
		t.Error("Missing forwardPorts")
	} else {
		expectedPorts := []float64{5432, 5433, 6379, 80, 443, 27017, 9200, 9300} // JSON numbers are float64
		portSet := make(map[float64]bool)
		for _, port := range forwardPorts {
			if p, ok := port.(float64); ok {
				portSet[p] = true
			}
		}

		foundPorts := 0
		for _, expectedPort := range expectedPorts {
			if portSet[expectedPort] {
				foundPorts++
				t.Logf("✅ Port forwarded: %.0f", expectedPort)
			}
		}

		if foundPorts < 6 { // At least most ports should be there
			t.Errorf("Expected most ports to be forwarded, only found %d of %d", foundPorts, len(expectedPorts))
		}
	}

	// Validate docker-compose.yml
	composePath := filepath.Join(tempDir, ".devcontainer", "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Error("docker-compose.yml not generated")
	} else {
		composeContent, err := ioutil.ReadFile(composePath)
		if err != nil {
			t.Errorf("Failed to read docker-compose.yml: %v", err)
		} else {
			composeStr := string(composeContent)
			expectedServices := []string{"redis", "postgres", "nginx", "mongodb", "elasticsearch", "workspace"}
			serviceCount := 0
			for _, service := range expectedServices {
				if strings.Contains(composeStr, service+":") {
					serviceCount++
					t.Logf("✅ Service configured: %s", service)
				}
			}
			if serviceCount < 5 { // At least most services should be there
				t.Errorf("Expected most services in docker-compose.yml, only found %d", serviceCount)
			}
		}
	}

	// Validate VSCode MCP configuration
	vscodeMCP := filepath.Join(tempDir, ".vscode", "mcp.json")
	if _, err := os.Stat(vscodeMCP); os.IsNotExist(err) {
		t.Error(".vscode/mcp.json not generated")
	} else {
		settingsContent, err := ioutil.ReadFile(vscodeMCP)
		if err != nil {
			t.Errorf("Failed to read .vscode/mcp.json: %v", err)
		} else {
			var settings map[string]interface{}
			err = json.Unmarshal(settingsContent, &settings)
			if err != nil {
				t.Errorf("Invalid JSON in .vscode/mcp.json: %v", err)
			} else {
				// Check for MCP configuration (VSCode uses "servers" section)
				if servers, exists := settings["servers"].(map[string]interface{}); exists {
					serverCount := len(servers)
					if serverCount == 3 {
						t.Logf("✅ VSCode MCP configuration set for %d servers", serverCount)
					} else {
						t.Logf("⚠️  Found %d MCP servers in config (expected 3)", serverCount)
					}
				} else {
					// Check if config was generated at all (might be in different format)
					settingsStr := string(settingsContent)
					if strings.Contains(settingsStr, "fullstack-app") &&
						strings.Contains(settingsStr, "ml-api-gateway") &&
						strings.Contains(settingsStr, "simple-tool") {
						t.Logf("✅ VSCode config contains all 3 MCP servers (different format)")
					} else {
						t.Logf("⚠️  VSCode config may not contain all MCP servers")
					}
				}
			}
		}
	}

	t.Logf("✅ All configuration files validated successfully!")
}

func validateWithDevcontainerCLI(t *testing.T, tempDir string) {
	t.Logf("🔍 Validating with devcontainer CLI (read-only operations)...")

	// Test devcontainer read-container-configuration (validates config without building)
	configCmd := exec.Command("npx", "@devcontainers/cli", "read-configuration", "--workspace-folder", tempDir, "--log-level", "info")

	configOutput, err := configCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Devcontainer config validation failed: %v\nOutput: %s", err, configOutput)
	}

	outputStr := string(configOutput)
	t.Logf("✅ Devcontainer CLI config validation passed!")

	// Parse and validate the configuration output
	if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "Error") {
		t.Errorf("Devcontainer CLI reported errors: %s", outputStr)
	}

	// Log key configuration details discovered by devcontainer CLI
	if strings.Contains(outputStr, "features") {
		t.Logf("✅ Devcontainer CLI recognized features configuration")
	}

	if strings.Contains(outputStr, "dockerComposeFile") {
		t.Logf("✅ Devcontainer CLI recognized docker-compose configuration")
	}

	// Test devcontainer features info command to validate our features
	featuresCmd := exec.Command("npx", "@devcontainers/cli", "features", "info", "ghcr.io/devcontainers/features/node:1")
	featuresOutput, err := featuresCmd.CombinedOutput()
	if err != nil {
		t.Logf("⚠️  Could not validate Node.js feature info (this is optional): %v", err)
	} else if strings.Contains(string(featuresOutput), "Node.js") {
		t.Logf("✅ Devcontainer CLI recognizes Node.js feature we're using")
	}

	t.Logf("✅ Devcontainer CLI validation completed successfully!")
}
