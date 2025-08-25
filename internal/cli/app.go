package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/servo/servo/clients/claude_code"
	"github.com/servo/servo/clients/cursor"
	"github.com/servo/servo/clients/vscode"
	"github.com/servo/servo/internal/cli/commands"
	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/internal/mcp"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/session"
)

// NewApp creates a new CLI application using urfave/cli
func NewApp(version string) (*cli.App, error) {
	// Initialize core managers
	projectManager := project.NewManager()

	// Session manager is created on-demand with project-specific servo directory
	// Config generation is handled by ConfigGeneratorManager in individual commands
	// Secrets management is handled by SecretsCommand directly with project integration

	clientRegistry := client.NewRegistry()
	parser := mcp.NewParser()
	validator := mcp.NewValidator()

	// Register built-in clients
	clientRegistry.Register(claude_code.New())
	clientRegistry.Register(vscode.New())
	clientRegistry.Register(cursor.New())

	app := &cli.App{
		Name:        "servo",
		Usage:       "MCP Server Project Manager",
		Version:     version,
		Description: "Servo provides project-focused tool for managing Model Context Protocol (MCP) servers in isolated, containerized development environments. Each project maintains its own MCP servers, dependencies, and configuration while supporting team collaboration through git-friendly configs.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "no-interactive",
				Usage:   "Disable interactive prompts (requires environment variables for passwords)",
				Aliases: []string{"n"},
				EnvVars: []string{"SERVO_NON_INTERACTIVE"},
			},
		},
		Before: func(c *cli.Context) error {
			// Set global environment variable if flag is set
			if c.Bool("no-interactive") {
				os.Setenv("SERVO_NON_INTERACTIVE", "1")
			}
			return nil
		},
		Commands: []*cli.Command{
			// Project management commands
			{
				Name:        "init",
				Usage:       "Initialize a new Servo project",
				Description: "Create a new Servo project in the current directory with session support",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "session",
						Usage:   "Default session name (default: \"default\")",
						Aliases: []string{"s"},
					},
					&cli.StringSliceFlag{
						Name:    "clients",
						Usage:   "Comma-separated list of MCP clients (vscode,claude-code,cursor)",
						Aliases: []string{"c"},
					},
				},
				Action: func(c *cli.Context) error {
					sessionName := c.String("session")
					if sessionName == "" {
						sessionName = "default"
					}
					clients := c.StringSlice("clients")

					fmt.Printf("🏗️  Initializing servo project...\n")

					// Initialize project structure
					projectManager := project.NewManager()
					project, err := projectManager.Init(sessionName, clients)
					if err != nil {
						return fmt.Errorf("failed to initialize project: %w", err)
					}

					// Create default session using session manager
					sessionManager := session.NewManager(".servo")
					_, err = sessionManager.Create(sessionName, fmt.Sprintf("Default session for project"), "")
					if err != nil {
						return fmt.Errorf("failed to create default session: %w", err)
					}

					// Activate the session
					err = sessionManager.Activate(sessionName)
					if err != nil {
						return fmt.Errorf("failed to activate default session: %w", err)
					}

					fmt.Printf("✅ Project initialized successfully!\n")
					fmt.Printf("📁 Project directory: %s\n", ".servo")
					fmt.Printf("📋 Default session: %s\n", project.DefaultSession)
					fmt.Printf("\nNext steps:\n")
					fmt.Printf("  • Install MCP servers: servo install <source>\n")
					fmt.Printf("  • Start working: servo work\n")

					return nil
				},
			},

			{
				Name:        "install",
				Usage:       "Install MCP server from source",
				Description: "Install MCP server from .servo file, git repository, or local directory",
				ArgsUsage:   "<source>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "session",
						Usage:   "Install to specific session",
						Aliases: []string{"s"},
					},
					&cli.StringSliceFlag{
						Name:    "clients",
						Usage:   "Target MCP clients",
						Aliases: []string{"c"},
					},
					&cli.BoolFlag{
						Name:    "update",
						Usage:   "Update server if it already exists",
						Aliases: []string{"u"},
					},
					// Git authentication flags
					&cli.StringFlag{
						Name:    "ssh-key",
						Usage:   "SSH private key path for git authentication",
						EnvVars: []string{"GIT_SSH_KEY"},
					},
					&cli.StringFlag{
						Name:    "ssh-password",
						Usage:   "SSH key passphrase (SSH config ~/.ssh/config used automatically)",
						EnvVars: []string{"GIT_SSH_PASSWORD"},
					},
					&cli.StringFlag{
						Name:    "http-token",
						Usage:   "HTTP token for git authentication (GitHub token, etc.)",
						EnvVars: []string{"GIT_TOKEN", "GITHUB_TOKEN"},
					},
					&cli.StringFlag{
						Name:    "http-username",
						Usage:   "HTTP username for git authentication",
						EnvVars: []string{"GIT_USERNAME"},
					},
					&cli.StringFlag{
						Name:    "http-password",
						Usage:   "HTTP password for git authentication",
						EnvVars: []string{"GIT_PASSWORD"},
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("source required")
					}

					// Configure parser with authentication credentials
					parser.SSHKeyPath = c.String("ssh-key")
					parser.SSHPassword = c.String("ssh-password")
					parser.HTTPToken = c.String("http-token")
					parser.HTTPUsername = c.String("http-username")
					parser.HTTPPassword = c.String("http-password")

					installCmd := commands.NewInstallCommand(parser, validator)

					// Pass arguments and options directly
					args := []string{c.Args().First()}
					clients := c.StringSlice("clients")
					session := c.String("session")
					update := c.Bool("update")

					return installCmd.ExecuteWithOptions(args, clients, session, update)
				},
			},

			{
				Name:        "status",
				Usage:       "Show status of servers and services",
				Description: "Display current project status including servers and configurations",
				Action: func(c *cli.Context) error {
					statusCmd := commands.NewStatusCommand()
					return statusCmd.Execute([]string{})
				},
			},

			{
				Name:        "work",
				Usage:       "Start development environment with MCP servers",
				Description: "Launch the development environment with devcontainer support",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "client",
						Usage:   "Target client for development",
						Aliases: []string{"c"},
					},
				},
				Action: func(c *cli.Context) error {
					workCmd := commands.NewWorkCommand()

					args := []string{}
					if client := c.String("client"); client != "" {
						args = append(args, "--client", client)
					}

					return workCmd.Execute(args)
				},
			},

			{
				Name:        "session",
				Usage:       "Manage project sessions",
				Description: "Create, list, activate, and manage project sessions",
				Subcommands: []*cli.Command{
					{
						Name:      "create",
						Usage:     "Create a new session",
						ArgsUsage: "<session-name>",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "description",
								Usage:   "Session description",
								Aliases: []string{"d"},
							},
						},
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("session name required")
							}

							sessionName := c.Args().First()
							description := c.String("description")
							if description == "" {
								description = fmt.Sprintf("Session: %s", sessionName)
							}

							sessionManager := session.NewManager(".servo")
							_, err := sessionManager.Create(sessionName, description, "")
							if err != nil {
								return fmt.Errorf("failed to create session: %w", err)
							}

							fmt.Printf("✅ Created session '%s'\n", sessionName)
							return nil
						},
					},
					{
						Name:  "list",
						Usage: "List all sessions",
						Action: func(c *cli.Context) error {
							sessionManager := session.NewManager(".servo")
							sessions, err := sessionManager.List()
							if err != nil {
								return fmt.Errorf("failed to list sessions: %w", err)
							}

							if len(sessions) == 0 {
								fmt.Println("No sessions found")
								return nil
							}

							fmt.Println("Sessions:")
							for _, sess := range sessions {
								status := ""
								if sess.Active {
									status = " (active)"
								}
								fmt.Printf("  • %s%s - %s\n", sess.Name, status, sess.Description)
							}
							return nil
						},
					},
					{
						Name:      "activate",
						Usage:     "Activate a session",
						ArgsUsage: "<session-name>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("session name required")
							}

							sessionName := c.Args().First()
							sessionManager := session.NewManager(".servo")
							err := sessionManager.Activate(sessionName)
							if err != nil {
								return fmt.Errorf("failed to activate session: %w", err)
							}

							fmt.Printf("✅ Activated session '%s'\n", sessionName)
							return nil
						},
					},
					{
						Name:      "delete",
						Usage:     "Delete a session",
						ArgsUsage: "<session-name>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("session name required")
							}

							sessionName := c.Args().First()
							sessionManager := session.NewManager(".servo")
							err := sessionManager.Delete(sessionName)
							if err != nil {
								return fmt.Errorf("failed to delete session: %w", err)
							}

							fmt.Printf("✅ Deleted session '%s'\n", sessionName)
							return nil
						},
					},
					{
						Name:      "rename",
						Usage:     "Rename a session",
						ArgsUsage: "<old-name> <new-name>",
						Action: func(c *cli.Context) error {
							if c.NArg() != 2 {
								return fmt.Errorf("both old and new session names required")
							}

							oldName := c.Args().Get(0)
							newName := c.Args().Get(1)
							sessionManager := session.NewManager(".servo")
							err := sessionManager.Rename(oldName, newName)
							if err != nil {
								return fmt.Errorf("failed to rename session: %w", err)
							}

							fmt.Printf("✅ Renamed session '%s' to '%s'\n", oldName, newName)
							return nil
						},
					},
				},
			},

			{
				Name:        "validate",
				Usage:       "Validate .servo file or source",
				Description: "Validate the structure and content of a .servo file",
				ArgsUsage:   "<source>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("source required")
					}

					validateCmd := commands.NewValidateCommand(parser, validator)
					return validateCmd.Execute([]string{c.Args().First()})
				},
			},

			{
				Name:        "client",
				Usage:       "Manage MCP client support for this project",
				Description: "Enable or disable MCP client support for the current project, and list available clients.",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "List available MCP clients",
						Action: func(c *cli.Context) error {
							clients := clientRegistry.List()
							fmt.Printf("Available Clients:\n")
							fmt.Printf("%-15s %-10s %s\n", "NAME", "INSTALLED", "DESCRIPTION")
							fmt.Printf("%-15s %-10s %s\n", "----", "---------", "-----------")
							for _, client := range clients {
								installed := "No"
								if client.IsInstalled() {
									installed = "Yes"
								}
								fmt.Printf("%-15s %-10s %s\n", client.Name(), installed, client.Description())
							}
							return nil
						},
					},
					{
						Name:      "enable",
						Usage:     "Enable support for one or more clients in the current project",
						ArgsUsage: "<client> [<client> ...]",
						Action: func(c *cli.Context) error {
							if !projectManager.IsProject() {
								return fmt.Errorf("not in a servo project directory")
							}
							if c.NArg() == 0 {
								return fmt.Errorf("at least one client name required (e.g. vscode, claude-code, cursor)")
							}
							var enabled []string
							var already []string
							var failed []string
							for i := 0; i < c.NArg(); i++ {
								clientName := c.Args().Get(i)
								err := projectManager.AddClient(clientName)
								if err != nil {
									if err.Error() == fmt.Sprintf("client '%s' already configured", clientName) {
										already = append(already, clientName)
									} else {
										failed = append(failed, fmt.Sprintf("%s (%v)", clientName, err))
									}
								} else {
									enabled = append(enabled, clientName)
								}
							}
							if len(enabled) > 0 {
								fmt.Printf("✅ Enabled client(s): %s\n", enabled)
							}
							if len(already) > 0 {
								fmt.Printf("⚠️  Already enabled: %s\n", already)
							}
							if len(failed) > 0 {
								fmt.Printf("❌ Failed to enable: %s\n", failed)
							}
							return nil
						},
					},
					{
						Name:      "disable",
						Usage:     "Disable support for one or more clients in the current project",
						ArgsUsage: "<client> [<client> ...]",
						Action: func(c *cli.Context) error {
							if !projectManager.IsProject() {
								return fmt.Errorf("not in a servo project directory")
							}
							if c.NArg() == 0 {
								return fmt.Errorf("at least one client name required (e.g. vscode, claude-code, cursor)")
							}
							var disabled []string
							var notfound []string
							var failed []string
							for i := 0; i < c.NArg(); i++ {
								clientName := c.Args().Get(i)
								err := projectManager.RemoveClient(clientName)
								if err != nil {
									if err.Error() == fmt.Sprintf("client '%s' not found in project", clientName) {
										notfound = append(notfound, clientName)
									} else {
										failed = append(failed, fmt.Sprintf("%s (%v)", clientName, err))
									}
								} else {
									disabled = append(disabled, clientName)
								}
							}
							if len(disabled) > 0 {
								fmt.Printf("✅ Disabled client(s): %s\n", disabled)
							}
							if len(notfound) > 0 {
								fmt.Printf("⚠️  Not enabled: %s\n", notfound)
							}
							if len(failed) > 0 {
								fmt.Printf("❌ Failed to disable: %s\n", failed)
							}
							return nil
						},
					},
				},
			},

			{
				Name:        "config",
				Usage:       "Manage project configuration",
				Description: "Get and set project-level configuration values",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "Show all project configuration",
						Action: func(c *cli.Context) error {
							if !projectManager.IsProject() {
								return fmt.Errorf("not in a servo project directory")
							}
							configCmd := commands.NewConfigCommand(projectManager)
							return configCmd.Execute([]string{"list"})
						},
					},
					{
						Name:      "get",
						Usage:     "Get a configuration value",
						ArgsUsage: "<key>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("configuration key required")
							}
							if !projectManager.IsProject() {
								return fmt.Errorf("not in a servo project directory")
							}
							configCmd := commands.NewConfigCommand(projectManager)
							return configCmd.Execute([]string{"get", c.Args().First()})
						},
					},
					{
						Name:      "set",
						Usage:     "Set a configuration value",
						ArgsUsage: "<key> <value>",
						Action: func(c *cli.Context) error {
							if c.NArg() < 2 {
								return fmt.Errorf("configuration key and value required")
							}
							if !projectManager.IsProject() {
								return fmt.Errorf("not in a servo project directory")
							}
							configCmd := commands.NewConfigCommand(projectManager)
							return configCmd.Execute([]string{"set", c.Args().Get(0), c.Args().Get(1)})
						},
					},
				},
			},

			{
				Name:        "secrets",
				Usage:       "Manage project secrets",
				Description: "Manage encrypted secrets for the current project",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "List all secrets in current project",
						Action: func(c *cli.Context) error {
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"list"})
						},
					},
					{
						Name:      "set",
						Usage:     "Set a secret value",
						ArgsUsage: "<key> <value>",
						Action: func(c *cli.Context) error {
							if c.NArg() < 2 {
								return fmt.Errorf("secret key and value required")
							}
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"set", c.Args().Get(0), c.Args().Get(1)})
						},
					},
					{
						Name:      "get",
						Usage:     "Get a secret value",
						ArgsUsage: "<key>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("secret key required")
							}
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"get", c.Args().First()})
						},
					},
					{
						Name:      "delete",
						Usage:     "Delete a secret",
						ArgsUsage: "<key>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("secret key required")
							}
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"delete", c.Args().First()})
						},
					},
					{
						Name:      "export",
						Usage:     "Export encrypted secrets to file",
						ArgsUsage: "<output-file>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("output file required")
							}
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"export", c.Args().First()})
						},
					},
					{
						Name:      "import",
						Usage:     "Import encrypted secrets from file",
						ArgsUsage: "<input-file>",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("input file required")
							}
							secretsCmd := commands.NewSecretsCommand(projectManager)
							return secretsCmd.Execute([]string{"import", c.Args().First()})
						},
					},
				},
			},
		},
	}

	return app, nil
}
