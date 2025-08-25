package mcp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/servo/servo/pkg"
	"gopkg.in/yaml.v3"
)

// Parser handles parsing .servo files from various sources
type Parser struct {
	// Authentication options
	SSHKeyPath   string
	SSHPassword  string
	HTTPUsername string
	HTTPPassword string
	HTTPToken    string
}

// NewParser creates a new servo file parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFromFile parses a .servo file from a local file path
func (p *Parser) ParseFromFile(filePath string) (*pkg.ServoDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return p.parseYAML(data)
}

// ParseFromURL parses a .servo file from a remote URL
func (p *Parser) ParseFromURL(urlStr string) (*pkg.ServoDefinition, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d when fetching %s", resp.StatusCode, urlStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return p.parseYAML(data)
}

// ParseFromGitRepo clones a git repository and parses a .servo file from it
// Supports SSH key authentication and all git hosting services
func (p *Parser) ParseFromGitRepo(repoURL string, subdirectory string) (*pkg.ServoDefinition, error) {
	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "servo-clone-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up authentication with preference: explicit options > environment > system defaults
	var auth interface{}

	switch {
	case strings.HasPrefix(repoURL, "git@") || strings.Contains(repoURL, "ssh://"):
		// SSH authentication - go-git handles most of this automatically

		// 1. Use explicit SSH key if provided
		if p.SSHKeyPath != "" {
			if sshAuth, err := ssh.NewPublicKeysFromFile("git", p.SSHKeyPath, p.SSHPassword); err == nil {
				auth = sshAuth
			}
		}

		// 2. Try SSH agent if no explicit key
		if auth == nil {
			if sshAuth, err := ssh.NewSSHAgentAuth("git"); err == nil {
				auth = sshAuth
			}
		}

		// 3. If no explicit auth set, go-git automatically handles:
		//    - Reading ~/.ssh/config for host-specific configurations
		//    - Trying common SSH key locations (~/.ssh/id_rsa, ~/.ssh/id_ed25519, etc.)
		//    - Using SSH agent
		//    - Handling known_hosts verification
		//    - SSH key passphrases via system prompts or agents

	case strings.HasPrefix(repoURL, "https://"):
		// HTTPS authentication with multiple options

		// 1. Use explicit credentials if provided
		switch {
		case p.HTTPToken != "":
			auth = &githttp.BasicAuth{
				Username: "token", // Standard token format for GitHub/GitLab
				Password: p.HTTPToken,
			}
		case p.HTTPUsername != "" && p.HTTPPassword != "":
			auth = &githttp.BasicAuth{
				Username: p.HTTPUsername,
				Password: p.HTTPPassword,
			}
		default:
			// 2. Check environment variables
			if token := os.Getenv("GITHUB_TOKEN"); token != "" {
				auth = &githttp.BasicAuth{
					Username: "token",
					Password: token,
				}
			} else if username := os.Getenv("GIT_USERNAME"); username != "" {
				if password := os.Getenv("GIT_PASSWORD"); password != "" {
					auth = &githttp.BasicAuth{
						Username: username,
						Password: password,
					}
				}
			}
		}

		// 3. If no explicit auth, go-git will automatically:
		//    - Use git credential helpers (credential-manager, etc.)
		//    - Check ~/.netrc for stored credentials
		//    - Use system keychain/credential store
	}

	// Clone the repository
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: nil, // Silent clone
		Depth:    1,   // Shallow clone for efficiency
	}
	if auth != nil {
		switch a := auth.(type) {
		case *githttp.BasicAuth:
			cloneOptions.Auth = a
		case ssh.AuthMethod:
			cloneOptions.Auth = a
		}
	}

	_, err = git.PlainClone(tempDir, false, cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
	}

	// Determine the directory to search for .servo files
	searchDir := tempDir
	if subdirectory != "" {
		searchDir = filepath.Join(tempDir, subdirectory)
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("subdirectory %s not found in repository", subdirectory)
		}
	}

	// Find and parse the .servo file
	return p.ParseFromDirectory(searchDir)
}

// ParseFromDirectory finds and parses a .servo file in a directory
func (p *Parser) ParseFromDirectory(dirPath string) (*pkg.ServoDefinition, error) {
	// Look for any .servo files in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var servoFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".servo") {
			servoFiles = append(servoFiles, filepath.Join(dirPath, entry.Name()))
		}
	}

	if len(servoFiles) == 0 {
		return nil, fmt.Errorf("no .servo files found in directory %s", dirPath)
	}

	if len(servoFiles) > 1 {
		return nil, fmt.Errorf("multiple .servo files found in directory %s, specify one: %v", dirPath, servoFiles)
	}

	return p.ParseFromFile(servoFiles[0])
}

// parseYAML parses YAML data into ServoDefinition
func (p *Parser) parseYAML(data []byte) (*pkg.ServoDefinition, error) {
	// First, try to parse with the new structure
	var servo pkg.ServoDefinition
	if err := yaml.Unmarshal(data, &servo); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Handle backward compatibility: check if we need to migrate from old metadata structure
	if servo.Name == "" {
		// Try parsing with the old structure to extract fields
		var legacyServo struct {
			pkg.ServoDefinition
			Metadata *struct {
				Name        string   `yaml:"name"`
				Version     string   `yaml:"version"`
				Description string   `yaml:"description"`
				Author      string   `yaml:"author"`
				License     string   `yaml:"license"`
				Homepage    string   `yaml:"homepage,omitempty"`
				Repository  string   `yaml:"repository,omitempty"`
				Tags        []string `yaml:"tags,omitempty"`
			} `yaml:"metadata"`
		}

		if err := yaml.Unmarshal(data, &legacyServo); err == nil && legacyServo.Metadata != nil {
			// Migrate fields from legacy structure
			if servo.Name == "" && legacyServo.Metadata.Name != "" {
				servo.Name = legacyServo.Metadata.Name
			}
			if servo.Version == "" && legacyServo.Metadata.Version != "" {
				servo.Version = legacyServo.Metadata.Version
			}
			if servo.Description == "" && legacyServo.Metadata.Description != "" {
				servo.Description = legacyServo.Metadata.Description
			}
			if servo.Author == "" && legacyServo.Metadata.Author != "" {
				servo.Author = legacyServo.Metadata.Author
			}
			if servo.License == "" && legacyServo.Metadata.License != "" {
				servo.License = legacyServo.Metadata.License
			}

			// Preserve optional metadata fields in the new structure
			if servo.Metadata == nil {
				servo.Metadata = &pkg.Metadata{}
			}
			if servo.Metadata.Homepage == "" && legacyServo.Metadata.Homepage != "" {
				servo.Metadata.Homepage = legacyServo.Metadata.Homepage
			}
			if servo.Metadata.Repository == "" && legacyServo.Metadata.Repository != "" {
				servo.Metadata.Repository = legacyServo.Metadata.Repository
			}
			if len(servo.Metadata.Tags) == 0 && len(legacyServo.Metadata.Tags) > 0 {
				servo.Metadata.Tags = legacyServo.Metadata.Tags
			}
		}
	}

	return &servo, nil
}

// convertGitToRawURL converts git repository URLs to raw content URLs
func (p *Parser) convertGitToRawURL(repoURL string, subdirectory string) (string, error) {
	// Parse the repository URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %w", err)
	}

	// Handle GitHub URLs
	if strings.Contains(parsedURL.Host, "github.com") {
		// Convert github.com/user/repo.git to raw.githubusercontent.com/user/repo/main/name.servo
		path := strings.TrimSuffix(parsedURL.Path, ".git")
		path = strings.TrimPrefix(path, "/")

		// Extract repository name for .servo file name
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return "", fmt.Errorf("invalid GitHub repository path: %s", path)
		}
		repoName := pathParts[len(pathParts)-1]

		servoPath := fmt.Sprintf("%s.servo", repoName)
		if subdirectory != "" {
			servoPath = filepath.Join(subdirectory, fmt.Sprintf("%s.servo", repoName))
		}

		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", path, servoPath)
		return rawURL, nil
	}

	// Handle other git hosting services here
	return "", fmt.Errorf("unsupported git hosting service: %s", parsedURL.Host)
}

// Validator handles validation of .servo files
type Validator struct {
}

// NewValidator creates a new servo file validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a ServoDefinition
func (v *Validator) Validate(servo *pkg.ServoDefinition) error {
	if servo == nil {
		return fmt.Errorf("servo definition cannot be nil")
	}

	// Validate servo version
	if err := v.validateServoVersion(servo.ServoVersion); err != nil {
		return err
	}

	// Validate required top-level fields
	if err := v.validateTopLevelFields(servo); err != nil {
		return err
	}

	// Validate optional metadata
	if servo.Metadata != nil {
		if err := v.validateMetadata(servo.Metadata); err != nil {
			return err
		}
	}

	// Validate requirements
	if servo.Requirements != nil {
		if err := v.validateRequirements(servo.Requirements); err != nil {
			return err
		}
	}

	// Validate install section
	if err := v.validateInstall(&servo.Install); err != nil {
		return err
	}

	// Validate dependencies
	if servo.Dependencies != nil {
		if err := v.validateDependencies(servo.Dependencies); err != nil {
			return err
		}
	}

	// Validate configuration schema
	if servo.ConfigurationSchema != nil {
		if err := v.validateConfigurationSchema(servo.ConfigurationSchema); err != nil {
			return err
		}
	}

	// Validate server section
	if err := v.validateServer(&servo.Server); err != nil {
		return err
	}

	// Validate clients section
	if servo.Clients != nil {
		if err := v.validateClients(servo.Clients); err != nil {
			return err
		}
	}

	return nil
}

// validateServoVersion validates the servo_version field
func (v *Validator) validateServoVersion(version string) error {
	if version == "" {
		return fmt.Errorf("servo_version is required")
	}

	validVersions := []string{"1.0"}
	for _, validVersion := range validVersions {
		if version == validVersion {
			return nil
		}
	}

	return fmt.Errorf("unsupported servo_version: %s, supported versions: %v", version, validVersions)
}

// validateTopLevelFields validates the required and optional top-level fields
func (v *Validator) validateTopLevelFields(servo *pkg.ServoDefinition) error {
	// Validate required name field
	if servo.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate name format (lowercase, hyphens only)
	nameRegex := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)
	if !nameRegex.MatchString(servo.Name) {
		return fmt.Errorf("name must be lowercase with hyphens only: %s", servo.Name)
	}

	// Validate optional version field if provided
	if servo.Version != "" {
		// Validate semantic version format
		semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$`)
		if !semverRegex.MatchString(servo.Version) {
			return fmt.Errorf("version must be valid semantic version: %s", servo.Version)
		}
	}

	// Validate optional description field if provided
	if servo.Description != "" && len(servo.Description) > 200 {
		return fmt.Errorf("description must be 200 characters or less")
	}

	return nil
}

// validateMetadata validates the optional metadata section
func (v *Validator) validateMetadata(metadata *pkg.Metadata) error {
	// Validate URLs if provided
	if metadata.Homepage != "" {
		if _, err := url.Parse(metadata.Homepage); err != nil {
			return fmt.Errorf("metadata.homepage must be valid URL: %w", err)
		}
	}

	if metadata.Repository != "" {
		if _, err := url.Parse(metadata.Repository); err != nil {
			return fmt.Errorf("metadata.repository must be valid URL: %w", err)
		}
	}

	// Validate tags
	tagRegex := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	for _, tag := range metadata.Tags {
		if !tagRegex.MatchString(tag) {
			return fmt.Errorf("invalid tag format: %s", tag)
		}
	}

	return nil
}

// validateRequirements validates the requirements section
func (v *Validator) validateRequirements(req *pkg.Requirements) error {
	// Validate system requirements
	for _, sysReq := range req.System {
		if sysReq.Name == "" {
			return fmt.Errorf("system requirement name is required")
		}
		if sysReq.Description == "" {
			return fmt.Errorf("system requirement description is required")
		}
		if sysReq.CheckCommand == "" {
			return fmt.Errorf("system requirement check_command is required")
		}
	}

	// Validate runtime requirements
	for _, runtime := range req.Runtimes {
		if runtime.Name == "" {
			return fmt.Errorf("runtime requirement name is required")
		}
		if runtime.Version == "" {
			return fmt.Errorf("runtime requirement version is required")
		}
	}

	return nil
}

// validateInstall validates the install section
func (v *Validator) validateInstall(install *pkg.Install) error {
	if install.Type == "" {
		return fmt.Errorf("install.type is required")
	}

	validTypes := []string{"git", "local", "file", "remote"}
	validType := false
	for _, vt := range validTypes {
		if install.Type == vt {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("install.type must be one of: %v", validTypes)
	}

	if install.Method == "" {
		return fmt.Errorf("install.method is required")
	}

	if install.Method != install.Type {
		return fmt.Errorf("install.method must match install.type")
	}

	// Type-specific validations
	if install.Type == "git" {
		if install.Repository == "" {
			return fmt.Errorf("install.repository is required for git type")
		}
		if _, err := url.Parse(install.Repository); err != nil {
			return fmt.Errorf("install.repository must be valid URL: %w", err)
		}
	}

	if len(install.SetupCommands) == 0 {
		return fmt.Errorf("install.setup_commands is required")
	}

	// Validate commands for safety
	for _, cmd := range install.SetupCommands {
		if err := v.validateCommand(cmd); err != nil {
			return fmt.Errorf("unsafe setup command: %w", err)
		}
	}

	for _, cmd := range install.BuildCommands {
		if err := v.validateCommand(cmd); err != nil {
			return fmt.Errorf("unsafe build command: %w", err)
		}
	}

	for _, cmd := range install.TestCommands {
		if err := v.validateCommand(cmd); err != nil {
			return fmt.Errorf("unsafe test command: %w", err)
		}
	}

	return nil
}

// validateDependencies validates the dependencies section
func (v *Validator) validateDependencies(deps *pkg.Dependencies) error {
	for serviceName, service := range deps.Services {
		if serviceName == "" {
			return fmt.Errorf("service name cannot be empty")
		}

		if service.Image == "" {
			return fmt.Errorf("service.image is required for service %s", serviceName)
		}

		// Validate Docker image format
		if !v.isValidDockerImage(service.Image) {
			return fmt.Errorf("invalid Docker image format: %s", service.Image)
		}

		// Validate ports
		for _, port := range service.Ports {
			if err := v.validatePort(port); err != nil {
				return fmt.Errorf("invalid port for service %s: %w", serviceName, err)
			}
		}

		// Validate health check
		if service.HealthCheck != nil {
			if err := v.validateHealthCheck(service.HealthCheck); err != nil {
				return fmt.Errorf("invalid health check for service %s: %w", serviceName, err)
			}
		}
	}

	return nil
}

// validateConfigurationSchema validates the configuration_schema section
func (v *Validator) validateConfigurationSchema(schema *pkg.ConfigurationSchema) error {
	// Validate secrets
	for secretName, secret := range schema.Secrets {
		if secret.Description == "" {
			return fmt.Errorf("secret %s: description is required", secretName)
		}
		if secret.Type == "" {
			return fmt.Errorf("secret %s: type is required", secretName)
		}
		if secret.EnvVar == "" {
			return fmt.Errorf("secret %s: env_var is required", secretName)
		}

		validSecretTypes := []string{"api_key", "password", "certificate", "url"}
		if !v.contains(validSecretTypes, secret.Type) {
			return fmt.Errorf("secret %s: invalid type %s", secretName, secret.Type)
		}
	}

	// Validate config
	for configName, config := range schema.Config {
		if config.Description == "" {
			return fmt.Errorf("config %s: description is required", configName)
		}
		if config.Type == "" {
			return fmt.Errorf("config %s: type is required", configName)
		}
		if config.EnvVar == "" {
			return fmt.Errorf("config %s: env_var is required", configName)
		}

		validConfigTypes := []string{"string", "integer", "boolean", "select", "multiselect", "file", "url"}
		if !v.contains(validConfigTypes, config.Type) {
			return fmt.Errorf("config %s: invalid type %s", configName, config.Type)
		}

		// Validate select type has options
		if config.Type == "select" || config.Type == "multiselect" {
			if len(config.Options) == 0 {
				return fmt.Errorf("config %s: options required for %s type", configName, config.Type)
			}
		}
	}

	return nil
}

// validateServer validates the server section
func (v *Validator) validateServer(server *pkg.Server) error {
	if server.Transport == "" {
		return fmt.Errorf("server.transport is required")
	}

	validTransports := []string{"stdio", "sse", "http"}
	if !v.contains(validTransports, server.Transport) {
		return fmt.Errorf("server.transport must be one of: %v", validTransports)
	}

	if server.Command == "" {
		return fmt.Errorf("server.command is required")
	}

	if len(server.Args) == 0 {
		return fmt.Errorf("server.args is required")
	}

	return nil
}

// validateClients validates the clients section
func (v *Validator) validateClients(clients *pkg.ClientInfo) error {
	// No specific validation needed for recommended/tested/excluded lists
	// They're just informational

	// Validate client requirements
	for clientName, req := range clients.Requirements {
		if req.MinimumVersion != "" {
			// Basic semantic version check
			semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)
			if !semverRegex.MatchString(req.MinimumVersion) {
				return fmt.Errorf("client %s: invalid minimum_version format", clientName)
			}
		}
	}

	return nil
}

// validateCommand validates that a command is safe to execute
func (v *Validator) validateCommand(cmd string) error {
	// Basic safety checks - prevent obviously dangerous commands
	dangerousPatterns := []string{
		"rm -rf",
		"sudo",
		"su ",
		"chmod 777",
		"wget http://",
		"curl http://",
		"bash -c",
		"sh -c",
		"eval",
		"exec",
	}

	cmdLower := strings.ToLower(cmd)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdLower, pattern) {
			return fmt.Errorf("potentially dangerous command: %s", cmd)
		}
	}

	return nil
}

// validatePort validates a port string
func (v *Validator) validatePort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535: %d", portNum)
	}

	return nil
}

// validateHealthCheck validates a health check configuration
func (v *Validator) validateHealthCheck(hc *pkg.HealthCheck) error {
	if len(hc.Test) == 0 {
		return fmt.Errorf("healthcheck.test is required")
	}

	// Validate duration formats
	if hc.Interval != "" && !v.isValidDuration(hc.Interval) {
		return fmt.Errorf("invalid healthcheck interval: %s", hc.Interval)
	}

	if hc.Timeout != "" && !v.isValidDuration(hc.Timeout) {
		return fmt.Errorf("invalid healthcheck timeout: %s", hc.Timeout)
	}

	if hc.Retries < 0 {
		return fmt.Errorf("healthcheck retries cannot be negative: %d", hc.Retries)
	}

	return nil
}

// isValidDockerImage checks if a string is a valid Docker image reference
func (v *Validator) isValidDockerImage(image string) bool {
	// Basic validation - more comprehensive validation could be added
	if image == "" {
		return false
	}

	// Check for basic format: [registry/]name[:tag]
	parts := strings.Split(image, ":")
	if len(parts) > 2 {
		return false
	}

	return true
}

// isValidDuration checks if a string is a valid duration format
func (v *Validator) isValidDuration(duration string) bool {
	durationRegex := regexp.MustCompile(`^\d+[smh]$`)
	return durationRegex.MatchString(duration)
}

// contains checks if a slice contains a string
func (v *Validator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
