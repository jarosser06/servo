package config

import (
	"testing"

	"github.com/servo/servo/internal/project"
)

func TestBaseGenerator_ServiceNeedsSecrets(t *testing.T) {
	generator := NewBaseGenerator("/tmp")

	availableSecrets := map[string]bool{
		"database_url": true,
		"api_key":      true,
		"redis_pass":   true,
	}

	tests := []struct {
		name            string
		serviceConfig   map[string]interface{}
		expectedSecrets []string
	}{
		{
			name: "service with environment variable secret references",
			serviceConfig: map[string]interface{}{
				"image": "app:latest",
				"environment": []string{
					"DATABASE_URL=${database_url}",
					"API_KEY=/run/secrets/api_key",
					"OTHER_VAR=value",
				},
			},
			expectedSecrets: []string{"database_url", "api_key"},
		},
		{
			name: "service with explicit secret label",
			serviceConfig: map[string]interface{}{
				"image": "worker:latest",
				"labels": map[string]string{
					"servo.secrets": "database_url,redis_pass",
				},
			},
			expectedSecrets: []string{"database_url", "redis_pass"},
		},
		{
			name: "service with no secret references",
			serviceConfig: map[string]interface{}{
				"image": "nginx:latest",
				"environment": []string{
					"PORT=80",
					"HOST=localhost",
				},
			},
			expectedSecrets: []string{},
		},
		{
			name: "service with case-insensitive secret references",
			serviceConfig: map[string]interface{}{
				"image": "app:latest",
				"environment": []string{
					"DATABASE_URL=${DATABASE_URL}",
				},
			},
			expectedSecrets: []string{"database_url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.serviceNeedsSecrets(tt.serviceConfig, availableSecrets)

			if len(result) != len(tt.expectedSecrets) {
				t.Errorf("Expected %d secrets, got %d: %v", len(tt.expectedSecrets), len(result), result)
				return
			}

			// Convert to map for easier comparison
			resultMap := make(map[string]bool)
			for _, secret := range result {
				resultMap[secret] = true
			}

			for _, expected := range tt.expectedSecrets {
				if !resultMap[expected] {
					t.Errorf("Expected secret %s not found in result", expected)
				}
			}
		})
	}
}

func TestInjectSecrets_BasicInjection(t *testing.T) {
	generator := NewBaseGenerator("/tmp")

	// Mock configured secrets
	configuredSecrets := map[string]bool{
		"database_url": true,
		"api_key":      true,
	}

	// Test base docker-compose configuration
	baseConfig := map[string]interface{}{
		"version": "3.8",
		"services": map[string]interface{}{
			"web": map[string]interface{}{
				"image": "nginx:latest",
				"environment": []string{
					"DATABASE_URL=${database_url}",
					"API_KEY=/run/secrets/api_key",
				},
			},
			"worker": map[string]interface{}{
				"image": "worker:latest",
				"labels": map[string]string{
					"servo.secrets": "database_url",
				},
			},
			"cache": map[string]interface{}{
				"image": "redis:latest",
			},
		},
	}

	// Mock project (simple struct)
	testProject := &project.Project{
		Clients:        []string{"vscode"},
		DefaultSession: "test",
	}

	// Create a custom generator with mocked project manager functionality
	result, err := generator.injectSecretsForTesting(baseConfig, configuredSecrets)
	if err != nil {
		t.Fatalf("Failed to inject secrets: %v", err)
	}

	// Verify secrets section was added
	secretsSection, ok := result["secrets"].(map[string]interface{})
	if !ok {
		t.Fatal("Secrets section not found in result")
	}

	if len(secretsSection) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(secretsSection))
	}

	// Check that secrets are marked as external
	for secretName := range configuredSecrets {
		if secretConfig, exists := secretsSection[secretName]; exists {
			secretMap := secretConfig.(map[string]interface{})
			if secretMap["external"] != true {
				t.Errorf("Secret %s should be external", secretName)
			}
		} else {
			t.Errorf("Secret %s not found in secrets section", secretName)
		}
	}

	// Verify services section
	services := result["services"].(map[string]interface{})

	// Check web service (should have secrets due to environment variables)
	webService := services["web"].(map[string]interface{})
	webSecrets, ok := webService["secrets"].([]interface{})
	if !ok {
		t.Fatal("Web service should have secrets")
	}

	if len(webSecrets) != 2 {
		t.Errorf("Expected 2 secrets for web service, got %d", len(webSecrets))
	}

	// Check worker service (should have secrets due to label)
	workerService := services["worker"].(map[string]interface{})
	workerSecrets, ok := workerService["secrets"].([]interface{})
	if !ok {
		t.Fatal("Worker service should have secrets")
	}

	if len(workerSecrets) != 1 {
		t.Errorf("Expected 1 secret for worker service, got %d", len(workerSecrets))
	}

	// Check cache service (should not have secrets)
	cacheService := services["cache"].(map[string]interface{})
	if _, hasSecrets := cacheService["secrets"]; hasSecrets {
		t.Error("Cache service should not have secrets")
	}

	_ = testProject // Use the variable to avoid unused error
}

func TestInjectSecrets_NoSecretsAvailable(t *testing.T) {
	generator := NewBaseGenerator("/tmp")

	baseConfig := map[string]interface{}{
		"version": "3.8",
		"services": map[string]interface{}{
			"web": map[string]interface{}{
				"image": "nginx:latest",
			},
		},
	}

	// No secrets configured
	result, err := generator.injectSecretsForTesting(baseConfig, map[string]bool{})
	if err != nil {
		t.Fatalf("Failed to inject secrets: %v", err)
	}

	// Should return unchanged config when no secrets
	if _, hasSecrets := result["secrets"]; hasSecrets {
		t.Error("Should not have secrets section when no secrets configured")
	}

	// Verify original structure is preserved
	services := result["services"].(map[string]interface{})
	webService := services["web"].(map[string]interface{})
	if webService["image"] != "nginx:latest" {
		t.Error("Original service configuration should be preserved")
	}
}
