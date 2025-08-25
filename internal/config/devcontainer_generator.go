package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/servo/servo/internal/override"
	"github.com/servo/servo/internal/runtime"
	"github.com/servo/servo/pkg"
)

// DevcontainerGenerator handles devcontainer.json generation
type DevcontainerGenerator struct {
	*BaseGenerator
	runtimeAnalyzer *runtime.RuntimeAnalyzer
}

// NewDevcontainerGenerator creates a new devcontainer generator
func NewDevcontainerGenerator(servoDir string) *DevcontainerGenerator {
	return &DevcontainerGenerator{
		BaseGenerator:   NewBaseGenerator(servoDir),
		runtimeAnalyzer: runtime.NewRuntimeAnalyzer(),
	}
}

// Generate generates .devcontainer/devcontainer.json for infrastructure setup
func (g *DevcontainerGenerator) Generate() error {
	project, activeSession, manifests, err := g.GetActiveSessionData()
	if err != nil {
		return err
	}

	// Validate secrets before generating configuration
	if err := g.ValidateSecretsBeforeGeneration(project, manifests); err != nil {
		return fmt.Errorf("secrets validation failed: %w", err)
	}

	// Setup override manager
	g.SetupOverrideManager(activeSession.Name)

	// Generate base devcontainer configuration (infrastructure only)
	devcontainerConfig := g.buildBaseDevcontainerConfig()

	// Apply overrides with precedence: session > project > defaults
	finalConfig := g.processDevcontainerOverrides(devcontainerConfig)

	// Create .devcontainer directory
	if err := os.MkdirAll(".devcontainer", 0755); err != nil {
		return fmt.Errorf("failed to create .devcontainer directory: %w", err)
	}

	// Write devcontainer.json
	data, err := json.MarshalIndent(finalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal devcontainer config: %w", err)
	}

	return os.WriteFile(".devcontainer/devcontainer.json", data, 0644)
}

// processDevcontainerOverrides applies override configurations to devcontainer config
func (g *DevcontainerGenerator) processDevcontainerOverrides(baseConfig map[string]interface{}) map[string]interface{} {
	if g.overrideManager == nil {
		return baseConfig
	}

	// Get devcontainer overrides from the override manager
	overrides, err := g.overrideManager.GetDevcontainerOverrides()
	if err != nil || overrides == nil {
		// If no overrides or error getting them, return base config
		return baseConfig
	}

	// Convert the typed override to generic map format and apply
	overrideMap := g.convertDevcontainerOverrideToMap(overrides)
	if overrideMap == nil {
		return baseConfig
	}

	return g.applyDevcontainerOverrides(baseConfig, overrideMap)
}

// convertDevcontainerOverrideToMap converts a typed DevcontainerOverride to a generic map format
func (g *DevcontainerGenerator) convertDevcontainerOverrideToMap(override *override.DevcontainerOverride) map[string]interface{} {
	result := make(map[string]interface{})

	// Add basic fields
	if override.Name != "" {
		result["name"] = override.Name
	}
	if override.Image != "" {
		result["image"] = override.Image
	}
	if override.RemoteUser != "" {
		result["remoteUser"] = override.RemoteUser
	}
	if override.WorkspaceFolder != "" {
		result["workspaceFolder"] = override.WorkspaceFolder
	}
	if override.PostCreateCommand != "" {
		result["postCreateCommand"] = override.PostCreateCommand
	}
	if override.PostStartCommand != "" {
		result["postStartCommand"] = override.PostStartCommand
	}

	// Add arrays
	if len(override.ForwardPorts) > 0 {
		result["forwardPorts"] = override.ForwardPorts
	}
	if len(override.Mounts) > 0 {
		result["mounts"] = override.Mounts
	}

	// Add features if specified
	if len(override.Features) > 0 {
		result["features"] = override.Features
	}

	// Add customizations if specified
	if len(override.Customizations) > 0 {
		result["customizations"] = override.Customizations
	}

	// Add any extra fields
	for key, value := range override.Extra {
		result[key] = value
	}

	return result
}

// buildBaseDevcontainerConfig creates the base infrastructure-only devcontainer configuration
func (g *DevcontainerGenerator) buildBaseDevcontainerConfig() map[string]interface{} {
	_, _, manifests, err := g.GetActiveSessionData()
	if err != nil {
		// Fallback to basic config if we can't get manifests
		return g.buildFallbackConfig()
	}

	config := map[string]interface{}{
		"name":              "Servo Development Environment",
		"dockerComposeFile": []string{"docker-compose.yml"},
		"service":           "workspace",
		"workspaceFolder":   "/workspace",
		"remoteUser":        "root",
	}

	// Extract runtime requirements and build features
	features := g.buildDevcontainerFeatures(manifests)

	// Extract ports from service definitions
	forwardPorts := g.extractForwardPorts(manifests)

	// Skip customizations - those are client-specific and should be handled by individual clients
	config["features"] = features
	config["forwardPorts"] = forwardPorts

	// Add setup commands for infrastructure
	config["onCreateCommand"] = g.buildOnCreateCommand(manifests)
	config["postStartCommand"] = g.buildPostStartCommand()

	return config
}

// buildFallbackConfig creates a basic config when manifests can't be loaded
func (g *DevcontainerGenerator) buildFallbackConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":              "Servo Development Environment",
		"dockerComposeFile": []string{"docker-compose.yml"},
		"service":           "workspace",
		"workspaceFolder":   "/workspace",
		"remoteUser":        "root",
		"features":          map[string]interface{}{},
		"forwardPorts":      []interface{}{},
		"customizations":    map[string]interface{}{},
		"onCreateCommand":   g.buildOnCreateCommand(nil),
		"postStartCommand":  g.buildPostStartCommand(),
	}
}

// buildDevcontainerFeatures analyzes manifests and builds devcontainer features
func (g *DevcontainerGenerator) buildDevcontainerFeatures(manifests map[string]*pkg.ServoDefinition) map[string]interface{} {
	features := make(map[string]interface{})

	// Use the runtime analyzer to get configured features
	runtimeFeatures := g.runtimeAnalyzer.AnalyzeManifests(manifests)

	// Add features for each detected runtime
	for _, runtimeFeature := range runtimeFeatures {
		featureConfig := make(map[string]interface{})

		// Copy base configuration
		for key, value := range runtimeFeature.Config() {
			featureConfig[key] = value
		}

		// Set the selected version
		selectedVersion := runtimeFeature.SelectVersion("")
		if selectedVersion != "" {
			featureConfig["version"] = selectedVersion
		}

		features[runtimeFeature.FeatureID()] = featureConfig
	}

	// Always add common development features
	commonFeatures := g.getCommonFeatures()
	for feature, config := range commonFeatures {
		features[feature] = config
	}

	return features
}

// extractForwardPorts extracts all port mappings from service definitions
func (g *DevcontainerGenerator) extractForwardPorts(manifests map[string]*pkg.ServoDefinition) []interface{} {
	var ports []interface{}
	seen := make(map[string]bool)

	for serverName, manifest := range manifests {
		if manifest == nil {
			continue
		}

		// Check both Dependencies.Services and direct Services fields
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
			for serviceName, service := range servicesToCheck {
				if service != nil {
					for _, portMapping := range service.Ports {
						// Handle different port mapping formats
						portStr := g.normalizePortMapping(portMapping, serverName, serviceName)
						if portStr != "" && !seen[portStr] {
							// Convert to integer for devcontainer forwardPorts
							if portInt, err := strconv.Atoi(portStr); err == nil {
								ports = append(ports, portInt)
								seen[portStr] = true
							}
						}
					}
				}
			}
		}
	}

	return ports
}

// normalizePortMapping converts various port mapping formats to devcontainer format
func (g *DevcontainerGenerator) normalizePortMapping(portMapping, serverName, serviceName string) string {
	// Handle formats like "5432", "5432:5432", "127.0.0.1:5433:5432"
	parts := strings.Split(portMapping, ":")

	switch len(parts) {
	case 1:
		// "5432" -> "5432"
		return parts[0]
	case 2:
		// "5432:5432" -> "5432"
		return parts[0]
	case 3:
		// "127.0.0.1:5433:5432" -> "5433"
		return parts[1]
	default:
		return ""
	}
}

// getCommonFeatures returns features that should always be included
func (g *DevcontainerGenerator) getCommonFeatures() map[string]interface{} {
	return map[string]interface{}{
		"ghcr.io/devcontainers/features/common-utils:2": map[string]interface{}{
			"installZsh":                 true,
			"configureZshAsDefaultShell": true,
			"installOhMyZsh":             true,
			"upgradePackages":            true,
			"username":                   "vscode",
		},
		"ghcr.io/devcontainers/features/docker-in-docker:2": map[string]interface{}{
			"moby":                     true,
			"dockerDashComposeVersion": "v2",
		},
		"ghcr.io/devcontainers/features/git:1": map[string]interface{}{
			"ppa":     true,
			"version": "latest",
		},
	}
}

// applyDevcontainerOverrides merges overrides into the base configuration
func (g *DevcontainerGenerator) applyDevcontainerOverrides(base map[string]interface{}, overrides interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base configuration
	for key, value := range base {
		result[key] = value
	}

	// Apply overrides if they exist and are valid
	if overrideMap, ok := overrides.(map[string]interface{}); ok {
		for key, value := range overrideMap {
			switch key {
			case "features":
				// Merge features
				if baseFeatures, ok := result["features"].(map[string]interface{}); ok {
					if overrideFeatures, ok := value.(map[string]interface{}); ok {
						for featureName, featureConfig := range overrideFeatures {
							baseFeatures[featureName] = featureConfig
						}
					}
				} else {
					result[key] = value
				}
			case "customizations":
				// Merge customizations
				if baseCust, ok := result["customizations"].(map[string]interface{}); ok {
					if overrideCust, ok := value.(map[string]interface{}); ok {
						for custKey, custValue := range overrideCust {
							baseCust[custKey] = custValue
						}
					}
				} else {
					result[key] = value
				}
			case "forwardPorts":
				// Merge forward ports arrays
				result[key] = g.mergeForwardPorts(result[key], value)
			default:
				// Direct override for other properties
				result[key] = value
			}
		}
	}

	return result
}

// buildOnCreateCommand builds the onCreateCommand for devcontainer infrastructure setup
func (g *DevcontainerGenerator) buildOnCreateCommand(manifests map[string]*pkg.ServoDefinition) string {
	commands := []string{
		"echo '🔧 Setting up development environment...'",
		"mkdir -p /workspace/.servo/services",
		"mkdir -p /workspace/.servo/logs",
	}

	// Add persistence directory creation commands
	persistenceDirs := g.extractPersistenceDirectories(manifests)
	for _, dir := range persistenceDirs {
		commands = append(commands, dir)
	}

	commands = append(commands,
		"echo '⚙️  Installing servo CLI...'",
		"if ! command -v servo &> /dev/null; then",
		"  cd /workspace &&",
		"  go build -o /usr/local/bin/servo ./cmd/servo &&",
		"  echo '✅ Servo CLI installed successfully'",
		"else",
		"  echo '✅ Servo CLI already available'",
		"fi",
		"echo '🎉 Development environment ready!'",
	)
	return strings.Join(commands, " && ")
}

// extractPersistenceDirectories extracts directory creation commands for volume persistence
func (g *DevcontainerGenerator) extractPersistenceDirectories(manifests map[string]*pkg.ServoDefinition) []string {
	var commands []string
	seen := make(map[string]bool)

	if manifests == nil {
		return commands
	}

	for serverName, manifest := range manifests {
		if manifest == nil {
			continue
		}

		// Check both Dependencies.Services and direct Services fields
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
			for serviceName, service := range servicesToCheck {
				if service != nil && len(service.Volumes) > 0 {
					serviceDir := fmt.Sprintf("mkdir -p /workspace/.servo/services/%s/%s", serverName, serviceName)
					logDir := fmt.Sprintf("mkdir -p /workspace/.servo/logs/%s/%s", serverName, serviceName)

					if !seen[serviceDir] {
						commands = append(commands, serviceDir)
						seen[serviceDir] = true
					}
					if !seen[logDir] {
						commands = append(commands, logDir)
						seen[logDir] = true
					}
				}
			}
		}
	}

	return commands
}

// buildPostStartCommand builds the postStartCommand for devcontainer
func (g *DevcontainerGenerator) buildPostStartCommand() string {
	return "echo 'Development container started. Services should be running via docker-compose.'; servo status || echo 'Servo not yet available - will be after setup completes'"
}

// mergeForwardPorts merges base forward ports with override forward ports
func (g *DevcontainerGenerator) mergeForwardPorts(base, override interface{}) interface{} {
	var basePorts []interface{}
	var overridePorts []interface{}

	if base != nil {
		if ports, ok := base.([]interface{}); ok {
			basePorts = ports
		}
	}

	if override != nil {
		if ports, ok := override.([]interface{}); ok {
			overridePorts = ports
		} else if ports, ok := override.([]int); ok {
			// Convert []int to []interface{}
			overridePorts = make([]interface{}, len(ports))
			for i, port := range ports {
				overridePorts[i] = port
			}
		}
	}

	// Use a map to deduplicate ports
	seen := make(map[interface{}]bool)
	var result []interface{}

	// Add base ports first
	for _, port := range basePorts {
		if !seen[port] {
			result = append(result, port)
			seen[port] = true
		}
	}

	// Add override ports
	for _, port := range overridePorts {
		if !seen[port] {
			result = append(result, port)
			seen[port] = true
		}
	}

	return result
}
