package errors

import (
	"fmt"
	"testing"
)

func TestServoError_Basic(t *testing.T) {
	err := New(CategoryValidation, "test error")
	
	if err.Category != CategoryValidation {
		t.Errorf("expected category %s, got %s", CategoryValidation, err.Category)
	}
	
	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", err.Message)
	}
	
	expected := "validation: test error"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestServoError_WithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := Wrap(cause, CategoryFileSystem, "wrapper error")
	
	if err.Cause != cause {
		t.Errorf("expected cause to be preserved")
	}
	
	expected := "filesystem: wrapper error: underlying error"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
	
	if err.Unwrap() != cause {
		t.Errorf("expected Unwrap to return cause")
	}
}

func TestServoError_WithContext(t *testing.T) {
	err := New(CategorySession, "test error").
		WithContext("session", "test-session").
		WithContext("operation", "create")
	
	if err.Context["session"] != "test-session" {
		t.Errorf("expected session context to be 'test-session'")
	}
	
	if err.Context["operation"] != "create" {
		t.Errorf("expected operation context to be 'create'")
	}
}

func TestServoError_WithCode(t *testing.T) {
	err := New(CategoryValidation, "test error").WithCode(CodeInvalidInput)
	
	if err.Code != CodeInvalidInput {
		t.Errorf("expected code %s, got %s", CodeInvalidInput, err.Code)
	}
}

func TestIs(t *testing.T) {
	err := New(CategoryValidation, "test error")
	
	if !Is(err, CategoryValidation) {
		t.Errorf("expected Is to return true for matching category")
	}
	
	if Is(err, CategoryFileSystem) {
		t.Errorf("expected Is to return false for non-matching category")
	}
	
	// Test with wrapped error
	wrappedErr := fmt.Errorf("outer: %w", err)
	if !Is(wrappedErr, CategoryValidation) {
		t.Errorf("expected Is to work with wrapped errors")
	}
}

func TestHasCode(t *testing.T) {
	err := New(CategoryValidation, "test error").WithCode(CodeInvalidInput)
	
	if !HasCode(err, CodeInvalidInput) {
		t.Errorf("expected HasCode to return true for matching code")
	}
	
	if HasCode(err, CodeNotFound) {
		t.Errorf("expected HasCode to return false for non-matching code")
	}
}

func TestGetContext(t *testing.T) {
	err := New(CategorySession, "test error").WithContext("session", "test-session")
	
	value, exists := GetContext(err, "session")
	if !exists {
		t.Errorf("expected context key to exist")
	}
	
	if value != "test-session" {
		t.Errorf("expected context value 'test-session', got %v", value)
	}
	
	_, exists = GetContext(err, "nonexistent")
	if exists {
		t.Errorf("expected nonexistent context key to return false")
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *ServoError
		category string
		code     string
	}{
		{
			name:     "ValidationError",
			err:      ValidationError("invalid input"),
			category: CategoryValidation,
			code:     CodeInvalidInput,
		},
		{
			name:     "SessionNotFoundError",
			err:      SessionNotFoundError("test-session"),
			category: CategorySession,
			code:     CodeNotFound,
		},
		{
			name:     "FileNotFoundError",
			err:      FileNotFoundError("/path/to/file"),
			category: CategoryFileSystem,
			code:     CodeNotFound,
		},
		{
			name:     "MCPServerExistsError",
			err:      MCPServerExistsError("test-server"),
			category: CategoryMCP,
			code:     CodeAlreadyExists,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Category != tt.category {
				t.Errorf("expected category %s, got %s", tt.category, tt.err.Category)
			}
			
			if tt.err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, tt.err.Code)
			}
		})
	}
}

func TestRequiredFieldError(t *testing.T) {
	err := RequiredFieldError("name")
	
	if err.Category != CategoryValidation {
		t.Errorf("expected validation category")
	}
	
	if err.Code != CodeInvalidInput {
		t.Errorf("expected invalid_input code")
	}
	
	field, exists := GetContext(err, "field")
	if !exists || field != "name" {
		t.Errorf("expected field context to be 'name'")
	}
}

func TestChainedErrors(t *testing.T) {
	// Create a chain of errors
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, CategoryFileSystem, "file operation failed")
	chainedErr := Wrap(wrappedErr, CategoryProject, "project operation failed")
	
	// Test that we can still identify the categories in the chain
	if !Is(chainedErr, CategoryProject) {
		t.Errorf("expected top-level category to be project")
	}
	
	// Test unwrapping
	if chainedErr.Unwrap() != wrappedErr {
		t.Errorf("expected first unwrap to return wrapped error")
	}
	
	if wrappedErr.Unwrap() != originalErr {
		t.Errorf("expected second unwrap to return original error")
	}
}