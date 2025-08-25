package commands

import (
	"fmt"

	"github.com/servo/servo/internal/project"
)

// ConfigCommand manages project configuration
type ConfigCommand struct {
	projectManager *project.Manager
}

func NewConfigCommand(projectManager *project.Manager) *ConfigCommand {
	return &ConfigCommand{
		projectManager: projectManager,
	}
}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Description() string {
	return "Manage project configuration"
}

func (c *ConfigCommand) Execute(args []string) error {
	if len(args) == 0 {
		return c.showCurrentConfig()
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: servo config get <key>")
		}
		return c.getConfig(args[1])
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: servo config set <key> <value>")
		}
		return c.setConfig(args[1], args[2])
	case "list":
		return c.showCurrentConfig()
	case "help", "--help", "-h":
		return c.showHelp()
	default:
		return fmt.Errorf("unknown config subcommand: %s", args[0])
	}
}

func (c *ConfigCommand) showCurrentConfig() error {
	if !c.projectManager.IsProject() {
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	fmt.Println("Project Configuration:")
	fmt.Printf("  default_session: %s\n", project.DefaultSession)
	fmt.Printf("  active_session: %s\n", project.ActiveSession)
	if len(project.Clients) > 0 {
		fmt.Printf("  clients: %v\n", project.Clients)
	}
	if len(project.MCPServers) > 0 {
		fmt.Printf("  mcp_servers:\n")
		for _, server := range project.MCPServers {
			fmt.Printf("    - %s: %s\n", server.Name, server.Source)
		}
	}

	return nil
}

func (c *ConfigCommand) getConfig(key string) error {
	if !c.projectManager.IsProject() {
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	switch key {
	case "default_session":
		fmt.Println(project.DefaultSession)
	case "active_session":
		fmt.Println(project.ActiveSession)
	case "clients":
		for _, client := range project.Clients {
			fmt.Println(client)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func (c *ConfigCommand) setConfig(key, value string) error {
	if !c.projectManager.IsProject() {
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	switch key {
	case "default_session":
		project.DefaultSession = value
	case "active_session":
		project.ActiveSession = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := c.projectManager.Save(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("✓ Project '%s' set to '%s'\n", key, value)

	// If we changed the default session, provide guidance
	if key == "default_session" {
		fmt.Printf("💡 The default session '%s' will be used for new installations\n", value)
		fmt.Printf("💡 Use 'servo config set active_session %s' to also switch the active session\n", value)
	}

	return nil
}

func (c *ConfigCommand) showHelp() error {
	fmt.Printf(`project config - Manage project configuration

USAGE:
    servo config [SUBCOMMAND]

SUBCOMMANDS:
    list           Show all project configuration
    get <key>      Get a configuration value
    set <key> <val> Set a configuration value
    help           Show this help message

AVAILABLE KEYS:
    name              Project name
    description       Project description
    default_session   Default session name (used for new installations)
    active_session    Currently active session

EXAMPLES:
    servo config list
    servo config get default_session
    servo config set default_session production
    servo config set active_session development
`)
	return nil
}
