package cli

import (
	"testing"

	"github.com/urfave/cli/v2"
)

func TestApp_Creation(t *testing.T) {
	// Create a test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Verify app properties
	if app.Name != "servo" {
		t.Errorf("Expected app name 'servo', got '%s'", app.Name)
	}

	if app.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", app.Version)
	}

	// Test that required commands are present
	requiredCommands := []string{"init", "install", "status", "work", "validate", "clients", "config", "secrets"}
	for _, cmdName := range requiredCommands {
		found := false
		for _, cmd := range app.Commands {
			if cmd.Name == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required command '%s' not found", cmdName)
		}
	}
}

func TestApp_CommandStructure(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Verify app properties without actually running commands (urfave/cli calls os.Exit)
	if app.Name != "servo" {
		t.Errorf("Expected app name 'servo', got '%s'", app.Name)
	}

	if app.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", app.Version)
	}

	// Test that each command has proper actions
	for _, cmd := range app.Commands {
		if cmd.Action == nil && len(cmd.Subcommands) == 0 {
			t.Errorf("Command '%s' missing action and subcommands", cmd.Name)
		}

		// Verify flags are properly structured
		for _, flag := range cmd.Flags {
			if flag == nil {
				t.Errorf("Command '%s' has nil flag", cmd.Name)
			}
		}
	}
}

func TestApp_CommandValidation(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Verify that each command has proper structure
	for _, cmd := range app.Commands {
		if cmd.Name == "" {
			t.Error("Command missing name")
		}
		if cmd.Usage == "" {
			t.Errorf("Command '%s' missing usage", cmd.Name)
		}
		if cmd.Action == nil && len(cmd.Subcommands) == 0 {
			t.Errorf("Command '%s' missing action and subcommands", cmd.Name)
		}
	}
}

func TestApp_CommandCount(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Verify we have expected number of commands
	if len(app.Commands) == 0 {
		t.Error("No commands registered")
	}

	// Count expected commands (should be at least the core ones)
	expectedMinCommands := 8 // init, install, status, work, validate, clients, config, secrets
	if len(app.Commands) < expectedMinCommands {
		t.Errorf("Expected at least %d commands, got %d", expectedMinCommands, len(app.Commands))
	}
}

func TestApp_SubcommandValidation(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find commands with subcommands and validate them
	for _, cmd := range app.Commands {
		if len(cmd.Subcommands) > 0 {
			for _, subcmd := range cmd.Subcommands {
				if subcmd.Name == "" {
					t.Errorf("Subcommand of '%s' missing name", cmd.Name)
				}
				if subcmd.Usage == "" {
					t.Errorf("Subcommand '%s' of '%s' missing usage", subcmd.Name, cmd.Name)
				}
			}
		}
	}
}

func TestApp_FlagValidation(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find commands with flags and validate them
	for _, cmd := range app.Commands {
		for _, flag := range cmd.Flags {
			if flag == nil {
				t.Errorf("Command '%s' has nil flag", cmd.Name)
				continue
			}
			// For string flags, check if they have names
			if stringFlag, ok := flag.(*cli.StringFlag); ok {
				if stringFlag.Name == "" {
					t.Errorf("Command '%s' has string flag without name", cmd.Name)
				}
			}
		}
	}
}
