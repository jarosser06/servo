# Servo Error Handling Standards

This package provides a standardized error handling system for the Servo project. It offers structured errors with categories, codes, and contextual information that enables better error handling and debugging.

## Quick Start

### Creating Errors

```go
// Use predefined error functions
err := ValidationError("invalid input provided")
err := SessionNotFoundError("my-session")
err := FileReadError("/path/to/file", cause)

// Or create custom errors
err := New(CategoryValidation, "custom validation message")
err := Wrap(cause, CategoryFileSystem, "file operation failed")
```

### Checking Errors

```go
if errors.Is(err, CategorySession) {
    // Handle session-related errors
}

if errors.HasCode(err, CodeNotFound) {
    // Handle not-found errors specifically
}

// Get context information
if sessionName, exists := errors.GetContext(err, "session"); exists {
    log.Printf("Error with session: %s", sessionName)
}
```

## Error Categories

| Category | Description | Common Use Cases |
|----------|-------------|------------------|
| `CategoryValidation` | Input validation errors | Invalid parameters, missing required fields |
| `CategoryFileSystem` | File system operations | File not found, permission denied, disk full |
| `CategoryNetwork` | Network operations | Connection failed, timeout, DNS resolution |
| `CategoryConfig` | Configuration errors | Invalid config, missing settings |
| `CategorySession` | Session management | Session not found, already exists, corrupted |
| `CategoryProject` | Project operations | Not in project directory, invalid project |
| `CategoryMCP` | MCP server operations | Parse errors, validation, installation |
| `CategoryClient` | Client operations | Unsupported client, config generation |
| `CategorySystem` | System-level errors | Dependencies, timeouts, state issues |

## Error Codes

| Code | Description | When to Use |
|------|-------------|-------------|
| `CodeNotFound` | Resource not found | Files, sessions, projects don't exist |
| `CodeAlreadyExists` | Resource already exists | Duplicate sessions, servers |
| `CodeInvalidInput` | Invalid input provided | Validation failures |
| `CodePermissionDenied` | Permission denied | File access, restricted operations |
| `CodeTimeout` | Operation timed out | Network operations, long processes |
| `CodeCorrupted` | Data is corrupted | Malformed files, invalid state |
| `CodeUnsupported` | Feature not supported | Unknown clients, missing features |
| `CodeDependencyFailed` | Dependency check failed | Missing requirements |
| `CodeConfigInvalid` | Configuration is invalid | Malformed config files |
| `CodeStateInconsistent` | Inconsistent state | Concurrent modification issues |

## Migration Guide

### Phase 1: Replace Error Creation

**Before:**
```go
if name == "" {
    return fmt.Errorf("session name cannot be empty")
}

if err := os.MkdirAll(path, 0755); err != nil {
    return fmt.Errorf("failed to create directory: %w", err)
}
```

**After:**
```go
if name == "" {
    return InvalidSessionNameError(name)
}

if err := os.MkdirAll(path, 0755); err != nil {
    return DirectoryCreateError(path, err)
}
```

### Phase 2: Update Error Handling

**Before:**
```go
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // handle not found
    } else if strings.Contains(err.Error(), "already exists") {
        // handle duplicate
    }
    return err
}
```

**After:**
```go
if err != nil {
    if errors.HasCode(err, CodeNotFound) {
        // handle not found
    } else if errors.HasCode(err, CodeAlreadyExists) {
        // handle duplicate
    }
    return err
}
```

### Phase 3: Add Context and Structured Handling

```go
// Add context for debugging
err := ValidationError("invalid name format").
       WithContext("provided_name", name).
       WithContext("expected_format", "alphanumeric with dashes")

// Structured error handling
func handleError(err error) {
    switch {
    case errors.Is(err, CategoryValidation):
        handleUserError(err)
    case errors.Is(err, CategoryFileSystem):
        handleSystemError(err)
    case errors.Is(err, CategoryNetwork):
        if errors.HasCode(err, CodeTimeout) {
            retryOperation()
        } else {
            handleNetworkError(err)
        }
    default:
        handleUnexpectedError(err)
    }
}
```

## Best Practices

### 1. Use Appropriate Categories

Choose the most specific category that describes the root cause:
- Use `CategoryValidation` for user input errors
- Use `CategoryFileSystem` for file operations
- Use `CategorySession` for session management errors

### 2. Add Meaningful Context

```go
err := FileReadError(path, cause).
       WithContext("operation", "loading config").
       WithContext("retry_count", attempts)
```

### 3. Preserve Error Chains

Always use `Wrap` or `Wrapf` to preserve the original error:
```go
if err := someOperation(); err != nil {
    return Wrap(err, CategoryProject, "project initialization failed")
}
```

### 4. Use Codes for Programmatic Handling

```go
// Good: Enables specific handling
if errors.HasCode(err, CodeAlreadyExists) {
    return updateExistingResource()
}

// Bad: Fragile string matching
if strings.Contains(err.Error(), "already exists") {
    return updateExistingResource()
}
```

### 5. Check Categories Before Codes

```go
// Good: Check category first, then code
if errors.Is(err, CategorySession) && errors.HasCode(err, CodeNotFound) {
    return createNewSession()
}

// Also good: Use helper functions
if IsSessionNotFound(err) {
    return createNewSession()
}
```

## Testing Error Handling

```go
func TestErrorHandling(t *testing.T) {
    err := CreateSession("")
    
    // Test error category
    if !errors.Is(err, CategorySession) {
        t.Error("Expected session error")
    }
    
    // Test error code
    if !errors.HasCode(err, CodeInvalidInput) {
        t.Error("Expected invalid input code")
    }
    
    // Test context
    if session, exists := errors.GetContext(err, "session"); !exists || session != "" {
        t.Error("Expected empty session name in context")
    }
}
```

## Logging and Monitoring

```go
func logError(err error, operation string) {
    logger := log.WithFields(log.Fields{
        "operation": operation,
    })
    
    if servoErr, ok := err.(*errors.ServoError); ok {
        logger = logger.WithFields(log.Fields{
            "category": servoErr.Category,
            "code":     servoErr.Code,
            "context":  servoErr.Context,
        })
    }
    
    logger.Error(err.Error())
}
```

## Common Patterns

### Retry Logic
```go
func withRetry(operation func() error) error {
    for attempt := 1; attempt <= 3; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Don't retry user errors
        if errors.Is(err, CategoryValidation) || 
           errors.HasCode(err, CodeInvalidInput) {
            return err
        }
        
        // Don't retry fatal system errors
        if errors.HasCode(err, CodePermissionDenied) {
            return err
        }
        
        time.Sleep(time.Duration(attempt) * time.Second)
    }
    
    return TimeoutError("max retries exceeded")
}
```

### Error Aggregation
```go
type ErrorCollector struct {
    errors []error
}

func (ec *ErrorCollector) Add(err error) {
    if err != nil {
        ec.errors = append(ec.errors, err)
    }
}

func (ec *ErrorCollector) Error() error {
    if len(ec.errors) == 0 {
        return nil
    }
    
    if len(ec.errors) == 1 {
        return ec.errors[0]
    }
    
    // Create aggregate error
    messages := make([]string, len(ec.errors))
    for i, err := range ec.errors {
        messages[i] = err.Error()
    }
    
    return New(CategorySystem, "multiple errors occurred").
           WithContext("errors", ec.errors).
           WithContext("count", len(ec.errors))
}
```