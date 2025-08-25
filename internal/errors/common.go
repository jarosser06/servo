package errors

// Common error codes
const (
	CodeNotFound         = "not_found"
	CodeAlreadyExists    = "already_exists"
	CodeInvalidInput     = "invalid_input"
	CodePermissionDenied = "permission_denied"
	CodeTimeout          = "timeout"
	CodeCorrupted        = "corrupted"
	CodeUnsupported      = "unsupported"
	CodeDependencyFailed = "dependency_failed"
	CodeConfigInvalid    = "config_invalid"
	CodeStateInconsistent = "state_inconsistent"
)

// Validation errors
func ValidationError(message string) *ServoError {
	return New(CategoryValidation, message).WithCode(CodeInvalidInput)
}

func ValidationErrorf(format string, args ...interface{}) *ServoError {
	return Newf(CategoryValidation, format, args...).WithCode(CodeInvalidInput)
}

func RequiredFieldError(field string) *ServoError {
	return New(CategoryValidation, "required field missing").
		WithCode(CodeInvalidInput).
		WithContext("field", field)
}

// File system errors
func FileNotFoundError(path string) *ServoError {
	return New(CategoryFileSystem, "file not found").
		WithCode(CodeNotFound).
		WithContext("path", path)
}

func FileReadError(path string, cause error) *ServoError {
	return Wrap(cause, CategoryFileSystem, "failed to read file").
		WithContext("path", path)
}

func FileWriteError(path string, cause error) *ServoError {
	return Wrap(cause, CategoryFileSystem, "failed to write file").
		WithContext("path", path)
}

func DirectoryCreateError(path string, cause error) *ServoError {
	return Wrap(cause, CategoryFileSystem, "failed to create directory").
		WithContext("path", path)
}

func PermissionError(path string, operation string) *ServoError {
	return New(CategoryFileSystem, "permission denied").
		WithCode(CodePermissionDenied).
		WithContext("path", path).
		WithContext("operation", operation)
}

// Session errors
func SessionNotFoundError(name string) *ServoError {
	return New(CategorySession, "session not found").
		WithCode(CodeNotFound).
		WithContext("session", name)
}

func SessionExistsError(name string) *ServoError {
	return New(CategorySession, "session already exists").
		WithCode(CodeAlreadyExists).
		WithContext("session", name)
}

func SessionCorruptedError(name string, cause error) *ServoError {
	return Wrap(cause, CategorySession, "session data corrupted").
		WithCode(CodeCorrupted).
		WithContext("session", name)
}

func InvalidSessionNameError(name string) *ServoError {
	return New(CategorySession, "invalid session name").
		WithCode(CodeInvalidInput).
		WithContext("session", name)
}

// Project errors
func ProjectNotFoundError() *ServoError {
	return New(CategoryProject, "not in a servo project directory").
		WithCode(CodeNotFound)
}

func ProjectConfigError(cause error) *ServoError {
	return Wrap(cause, CategoryProject, "project configuration error").
		WithCode(CodeConfigInvalid)
}

// MCP errors
func MCPParseError(path string, cause error) *ServoError {
	return Wrap(cause, CategoryMCP, "failed to parse MCP server definition").
		WithContext("path", path)
}

func MCPValidationError(path string, message string) *ServoError {
	return New(CategoryMCP, "MCP server validation failed").
		WithCode(CodeInvalidInput).
		WithContext("path", path).
		WithContext("validation_error", message)
}

func MCPServerExistsError(name string) *ServoError {
	return New(CategoryMCP, "MCP server already installed").
		WithCode(CodeAlreadyExists).
		WithContext("server", name)
}

func MCPServerNotFoundError(name string) *ServoError {
	return New(CategoryMCP, "MCP server not found").
		WithCode(CodeNotFound).
		WithContext("server", name)
}

// Client errors
func ClientNotSupportedError(client string) *ServoError {
	return New(CategoryClient, "client not supported").
		WithCode(CodeUnsupported).
		WithContext("client", client)
}

func ClientConfigError(client string, cause error) *ServoError {
	return Wrap(cause, CategoryClient, "failed to generate client configuration").
		WithContext("client", client)
}

// Configuration errors
func ConfigGenerationError(component string, cause error) *ServoError {
	return Wrap(cause, CategoryConfig, "failed to generate configuration").
		WithContext("component", component)
}

func ConfigValidationError(component string, message string) *ServoError {
	return New(CategoryConfig, "configuration validation failed").
		WithCode(CodeInvalidInput).
		WithContext("component", component).
		WithContext("validation_error", message)
}

// System errors
func DependencyError(dependency string, cause error) *ServoError {
	return Wrap(cause, CategorySystem, "dependency check failed").
		WithCode(CodeDependencyFailed).
		WithContext("dependency", dependency)
}

func TimeoutError(operation string) *ServoError {
	return New(CategorySystem, "operation timed out").
		WithCode(CodeTimeout).
		WithContext("operation", operation)
}

// Network errors
func NetworkError(operation string, cause error) *ServoError {
	return Wrap(cause, CategoryNetwork, "network operation failed").
		WithContext("operation", operation)
}

// State errors
func StateInconsistentError(message string) *ServoError {
	return New(CategorySystem, message).
		WithCode(CodeStateInconsistent)
}