package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/pkg"
)

// StatusCommand handles project status display
type StatusCommand struct {
	projectManager *project.Manager
	clientRegistry *client.Registry
}

// NewStatusCommand creates a new status command
func NewStatusCommand() *StatusCommand {
	deps := NewBaseCommandDependencies()
	
	return &StatusCommand{
		projectManager: deps.ProjectManager,
		clientRegistry: deps.ClientRegistry,
	}
}

// Name returns the command name
func (c *StatusCommand) Name() string {
	return "status"
}

// Description returns the command description
func (c *StatusCommand) Description() string {
	return "Show current project information"
}

// Execute runs the status command
func (c *StatusCommand) Execute(args []string) error {
	if !c.projectManager.IsProject() {
		fmt.Fprintf(os.Stderr, "Not in a servo project directory. Run 'servo init' to initialize.\n")
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	projectPath, _ := c.projectManager.GetProjectPath()

	fmt.Printf("Servo Project Status\n")
	fmt.Printf("===================\n")
	fmt.Printf("Path:        %s\n", projectPath)

	if len(project.Clients) > 0 {
		fmt.Printf("Clients:     %s\n", strings.Join(project.Clients, ", "))
	} else {
		fmt.Printf("Clients:     (none configured)\n")
	}

	if project.ActiveSession != "" {
		fmt.Printf("Active Session: %s\n", project.ActiveSession)
	}
	fmt.Printf("Default Session: %s\n", project.DefaultSession)

	// Show MCP servers
	fmt.Println()
	if len(project.MCPServers) > 0 {
		fmt.Printf("MCP Servers: %d configured\n", len(project.MCPServers))
		for _, server := range project.MCPServers {
			clientList := "all clients"
			if len(server.Clients) > 0 {
				clientList = strings.Join(server.Clients, ", ")
			}
			fmt.Printf("  • %s (%s)\n", server.Name, clientList)
		}
	} else {
		fmt.Printf("MCP Servers: (none configured)\n")
	}

	// Show devcontainer status
	fmt.Println()
	devcontainerExists := c.checkDevcontainerExists()
	if devcontainerExists {
		fmt.Printf("Devcontainer: ✅ Configured\n")
		fmt.Printf("  • .devcontainer/devcontainer.json\n")
		fmt.Printf("  • .devcontainer/docker-compose.yml\n")
	} else {
		fmt.Printf("Devcontainer: ❌ Not configured\n")
		fmt.Printf("  • Run 'servo install <server>' to generate devcontainer\n")
	}

	// Show client configurations
	fmt.Println()
	fmt.Printf("Client Configurations:\n")
	c.showClientConfigurations(project)

	// Show secrets status
	fmt.Println()
	c.showSecretsStatus(project)

	fmt.Println()
	fmt.Printf("Configuration: %s\n", c.projectManager.GetServoDir())

	return nil
}

func (c *StatusCommand) checkDevcontainerExists() bool {
	_, err := os.Stat(".devcontainer/devcontainer.json")
	return err == nil
}

func (c *StatusCommand) showClientConfigurations(project *project.Project) {
	// Only show status for clients that are configured in the project
	if len(project.Clients) == 0 {
		fmt.Printf("  • No clients configured\n")
		return
	}

	hasUnconfigured := false
	for _, clientName := range project.Clients {
		client, err := c.clientRegistry.Get(clientName)
		if err != nil {
			fmt.Printf("  • %s: ❌ Unknown client\n", clientName)
			continue
		}

		// Check if client is installed
		if !client.IsInstalled() {
			fmt.Printf("  • %s: ❌ Not installed\n", client.Name())
			continue
		}

		// Check if client has configuration
		hasConfig := c.checkClientConfig(client)
		if hasConfig {
			fmt.Printf("  • %s: ✅ Configured\n", client.Name())
		} else {
			fmt.Printf("  • %s: ⚠️  Available but not configured\n", client.Name())
			hasUnconfigured = true
		}
	}

	// Show helpful tip if there are unconfigured clients
	if hasUnconfigured {
		fmt.Printf("  💡 Generate client configurations: servo configure\n")
	}
}

func (c *StatusCommand) checkClientConfig(client pkg.Client) bool {
	// Check for common configuration files based on client type
	switch client.Name() {
	case "vscode":
		_, err := os.Stat(".vscode/mcp.json")
		return err == nil
	case "claude-code":
		_, err := os.Stat(".mcp.json")
		return err == nil
	case "cursor":
		_, err := os.Stat(".cursor/mcp.json")
		return err == nil
	default:
		// For unknown clients, try to get current config
		config, err := client.GetCurrentConfig("local")
		if err != nil {
			return false
		}
		return len(config.Servers) > 0
	}
}

func (c *StatusCommand) showSecretsStatus(project *project.Project) {
	fmt.Printf("Secrets Status:\n")

	if len(project.RequiredSecrets) == 0 {
		fmt.Printf("  • No secrets required by configured servers\n")
		return
	}

	// Get missing secrets from project manager (uses proper decryption)
	missingSecrets, err := c.projectManager.GetMissingSecrets()
	if err != nil {
		fmt.Printf("  ⚠️  Warning: Failed to check secrets status: %v\n", err)
		return
	}

	// Build set of missing secret names
	missingSet := make(map[string]bool)
	for _, secret := range missingSecrets {
		missingSet[secret.Name] = true
	}

	// Show configured and missing secrets
	var configuredSecrets []string

	for _, required := range project.RequiredSecrets {
		if missingSet[required.Name] {
			// This is a missing secret - will be shown below
		} else {
			configuredSecrets = append(configuredSecrets, required.Name)
		}
	}

	if len(configuredSecrets) > 0 {
		fmt.Printf("  ✅ Configured secrets: %d\n", len(configuredSecrets))
		for _, secret := range configuredSecrets {
			fmt.Printf("    • %s\n", secret)
		}
	}

	if len(missingSecrets) > 0 {
		fmt.Printf("  ❌ Missing secrets: %d\n", len(missingSecrets))
		for _, secret := range missingSecrets {
			fmt.Printf("    • %s: %s\n", secret.Name, secret.Description)
		}
		fmt.Printf("  💡 Set secrets with: servo secrets set <key> <value>\n")
	}

	if len(missingSecrets) == 0 && len(configuredSecrets) > 0 {
		fmt.Printf("  🎉 All required secrets are configured!\n")
	}
}
