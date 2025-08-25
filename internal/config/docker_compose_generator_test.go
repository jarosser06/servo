package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

func TestDockerComposeGenerator_EnvironmentVariableIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	servoDir := filepath.Join(tempDir, ".servo")
	os.MkdirAll(servoDir, 0755)

	// Create env.yaml with test environment variables
	envData := EnvData{
		Version: "1.0",
		Env: map[string]string{
			"API_BASE_URL": "https://test-api.example.com",
			"DEBUG_MODE":   "true",
			"MAX_RETRIES":  "5",
		},
	}

	envBytes, err := yaml.Marshal(envData)
	if err != nil {
		t.Fatalf("Failed to marshal env data: %v", err)
	}

	envPath := filepath.Join(servoDir, "env.yaml")
	if err := os.WriteFile(envPath, envBytes, 0644); err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Create generator
	generator := NewDockerComposeGenerator(servoDir)

	// Create a test manifest with services that have environment variables
	manifests := map[string]*pkg.ServoDefinition{
		"test-server": {
			Name:        "test-server",
			Description: "Test server for environment variable testing",
			Version:     "1.0.0",
			Services: map[string]*pkg.ServiceDependency{
				"web": {
					Image: "nginx:latest",
					Environment: map[string]string{
						"SERVICE_ENV": "from-service",
						"PORT":        "8080",
					},
					Ports: []string{"8080:80"},
				},
			},
		},
	}

	// Test the addServicesFromManifests method with environment variables
	config := map[string]interface{}{
		"services": make(map[string]interface{}),
	}

	// Change to temp directory to make the generator work properly
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	err = generator.addServicesFromManifests(config, manifests)
	if err != nil {
		t.Fatalf("Failed to add services from manifests: %v", err)
	}

	// Verify the service was added
	services := config["services"].(map[string]interface{})
	if len(services) == 0 {
		t.Fatal("Expected services to be added, but found none")
	}

	// Check that the test-server-web service exists
	serviceName := "test-server-web"
	service, exists := services[serviceName]
	if !exists {
		t.Fatalf("Expected service %s to exist, but it doesn't", serviceName)
	}

	serviceConfig := service.(map[string]interface{})

	// Verify environment variables are merged correctly
	envVars, hasEnv := serviceConfig["environment"]
	if !hasEnv {
		t.Fatal("Expected environment variables in service config")
	}

	envSlice := envVars.([]string)
	envMap := make(map[string]string)
	for _, env := range envSlice {
		parts := splitEnvVar(env)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Check project-level environment variables are included
	expectedProjectEnvs := map[string]string{
		"API_BASE_URL": "https://test-api.example.com",
		"DEBUG_MODE":   "true",
		"MAX_RETRIES":  "5",
	}

	for key, expected := range expectedProjectEnvs {
		if actual, exists := envMap[key]; !exists {
			t.Errorf("Expected project environment variable %s to be present", key)
		} else if actual != expected {
			t.Errorf("Expected project environment variable %s=%s, got %s", key, expected, actual)
		}
	}

	// Check service-specific environment variables are included and can override project-level ones
	expectedServiceEnvs := map[string]string{
		"SERVICE_ENV": "from-service",
		"PORT":        "8080",
	}

	for key, expected := range expectedServiceEnvs {
		if actual, exists := envMap[key]; !exists {
			t.Errorf("Expected service environment variable %s to be present", key)
		} else if actual != expected {
			t.Errorf("Expected service environment variable %s=%s, got %s", key, expected, actual)
		}
	}

	// Verify other service properties are preserved
	if image := serviceConfig["image"]; image != "nginx:latest" {
		t.Errorf("Expected image nginx:latest, got %v", image)
	}

	if ports := serviceConfig["ports"]; len(ports.([]string)) != 1 || ports.([]string)[0] != "8080:80" {
		t.Errorf("Expected ports [8080:80], got %v", ports)
	}
}

func TestDockerComposeGenerator_NoEnvironmentFile(t *testing.T) {
	// Create temporary directory for test (without env.yaml)
	tempDir := t.TempDir()
	servoDir := filepath.Join(tempDir, ".servo")
	os.MkdirAll(servoDir, 0755)

	// Create generator
	generator := NewDockerComposeGenerator(servoDir)

	// Create a test manifest with services that have environment variables
	manifests := map[string]*pkg.ServoDefinition{
		"test-server": {
			Name:        "test-server",
			Description: "Test server",
			Version:     "1.0.0",
			Services: map[string]*pkg.ServiceDependency{
				"web": {
					Image: "nginx:latest",
					Environment: map[string]string{
						"SERVICE_ENV": "from-service",
					},
				},
			},
		},
	}

	config := map[string]interface{}{
		"services": make(map[string]interface{}),
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	err := generator.addServicesFromManifests(config, manifests)
	if err != nil {
		t.Fatalf("Failed to add services from manifests: %v", err)
	}

	// Verify the service was added with only service-specific env vars
	services := config["services"].(map[string]interface{})
	serviceName := "test-server-web"
	service := services[serviceName].(map[string]interface{})
	
	envVars := service["environment"].([]string)
	if len(envVars) != 1 || envVars[0] != "SERVICE_ENV=from-service" {
		t.Errorf("Expected only service environment variables, got %v", envVars)
	}
}

// Helper function to split environment variable strings
func splitEnvVar(env string) []string {
	for i, char := range env {
		if char == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
