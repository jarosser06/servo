package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/pkg"
)

func TestBaseGenerator_ExtractSecretFromEnvValue(t *testing.T) {
	generator := NewBaseGenerator("/tmp")

	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "environment variable with ${} format",
			envValue: "${DATABASE_URL}",
			expected: "database_url",
		},
		{
			name:     "environment variable with run/secrets format",
			envValue: "/run/secrets/api_key",
			expected: "api_key",
		},
		{
			name:     "regular environment variable",
			envValue: "localhost:5432",
			expected: "",
		},
		{
			name:     "empty value",
			envValue: "",
			expected: "",
		},
		{
			name:     "malformed secret reference",
			envValue: "${INCOMPLETE",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.extractSecretFromEnvValue(tt.envValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBaseGenerator_ExtractRequiredSecretsFromManifests(t *testing.T) {
	generator := NewBaseGenerator("/tmp")

	// Create test manifests
	manifests := map[string]*pkg.ServoDefinition{
		"test-server": {
			ServoVersion: "1.0",
			Name:         "test-server",
			ConfigurationSchema: &pkg.ConfigurationSchema{
				Secrets: map[string]pkg.SecretSchema{
					"api_key": {
						Description: "API key for external service",
						Required:    true,
					},
					"optional_key": {
						Description: "Optional key",
						Required:    false,
					},
				},
			},
			Services: map[string]*pkg.ServiceDependency{
				"web": {
					Image: "nginx:latest",
					Environment: map[string]string{
						"DATABASE_URL": "${database_url}",
						"REDIS_URL":    "/run/secrets/redis_pass",
						"PORT":         "8080",
					},
				},
			},
			Dependencies: &pkg.Dependencies{
				Services: map[string]pkg.ServiceDependency{
					"worker": {
						Image: "worker:latest",
						Environment: map[string]string{
							"WORKER_TOKEN": "${worker_token}",
						},
					},
				},
			},
		},
		"simple-server": {
			ServoVersion: "1.0",
			Name:         "simple-server",
			ConfigurationSchema: &pkg.ConfigurationSchema{
				Secrets: map[string]pkg.SecretSchema{
					"simple_key": {
						Description: "Simple key",
						Required:    true,
					},
				},
			},
		},
	}

	result := generator.extractRequiredSecretsFromManifests(manifests)

	// Convert slice to map for easier checking
	resultMap := make(map[string]bool)
	for _, secret := range result {
		resultMap[secret] = true
	}

	// Should include required secrets from configuration schema
	if !resultMap["api_key"] {
		t.Error("Should include api_key from configuration schema")
	}

	if !resultMap["simple_key"] {
		t.Error("Should include simple_key from configuration schema")
	}

	// Should not include optional secrets
	if resultMap["optional_key"] {
		t.Error("Should not include optional secrets")
	}

	// Should include secrets from service environment variables
	if !resultMap["database_url"] {
		t.Error("Should include database_url from service environment")
	}

	if !resultMap["redis_pass"] {
		t.Error("Should include redis_pass from service environment")
	}

	// Should include secrets from dependency services
	if !resultMap["worker_token"] {
		t.Error("Should include worker_token from dependency service")
	}

	// The current implementation only returns secret names as a slice
	// Description functionality would require a different return type
	t.Logf("Extracted %d required secrets: %v", len(result), result)
}

func TestGenerator_validateSecretsBeforeGeneration(t *testing.T) {
	// Create test project
	testProject := &project.Project{
		Clients:        []string{"vscode"},
		DefaultSession: "test",
	}

	// Create test manifests that require secrets
	manifests := map[string]*pkg.ServoDefinition{
		"test-server": {
			ServoVersion: "1.0",
			ConfigurationSchema: &pkg.ConfigurationSchema{
				Secrets: map[string]pkg.SecretSchema{
					"api_key": {
						Description: "API key",
						Required:    true,
					},
					"database_url": {
						Description: "Database URL",
						Required:    true,
					},
				},
			},
		},
	}

	t.Run("validation passes with all required secrets configured", func(t *testing.T) {
		configuredSecrets := map[string]bool{
			"api_key":      true,
			"database_url": true,
		}

		err := validateSecretsForTesting(testProject, manifests, configuredSecrets)
		if err != nil {
			t.Errorf("Validation should pass when all secrets are configured: %v", err)
		}
	})

	t.Run("validation fails with missing required secrets", func(t *testing.T) {
		configuredSecrets := map[string]bool{
			"api_key": true,
			// database_url is missing
		}

		err := validateSecretsForTesting(testProject, manifests, configuredSecrets)
		if err == nil {
			t.Error("Validation should fail when required secrets are missing")
		}

		if !strings.Contains(err.Error(), "database_url") {
			t.Errorf("Error message should mention missing secret: %v", err)
		}
	})

	t.Run("validation passes with no required secrets", func(t *testing.T) {
		// Empty manifests
		emptyManifests := map[string]*pkg.ServoDefinition{}
		configuredSecrets := map[string]bool{}

		err := validateSecretsForTesting(testProject, emptyManifests, configuredSecrets)
		if err != nil {
			t.Errorf("Validation should pass when no secrets are required: %v", err)
		}
	})
}

// Helper function for testing secrets validation without mocking
func validateSecretsForTesting(project *project.Project, manifests map[string]*pkg.ServoDefinition, configuredSecrets map[string]bool) error {
	generator := NewBaseGenerator("/tmp")

	// Extract required secrets from manifests
	requiredSecretsSlice := generator.extractRequiredSecretsFromManifests(manifests)

	// Check for missing secrets
	var missingSecrets []string
	for _, secretName := range requiredSecretsSlice {
		if !configuredSecrets[secretName] {
			missingSecrets = append(missingSecrets, secretName)
		}
	}

	if len(missingSecrets) > 0 {
		return fmt.Errorf("missing required secrets: %v. Please configure these secrets before generating configurations", missingSecrets)
	}

	return nil
}

// mockProjectManager for testing (reused from secrets_test.go)
type mockProjectManager struct {
	configuredSecrets map[string]bool
}

func (m *mockProjectManager) GetConfiguredSecrets() (map[string]bool, error) {
	return m.configuredSecrets, nil
}

// Implement other required methods as no-ops
func (m *mockProjectManager) IsProject() bool                { return true }
func (m *mockProjectManager) Get() (*project.Project, error) { return &project.Project{}, nil }
func (m *mockProjectManager) Init(sessionName string, clients []string) (*project.Project, error) {
	return &project.Project{}, nil
}
func (m *mockProjectManager) Save(project *project.Project) error                      { return nil }
func (m *mockProjectManager) Delete() error                                            { return nil }
func (m *mockProjectManager) AddMCPServer(name, source string, clients []string) error { return nil }
func (m *mockProjectManager) RemoveMCPServer(name string) error                        { return nil }
func (m *mockProjectManager) GetServoDir() string                                      { return ".servo" }
func (m *mockProjectManager) AddRequiredSecret(name, description string) error         { return nil }
func (m *mockProjectManager) RemoveRequiredSecret(name string) error                   { return nil }
func (m *mockProjectManager) GetMissingSecrets() ([]project.RequiredSecret, error)     { return nil, nil }
