package config

import (
	"encoding/base64"
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/pkg"
)

func TestDockerComposeGeneration_WithRealSecrets(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Set up test environment with base64 secrets

	if err := setupTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createSecretsManifest(); err != nil {
		t.Fatalf("Failed to create secrets manifest: %v", err)
	}

	// Create base64 encoded secrets file
	secrets := map[string]string{
		"database_url": "postgres://user:pass@localhost/db",
		"api_key":      "secret-api-key",
	}
	if err := createBase64Secrets(secrets); err != nil {
		t.Fatalf("Failed to create base64 secrets: %v", err)
	}

	// Generate docker-compose
	manager := NewConfigGeneratorManager(".servo")
	if err := manager.GenerateDockerCompose(); err != nil {
		t.Fatalf("Failed to generate docker-compose: %v", err)
	}

	// Verify secrets were injected
	verifyDockerComposeWithSecrets(t)
}

func TestDockerComposeGeneration_MissingSecretsFixed(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "servo-missing-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := setupTestProject(); err != nil {
		t.Fatalf("Failed to setup test project: %v", err)
	}

	if err := createSecretsManifest(); err != nil {
		t.Fatalf("Failed to create secrets manifest: %v", err)
	}

	// Generate should fail
	manager := NewConfigGeneratorManager(".servo")
	err = manager.GenerateDockerCompose()
	if err == nil {
		t.Fatal("Expected generation to fail with missing secrets")
	}

	expectedError := "secrets validation failed:"
	if len(err.Error()) < len(expectedError) || err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error about missing secrets, got: %v", err)
	}
}

func createBase64Secrets(secrets map[string]string) error {
	// Encode secrets with base64 for basic obscurity
	encodedSecrets := make(map[string]string)
	for key, value := range secrets {
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		encodedSecrets[key] = encodedValue
	}

	// Create final data structure
	fileData := map[string]interface{}{
		"version": "1.0",
		"secrets": encodedSecrets,
	}

	output, err := yaml.Marshal(fileData)
	if err != nil {
		return err
	}

	// Create .servo directory if it doesn't exist
	if err := os.MkdirAll(".servo", 0755); err != nil {
		return err
	}

	// Write with restricted permissions
	return os.WriteFile(".servo/secrets.yaml", output, 0600)
}

func verifyDockerComposeWithSecrets(t *testing.T) {
	data, err := os.ReadFile(".devcontainer/docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse docker-compose.yml: %v", err)
	}

	// Verify secrets section exists
	secrets, ok := config["secrets"].(map[string]interface{})
	if !ok {
		t.Fatal("Secrets section not found")
	}

	// Should have database_url and api_key secrets
	expectedSecrets := []string{"database_url", "api_key"}
	for _, secretName := range expectedSecrets {
		secret, exists := secrets[secretName]
		if !exists {
			t.Errorf("Secret %s not found", secretName)
			continue
		}

		secretConfig := secret.(map[string]interface{})
		if secretConfig["external"] != true {
			t.Errorf("Secret %s should be external", secretName)
		}
	}

	// Verify services have secrets attached
	services := config["services"].(map[string]interface{})

	appService, serviceExists := services["secure-app-app"] // Service name is prefixed with manifest name
	if !serviceExists {
		t.Fatal("secure-app-app service not found")
	}

	appServiceMap := appService.(map[string]interface{})

	serviceSecrets, exists := appServiceMap["secrets"]
	if !exists {
		t.Fatal("App service should have secrets")
	}

	secretsList := serviceSecrets.([]interface{})
	if len(secretsList) != 2 {
		t.Errorf("Expected 2 secrets for app service, got %d", len(secretsList))
	}
}

// Helper functions

func setupTestProject() error {
	// Create .servo directory structure
	dirs := []string{
		".servo",
		".servo/sessions",
		".servo/sessions/test",
		".servo/sessions/test/manifests",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create project.yaml
	project := &project.Project{
		Clients:        []string{"vscode"},
		DefaultSession: "test",
		ActiveSession:  "test",
	}

	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".servo/project.yaml", data, 0644); err != nil {
		return err
	}

	// Create active session file
	if err := os.WriteFile(".servo/active_session", []byte("test"), 0644); err != nil {
		return err
	}

	// Create project session file (simpler approach)
	// The session manager looks for .servo/sessions/{name}/session.yaml for project sessions
	sessionInfo := map[string]interface{}{
		"name":        "test",
		"description": "Test session",
		"active":      true,
		"created_at":  "2024-01-01T00:00:00Z",
	}

	sessionData, err := yaml.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/session.yaml", sessionData, 0644)
}

func createSecretsManifest() error {
	manifest := &pkg.ServoDefinition{
		ServoVersion: "1.0",
		Name:         "secure-app",
		ConfigurationSchema: &pkg.ConfigurationSchema{
			Secrets: map[string]pkg.SecretSchema{
				"database_url": {
					Description: "Database connection URL",
					Required:    true,
				},
				"api_key": {
					Description: "API key for external service",
					Required:    true,
				},
			},
		},
		Services: map[string]*pkg.ServiceDependency{
			"app": {
				Image: "myapp:latest",
				Environment: map[string]string{
					"DATABASE_URL": "${database_url}",
					"API_KEY":      "/run/secrets/api_key",
				},
				Ports: []string{"8080:8080"},
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(".servo/sessions/test/manifests/secure-app.servo", data, 0644)
}
