package runtime

import (
	"strings"

	"github.com/servo/servo/pkg"
)

// RuntimeFeature defines how a runtime maps to devcontainer features
type RuntimeFeature interface {
	Name() string
	FeatureID() string
	DefaultVersion() string
	Config() map[string]interface{}
	SelectVersion(requested string) string
	SupportsVersion(version string) bool
}

// RuntimeFeatureRegistry manages runtime feature mappings
type RuntimeFeatureRegistry interface {
	Register(feature RuntimeFeature)
	GetFeature(runtimeName string) (RuntimeFeature, bool)
	ListFeatures() []RuntimeFeature
}

// DefaultRuntimeFeatureRegistry is the default implementation
type DefaultRuntimeFeatureRegistry struct {
	features map[string]RuntimeFeature
}

// NewRuntimeFeatureRegistry creates a new registry with default features
func NewRuntimeFeatureRegistry() RuntimeFeatureRegistry {
	registry := &DefaultRuntimeFeatureRegistry{
		features: make(map[string]RuntimeFeature),
	}

	// Register default features
	registry.Register(NewPythonFeature())
	registry.Register(NewNodeFeature())
	registry.Register(NewGoFeature())
	registry.Register(NewDockerFeature())

	return registry
}

func (r *DefaultRuntimeFeatureRegistry) Register(feature RuntimeFeature) {
	r.features[feature.Name()] = feature
}

func (r *DefaultRuntimeFeatureRegistry) GetFeature(runtimeName string) (RuntimeFeature, bool) {
	feature, exists := r.features[runtimeName]
	return feature, exists
}

func (r *DefaultRuntimeFeatureRegistry) ListFeatures() []RuntimeFeature {
	var features []RuntimeFeature
	for _, feature := range r.features {
		features = append(features, feature)
	}
	return features
}

// BaseRuntimeFeature provides common functionality
type BaseRuntimeFeature struct {
	RuntimeName   string
	Feature       string
	DefaultVer    string
	Versions      []string
	DefaultConfig map[string]interface{}
}

func (b *BaseRuntimeFeature) Name() string {
	return b.RuntimeName
}

func (b *BaseRuntimeFeature) FeatureID() string {
	return b.Feature
}

func (b *BaseRuntimeFeature) DefaultVersion() string {
	return b.DefaultVer
}

func (b *BaseRuntimeFeature) Config() map[string]interface{} {
	return b.DefaultConfig
}

func (b *BaseRuntimeFeature) SelectVersion(requested string) string {
	if requested == "" {
		return b.DefaultVersion()
	}

	// Find the best matching version
	for _, version := range b.Versions {
		if strings.HasPrefix(requested, version) {
			// Return the requested version if it's more specific than the supported version
			if len(requested) > len(version) && strings.HasPrefix(requested, version+".") {
				return requested
			}
			return version
		}
	}

	return b.DefaultVersion()
}

func (b *BaseRuntimeFeature) SupportsVersion(version string) bool {
	for _, v := range b.Versions {
		if v == version {
			return true
		}
	}
	return false
}

// Specific runtime implementations

// PythonFeature handles Python runtime
type PythonFeature struct {
	*BaseRuntimeFeature
}

func NewPythonFeature() RuntimeFeature {
	return &PythonFeature{
		BaseRuntimeFeature: &BaseRuntimeFeature{
			RuntimeName: "python",
			Feature:     "ghcr.io/devcontainers/features/python:1",
			DefaultVer:  "3.11",
			Versions:    []string{"3.12", "3.11", "3.10", "3.9"},
			DefaultConfig: map[string]interface{}{
				"installTools":      true,
				"installJupyterlab": false,
			},
		},
	}
}

// NodeFeature handles Node.js runtime
type NodeFeature struct {
	*BaseRuntimeFeature
}

func NewNodeFeature() RuntimeFeature {
	return &NodeFeature{
		BaseRuntimeFeature: &BaseRuntimeFeature{
			RuntimeName: "node",
			Feature:     "ghcr.io/devcontainers/features/node:1",
			DefaultVer:  "18",
			Versions:    []string{"20", "18", "16"},
			DefaultConfig: map[string]interface{}{
				"nodeGypDependencies": true,
			},
		},
	}
}

// GoFeature handles Go runtime
type GoFeature struct {
	*BaseRuntimeFeature
}

func NewGoFeature() RuntimeFeature {
	return &GoFeature{
		BaseRuntimeFeature: &BaseRuntimeFeature{
			RuntimeName:   "go",
			Feature:       "ghcr.io/devcontainers/features/go:1",
			DefaultVer:    "1.21",
			Versions:      []string{"1.22", "1.21", "1.20"},
			DefaultConfig: map[string]interface{}{},
		},
	}
}

// DockerFeature handles Docker runtime
type DockerFeature struct {
	*BaseRuntimeFeature
}

func NewDockerFeature() RuntimeFeature {
	return &DockerFeature{
		BaseRuntimeFeature: &BaseRuntimeFeature{
			RuntimeName: "docker",
			Feature:     "ghcr.io/devcontainers/features/docker-outside-of-docker:1",
			DefaultVer:  "latest",
			Versions:    []string{"latest", "24", "23", "20"},
			DefaultConfig: map[string]interface{}{
				"version":                  "latest",
				"moby":                     true,
				"dockerDashComposeVersion": "v2",
			},
		},
	}
}

// RuntimeAnalyzer analyzes servo manifests for runtime requirements
type RuntimeAnalyzer struct {
	registry RuntimeFeatureRegistry
}

func NewRuntimeAnalyzer() *RuntimeAnalyzer {
	return &RuntimeAnalyzer{
		registry: NewRuntimeFeatureRegistry(),
	}
}

func (a *RuntimeAnalyzer) AnalyzeManifests(manifests map[string]*pkg.ServoDefinition) map[string]RuntimeFeature {
	runtimeFeatures := make(map[string]RuntimeFeature)
	runtimeVersions := make(map[string]string)

	// Extract runtime requirements from all manifests
	for _, manifest := range manifests {
		if manifest != nil && manifest.Requirements != nil && manifest.Requirements.Runtimes != nil {
			for _, runtime := range manifest.Requirements.Runtimes {
				// Keep the highest version requirement for each runtime
				existingVersion := runtimeVersions[runtime.Name]
				if existingVersion == "" || a.isHigherVersion(runtime.Version, existingVersion) {
					runtimeVersions[runtime.Name] = runtime.Version
				}
			}
		}
	}

	// Map to features
	for runtimeName, version := range runtimeVersions {
		if feature, exists := a.registry.GetFeature(runtimeName); exists {
			// Create a configured feature
			runtimeFeatures[runtimeName] = &ConfiguredRuntimeFeature{
				RuntimeFeature:  feature,
				selectedVersion: feature.SelectVersion(version),
			}
		}
	}

	return runtimeFeatures
}

func (a *RuntimeAnalyzer) isHigherVersion(v1, v2 string) bool {
	// Simple string comparison for now - in production would use proper semver
	return v1 > v2
}

// ConfiguredRuntimeFeature wraps a RuntimeFeature with a specific selected version
type ConfiguredRuntimeFeature struct {
	RuntimeFeature
	selectedVersion string
}

func (c *ConfiguredRuntimeFeature) SelectVersion(requested string) string {
	return c.selectedVersion
}
