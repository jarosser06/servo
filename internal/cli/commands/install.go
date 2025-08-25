package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/servo/servo/clients/claude_code"
	"github.com/servo/servo/clients/cursor"
	"github.com/servo/servo/clients/vscode"
	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/internal/config"
	"github.com/servo/servo/internal/manifest"
	"github.com/servo/servo/internal/mcp"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/session"
	"github.com/servo/servo/pkg"
)

// InstallCommand handles MCP server installation for projects
type InstallCommand struct {
	projectManager *project.Manager
	sessionManager *session.Manager
	configManager  *config.ConfigGeneratorManager
	clientRegistry pkg.ClientRegistry
	parser         *mcp.Parser
	validator      *mcp.Validator
}

// NewInstallCommand creates a new project install command
func NewInstallCommand(parser *mcp.Parser, validator *mcp.Validator) *InstallCommand {
	servoDir := os.Getenv("SERVO_DIR")
	if servoDir == "" {
		servoDir = ".servo" // Use project-local servo directory
	}

	// Create client registry and register default clients
	clientRegistry := client.NewRegistry()
	clientRegistry.Register(vscode.New())
	clientRegistry.Register(cursor.New())
	clientRegistry.Register(claude_code.New())

	return &InstallCommand{
		projectManager: project.NewManager(),
		sessionManager: session.NewManager(servoDir),
		configManager:  config.NewConfigGeneratorManager(servoDir),
		clientRegistry: clientRegistry,
		parser:         parser,
		validator:      validator,
	}
}

// Name returns the command name
func (c *InstallCommand) Name() string {
	return "install"
}

// Description returns the command description
func (c *InstallCommand) Description() string {
	return "Install MCP server to current project"
}

// Execute runs the install command
func (c *InstallCommand) Execute(args []string) error {
	return c.ExecuteWithOptions(args, nil, "", false)
}

// ExecuteWithOptions runs the install command with specific options
func (c *InstallCommand) ExecuteWithOptions(args []string, clients []string, sessionName string, forceUpdate bool) error {
	if !c.projectManager.IsProject() {
		fmt.Fprintf(os.Stderr, "Not in a servo project directory. Run 'servo init' to initialize.\n")
		return fmt.Errorf("not in a servo project directory")
	}

	if len(args) == 0 {
		return fmt.Errorf("server source is required\nUsage: servo install <source>")
	}

	source := args[0]

	// Validate and cleanup clients list - only support devcontainer-compatible clients
	clients = c.validateClients(clients)

	// Get project configuration to determine session
	project, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get project configuration: %w", err)
	}

	// Use specified session or fall back to active/default session
	targetSession := sessionName
	explicitSession := targetSession != "" // Track if user explicitly specified a session
	if targetSession == "" {
		// Check for active session first (from session manager, not project)
		activeSession, err := c.sessionManager.GetActive()
		if err != nil {
			return fmt.Errorf("failed to get active session: %w", err)
		}

		if activeSession != nil {
			targetSession = activeSession.Name
		} else {
			targetSession = project.DefaultSession
		}
	}

	// Parse the source to get server name (could be file, URL, or repo)
	serverName, err := c.extractServerName(source)
	if err != nil {
		return fmt.Errorf("failed to determine server name: %w", err)
	}

	fmt.Printf("📦 Adding MCP server '%s' to project (session: %s)...\n", serverName, targetSession)

	// Ensure session directories exist (only check if explicitly specified)
	if explicitSession {
		if err := c.validateSessionExists(targetSession); err != nil {
			return err
		}
	}

	// Add server to project configuration for specific session
	if err := c.projectManager.AddMCPServerToSession(serverName, source, clients, targetSession, forceUpdate); err != nil {
		// Handle the special case where server already exists and no update was requested
		if strings.Contains(err.Error(), "already exists in session") {
			fmt.Printf("⚠️  %s\n", err.Error())
			fmt.Printf("   Use --update flag to update the existing server.\n")
			fmt.Printf("   Nothing to do.\n")
			return nil // Not an error - just nothing to do
		}
		return fmt.Errorf("failed to add server to project: %w", err)
	}

	// Extract and add required secrets from the servo file
	if err := c.addRequiredSecretsFromSource(source); err != nil {
		return fmt.Errorf("failed to extract required secrets: %w", err)
	}

	// Store manifest in session and generate configurations dynamically
	if err := c.storeManifestAndGenerateConfigs(serverName, source, targetSession); err != nil {
		return fmt.Errorf("failed to store manifest and generate configurations: %w", err)
	}

	fmt.Printf("✅ Added server '%s' to project\n", serverName)
	fmt.Println()
	fmt.Println("Updated files:")
	fmt.Println("  • .servo/project.yaml (server declaration)")
	fmt.Println("  • .devcontainer/devcontainer.json (installation commands)")
	fmt.Println("  • .vscode/mcp.json (MCP configuration)")
	if c.hasClaudeCodeClient(clients) {
		fmt.Println("  • .mcp.json (Claude Code configuration)")
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Open in VSCode Dev Container")
	fmt.Println("  • MCP servers will be installed and configured automatically")

	return nil
}

// validateClients ensures only supported devcontainer-compatible clients are included
func (c *InstallCommand) validateClients(clients []string) []string {
	supportedClients := map[string]bool{
		"vscode":      true,
		"claude-code": true,
		"cursor":      true,
	}

	// If no clients specified, default to all supported clients
	if len(clients) == 0 {
		return []string{"vscode", "claude-code", "cursor"}
	}

	var validClients []string
	for _, client := range clients {
		client = strings.TrimSpace(strings.ToLower(client))
		if supportedClients[client] {
			validClients = append(validClients, client)
		} else {
			fmt.Printf("⚠️  Skipping unsupported client '%s' (only vscode, claude-code, cursor supported)\n", client)
		}
	}

	return validClients
}

// extractServerName extracts a server name from various source formats
func (c *InstallCommand) extractServerName(source string) (string, error) {
	// If it's a file path, parse it
	if strings.HasSuffix(source, ".servo") {
		// Use parser to extract name from .servo file
		servoDef, err := c.parser.ParseFromFile(source)
		if err != nil {
			return "", fmt.Errorf("failed to parse servo file: %w", err)
		}
		if servoDef.Name == "" {
			return "", fmt.Errorf("servo file missing name field")
		}
		return servoDef.Name, nil
	}

	// Handle URLs (git repos, direct URLs)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Try parsing as URL first
		servoDef, err := c.parser.ParseFromURL(source)
		if err == nil {
			return servoDef.Name, nil
		}

		// If URL parsing fails, try as git repo
		if strings.Contains(source, "github.com") || strings.Contains(source, ".git") {
			servoDef, err := c.parser.ParseFromGitRepo(source, "")
			if err == nil {
				return servoDef.Name, nil
			}
		}

		// Fallback: extract from URL path
		if strings.Contains(source, "/") {
			parts := strings.Split(source, "/")
			name := parts[len(parts)-1]
			if strings.Contains(name, ".") {
				name = strings.Split(name, ".")[0]
			}
			return name, nil
		}
	}

	// Handle git SSH URLs
	if strings.HasPrefix(source, "git@") || strings.Contains(source, "ssh://") {
		servoDef, err := c.parser.ParseFromGitRepo(source, "")
		if err == nil {
			return servoDef.Name, nil
		}

		// Fallback: extract repo name from SSH URL
		// e.g., git@github.com:user/repo.git -> repo
		parts := strings.Split(source, "/")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			name = strings.TrimSuffix(name, ".git")
			return name, nil
		}
	}

	// Handle directory paths
	if strings.Contains(source, "/") {
		// Try parsing from directory first
		servoDef, err := c.parser.ParseFromDirectory(source)
		if err == nil {
			return servoDef.Name, nil
		}

		// Fallback: extract from path
		parts := strings.Split(source, "/")
		name := parts[len(parts)-1]
		if strings.Contains(name, ".") {
			name = strings.Split(name, ".")[0]
		}
		return name, nil
	}

	// Default: assume it's already a server name
	return source, nil
}

// hasClaudeCodeClient checks if Claude Code is in the clients list
func (c *InstallCommand) hasClaudeCodeClient(clients []string) bool {
	for _, client := range clients {
		if client == "claude-code" {
			return true
		}
	}
	return false
}

// storeManifestAndGenerateConfigs stores the server manifest and generates all configurations dynamically
func (c *InstallCommand) storeManifestAndGenerateConfigs(serverName, source, sessionName string) error {
	// Get session directory
	sessionDir := c.sessionManager.GetSessionDir(sessionName)

	// Create manifest store for this session
	store := manifest.NewStore(sessionDir, c.parser)

	// Store the manifest
	if err := store.StoreManifest(serverName, source); err != nil {
		return fmt.Errorf("failed to store manifest: %w", err)
	}

	// Generate all configurations dynamically from manifests
	if err := c.configManager.GenerateDevcontainer(); err != nil {
		return fmt.Errorf("failed to generate devcontainer: %w", err)
	}

	if err := c.configManager.GenerateDockerCompose(); err != nil {
		return fmt.Errorf("failed to generate docker-compose: %w", err)
	}

	// Generate MCP configurations for all installed clients based on target session
	if err := c.generateMCPConfigurationsForSession(sessionName); err != nil {
		return fmt.Errorf("failed to generate MCP configs: %w", err)
	}

	return nil
}

// generateMCPConfigurations generates MCP configurations for all registered clients using active session
func (c *InstallCommand) generateMCPConfigurations() error {
	// Get active session and manifests
	activeSession, err := c.sessionManager.GetActive()
	if err != nil {
		return fmt.Errorf("failed to get active session: %w", err)
	}

	if activeSession == nil {
		return fmt.Errorf("no active session found")
	}

	return c.generateMCPConfigurationsForSession(activeSession.Name)
}

// generateMCPConfigurationsForSession generates MCP configurations for a specific session
func (c *InstallCommand) generateMCPConfigurationsForSession(sessionName string) error {
	// Get project configuration
	_, err := c.projectManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get manifests from specified session
	store := manifest.NewStore(c.sessionManager.GetSessionDir(sessionName), c.parser)
	manifestsMap, err := store.ListManifests()
	if err != nil {
		return fmt.Errorf("failed to list manifests: %w", err)
	}

	// Convert to slice format for client interface
	manifests := make([]pkg.ServoDefinition, 0, len(manifestsMap))
	for _, manifest := range manifestsMap {
		manifests = append(manifests, *manifest)
	}

	// Create secrets provider
	configuredSecrets, err := c.projectManager.GetConfiguredSecrets()
	if err != nil {
		return fmt.Errorf("failed to get configured secrets: %w", err)
	}

	secretsProvider := func(secretName string) (string, error) {
		if !configuredSecrets[secretName] {
			return "", fmt.Errorf("secret '%s' is not configured", secretName)
		}
		// Return placeholder - in real usage, this would decrypt the actual secret
		return "", nil
	}

	// Generate configurations for all clients
	clients := c.clientRegistry.List()
	for _, client := range clients {
		if client.IsInstalled() { // Only generate for installed clients
			if err := client.GenerateConfig(manifests, secretsProvider); err != nil {
				return fmt.Errorf("failed to generate config for %s: %w", client.Name(), err)
			}
		}
	}

	return nil
}

// addRequiredSecretsFromSource extracts and adds required secrets from a servo source
func (c *InstallCommand) addRequiredSecretsFromSource(source string) error {
	// Parse the source to extract required secrets
	var servoDef *pkg.ServoDefinition
	var err error

	switch {
	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		servoDef, err = c.parser.ParseFromURL(source)
	case strings.Contains(source, "@") || strings.Contains(source, "git"):
		servoDef, err = c.parser.ParseFromGitRepo(source, "")
	default:
		servoDef, err = c.parser.ParseFromFile(source)
	}

	if err != nil {
		return fmt.Errorf("failed to parse servo file: %w", err)
	}

	// Extract required secrets from configuration schema
	if servoDef.ConfigurationSchema != nil && servoDef.ConfigurationSchema.Secrets != nil {
		for secretName, secretSchema := range servoDef.ConfigurationSchema.Secrets {
			if secretSchema.Required {
				err := c.projectManager.AddRequiredSecret(secretName, secretSchema.Description)
				if err != nil {
					// If secret already exists, that's fine
					if !strings.Contains(err.Error(), "already exists") {
						return fmt.Errorf("failed to add required secret %s: %w", secretName, err)
					}
				}
			}
		}
	}

	return nil
}

// validateSessionExists ensures the session exists (fails if it doesn't)
func (c *InstallCommand) validateSessionExists(sessionName string) error {
	// Check if session exists
	exists, err := c.sessionManager.Exists(sessionName)
	if err != nil {
		return fmt.Errorf("failed to check if session exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("session '%s' does not exist. Create it first with: servo session create %s", sessionName, sessionName)
	}

	return nil
}
