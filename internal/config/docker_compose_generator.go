package config

import (
	"fmt"
	"strings"

	"github.com/servo/servo/internal/override"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/utils"
	"github.com/servo/servo/pkg"
)

// DockerComposeGenerator handles docker-compose.yml generation
type DockerComposeGenerator struct {
	*BaseGenerator
}

// NewDockerComposeGenerator creates a new docker-compose generator
func NewDockerComposeGenerator(servoDir string) *DockerComposeGenerator {
	return &DockerComposeGenerator{
		BaseGenerator: NewBaseGenerator(servoDir),
	}
}

// Generate creates a docker-compose.yml file from active session manifests.
//
// This method implements a multi-stage configuration generation process:
//
// 1. Data Collection: Retrieves project, active session, and all manifests
// 2. Validation: Ensures all required secrets are configured
// 3. Base Configuration: Creates infrastructure-only docker-compose structure
// 4. Service Integration: Adds MCP server services from manifest definitions
// 5. Override Processing: Applies configuration overrides (session > project > defaults)
// 6. Secret Injection: Replaces secret placeholders with encrypted values
// 7. File Generation: Writes final docker-compose.yml to .devcontainer/
//
// The generated configuration supports both development and production deployment
// scenarios with proper volume mounting, networking, and service dependencies.
func (g *DockerComposeGenerator) Generate() error {
	project, activeSession, manifests, err := g.GetActiveSessionData()
	if err != nil {
		return err
	}

	if err := g.ValidateSecretsBeforeGeneration(project, manifests); err != nil {
		return fmt.Errorf("secrets validation failed: %w", err)
	}

	g.SetupOverrideManager(activeSession.Name)

	// Build the complete configuration through staged composition
	dockerComposeConfig := g.buildBaseDockerComposeConfig()
	if err := g.addServicesFromManifests(dockerComposeConfig, manifests); err != nil {
		return fmt.Errorf("failed to add services from manifests: %w", err)
	}
	finalConfig := g.processDockerComposeOverrides(dockerComposeConfig)

	if err := g.injectSecrets(finalConfig, project); err != nil {
		return fmt.Errorf("failed to inject secrets: %w", err)
	}

	// Create .devcontainer directory
	if err := utils.EnsureDirectoryStructure([]string{".devcontainer"}); err != nil {
		return fmt.Errorf("failed to create .devcontainer directory: %w", err)
	}

	// Write docker-compose.yml
	return utils.WriteYAMLFile(".devcontainer/docker-compose.yml", finalConfig)
}

// buildBaseDockerComposeConfig creates the base infrastructure-only docker-compose configuration
func (g *DockerComposeGenerator) buildBaseDockerComposeConfig() map[string]interface{} {
	config := map[string]interface{}{
		"version":  "3.8",
		"services": map[string]interface{}{},
		"volumes": map[string]interface{}{
			"workspace-data": nil,
		},
	}

	services := config["services"].(map[string]interface{})

	// Add workspace service (infrastructure only)
	workspaceService := g.buildWorkspaceService()
	services["workspace"] = workspaceService

	return config
}

// addServicesFromManifests adds services from manifests to docker-compose config
func (g *DockerComposeGenerator) addServicesFromManifests(config map[string]interface{}, manifests map[string]*pkg.ServoDefinition) error {
	services := config["services"].(map[string]interface{})

	for manifestName, manifest := range manifests {
		if manifest == nil {
			continue
		}

		// Check both Dependencies.Services and direct Services fields
		var servicesToAdd map[string]*pkg.ServiceDependency
		if manifest.Dependencies != nil && manifest.Dependencies.Services != nil {
			servicesToAdd = make(map[string]*pkg.ServiceDependency)
			for name, service := range manifest.Dependencies.Services {
				servicesToAdd[name] = &service
			}
		}
		if manifest.Services != nil {
			if servicesToAdd == nil {
				servicesToAdd = manifest.Services
			} else {
				// Merge both if present
				for name, service := range manifest.Services {
					servicesToAdd[name] = service
				}
			}
		}

		if servicesToAdd != nil {
			for serviceName, service := range servicesToAdd {
				// Add service with prefix to avoid conflicts
				prefixedName := fmt.Sprintf("%s-%s", manifestName, serviceName)
				serviceConfig := make(map[string]interface{})

				// Copy service configuration
				if service.Image != "" {
					serviceConfig["image"] = service.Image
				}
				if len(service.Ports) > 0 {
					serviceConfig["ports"] = service.Ports
				}
				
				// Merge environment variables from multiple sources
				envSlice := []string{}
				
				// 1. Add project-level environment variables first
				projectEnv, err := g.LoadProjectEnvironmentVariables()
				if err != nil {
					return fmt.Errorf("failed to load project environment variables: %w", err)
				}
				for key, value := range projectEnv {
					envSlice = append(envSlice, key+"="+value)
				}
				
				// 2. Add service-specific environment variables (these can override project-level)
				if service.Environment != nil {
					for key, value := range service.Environment {
						envSlice = append(envSlice, key+"="+value)
					}
				}
				
				// Set environment if we have any variables
				if len(envSlice) > 0 {
					serviceConfig["environment"] = envSlice
				}
				if len(service.Volumes) > 0 {
					// Transform volumes to use host paths for persistence
					transformedVolumes := make([]string, 0, len(service.Volumes))
					for _, volume := range service.Volumes {
						if strings.Contains(volume, ":") {
							parts := strings.Split(volume, ":")
							volumeName := parts[0]
							if !strings.HasPrefix(volumeName, "/") && !strings.HasPrefix(volumeName, ".") {
								// Named volume - transform to host path
								hostPath := fmt.Sprintf("../.servo/services/%s/%s/%s", manifestName, serviceName, volumeName)
								transformedVolume := hostPath + ":" + strings.Join(parts[1:], ":")
								transformedVolumes = append(transformedVolumes, transformedVolume)
							} else {
								// Already a host path or absolute path - keep as-is
								transformedVolumes = append(transformedVolumes, volume)
							}
						} else {
							// Simple named volume - transform to host path with default mount point
							hostPath := fmt.Sprintf("../.servo/services/%s/%s/%s", manifestName, serviceName, volume)
							transformedVolume := hostPath + ":/data"
							transformedVolumes = append(transformedVolumes, transformedVolume)
						}
					}
					serviceConfig["volumes"] = transformedVolumes
				}
				if len(service.Command) > 0 {
					serviceConfig["command"] = service.Command
				}

				services[prefixedName] = serviceConfig
			}
		}
	}
	return nil
}

// buildWorkspaceService creates the main workspace service
func (g *DockerComposeGenerator) buildWorkspaceService() map[string]interface{} {
	return map[string]interface{}{
		"build": map[string]interface{}{
			"dockerfile": "Dockerfile",
			"context":    "..",
		},
		"volumes": []interface{}{
			"../..:/workspaces:cached",
			"workspace-data:/workspace/.servo",
		},
		"command": "/bin/sh -c \"while sleep 1000; do :; done\"",
		"environment": []string{
			"SERVO_DEV_MODE=1",
		},
		"working_dir": "/workspaces",
	}
}

// processServiceConfig processes a service configuration and injects secrets
func (g *DockerComposeGenerator) processServiceConfig(serviceConfig interface{}, secretProvider func(string) (string, error)) interface{} {
	switch v := serviceConfig.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = g.processServiceConfig(value, secretProvider)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = g.processServiceConfig(item, secretProvider)
		}
		return result
	case string:
		return g.expandSecrets(v, secretProvider)
	default:
		return v
	}
}

// applyDockerComposeOverrides merges overrides into the base configuration
func (g *DockerComposeGenerator) applyDockerComposeOverrides(base map[string]interface{}, overrides map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base configuration
	for key, value := range base {
		result[key] = value
	}

	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "services":
			// Merge services
			if baseServices, ok := result["services"].(map[string]interface{}); ok {
				if overrideServices, ok := value.(map[string]interface{}); ok {
					for serviceName, serviceConfig := range overrideServices {
						if existingService, exists := baseServices[serviceName]; exists {
							// Merge existing service
							baseServices[serviceName] = g.mergeServiceConfigs(existingService, serviceConfig)
						} else {
							// Add new service
							baseServices[serviceName] = serviceConfig
						}
					}
				}
			} else {
				result[key] = value
			}
		case "volumes":
			// Merge volumes
			if baseVolumes, ok := result["volumes"].(map[string]interface{}); ok {
				if overrideVolumes, ok := value.(map[string]interface{}); ok {
					for volumeName, volumeConfig := range overrideVolumes {
						baseVolumes[volumeName] = volumeConfig
					}
				}
			} else {
				result[key] = value
			}
		default:
			// Direct override for other properties
			result[key] = value
		}
	}

	return result
}

// mergeServiceConfigs merges two service configurations
func (g *DockerComposeGenerator) mergeServiceConfigs(base, override interface{}) interface{} {
	baseMap, baseOk := base.(map[string]interface{})
	overrideMap, overrideOk := override.(map[string]interface{})

	if !baseOk || !overrideOk {
		// If either is not a map, override completely
		return override
	}

	result := make(map[string]interface{})

	// Copy base
	for key, value := range baseMap {
		result[key] = value
	}

	// Apply overrides
	for key, value := range overrideMap {
		switch key {
		case "environment":
			// Merge environment variables
			result[key] = g.mergeEnvironmentVars(result[key], value)
		case "volumes":
			// Merge volume mounts
			result[key] = g.mergeSlices(result[key], value)
		case "ports":
			// Override port mappings completely
			result[key] = value
		default:
			// Direct override
			result[key] = value
		}
	}

	return result
}

// mergeEnvironmentVars merges environment variable configurations
func (g *DockerComposeGenerator) mergeEnvironmentVars(base, override interface{}) interface{} {
	// Handle both map[string]string and []string formats
	baseEnv := g.normalizeEnvVars(base)
	overrideEnv := g.normalizeEnvVars(override)

	// Merge
	for key, value := range overrideEnv {
		baseEnv[key] = value
	}

	// Convert back to slice format (preferred for docker-compose)
	result := make([]string, 0, len(baseEnv))
	for key, value := range baseEnv {
		result = append(result, key+"="+value)
	}

	return result
}

// normalizeEnvVars converts environment variables to map[string]string format
func (g *DockerComposeGenerator) normalizeEnvVars(env interface{}) map[string]string {
	result := make(map[string]string)

	switch v := env.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if str, ok := value.(string); ok {
				result[key] = str
			}
		}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				parts := strings.SplitN(str, "=", 2)
				if len(parts) == 2 {
					result[parts[0]] = parts[1]
				} else {
					result[parts[0]] = ""
				}
			}
		}
	case []string:
		for _, str := range v {
			parts := strings.SplitN(str, "=", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			} else {
				result[parts[0]] = ""
			}
		}
	}

	return result
}

// mergeSlices merges two slices
func (g *DockerComposeGenerator) mergeSlices(base, override interface{}) interface{} {
	var baseSlice []interface{}
	var overrideSlice []interface{}

	if base != nil {
		if slice, ok := base.([]interface{}); ok {
			baseSlice = slice
		} else if stringSlice, ok := base.([]string); ok {
			// Convert []string to []interface{}
			baseSlice = make([]interface{}, len(stringSlice))
			for i, s := range stringSlice {
				baseSlice[i] = s
			}
		}
	}

	if override != nil {
		if slice, ok := override.([]interface{}); ok {
			overrideSlice = slice
		} else if stringSlice, ok := override.([]string); ok {
			// Convert []string to []interface{}
			overrideSlice = make([]interface{}, len(stringSlice))
			for i, s := range stringSlice {
				overrideSlice[i] = s
			}
		}
	}

	result := make([]interface{}, len(baseSlice)+len(overrideSlice))
	copy(result, baseSlice)
	copy(result[len(baseSlice):], overrideSlice)

	return result
}

// injectSecrets injects secrets into docker-compose configuration
func (g *DockerComposeGenerator) injectSecrets(config map[string]interface{}, project *project.Project) error {
	// Get configured secrets
	configuredSecrets, err := g.projectManager.GetConfiguredSecrets()
	if err != nil {
		return fmt.Errorf("failed to get configured secrets: %w", err)
	}

	// If no secrets are configured, skip injection
	if len(configuredSecrets) == 0 {
		return nil
	}

	// Track all secrets that need to be added to the top-level secrets section
	allNeededSecrets := make(map[string]bool)

	// Check services section for secrets injection
	if services, ok := config["services"].(map[string]interface{}); ok {
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
		config["secrets"] = secretsSection
	}

	return nil
}

// expandSecrets expands secret placeholders in a string
func (g *DockerComposeGenerator) expandSecrets(value string, secretProvider func(string) (string, error)) string {
	if !strings.Contains(value, "${") {
		return value
	}

	// Simple secret expansion - replace ${SECRET_NAME} with actual values
	result := value
	start := strings.Index(result, "${")
	for start != -1 {
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		secretName := result[start+2 : end]
		secretValue, err := secretProvider(secretName)
		if err != nil || secretValue == "" {
			// Keep placeholder if secret not found
			start = strings.Index(result[end+1:], "${")
			if start != -1 {
				start += end + 1
			}
			continue
		}

		result = result[:start] + secretValue + result[end+1:]
		start = strings.Index(result[start+len(secretValue):], "${")
		if start != -1 {
			start += len(secretValue)
		}
	}

	return result
}

// processDockerComposeOverrides applies override configurations to docker-compose config
func (g *DockerComposeGenerator) processDockerComposeOverrides(baseConfig map[string]interface{}) map[string]interface{} {
	if g.overrideManager == nil {
		return baseConfig
	}

	// Get docker-compose overrides from the override manager
	overrides, err := g.overrideManager.GetDockerComposeOverrides()
	if err != nil || overrides == nil {
		// If no overrides or error getting them, return base config
		return baseConfig
	}

	// Convert the typed override to generic map format and apply
	overrideMap := g.convertOverrideToMap(overrides)
	if overrideMap == nil {
		return baseConfig
	}

	return g.applyDockerComposeOverrides(baseConfig, overrideMap)
}

// convertOverrideToMap converts a typed DockerComposeOverride to a generic map format
func (g *DockerComposeGenerator) convertOverrideToMap(override *override.DockerComposeOverride) map[string]interface{} {
	result := make(map[string]interface{})

	// Add version if specified
	if override.Version != "" {
		result["version"] = override.Version
	}

	// Convert services
	if len(override.Services) > 0 {
		services := make(map[string]interface{})
		for name, service := range override.Services {
			serviceMap := make(map[string]interface{})

			if service.Image != "" {
				serviceMap["image"] = service.Image
			}
			if len(service.Environment) > 0 {
				// Convert environment map to slice format (preferred for docker-compose)
				envSlice := make([]string, 0, len(service.Environment))
				for key, value := range service.Environment {
					envSlice = append(envSlice, key+"="+value)
				}
				serviceMap["environment"] = envSlice
			}
			if len(service.Ports) > 0 {
				serviceMap["ports"] = service.Ports
			}
			if len(service.Volumes) > 0 {
				serviceMap["volumes"] = service.Volumes
			}
			if len(service.Command) > 0 {
				serviceMap["command"] = service.Command
			}
			if len(service.DependsOn) > 0 {
				serviceMap["depends_on"] = service.DependsOn
			}
			if len(service.Networks) > 0 {
				serviceMap["networks"] = service.Networks
			}
			if len(service.Labels) > 0 {
				serviceMap["labels"] = service.Labels
			}
			// Add any extra fields
			for key, value := range service.Extra {
				serviceMap[key] = value
			}

			services[name] = serviceMap
		}
		result["services"] = services
	}

	// Add networks if specified
	if len(override.Networks) > 0 {
		result["networks"] = override.Networks
	}

	// Add volumes if specified
	if len(override.Volumes) > 0 {
		result["volumes"] = override.Volumes
	}

	// Add secrets if specified
	if len(override.Secrets) > 0 {
		result["secrets"] = override.Secrets
	}

	return result
}
