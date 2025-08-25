package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/servo/servo/internal/constants"
	"github.com/servo/servo/internal/errors"
)

// Validator interface for consistent validation across the codebase
type Validator interface {
	Validate() error
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid bool
	Errors  []*errors.ServoError
	Warnings []string
}

// Add adds an error to the validation result
func (vr *ValidationResult) Add(err *errors.ServoError) {
	vr.Errors = append(vr.Errors, err)
	vr.IsValid = false
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// Error returns the first error if any exist
func (vr *ValidationResult) Error() error {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return nil
}

// AllErrors returns all validation errors as a single error
func (vr *ValidationResult) AllErrors() error {
	if len(vr.Errors) == 0 {
		return nil
	}
	
	if len(vr.Errors) == 1 {
		return vr.Errors[0]
	}
	
	var messages []string
	for _, err := range vr.Errors {
		messages = append(messages, err.Error())
	}
	
	return errors.New(errors.CategoryValidation, strings.Join(messages, "; "))
}

// Name validation functions

// ValidateName validates a general name field
func ValidateName(name string, fieldName string) *errors.ServoError {
	if name == "" {
		return errors.RequiredFieldError(fieldName)
	}
	
	if len(name) < constants.MinNameLength {
		return errors.ValidationErrorf("%s must be at least %d character", fieldName, constants.MinNameLength)
	}
	
	if len(name) > constants.MaxNameLength {
		return errors.ValidationErrorf("%s must be no more than %d characters", fieldName, constants.MaxNameLength)
	}
	
	return nil
}

// ValidateSessionName validates a session name with specific rules
func ValidateSessionName(name string) *errors.ServoError {
	if err := ValidateName(name, "session name"); err != nil {
		return err
	}
	
	// Session names should only contain alphanumeric, hyphens, and underscores
	sessionNameRegex := regexp.MustCompile(constants.PatternSessionName)
	if !sessionNameRegex.MatchString(name) {
		return errors.InvalidSessionNameError(name).
			WithContext("pattern", constants.PatternSessionName)
	}
	
	// Reserved names
	reservedNames := []string{"sessions", "config", "logs", "volumes", "manifests"}
	for _, reserved := range reservedNames {
		if name == reserved {
			return errors.ValidationErrorf("session name '%s' is reserved", name)
		}
	}
	
	return nil
}

// ValidateServerName validates MCP server names
func ValidateServerName(name string) *errors.ServoError {
	if err := ValidateName(name, "server name"); err != nil {
		return err
	}
	
	// Server names should be lowercase with hyphens only
	// Single character names are allowed
	if len(name) == 1 {
		if name[0] < 'a' || name[0] > 'z' {
			return errors.ValidationErrorf("server name must be lowercase: %s", name)
		}
	} else {
		serverNameRegex := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)
		if !serverNameRegex.MatchString(name) {
			return errors.ValidationErrorf("server name must be lowercase with hyphens only: %s", name)
		}
	}
	
	return nil
}

// Version validation

// ValidateVersion validates a semantic version string
func ValidateVersion(version string) *errors.ServoError {
	if version == "" {
		return errors.RequiredFieldError("version")
	}
	
	versionRegex := regexp.MustCompile(constants.PatternVersion)
	if !versionRegex.MatchString(version) {
		return errors.ValidationErrorf("invalid version format: %s (expected format: x.y.z)", version)
	}
	
	return nil
}

// ValidateServoVersion validates the servo_version field specifically
func ValidateServoVersion(version string) *errors.ServoError {
	if version == "" {
		return errors.RequiredFieldError("servo_version")
	}
	
	supportedVersions := []string{constants.DefaultServoVersion}
	for _, supported := range supportedVersions {
		if version == supported {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported servo_version: %s, supported versions: %v", version, supportedVersions)
}

// Description validation

// ValidateDescription validates description fields
func ValidateDescription(description string, required bool) *errors.ServoError {
	if required && description == "" {
		return errors.RequiredFieldError("description")
	}
	
	if len(description) > constants.MaxDescriptionLength {
		return errors.ValidationErrorf("description must be no more than %d characters", constants.MaxDescriptionLength)
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(description) {
		return errors.ValidationError("description must be valid UTF-8")
	}
	
	return nil
}

// URL validation

// ValidateURL validates URL fields
func ValidateURL(urlStr string, fieldName string, required bool) *errors.ServoError {
	if required && urlStr == "" {
		return errors.RequiredFieldError(fieldName)
	}
	
	if urlStr == "" {
		return nil // Optional field
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.ValidationErrorf("invalid %s URL: %v", fieldName, err)
	}
	
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return errors.ValidationErrorf("%s must use http or https scheme", fieldName)
	}
	
	return nil
}

// ValidateRepositoryURL validates Git repository URLs
func ValidateRepositoryURL(repoURL string) *errors.ServoError {
	if repoURL == "" {
		return errors.RequiredFieldError("repository")
	}
	
	// Support various Git URL formats
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^https://github\.com/[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+(?:\.git)?/?$`),
		regexp.MustCompile(`^git@github\.com:[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+(?:\.git)?$`),
		regexp.MustCompile(`^https://gitlab\.com/[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+(?:\.git)?/?$`),
		regexp.MustCompile(`^https://bitbucket\.org/[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+(?:\.git)?/?$`),
	}
	
	for _, pattern := range validPatterns {
		if pattern.MatchString(repoURL) {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported repository URL format: %s", repoURL)
}

// Client validation

// ValidateClient validates client names
func ValidateClient(client string) *errors.ServoError {
	if client == "" {
		return errors.RequiredFieldError("client")
	}
	
	for _, supported := range constants.SupportedClients {
		if client == supported {
			return nil
		}
	}
	
	return errors.ClientNotSupportedError(client)
}

// ValidateClients validates a list of client names
func ValidateClients(clients []string) *ValidationResult {
	result := &ValidationResult{IsValid: true}
	
	if len(clients) == 0 {
		result.AddWarning("no clients specified")
		return result
	}
	
	// Check for duplicates
	seen := make(map[string]bool)
	for _, client := range clients {
		if seen[client] {
			result.Add(errors.ValidationErrorf("duplicate client: %s", client))
			continue
		}
		seen[client] = true
		
		if err := ValidateClient(client); err != nil {
			result.Add(err)
		}
	}
	
	return result
}

// Transport validation

// ValidateTransport validates transport types
func ValidateTransport(transport string) *errors.ServoError {
	if transport == "" {
		return errors.RequiredFieldError("transport")
	}
	
	validTransports := []string{
		constants.TransportStdio,
		constants.TransportHTTP,
		constants.TransportSSE,
	}
	
	for _, valid := range validTransports {
		if transport == valid {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported transport: %s, supported: %v", transport, validTransports)
}

// Install method validation

// ValidateInstallMethod validates install methods
func ValidateInstallMethod(method string) *errors.ServoError {
	if method == "" {
		return errors.RequiredFieldError("install method")
	}
	
	validMethods := []string{
		constants.InstallMethodGit,
		constants.InstallMethodLocal,
		constants.InstallMethodDocker,
		constants.InstallMethodURL,
	}
	
	for _, valid := range validMethods {
		if method == valid {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported install method: %s, supported: %v", method, validMethods)
}

// Path validation

// ValidatePath validates file/directory paths
func ValidatePath(path string, fieldName string, required bool) *errors.ServoError {
	if required && path == "" {
		return errors.RequiredFieldError(fieldName)
	}
	
	if path == "" {
		return nil // Optional field
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(path) {
		return errors.ValidationErrorf("%s must be valid UTF-8", fieldName)
	}
	
	// Check for dangerous patterns
	dangerousPatterns := []string{"..", "//", "\\"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			return errors.ValidationErrorf("%s contains unsafe pattern: %s", fieldName, pattern)
		}
	}
	
	return nil
}

// Environment validation

// ValidateEnvironmentName validates environment names (development, production, etc.)
func ValidateEnvironmentName(env string) *errors.ServoError {
	if env == "" {
		return errors.RequiredFieldError("environment")
	}
	
	validEnvironments := []string{
		constants.EnvDevelopment,
		constants.EnvStaging,
		constants.EnvProduction,
		constants.EnvTesting,
	}
	
	for _, valid := range validEnvironments {
		if env == valid {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported environment: %s, supported: %v", env, validEnvironments)
}

// ValidateEnvironmentVariable validates environment variable names
func ValidateEnvironmentVariable(name string) *errors.ServoError {
	if name == "" {
		return errors.RequiredFieldError("environment variable name")
	}
	
	// Environment variables should be uppercase with underscores
	envVarRegex := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	if !envVarRegex.MatchString(name) {
		return errors.ValidationErrorf("environment variable name must be uppercase with underscores: %s", name)
	}
	
	return nil
}

// Platform validation

// ValidatePlatform validates platform names
func ValidatePlatform(platform string) *errors.ServoError {
	if platform == "" {
		return errors.RequiredFieldError("platform")
	}
	
	validPlatforms := []string{
		constants.PlatformDarwin,
		constants.PlatformLinux,
		constants.PlatformWindows,
	}
	
	for _, valid := range validPlatforms {
		if platform == valid {
			return nil
		}
	}
	
	return errors.ValidationErrorf("unsupported platform: %s, supported: %v", platform, validPlatforms)
}

// File extension validation

// ValidateFileExtension validates file extensions
func ValidateFileExtension(filename string, expectedExt string) *errors.ServoError {
	if !strings.HasSuffix(filename, expectedExt) {
		return errors.ValidationErrorf("file must have %s extension: %s", expectedExt, filename)
	}
	
	return nil
}

// Port validation

// ValidatePort validates port numbers
func ValidatePort(port int) *errors.ServoError {
	if port < 1 || port > 65535 {
		return errors.ValidationErrorf("port must be between 1 and 65535: %d", port)
	}
	
	return nil
}

// ValidatePortString validates port numbers from strings
func ValidatePortString(portStr string) *errors.ServoError {
	if portStr == "" {
		return errors.RequiredFieldError("port")
	}
	
	// Try to parse as integer
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return errors.ValidationErrorf("invalid port format: %s", portStr)
	}
	
	return ValidatePort(port)
}

// Composite validation functions

// ValidationContext holds context for validation operations
type ValidationContext struct {
	FieldPath   string
	AllowEmpty  bool
	CustomRules []func(interface{}) *errors.ServoError
}

// ValidateStruct validates a struct using reflection and field tags
// This is a placeholder for more advanced struct validation
func ValidateStruct(v interface{}, ctx *ValidationContext) *ValidationResult {
	result := &ValidationResult{IsValid: true}
	
	// TODO: Implement reflection-based struct validation using field tags
	// This would allow for validation rules like:
	// type Server struct {
	//     Name string `validate:"required,min=1,max=255,pattern=^[a-z][a-z0-9-]*$"`
	//     Port int    `validate:"required,min=1,max=65535"`
	// }
	
	return result
}