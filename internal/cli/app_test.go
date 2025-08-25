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
	requiredCommands := []string{"init", "install", "status", "work", "validate", "client", "config", "secrets"}
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
	expectedMinCommands := 8 // init, install, status, work, validate, client, config, secrets
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

func TestApp_ClientCommand(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find client command
	var clientCmd *cli.Command
	for _, cmd := range app.Commands {
		if cmd.Name == "client" {
			clientCmd = cmd
			break
		}
	}

	if clientCmd == nil {
		t.Fatal("Client command not found")
	}

	// Verify client command structure
	if clientCmd.Usage != "Manage MCP client support for this project" {
		t.Errorf("Expected client command usage to be 'Manage MCP client support for this project', got '%s'", clientCmd.Usage)
	}

	// Verify required subcommands are present
	requiredSubcommands := []string{"list", "enable", "disable"}
	for _, subcmdName := range requiredSubcommands {
		found := false
		for _, subcmd := range clientCmd.Subcommands {
			if subcmd.Name == subcmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required subcommand '%s' not found in client command", subcmdName)
		}
	}

	// Verify subcommand structure
	for _, subcmd := range clientCmd.Subcommands {
		if subcmd.Name == "" {
			t.Error("Client subcommand missing name")
		}
		if subcmd.Usage == "" {
			t.Errorf("Client subcommand '%s' missing usage", subcmd.Name)
		}
		if subcmd.Action == nil {
			t.Errorf("Client subcommand '%s' missing action", subcmd.Name)
		}
	}
}

func TestApp_ClientEnableCommand(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find client command and enable subcommand
	var enableCmd *cli.Command
	for _, cmd := range app.Commands {
		if cmd.Name == "client" {
			for _, subcmd := range cmd.Subcommands {
				if subcmd.Name == "enable" {
					enableCmd = subcmd
					break
				}
			}
			break
		}
	}

	if enableCmd == nil {
		t.Fatal("Client enable command not found")
	}

	// Verify enable command structure
	if enableCmd.Usage != "Enable support for one or more clients in the current project" {
		t.Errorf("Expected enable command usage to describe enabling clients, got '%s'", enableCmd.Usage)
	}

	if enableCmd.ArgsUsage != "<client> [<client> ...]" {
		t.Errorf("Expected enable command args usage to be '<client> [<client> ...]', got '%s'", enableCmd.ArgsUsage)
	}

	if enableCmd.Action == nil {
		t.Error("Enable command missing action")
	}
}

func TestApp_ClientDisableCommand(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find client command and disable subcommand
	var disableCmd *cli.Command
	for _, cmd := range app.Commands {
		if cmd.Name == "client" {
			for _, subcmd := range cmd.Subcommands {
				if subcmd.Name == "disable" {
					disableCmd = subcmd
					break
				}
			}
			break
		}
	}

	if disableCmd == nil {
		t.Fatal("Client disable command not found")
	}

	// Verify disable command structure
	if disableCmd.Usage != "Disable support for one or more clients in the current project" {
		t.Errorf("Expected disable command usage to describe disabling clients, got '%s'", disableCmd.Usage)
	}

	if disableCmd.ArgsUsage != "<client> [<client> ...]" {
		t.Errorf("Expected disable command args usage to be '<client> [<client> ...]', got '%s'", disableCmd.ArgsUsage)
	}

	if disableCmd.Action == nil {
		t.Error("Disable command missing action")
	}
}

func TestApp_ClientListCommand(t *testing.T) {
	// Create test app
	app, err := NewApp("test-version")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Find client command and list subcommand
	var listCmd *cli.Command
	for _, cmd := range app.Commands {
		if cmd.Name == "client" {
			for _, subcmd := range cmd.Subcommands {
				if subcmd.Name == "list" {
					listCmd = subcmd
					break
				}
			}
			break
		}
	}

	if listCmd == nil {
		t.Fatal("Client list command not found")
	}

	// Verify list command structure
	if listCmd.Usage != "List available MCP clients" {
		t.Errorf("Expected list command usage to describe listing clients, got '%s'", listCmd.Usage)
	}

	if listCmd.Action == nil {
		t.Error("List command missing action")
	}
}
