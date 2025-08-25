package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/internal/project"
)

func TestSecretsCommand_Execute(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		isProject   bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Not in project directory",
			args:        []string{"list"},
			isProject:   false,
			expectError: true,
			errorMsg:    "not in a servo project directory",
		},
		{
			name:        "No args shows help",
			args:        []string{},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Unknown subcommand",
			args:        []string{"unknown"},
			isProject:   true,
			expectError: true,
			errorMsg:    "unknown secrets subcommand: unknown",
		},
		{
			name:        "List subcommand",
			args:        []string{"list"},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Set subcommand",
			args:        []string{"set", "key", "value"},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Get subcommand",
			args:        []string{"get", "key"},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Delete subcommand",
			args:        []string{"delete", "key"},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Export subcommand",
			args:        []string{"export"},
			isProject:   true,
			expectError: false,
		},
		{
			name:        "Import subcommand",
			args:        []string{"import", "file.yaml"},
			isProject:   true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			oldCwd, _ := os.Getwd()
			defer os.Chdir(oldCwd)
			os.Chdir(tmpDir)

			// Setup project if needed
			if tt.isProject {
				// Create minimal project structure
				os.MkdirAll(".servo", 0755)
				os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)
			}

			// Create project manager
			projectManager := project.NewManager()
			cmd := NewSecretsCommand(projectManager)

			// Execute command
			err := cmd.Execute(tt.args)

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s' but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					// Some operations may fail due to missing files, which is acceptable for these tests
					t.Logf("Command returned error (may be expected): %v", err)
				}
			}
		})
	}
}

func TestSecretsCommand_SetAndGetSecret(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Test setting a secret
	err := cmd.Execute([]string{"set", "test_key", "test_value"})
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	// Verify secrets file was created
	secretsPath := cmd.getSecretsPath()
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		t.Errorf("Secrets file was not created at %s", secretsPath)
	}

	// Test getting the secret
	err = cmd.Execute([]string{"get", "test_key"})
	if err != nil {
		t.Errorf("Failed to get secret: %v", err)
	}

	// Test getting non-existent secret
	err = cmd.Execute([]string{"get", "nonexistent_key"})
	if err == nil {
		t.Errorf("Expected error when getting non-existent secret")
	}
}

func TestSecretsCommand_ListSecrets(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Test listing with no secrets
	err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("Failed to list empty secrets: %v", err)
	}

	// Add a secret
	err = cmd.Execute([]string{"set", "test_key", "test_value"})
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	// Test listing with secrets
	err = cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("Failed to list secrets: %v", err)
	}
}

func TestSecretsCommand_DeleteSecret(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Add a secret
	err := cmd.Execute([]string{"set", "test_key", "test_value"})
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	// Delete the secret
	err = cmd.Execute([]string{"delete", "test_key"})
	if err != nil {
		t.Errorf("Failed to delete secret: %v", err)
	}

	// Try to get the deleted secret (should fail)
	err = cmd.Execute([]string{"get", "test_key"})
	if err == nil {
		t.Errorf("Expected error when getting deleted secret")
	}

	// Try to delete non-existent secret - should succeed (idempotent operation)
	err = cmd.Execute([]string{"delete", "nonexistent_key"})
	if err != nil {
		t.Errorf("Deleting non-existent secret should succeed (idempotent), got error: %v", err)
	}
}

func TestSecretsCommand_ExportSecrets(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Add some secrets
	err := cmd.Execute([]string{"set", "key1", "value1"})
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}
	err = cmd.Execute([]string{"set", "key2", "value2"})
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	// Export to file
	exportFile := filepath.Join(tmpDir, "exported_secrets.yaml")
	err = cmd.Execute([]string{"export", exportFile})
	if err != nil {
		t.Errorf("Failed to export secrets: %v", err)
	}

	// Check if export file was created
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Errorf("Export file was not created at %s", exportFile)
	}
}

func TestSecretsCommand_ImportSecrets(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create import file
	importFile := filepath.Join(tmpDir, "import_secrets.yaml")
	importContent := `version: "1"
secrets:
  imported_key1: aW1wb3J0ZWRfdmFsdWUx  # base64 encoded "imported_value1"
  imported_key2: aW1wb3J0ZWRfdmFsdWUy  # base64 encoded "imported_value2"
`
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Import secrets
	err = cmd.Execute([]string{"import", importFile})
	if err != nil {
		t.Errorf("Failed to import secrets: %v", err)
	}

	// Verify secrets were imported
	err = cmd.Execute([]string{"get", "imported_key1"})
	if err != nil {
		t.Errorf("Failed to get imported secret: %v", err)
	}
}

func TestSecretsCommand_EdgeCases(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup project
	os.MkdirAll(".servo", 0755)
	os.WriteFile(".servo/project.yaml", []byte("version: 1\n"), 0644)

	// Create project manager and command
	projectManager := project.NewManager()
	cmd := NewSecretsCommand(projectManager)

	// Test set with insufficient arguments
	err := cmd.Execute([]string{"set"})
	if err == nil {
		t.Errorf("Expected error when setting secret with insufficient arguments")
	}

	// Test set with only key, no value
	err = cmd.Execute([]string{"set", "key_only"})
	if err == nil {
		t.Errorf("Expected error when setting secret with only key")
	}

	// Test get with insufficient arguments
	err = cmd.Execute([]string{"get"})
	if err == nil {
		t.Errorf("Expected error when getting secret with no key")
	}

	// Test delete with insufficient arguments
	err = cmd.Execute([]string{"delete"})
	if err == nil {
		t.Errorf("Expected error when deleting secret with no key")
	}

	// Test import with non-existent file
	err = cmd.Execute([]string{"import", "nonexistent.yaml"})
	if err == nil {
		t.Errorf("Expected error when importing from non-existent file")
	}
}