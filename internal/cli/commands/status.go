package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/servo/servo/internal/project"
)

// StatusCommand handles project status display
type StatusCommand struct {
	projectManager *project.Manager
}

// NewStatusCommand creates a new status command
func NewStatusCommand() *StatusCommand {
	return &StatusCommand{
		projectManager: project.NewManager(),
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
	if c.checkVSCodeConfigExists() {
		fmt.Printf("  • VSCode: ✅ .vscode/mcp.json\n")
	} else {
		fmt.Printf("  • VSCode: ❌ Not configured\n")
	}

	if c.checkClaudeCodeConfigExists() {
		fmt.Printf("  • Claude Code: ✅ .mcp.json\n")
	} else {
		fmt.Printf("  • Claude Code: ❌ Not configured\n")
	}

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

func (c *StatusCommand) checkVSCodeConfigExists() bool {
	_, err := os.Stat(".vscode/mcp.json")
	return err == nil
}

func (c *StatusCommand) checkClaudeCodeConfigExists() bool {
	_, err := os.Stat(".mcp.json")
	return err == nil
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
