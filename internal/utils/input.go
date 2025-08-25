package utils

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// PromptForPassword prompts the user for a password, respecting the global --no-interactive flag
// If SERVO_NON_INTERACTIVE is set, it will not prompt and instead return an error if no env var is available
func PromptForPassword(envVarName, promptMessage string) (string, error) {
	// Check for environment variable first
	if password := os.Getenv(envVarName); password != "" {
		return password, nil
	}

	// Check if in non-interactive mode
	if os.Getenv("SERVO_NON_INTERACTIVE") != "" {
		return "", fmt.Errorf("%s environment variable required in non-interactive mode", envVarName)
	}

	// Prompt user for password
	fmt.Print(promptMessage)
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	return string(password), nil
}

// PromptForInput prompts the user for general input, respecting the global --no-interactive flag
// If SERVO_NON_INTERACTIVE is set, it will return the defaultValue if provided, or error if not
func PromptForInput(promptMessage string, defaultValue *string) (string, error) {
	// Check if in non-interactive mode
	if os.Getenv("SERVO_NON_INTERACTIVE") != "" {
		if defaultValue != nil {
			return *defaultValue, nil
		}
		return "", fmt.Errorf("user input required but running in non-interactive mode")
	}

	// Prompt user for input
	fmt.Print(promptMessage)
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Use default if input is empty and default is provided
	if input == "" && defaultValue != nil {
		return *defaultValue, nil
	}

	return input, nil
}

// PromptForConfirmation prompts the user for yes/no confirmation, respecting the global --no-interactive flag
// If SERVO_NON_INTERACTIVE is set, it will return the defaultValue
func PromptForConfirmation(promptMessage string, defaultValue bool) (bool, error) {
	// Check if in non-interactive mode
	if os.Getenv("SERVO_NON_INTERACTIVE") != "" {
		return defaultValue, nil
	}

	// Prompt user for confirmation
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", promptMessage, defaultStr)
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return defaultValue, nil // Return default on scan error (like just pressing enter)
	}

	switch input {
	case "y", "Y", "yes", "Yes", "YES":
		return true, nil
	case "n", "N", "no", "No", "NO":
		return false, nil
	default:
		return defaultValue, nil
	}
}
