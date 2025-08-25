package errors

import (
	"fmt"
	"os"
)

// This file demonstrates how to migrate existing error handling to use standardized patterns

// Example: Before migration (old pattern)
func createSessionOld(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}
	
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("session '%s' already exists", name)
	}
	
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}
	
	return nil
}

// Example: After migration (new pattern) 
func createSessionNew(name string) error {
	if name == "" {
		return InvalidSessionNameError(name)
	}
	
	if _, err := os.Stat(name); err == nil {
		return SessionExistsError(name)
	}
	
	if err := os.MkdirAll(name, 0755); err != nil {
		return DirectoryCreateError(name, err)
	}
	
	return nil
}

// Example: Error handling in calling code

// Before migration - less structured error handling
func handleSessionCreationOld(name string) {
	if err := createSessionOld(name); err != nil {
		// Hard to distinguish between different error types
		// Must parse error strings to determine type
		fmt.Printf("Error: %v\n", err)
	}
}

// After migration - structured error handling
func handleSessionCreationNew(name string) {
	if err := createSessionNew(name); err != nil {
		// Can check error categories and codes programmatically
		if Is(err, CategorySession) {
			if HasCode(err, CodeAlreadyExists) {
				fmt.Println("Session already exists, continuing...")
				return
			}
			if HasCode(err, CodeInvalidInput) {
				fmt.Println("Invalid session name provided")
				return
			}
		}
		
		if Is(err, CategoryFileSystem) {
			fmt.Printf("File system error: %v\n", err)
			return
		}
		
		// Can extract context information
		if sessionName, exists := GetContext(err, "session"); exists {
			fmt.Printf("Error with session '%s': %v\n", sessionName, err)
		}
		
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Example of how to gradually migrate large codebases

// Phase 1: Create wrapper functions that return standardized errors
func wrapSessionError(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	// Convert common error patterns
	errStr := err.Error()
	
	if errStr == "session name cannot be empty" {
		return InvalidSessionNameError("")
	}
	
	if contains(errStr, "already exists") {
		// Extract session name from error message if possible
		return SessionExistsError("unknown")
	}
	
	if contains(errStr, "failed to create") {
		return DirectoryCreateError("unknown", err)
	}
	
	// Default: wrap as generic session error
	return Wrap(err, CategorySession, operation)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || (len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// Phase 2: Update calling code to use error checking helpers
func isRetryableError(err error) bool {
	// Network errors are typically retryable
	if Is(err, CategoryNetwork) {
		return true
	}
	
	// Temporary file system issues might be retryable
	if Is(err, CategoryFileSystem) && HasCode(err, CodePermissionDenied) {
		return false // Don't retry permission errors
	}
	
	if Is(err, CategoryFileSystem) {
		return true // Other filesystem errors might be temporary
	}
	
	return false
}

func isUserError(err error) bool {
	// Validation errors are user errors
	if Is(err, CategoryValidation) {
		return true
	}
	
	// Input-related errors
	if HasCode(err, CodeInvalidInput) {
		return true
	}
	
	return false
}

// Phase 3: Create domain-specific error checking functions
func IsSessionError(err error) bool {
	return Is(err, CategorySession)
}

func IsSessionNotFound(err error) bool {
	return Is(err, CategorySession) && HasCode(err, CodeNotFound)
}

func IsProjectNotInDirectory(err error) bool {
	return Is(err, CategoryProject) && HasCode(err, CodeNotFound)
}

// Example of creating error context for debugging
func enrichErrorContext(err error, operation string) error {
	if servoErr, ok := err.(*ServoError); ok {
		return servoErr.WithContext("operation", operation).
					  WithContext("timestamp", "2024-01-01T00:00:00Z")
	}
	
	// For non-ServoError, create new one
	return Wrap(err, CategorySystem, operation).
		   WithContext("timestamp", "2024-01-01T00:00:00Z")
}