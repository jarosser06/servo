package errors

import (
	"fmt"
)

// Error categories for consistent error handling
const (
	// Input/Validation errors
	CategoryValidation = "validation"
	
	// File system errors
	CategoryFileSystem = "filesystem"
	
	// Network/External service errors
	CategoryNetwork = "network"
	
	// Configuration errors
	CategoryConfig = "config"
	
	// Session/State management errors
	CategorySession = "session"
	
	// Project management errors
	CategoryProject = "project"
	
	// MCP server errors
	CategoryMCP = "mcp"
	
	// Client errors
	CategoryClient = "client"
	
	// System/Runtime errors
	CategorySystem = "system"
)

// ServoError represents a structured error with category and context
type ServoError struct {
	Category string
	Code     string
	Message  string
	Cause    error
	Context  map[string]interface{}
}

func (e *ServoError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Category, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Category, e.Message)
}

func (e *ServoError) Unwrap() error {
	return e.Cause
}

func (e *ServoError) WithContext(key string, value interface{}) *ServoError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *ServoError) WithCode(code string) *ServoError {
	e.Code = code
	return e
}

// New creates a new ServoError
func New(category, message string) *ServoError {
	return &ServoError{
		Category: category,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with ServoError
func Wrap(err error, category, message string) *ServoError {
	return &ServoError{
		Category: category,
		Message:  message,
		Cause:    err,
		Context:  make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, category, format string, args ...interface{}) *ServoError {
	return &ServoError{
		Category: category,
		Message:  fmt.Sprintf(format, args...),
		Cause:    err,
		Context:  make(map[string]interface{}),
	}
}

// Newf creates a new ServoError with formatted message
func Newf(category, format string, args ...interface{}) *ServoError {
	return &ServoError{
		Category: category,
		Message:  fmt.Sprintf(format, args...),
		Context:  make(map[string]interface{}),
	}
}

// Is checks if an error is of a specific category
func Is(err error, category string) bool {
	if err == nil {
		return false
	}
	
	// Handle wrapped ServoError
	for err != nil {
		if servoErr, ok := err.(*ServoError); ok {
			return servoErr.Category == category
		}
		
		// Try unwrapping
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	
	return false
}

// HasCode checks if an error has a specific error code
func HasCode(err error, code string) bool {
	if err == nil {
		return false
	}
	
	// Handle wrapped ServoError
	for err != nil {
		if servoErr, ok := err.(*ServoError); ok {
			return servoErr.Code == code
		}
		
		// Try unwrapping
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	
	return false
}

// GetContext retrieves context from ServoError
func GetContext(err error, key string) (interface{}, bool) {
	if err == nil {
		return nil, false
	}
	
	// Handle wrapped ServoError
	for err != nil {
		if servoErr, ok := err.(*ServoError); ok {
			if servoErr.Context != nil {
				if value, exists := servoErr.Context[key]; exists {
					return value, true
				}
			}
		}
		
		// Try unwrapping
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	
	return nil, false
}