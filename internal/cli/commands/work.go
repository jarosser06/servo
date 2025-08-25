package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/servo/servo/internal/config"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/session"
)

// WorkCommand handles starting the development environment
type WorkCommand struct {
	projectManager *project.Manager
	sessionManager *session.Manager
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

	// Step 3: Launch client if requested
	if shouldLaunchClient {
		fmt.Printf("🖥️  Launching %s client...\n", client)
		if err := c.launchClient(client); err != nil {
			fmt.Printf("⚠️  Failed to launch %s: %v\n", client, err)
			fmt.Printf("💡 You can manually launch %s and open this project\n", client)
		} else {
			fmt.Printf("✅ %s launched successfully\n", client)
		}
	} else {
		fmt.Println("💡 Development environment is ready!")
		fmt.Println("   Next steps:")
		fmt.Println("   • Open VSCode: servo work --vscode")
		fmt.Println("   • Use Claude Code: servo work --claude-code")
		fmt.Println("   • Or manually open your preferred client")
		fmt.Println()
		fmt.Println("   The devcontainer will automatically start services when opened.")
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

// launchClient launches the specified client
func (c *WorkCommand) launchClient(clientName string) error {
	switch clientName {
	case "vscode":
		return c.launchVSCode()
	case "claude-code":
		return c.launchClaudeCode()
	default:
		return fmt.Errorf("unsupported client: %s", clientName)
	}
}

// launchVSCode launches VSCode with devcontainer
func (c *WorkCommand) launchVSCode() error {
	// Try different VSCode command variations
	commands := []string{"code", "/usr/local/bin/code", "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"}

	for _, command := range commands {
		if c.commandExists(command) {
			cmd := exec.Command(command, ".")
			if err := cmd.Start(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("VSCode not found in PATH")
}

// launchClaudeCode launches Claude Code CLI
func (c *WorkCommand) launchClaudeCode() error {
	if !c.commandExists("claude") {
		return fmt.Errorf("Claude Code CLI not found - install from https://claude.ai/code")
	}

	cmd := exec.Command("claude")
	cmd.Dir = "."
	return cmd.Start()
}

// Helper methods
func (c *WorkCommand) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
