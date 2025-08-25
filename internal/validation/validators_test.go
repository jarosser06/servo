package validation

import (
	"strings"
	"testing"

	"github.com/servo/servo/internal/constants"
	"github.com/servo/servo/internal/errors"
)

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{IsValid: true}
	
	// Test initial state
	if result.HasErrors() {
		t.Error("New ValidationResult should not have errors")
	}
	
	if result.Error() != nil {
		t.Error("New ValidationResult should not return error")
	}
	
	// Add an error
	err := errors.ValidationError("test error")
	result.Add(err)
	
	if !result.HasErrors() {
		t.Error("ValidationResult should have errors after adding one")
	}
	
	if result.IsValid {
		t.Error("ValidationResult should not be valid after adding error")
	}
	
	if result.Error() != err {
		t.Error("ValidationResult should return the added error")
	}
	
	// Add a warning
	result.AddWarning("test warning")
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantError bool
	}{
		{"valid name", "test-name", "name", false},
		{"empty name", "", "name", true},
		{"very long name", strings.Repeat("a", 300), "name", true},
		{"single character", "a", "name", false},
		{"unicode name", "测试", "name", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input, tt.fieldName)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSessionName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid session name", "test-session", false},
		{"valid with underscores", "test_session", false},
		{"valid alphanumeric", "session123", false},
		{"empty name", "", true},
		{"with spaces", "test session", true},
		{"with special chars", "test@session", true},
		{"reserved name", "sessions", true},
		{"reserved name config", "config", true},
		{"valid long name", "very-long-session-name-that-is-valid", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionName(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateServerName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid server name", "test-server", false},
		{"valid single char", "a", false},
		{"valid with numbers", "server123", false},
		{"empty name", "", true},
		{"uppercase", "Test-Server", true},
		{"underscore", "test_server", true},
		{"starting with number", "1test", true},
		{"ending with hyphen", "test-", true},
		{"special characters", "test@server", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerName(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid version", "1.0.0", false},
		{"valid short version", "1.0", false},
		{"valid with prerelease", "1.0.0-beta", false},
		{"valid with build", "1.0.0-alpha.1", false},
		{"empty version", "", true},
		{"invalid format", "1", true},
		{"invalid format v prefix", "v1.0.0", true},
		{"invalid format extra parts", "1.0.0.0", true},
		{"non-numeric", "abc.def", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateServoVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"supported version", constants.DefaultServoVersion, false},
		{"empty version", "", true},
		{"unsupported version", "2.0", true},
		{"invalid format", "1.0.0", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServoVersion(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		required  bool
		wantError bool
	}{
		{"valid description", "A test description", false, false},
		{"empty optional", "", false, false},
		{"empty required", "", true, true},
		{"very long description", strings.Repeat("a", 1500), false, true},
		{"unicode description", "测试描述", false, false},
		{"invalid UTF-8", string([]byte{0xff, 0xfe, 0xfd}), false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.input, tt.required)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		required  bool
		wantError bool
	}{
		{"valid https URL", "https://example.com", "homepage", false, false},
		{"valid http URL", "http://example.com", "homepage", false, false},
		{"empty optional", "", "homepage", false, false},
		{"empty required", "", "homepage", true, true},
		{"invalid scheme", "ftp://example.com", "homepage", false, true},
		{"invalid URL", "not-a-url", "homepage", false, true},
		{"relative URL", "/path/to/page", "homepage", false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input, tt.fieldName, tt.required)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRepositoryURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid GitHub HTTPS", "https://github.com/user/repo", false},
		{"valid GitHub SSH", "git@github.com:user/repo.git", false},
		{"valid GitLab", "https://gitlab.com/user/repo", false},
		{"valid Bitbucket", "https://bitbucket.org/user/repo", false},
		{"empty URL", "", true},
		{"invalid domain", "https://example.com/user/repo", true},
		{"invalid format", "not-a-repo-url", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepositoryURL(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateClient(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid VSCode", constants.ClientVSCode, false},
		{"valid Claude Code", constants.ClientClaudeCode, false},
		{"valid Cursor", constants.ClientCursor, false},
		{"empty client", "", true},
		{"unsupported client", "unsupported-client", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClient(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateClients(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantError bool
		warnings  int
	}{
		{"valid clients", []string{constants.ClientVSCode, constants.ClientClaudeCode}, false, 0},
		{"empty list", []string{}, false, 1}, // Should have warning
		{"duplicate clients", []string{constants.ClientVSCode, constants.ClientVSCode}, true, 0},
		{"invalid client", []string{"unsupported"}, true, 0},
		{"mixed valid/invalid", []string{constants.ClientVSCode, "unsupported"}, true, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateClients(tt.input)
			
			if tt.wantError && !result.HasErrors() {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && result.HasErrors() {
				t.Errorf("Unexpected error: %v", result.Error())
			}
			
			if len(result.Warnings) != tt.warnings {
				t.Errorf("Expected %d warnings, got %d", tt.warnings, len(result.Warnings))
			}
		})
	}
}

func TestValidateTransport(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid stdio", constants.TransportStdio, false},
		{"valid http", constants.TransportHTTP, false},
		{"valid sse", constants.TransportSSE, false},
		{"empty transport", "", true},
		{"unsupported transport", "websocket", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransport(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateInstallMethod(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid git", constants.InstallMethodGit, false},
		{"valid local", constants.InstallMethodLocal, false},
		{"valid docker", constants.InstallMethodDocker, false},
		{"empty method", "", true},
		{"unsupported method", "pip", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstallMethod(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		required  bool
		wantError bool
	}{
		{"valid path", "/path/to/file", "path", false, false},
		{"empty optional", "", "path", false, false},
		{"empty required", "", "path", true, true},
		{"path with ..", "/path/../other", "path", false, true},
		{"path with double slash", "/path//file", "path", false, true},
		{"windows backslash", "C:\\path\\file", "path", false, true},
		{"unicode path", "/路径/文件", "path", false, false},
		{"invalid UTF-8", string([]byte{'/', 0xff, 0xfe}), "path", false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.input, tt.fieldName, tt.required)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEnvironmentName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid development", constants.EnvDevelopment, false},
		{"valid staging", constants.EnvStaging, false},
		{"valid production", constants.EnvProduction, false},
		{"valid testing", constants.EnvTesting, false},
		{"empty environment", "", true},
		{"unsupported environment", "custom", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironmentName(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid env var", "API_KEY", false},
		{"valid single char", "X", false},
		{"valid with numbers", "VAR123", false},
		{"empty name", "", true},
		{"lowercase", "api_key", true},
		{"with spaces", "API KEY", true},
		{"with hyphens", "API-KEY", true},
		{"starting with number", "123VAR", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironmentVariable(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid darwin", constants.PlatformDarwin, false},
		{"valid linux", constants.PlatformLinux, false},
		{"valid windows", constants.PlatformWindows, false},
		{"empty platform", "", true},
		{"unsupported platform", "freebsd", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlatform(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		ext       string
		wantError bool
	}{
		{"valid servo file", "server.servo", constants.ExtServo, false},
		{"valid json file", "config.json", constants.ExtJSON, false},
		{"invalid extension", "file.txt", constants.ExtServo, true},
		{"no extension", "file", constants.ExtJSON, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.filename, tt.ext)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		wantError bool
	}{
		{"valid port 80", 80, false},
		{"valid port 443", 443, false},
		{"valid port 8080", 8080, false},
		{"valid port 65535", 65535, false},
		{"invalid port 0", 0, true},
		{"invalid port -1", -1, true},
		{"invalid port 65536", 65536, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePortString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid port string", "8080", false},
		{"valid port 1", "1", false},
		{"valid port 65535", "65535", false},
		{"empty port", "", true},
		{"invalid format", "abc", true},
		{"invalid port 0", "0", true},
		{"invalid port negative", "-1", true},
		{"invalid port too high", "65536", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortString(tt.input)
			
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}