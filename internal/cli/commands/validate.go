package commands

import (
	"fmt"
	"strings"

	"github.com/servo/servo/internal/mcp"
	"github.com/servo/servo/pkg"
)

// ValidateCommand handles validation of .servo files
type ValidateCommand struct {
	parser    *mcp.Parser
	validator *mcp.Validator
}

// NewValidateCommand creates a new validate command
func NewValidateCommand(parser *mcp.Parser, validator *mcp.Validator) *ValidateCommand {
	return &ValidateCommand{
		parser:    parser,
		validator: validator,
	}
}

// Name returns the command name
func (c *ValidateCommand) Name() string {
	return "validate"
}

// Description returns the command description
func (c *ValidateCommand) Description() string {
	return "Validate .servo file or source"
}

// Execute runs the validate command
func (c *ValidateCommand) Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("source is required\nUsage: servo validate <source>")
	}

	// Handle help first
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return c.showHelp()
	}

	source := args[0]

	fmt.Printf("Validating: %s\n", source)

	// Parse the source
	servoFile, err := c.parseSource(source)
	if err != nil {
		fmt.Printf("❌ Failed to parse source: %v\n", err)
		return err
	}

	fmt.Printf("✓ Successfully parsed .servo file\n")

	// Validate the servo file
	if err := c.validator.Validate(servoFile); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Validation passed!\n\n")

	// Display summary
	fmt.Printf("Package Information:\n")
	fmt.Printf("  Name: %s\n", servoFile.Name)
	fmt.Printf("  Version: %s\n", servoFile.Version)
	fmt.Printf("  Description: %s\n", servoFile.Description)
	fmt.Printf("  Author: %s\n", servoFile.Author)
	fmt.Printf("  License: %s\n", servoFile.License)

	if servoFile.Metadata != nil && len(servoFile.Metadata.Tags) > 0 {
		fmt.Printf("  Tags: ")
		for i, tag := range servoFile.Metadata.Tags {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", tag)
		}
		fmt.Printf("\n")
	}

	// Display requirements
	if servoFile.Requirements != nil {
		fmt.Printf("\nRequirements:\n")

		if len(servoFile.Requirements.System) > 0 {
			fmt.Printf("  System:\n")
			for _, req := range servoFile.Requirements.System {
				fmt.Printf("    - %s: %s\n", req.Name, req.Description)
			}
		}

		if len(servoFile.Requirements.Runtimes) > 0 {
			fmt.Printf("  Runtimes:\n")
			for _, req := range servoFile.Requirements.Runtimes {
				fmt.Printf("    - %s %s\n", req.Name, req.Version)
			}
		}
	}

	// Display dependencies
	if servoFile.Dependencies != nil && len(servoFile.Dependencies.Services) > 0 {
		fmt.Printf("\nServices:\n")
		for name, service := range servoFile.Dependencies.Services {
			fmt.Printf("  - %s: %s\n", name, service.Image)
		}
	}

	// Display configuration requirements
	if servoFile.ConfigurationSchema != nil {
		if len(servoFile.ConfigurationSchema.Secrets) > 0 {
			fmt.Printf("\nRequired Secrets:\n")
			for name, secret := range servoFile.ConfigurationSchema.Secrets {
				required := ""
				if secret.Required {
					required = " (required)"
				}
				fmt.Printf("  - %s: %s%s\n", name, secret.Description, required)
			}
		}

		if len(servoFile.ConfigurationSchema.Config) > 0 {
			fmt.Printf("\nConfiguration Options:\n")
			for name, config := range servoFile.ConfigurationSchema.Config {
				defaultVal := ""
				if config.Default != nil {
					defaultVal = fmt.Sprintf(" (default: %v)", config.Default)
				}
				fmt.Printf("  - %s: %s%s\n", name, config.Description, defaultVal)
			}
		}
	}

	// Display client compatibility
	if servoFile.Clients != nil {
		if len(servoFile.Clients.Recommended) > 0 {
			fmt.Printf("\nRecommended Clients: ")
			for i, client := range servoFile.Clients.Recommended {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", client)
			}
			fmt.Printf("\n")
		}

		if len(servoFile.Clients.Tested) > 0 {
			fmt.Printf("Tested Clients: ")
			for i, client := range servoFile.Clients.Tested {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", client)
			}
			fmt.Printf("\n")
		}
	}

	return nil
}

// parseSource parses a source based on its format
func (c *ValidateCommand) parseSource(source string) (*pkg.ServoDefinition, error) {
	switch {
	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		if strings.Contains(source, "github.com") && !strings.HasSuffix(source, ".servo") {
			return c.parser.ParseFromGitRepo(source, "")
		} else {
			return c.parser.ParseFromURL(source)
		}
	case strings.HasSuffix(source, ".servo"):
		return c.parser.ParseFromFile(source)
	default:
		return c.parser.ParseFromDirectory(source)
	}
}

func (c *ValidateCommand) showHelp() error {
	fmt.Printf(`validate - Validate .servo file or source

USAGE:
    servo validate <source>

ARGUMENTS:
    <source>    .servo file, git repository URL, or local directory containing .servo files

EXAMPLES:
    servo validate ./graphiti.servo
    servo validate https://github.com/user/repo.git
    servo validate ./local-directory
`)
	return nil
}
