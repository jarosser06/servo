package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/servo/servo/internal/config"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/registry"
	"github.com/servo/servo/internal/session"
	"github.com/servo/servo/pkg"
)

// WorkCommand handles starting the development environment
type WorkCommand struct {
	projectManager *project.Manager
	sessionManager *session.Manager
	clientRegistry pkg.ClientRegistry
}

// NewWorkCommand creates a new work command
func NewWorkCommand() *WorkCommand {
	servoDir := os.Getenv("SERVO_DIR")
	if servoDir == "" {
		servoDir = ".servo"
	}

	return &WorkCommand{
		projectManager: project.NewManager(),
		sessionManager: session.NewManager(servoDir),
		clientRegistry: registry.GetDefaultRegistry(),
	}
}

// Name returns the command name
func (c *WorkCommand) Name() string {
	return "work"
}

// Description returns the command description
func (c *WorkCommand) Description() string {
	return "Start development environment for project"
}

// Execute runs the work command
func (c *WorkCommand) Execute(args []string) error {
	if !c.projectManager.IsProject() {
		fmt.Fprintf(os.Stderr, "Not in a servo project directory. Run 'servo init' to initialize.\n")
		return fmt.Errorf("not in a servo project directory")
	}

	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Parse command line options
	var client string
	var shouldLaunchClient bool

	for i, arg := range args {
		switch arg {
		case "--client":
			if i+1 < len(args) {
				client = args[i+1]
				shouldLaunchClient = true
			}
		case "--vscode":
			client = "vscode"
			shouldLaunchClient = true
		case "--claude-code":
			client = "claude-code"
			shouldLaunchClient = true
		}
	}

	projectName, err := c.projectManager.GetProjectName()
	if err != nil {
		projectName = "unknown"
	}
	fmt.Printf("🚀 Starting development environment for project: %s\n", projectName)
	fmt.Printf("📍 Active session: %s\n", project.ActiveSession)
	fmt.Println()

	// Step 1: Generate configurations
	fmt.Println("⚙️  Generating configurations...")
	if err := c.generateConfigurations(); err != nil {
		return fmt.Errorf("failed to generate configurations: %w", err)
	}
	fmt.Println("✅ Configurations generated")
	fmt.Println()

	// Step 2: Configuration ready
	fmt.Println("✅ Development environment configured")
	fmt.Printf("   → Devcontainer: .devcontainer/devcontainer.json\n")
	fmt.Printf("   → Services: .devcontainer/docker-compose.yml\n")
	fmt.Println()

	// Step 3: Show MCP server status
	if len(project.MCPServers) > 0 {
		fmt.Println("📦 MCP Servers configured:")
		for _, server := range project.MCPServers {
			fmt.Printf("   • %s (clients: %s)\n", server.Name, strings.Join(server.Clients, ", "))
		}
		fmt.Println()
	}

	// Step 3: Show launch commands if client requested
	if shouldLaunchClient {
		launchCmd := c.getLaunchCommand(client)
		if launchCmd != "" {
			fmt.Printf("# Launch command for %s:\n", client)
			fmt.Println(launchCmd)
		} else {
			fmt.Printf("# %s client not found or unsupported\n", client)
		}
	} else {
		fmt.Println("# Development environment is ready!")
		fmt.Println("# Available launch commands:")
		
		// Get current working directory for launch commands
		pwd, err := os.Getwd()
		if err != nil {
			pwd = "."
		}
		
		// Show available clients with their launch commands
		clients := c.clientRegistry.List()
		for _, client := range clients {
			if client.IsInstalled() {
				launchCmd := client.GetLaunchCommand(pwd)
				fmt.Printf("# %s:\n%s\n", client.Name(), launchCmd)
				fmt.Println()
			}
		}
	}

	return nil
}

// generateConfigurations generates all necessary configuration files
func (c *WorkCommand) generateConfigurations() error {
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

// getLaunchCommand returns the command to launch the specified client
func (c *WorkCommand) getLaunchCommand(clientName string) string {
	client, err := c.clientRegistry.Get(clientName)
	if err != nil {
		return "" // Client not found
	}
	
	if !client.IsInstalled() {
		return "" // Client not installed
	}
	
	// Get current working directory for launch command
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}
	
	return client.GetLaunchCommand(pwd)
}

