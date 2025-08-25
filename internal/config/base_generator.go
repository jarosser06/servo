package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/servo/servo/internal/manifest"
	"github.com/servo/servo/internal/override"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/session"
	"github.com/servo/servo/pkg"
)

// BaseGenerator provides common functionality for all configuration generators.
//
// This base class encapsulates the core logic for retrieving project data, managing
// active sessions, and validating secrets across different configuration output formats.
// It serves as the foundation for devcontainer, docker-compose, and other specific
// configuration generators.
//
// The generator maintains connections to project management, session management, and
// override management systems, providing a unified interface for configuration generation
// workflows.
type BaseGenerator struct {
	projectManager  *project.Manager
	sessionManager  *session.Manager
	overrideManager *override.Manager
	servoDir        string
}

// NewBaseGenerator creates a new base generator
func NewBaseGenerator(servoDir string) *BaseGenerator {
	projectDir, _ := os.Getwd() // Get current working directory as project directory

	return &BaseGenerator{
		projectManager:  project.NewManager(),
		sessionManager:  session.NewManager(servoDir),
		overrideManager: override.NewManager("", projectDir), // Session dir will be set dynamically
		servoDir:        servoDir,
	}
}

// GetActiveSessionData returns project, active session, and manifests
func (g *BaseGenerator) GetActiveSessionData() (*project.Project, *session.Session, map[string]*pkg.ServoDefinition, error) {
	project, err := g.projectManager.Get()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get project: %w", err)
	}

	activeSession, err := g.sessionManager.GetActive()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get active session: %w", err)
	}

	if activeSession == nil {
		return nil, nil, nil, fmt.Errorf("no active session found")
	}

	// Get manifests from session
	store := manifest.NewStore(g.sessionManager.GetSessionDir(activeSession.Name), nil)
	manifests, err := store.ListManifests()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list manifests: %w", err)
	}

	return project, activeSession, manifests, nil
}

// SetupOverrideManager updates the override manager with session directory
func (g *BaseGenerator) SetupOverrideManager(sessionName string) {
	sessionDir := g.sessionManager.GetSessionDir(sessionName)
	projectDir, _ := os.Getwd()
	g.overrideManager = override.NewManager(sessionDir, projectDir)
}

// ValidateSecretsBeforeGeneration ensures all required secrets are configured before generation.
//
// This validation prevents configuration generation with missing secrets, which would
// result in broken MCP server configurations. The method scans all manifests for
// required secret dependencies and cross-references them with the project's configured
// secrets to identify any missing dependencies.
//
// Returns an error if any required secrets are missing from the project configuration.
func (g *BaseGenerator) ValidateSecretsBeforeGeneration(project *project.Project, manifests map[string]*pkg.ServoDefinition) error {

	// Scan all manifests to build a comprehensive list of required secrets
	// This ensures we validate against the complete secret dependency graph
	requiredSecrets := make(map[string]bool)
	for _, manifest := range manifests {
		if manifest != nil && manifest.ConfigurationSchema != nil && manifest.ConfigurationSchema.Secrets != nil {
			for secretName, secretSchema := range manifest.ConfigurationSchema.Secrets {
				if secretSchema.Required {
					requiredSecrets[secretName] = true
				}
			}
		}
	}

	// If no required secrets, validation passes
	if len(requiredSecrets) == 0 {
		return nil
	}

	// Get configured secrets
	configuredSecrets, err := g.projectManager.GetConfiguredSecrets()
	if err != nil {
		return fmt.Errorf("failed to get configured secrets: %w", err)
	}

	// Check which required secrets are missing
	var missingSecrets []string
	for secretName := range requiredSecrets {
		if !configuredSecrets[secretName] {
			missingSecrets = append(missingSecrets, secretName)
		}
	}

	if len(missingSecrets) > 0 {
		return fmt.Errorf("required secrets not configured: %v. Set them using: servo secrets set <name> <value>", missingSecrets)
	}

	return nil
}

// CreateSecretProvider creates a function to retrieve secrets for template expansion
func (g *BaseGenerator) CreateSecretProvider(project *project.Project) func(string) (string, error) {
	// Get the configured secrets map
	configuredSecrets, err := g.projectManager.GetConfiguredSecrets()
	if err != nil {
		// Return a provider that always fails if we can't get configured secrets
		return func(secretName string) (string, error) {
			return "", fmt.Errorf("failed to access secrets: %w", err)
		}
	}

	return func(secretName string) (string, error) {
		if !configuredSecrets[secretName] {
			return "", fmt.Errorf("secret '%s' is not configured", secretName)
		}
		// For now, return empty string as placeholder since actual decryption requires password
		// In real usage, this would need to integrate with the secrets manager properly
		return "", nil
	}
}

// extractSecretFromEnvValue extracts secret names from environment value templates
func (g *BaseGenerator) extractSecretFromEnvValue(envValue string) string {
	if envValue == "" {
		return ""
	}

	// Check for ${SECRET_NAME} format
	if len(envValue) > 3 && envValue[0] == '$' && envValue[1] == '{' && envValue[len(envValue)-1] == '}' {
		secretName := envValue[2 : len(envValue)-1]
		return strings.ToLower(secretName)
	}

	// Check for /run/secrets/secret_name format
	if len(envValue) > 13 && envValue[:13] == "/run/secrets/" {
		secretName := envValue[13:]
		return strings.ToLower(secretName)
	}

	// Not a secret reference
	return ""
}

// extractRequiredSecretsFromManifests extracts all required secrets from servo manifests
func (g *BaseGenerator) extractRequiredSecretsFromManifests(manifests map[string]*pkg.ServoDefinition) []string {
	secretsSet := make(map[string]bool)

	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}

		// Extract from server environment variables
		if manifest.Server.Environment != nil {
			for _, envValue := range manifest.Server.Environment {
				if secretName := g.extractSecretFromEnvValue(envValue); secretName != "" {
					secretsSet[secretName] = true
				}
			}
		}

		// Extract from service definitions - check both Dependencies.Services and direct Services
		var servicesToCheck map[string]*pkg.ServiceDependency
		if manifest.Dependencies != nil && manifest.Dependencies.Services != nil {
			servicesToCheck = make(map[string]*pkg.ServiceDependency)
			for name, service := range manifest.Dependencies.Services {
				servicesToCheck[name] = &service
			}
		}
		if manifest.Services != nil {
			if servicesToCheck == nil {
				servicesToCheck = manifest.Services
			} else {
				// Merge both if present
				for name, service := range manifest.Services {
					servicesToCheck[name] = service
				}
			}
		}

		if servicesToCheck != nil {
			for _, service := range servicesToCheck {
				if service != nil && service.Environment != nil {
					for _, envValue := range service.Environment {
						if secretName := g.extractSecretFromEnvValue(envValue); secretName != "" {
							secretsSet[secretName] = true
						}
					}
				}
			}
		}

		// Extract from configuration schema secrets (only required ones)
		if manifest.ConfigurationSchema != nil && manifest.ConfigurationSchema.Secrets != nil {
			for secretName, secretSchema := range manifest.ConfigurationSchema.Secrets {
				if secretSchema.Required {
					secretsSet[secretName] = true
				}
			}
		}
	}

	// Convert set to slice
	var secrets []string
	for secretName := range secretsSet {
		secrets = append(secrets, secretName)
	}

	return secrets
}

// serviceNeedsSecrets checks if a service configuration needs secrets injection
func (g *BaseGenerator) serviceNeedsSecrets(serviceConfig map[string]interface{}, availableSecrets map[string]bool) []string {
	var neededSecrets []string
	secretsSet := make(map[string]bool)

	// Check environment section - handle both map and array formats
	if env, ok := serviceConfig["environment"].(map[string]interface{}); ok {
		// Map format: {"KEY": "value"}
		for _, value := range env {
			if envValue, ok := value.(string); ok {
				if secretName := g.extractSecretFromEnvValue(envValue); secretName != "" {
					secretsSet[secretName] = true
				}
			}
		}
	} else if envArray, ok := serviceConfig["environment"].([]interface{}); ok {
		// Array format: ["KEY=value", "KEY2=value2"]
		for _, envItem := range envArray {
			if envLine, ok := envItem.(string); ok {
				// Extract value part from KEY=VALUE format
				if parts := strings.SplitN(envLine, "=", 2); len(parts) == 2 {
					if secretName := g.extractSecretFromEnvValue(parts[1]); secretName != "" {
						secretsSet[secretName] = true
					}
				}
			}
		}
	} else if envStringArray, ok := serviceConfig["environment"].([]string); ok {
		// String array format: ["KEY=value", "KEY2=value2"]
		for _, envLine := range envStringArray {
			// Extract value part from KEY=VALUE format
			if parts := strings.SplitN(envLine, "=", 2); len(parts) == 2 {
				if secretName := g.extractSecretFromEnvValue(parts[1]); secretName != "" {
					secretsSet[secretName] = true
				}
			}
		}
	}

	// Check labels for explicit secret declarations
	if labels, ok := serviceConfig["labels"].(map[string]interface{}); ok {
		if secretsLabel, ok := labels["servo.secrets"].(string); ok {
			// Parse comma-separated secret names
			for _, secretName := range strings.Split(secretsLabel, ",") {
				secretName = strings.TrimSpace(secretName)
				if secretName != "" {
					secretsSet[secretName] = true
				}
			}
		}
	} else if labels, ok := serviceConfig["labels"].(map[string]string); ok {
		if secretsLabel, ok := labels["servo.secrets"]; ok {
			// Parse comma-separated secret names
			for _, secretName := range strings.Split(secretsLabel, ",") {
				secretName = strings.TrimSpace(secretName)
				if secretName != "" {
					secretsSet[secretName] = true
				}
			}
		}
	}

	// Convert set to slice
	for secretName := range secretsSet {
		neededSecrets = append(neededSecrets, secretName)
	}

	return neededSecrets
}

// injectSecretsForTesting is a helper method for testing secrets injection
func (g *BaseGenerator) injectSecretsForTesting(baseConfig map[string]interface{}, configuredSecrets map[string]bool) (map[string]interface{}, error) {
	// This is a simplified version for testing - in real implementation this would be more complex
	result := make(map[string]interface{})

	// Copy base configuration
	for key, value := range baseConfig {
		result[key] = value
	}

	// Track all secrets that need to be added to the top-level secrets section
	allNeededSecrets := make(map[string]bool)

	// Check services section for secrets injection
	if services, ok := result["services"].(map[string]interface{}); ok {
		for serviceName, serviceConfig := range services {
			if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
				neededSecrets := g.serviceNeedsSecrets(serviceMap, configuredSecrets)
				if len(neededSecrets) > 0 {
					// Add secrets section to service
					if serviceMap["secrets"] == nil {
						serviceMap["secrets"] = []interface{}{}
					}
					secrets := serviceMap["secrets"].([]interface{})
					for _, secretName := range neededSecrets {
						secrets = append(secrets, map[string]interface{}{
							"source": secretName,
							"target": secretName,
						})
						// Track this secret for the top-level secrets section
						allNeededSecrets[secretName] = true
					}
					serviceMap["secrets"] = secrets
					services[serviceName] = serviceMap
				}
			}
		}
	}

	// Create top-level secrets section if we have any secrets
	if len(allNeededSecrets) > 0 {
		secretsSection := make(map[string]interface{})
		for secretName := range allNeededSecrets {
			secretsSection[secretName] = map[string]interface{}{
				"external": true,
			}
		}
		result["secrets"] = secretsSection
	}

	return result, nil
}
