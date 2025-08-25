package commands

import (
	"fmt"
	"os"

	"github.com/servo/servo/internal/config"
	"github.com/servo/servo/internal/project"
)

// ConfigureCommand handles generating MCP client configurations
type ConfigureCommand struct {
	projectManager *project.Manager
}

// NewConfigureCommand creates a new configure command
func NewConfigureCommand() *ConfigureCommand {
	return &ConfigureCommand{
		projectManager: project.NewManager(),
	}
}

// Name returns the command name
func (c *ConfigureCommand) Name() string {
	return "configure"
}

// Description returns the command description
func (c *ConfigureCommand) Description() string {
	return "Generate MCP client configurations for the current project"
}

// Execute runs the configure command
func (c *ConfigureCommand) Execute(args []string) error {
	if !c.projectManager.IsProject() {
		fmt.Fprintf(os.Stderr, "Not in a servo project directory. Run 'servo init' to initialize.\n")
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	fmt.Printf("🔧 Generating MCP client configurations...\n")

	if len(project.MCPServers) == 0 {
		fmt.Printf("⚠️  No MCP servers configured. Install servers first with: servo install <source>\n")
		return nil
	}

	// Generate configurations
	if err := c.generateConfigurations(); err != nil {
		return fmt.Errorf("failed to generate configurations: %w", err)
	}

	fmt.Printf("✅ Configuration files generated successfully!\n")
	fmt.Printf("   → Devcontainer: .devcontainer/devcontainer.json\n")
	fmt.Printf("   → Services: .devcontainer/docker-compose.yml\n")

	// Show which clients were configured
	if len(project.Clients) > 0 {
		fmt.Printf("   → Client configs:\n")
		for _, clientName := range project.Clients {
			switch clientName {
			case "vscode":
				fmt.Printf("     • VSCode: .vscode/mcp.json\n")
			case "claude-code":
				fmt.Printf("     • Claude Code: .mcp.json\n")
			case "cursor":
				fmt.Printf("     • Cursor: .cursor/mcp.json\n")
			}
		}
	}

	fmt.Printf("\n💡 Your MCP clients should now be configured and ready to use!\n")
	return nil
}

// generateConfigurations generates all necessary configuration files
func (c *ConfigureCommand) generateConfigurations() error {
	servoDir := c.projectManager.GetServoDir()
	configManager := config.NewConfigGeneratorManager(servoDir)

	// Generate devcontainer configuration
	if err := configManager.GenerateDevcontainer(); err != nil {
		return fmt.Errorf("failed to generate devcontainer: %w", err)
	}

	// Generate docker-compose configuration
	if err := configManager.GenerateDockerCompose(); err != nil {
		return fmt.Errorf("failed to generate docker-compose: %w", err)
	}

	return nil
}
