package test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// SecretsData represents the structure of base64-encoded secrets file
type SecretsData struct {
	Version string            `yaml:"version"`
	Secrets map[string]string `yaml:"secrets"`
}

// SetupTestSecrets creates base64-encoded secrets file with the given secrets
func SetupTestSecrets(t *testing.T, projectDir string, secrets map[string]string) {
	t.Helper()

	// Ensure .servo directory exists
	servoDir := filepath.Join(projectDir, ".servo")
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		t.Fatalf("Failed to create .servo directory: %v", err)
	}

	secretsPath := filepath.Join(servoDir, "secrets.yaml")

	// Encode secrets with base64 for basic obscurity
	encodedSecrets := make(map[string]string)
	for key, value := range secrets {
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		encodedSecrets[key] = encodedValue
	}

	// Create final data structure
	fileData := SecretsData{
		Version: "1.0",
		Secrets: encodedSecrets,
	}

	output, err := yaml.Marshal(fileData)
	if err != nil {
		t.Fatalf("Failed to marshal file data: %v", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(secretsPath, output, 0600); err != nil {
		t.Fatalf("Failed to write secrets file: %v", err)
	}

	t.Logf("✅ Created test secrets file with %d secrets at %s", len(secrets), secretsPath)
}

// CleanupTestSecrets removes any test environment variables (placeholder for consistency)
func CleanupTestSecrets(t *testing.T) {
	t.Helper()
	// No environment cleanup needed with simple base64 approach
}

// RequiredSecretsForTests returns the standard set of secrets needed for integration tests
func RequiredSecretsForTests() map[string]string {
	return map[string]string{
		"database_url":   "postgresql://testuser:testpass@localhost:5432/testdb",
		"openai_api_key": "sk-test-api-key-12345",
		"api_key":        "test-api-key-67890",
		"jwt_secret":     "test-jwt-secret-abcdef",
	}
}
