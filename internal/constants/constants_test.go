package constants

import (
	"regexp"
	"strings"
	"testing"
)

func TestFileExtensions(t *testing.T) {
	extensions := []string{ExtServo, ExtYAML, ExtYML, ExtJSON, ExtMarkdown}
	
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			t.Errorf("Extension %s should start with '.'", ext)
		}
		if len(ext) < 2 {
			t.Errorf("Extension %s is too short", ext)
		}
	}
}

func TestDirectoryNames(t *testing.T) {
	dirs := []string{DirServo, DirSessions, DirManifests, DirConfig, DirVolumes, DirLogs}
	
	for _, dir := range dirs {
		if strings.Contains(dir, "/") || strings.Contains(dir, "\\") {
			t.Errorf("Directory name %s should not contain path separators", dir)
		}
		if len(dir) == 0 {
			t.Errorf("Directory name should not be empty")
		}
	}
}

func TestClientNames(t *testing.T) {
	clients := []string{ClientVSCode, ClientClaudeCode, ClientCursor, ClientClaudeApp}
	
	for _, client := range clients {
		if len(client) == 0 {
			t.Errorf("Client name should not be empty")
		}
		if strings.Contains(client, " ") {
			t.Errorf("Client name %s should not contain spaces", client)
		}
	}
	
	// Test that SupportedClients contains all client constants
	expectedClients := map[string]bool{
		ClientVSCode:     false,
		ClientClaudeCode: false,
		ClientCursor:     false,
		ClientClaudeApp:  false,
	}
	
	for _, client := range SupportedClients {
		if _, exists := expectedClients[client]; exists {
			expectedClients[client] = true
		} else {
			t.Errorf("Unknown client in SupportedClients: %s", client)
		}
	}
	
	for client, found := range expectedClients {
		if !found {
			t.Errorf("Client %s not found in SupportedClients", client)
		}
	}
}

func TestTransportTypes(t *testing.T) {
	transports := []string{TransportStdio, TransportHTTP, TransportSSE}
	
	for _, transport := range transports {
		if len(transport) == 0 {
			t.Errorf("Transport type should not be empty")
		}
	}
	
	if DefaultTransport != TransportStdio {
		t.Errorf("Expected default transport to be %s, got %s", TransportStdio, DefaultTransport)
	}
}

func TestInstallMethods(t *testing.T) {
	methods := []string{InstallMethodGit, InstallMethodLocal, InstallMethodDocker, InstallMethodURL}
	types := []string{InstallTypeGit, InstallTypeLocal, InstallTypeDocker}
	
	for _, method := range methods {
		if len(method) == 0 {
			t.Errorf("Install method should not be empty")
		}
	}
	
	for _, installType := range types {
		if len(installType) == 0 {
			t.Errorf("Install type should not be empty")
		}
	}
	
	if DefaultInstallMethod != InstallMethodGit {
		t.Errorf("Expected default install method to be %s, got %s", InstallMethodGit, DefaultInstallMethod)
	}
}

func TestEnvironmentNames(t *testing.T) {
	envs := []string{EnvDevelopment, EnvStaging, EnvProduction, EnvTesting}
	
	for _, env := range envs {
		if len(env) == 0 {
			t.Errorf("Environment name should not be empty")
		}
		if strings.Contains(env, " ") {
			t.Errorf("Environment name %s should not contain spaces", env)
		}
	}
}

func TestDefaultValues(t *testing.T) {
	if DefaultSessionName == "" {
		t.Error("DefaultSessionName should not be empty")
	}
	
	if DefaultServoVersion == "" {
		t.Error("DefaultServoVersion should not be empty")
	}
	
	// Validate version format
	versionPattern := regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	if !versionPattern.MatchString(DefaultServoVersion) {
		t.Errorf("DefaultServoVersion %s should match version pattern", DefaultServoVersion)
	}
}

func TestYAMLFieldNames(t *testing.T) {
	fields := []string{
		FieldName, FieldVersion, FieldDescription, FieldAuthor, FieldLicense,
		FieldServoVersion, FieldInstall, FieldServer, FieldClients,
		FieldRequirements, FieldRuntimes, FieldServices, FieldDependencies,
	}
	
	for _, field := range fields {
		if len(field) == 0 {
			t.Error("YAML field name should not be empty")
		}
		if strings.Contains(field, " ") {
			t.Errorf("YAML field name %s should not contain spaces", field)
		}
	}
}

func TestRegexPatterns(t *testing.T) {
	patterns := map[string]string{
		"session name": PatternSessionName,
		"version":      PatternVersion,
		"api key":      PatternAPIKey,
	}
	
	for name, pattern := range patterns {
		_, err := regexp.Compile(pattern)
		if err != nil {
			t.Errorf("Invalid regex pattern for %s: %v", name, err)
		}
	}
	
	// Test pattern matching
	sessionNameRegex := regexp.MustCompile(PatternSessionName)
	validNames := []string{"test", "test-session", "test_session", "session123"}
	invalidNames := []string{"test session", "test.session", "test@session", ""}
	
	for _, name := range validNames {
		if !sessionNameRegex.MatchString(name) {
			t.Errorf("Valid session name %s should match pattern", name)
		}
	}
	
	for _, name := range invalidNames {
		if sessionNameRegex.MatchString(name) {
			t.Errorf("Invalid session name %s should not match pattern", name)
		}
	}
	
	versionRegex := regexp.MustCompile(PatternVersion)
	validVersions := []string{"1.0", "1.0.0", "2.1.3", "1.0.0-beta"}
	invalidVersions := []string{"1", "v1.0", "1.0.0.0", ""}
	
	for _, version := range validVersions {
		if !versionRegex.MatchString(version) {
			t.Errorf("Valid version %s should match pattern", version)
		}
	}
	
	for _, version := range invalidVersions {
		if versionRegex.MatchString(version) {
			t.Errorf("Invalid version %s should not match pattern", version)
		}
	}
}

func TestSecretConstants(t *testing.T) {
	if !strings.Contains(SecretPrefix, "SERVO_SECRET") {
		t.Error("SecretPrefix should contain 'SERVO_SECRET'")
	}
	
	if !strings.Contains(SecretPlaceholder, "%"+"s") {
		t.Error("SecretPlaceholder should contain placeholder for formatting")
	}
}

func TestPlatformConstants(t *testing.T) {
	platforms := []string{PlatformDarwin, PlatformLinux, PlatformWindows}
	
	for _, platform := range platforms {
		if len(platform) == 0 {
			t.Errorf("Platform name should not be empty")
		}
	}
}

func TestFilePermissions(t *testing.T) {
	if FilePermReadWrite == 0 {
		t.Error("FilePermReadWrite should not be zero")
	}
	
	if FilePermExecute == 0 {
		t.Error("FilePermExecute should not be zero")
	}
	
	if DirPermDefault == 0 {
		t.Error("DirPermDefault should not be zero")
	}
}

func TestEnvironmentVariableNames(t *testing.T) {
	envVars := []string{
		EnvServoDir, EnvServoMasterPassword, EnvServoNonInteractive,
		EnvNodeEnv, EnvAPIKey, EnvDatabaseURL, EnvLogLevel,
	}
	
	for _, envVar := range envVars {
		if len(envVar) == 0 {
			t.Errorf("Environment variable name should not be empty")
		}
		if strings.ToUpper(envVar) != envVar {
			t.Errorf("Environment variable %s should be uppercase", envVar)
		}
	}
}

func TestTimeoutsAndLimits(t *testing.T) {
	if DefaultTimeout <= 0 {
		t.Error("DefaultTimeout should be positive")
	}
	
	if MaxRetries <= 0 {
		t.Error("MaxRetries should be positive")
	}
	
	if RetryDelay <= 0 {
		t.Error("RetryDelay should be positive")
	}
	
	if MaxFileSize <= 0 {
		t.Error("MaxFileSize should be positive")
	}
	
	if MaxManifestSize <= 0 {
		t.Error("MaxManifestSize should be positive")
	}
	
	if MaxManifestSize > MaxFileSize {
		t.Error("MaxManifestSize should not be larger than MaxFileSize")
	}
}

func TestLogLevels(t *testing.T) {
	levels := []string{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
	
	for _, level := range levels {
		if len(level) == 0 {
			t.Errorf("Log level should not be empty")
		}
		if strings.ToUpper(level) == level {
			t.Errorf("Log level %s should be lowercase", level)
		}
	}
}

func TestBuildInformation(t *testing.T) {
	// These are variables that can be set at build time
	// Test that they have default values
	if Version == "" {
		t.Error("Version should have a default value")
	}
	
	if GitCommit == "" {
		t.Error("GitCommit should have a default value")
	}
	
	if BuildDate == "" {
		t.Error("BuildDate should have a default value")
	}
}