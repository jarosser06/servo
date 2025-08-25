package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/servo/servo/internal/project"
	"gopkg.in/yaml.v3"
)

// EnvCommand handles project-based environment variables management
type EnvCommand struct {
	projectManager *project.Manager
}

// NewEnvCommand creates a new project environment variables command
func NewEnvCommand(projectManager *project.Manager) *EnvCommand {
	return &EnvCommand{
		projectManager: projectManager,
	}
}

// EnvData represents the structure of environment variables file
type EnvData struct {
	Version string            `yaml:"version"`
	Env     map[string]string `yaml:"env"`
}

// Execute runs the env command
func (c *EnvCommand) Execute(args []string) error {
	if !c.projectManager.IsProject() {
		return fmt.Errorf("not in a servo project directory")
	}

	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list":
		return c.listEnvVars(subArgs)
	case "set":
		return c.setEnvVar(subArgs)
	case "get":
		return c.getEnvVar(subArgs)
	case "delete", "del":
		return c.deleteEnvVar(subArgs)
	case "export":
		return c.exportEnvVars(subArgs)
	case "import":
		return c.importEnvVars(subArgs)
	case "help", "--help", "-h":
		return c.showHelp()
	default:
		return fmt.Errorf("unknown env subcommand: %s", subcommand)
	}
}

func (c *EnvCommand) listEnvVars(args []string) error {
	envData, err := c.loadEnvData()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No environment variables configured")
			return nil
		}
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	if len(envData.Env) == 0 {
		fmt.Println("No environment variables configured")
		return nil
	}

	fmt.Printf("Environment Variables (%d configured):\n", len(envData.Env))
	for key, value := range envData.Env {
		fmt.Printf("  %s = %s\n", key, value)
	}

	return nil
}

func (c *EnvCommand) setEnvVar(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: servo env set <key> <value>")
	}

	key := args[0]
	value := args[1]

	envData, err := c.loadEnvData()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	if envData == nil {
		envData = &EnvData{
			Version: "1.0",
			Env:     make(map[string]string),
		}
	}

	envData.Env[key] = value

	if err := c.saveEnvData(envData); err != nil {
		return fmt.Errorf("failed to save environment variables: %w", err)
	}

	fmt.Printf("✅ Set environment variable: %s\n", key)
	return nil
}

func (c *EnvCommand) getEnvVar(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo env get <key>")
	}

	key := args[0]

	envData, err := c.loadEnvData()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("environment variable '%s' not found", key)
		}
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	value, exists := envData.Env[key]
	if !exists {
		return fmt.Errorf("environment variable '%s' not found", key)
	}

	fmt.Println(value)
	return nil
}

func (c *EnvCommand) deleteEnvVar(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo env delete <key>")
	}

	key := args[0]

	envData, err := c.loadEnvData()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("environment variable '%s' not found", key)
		}
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	if _, exists := envData.Env[key]; !exists {
		return fmt.Errorf("environment variable '%s' not found", key)
	}

	delete(envData.Env, key)

	if err := c.saveEnvData(envData); err != nil {
		return fmt.Errorf("failed to save environment variables: %w", err)
	}

	fmt.Printf("✅ Deleted environment variable: %s\n", key)
	return nil
}

func (c *EnvCommand) exportEnvVars(args []string) error {
	envData, err := c.loadEnvData()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("# No environment variables configured")
			return nil
		}
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	fmt.Println("# Environment variables export")
	for key, value := range envData.Env {
		fmt.Printf("export %s=%q\n", key, value)
	}

	return nil
}

func (c *EnvCommand) importEnvVars(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo env import <file>")
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var importData EnvData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse environment variables file: %w", err)
	}

	envData, err := c.loadEnvData()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing environment variables: %w", err)
	}

	if envData == nil {
		envData = &EnvData{
			Version: "1.0",
			Env:     make(map[string]string),
		}
	}

	// Merge imported env vars
	for key, value := range importData.Env {
		envData.Env[key] = value
	}

	if err := c.saveEnvData(envData); err != nil {
		return fmt.Errorf("failed to save environment variables: %w", err)
	}

	fmt.Printf("✅ Imported %d environment variables from %s\n", len(importData.Env), filename)
	return nil
}

func (c *EnvCommand) loadEnvData() (*EnvData, error) {
	servoDir := c.projectManager.GetServoDir()
	envPath := filepath.Join(servoDir, "env.yaml")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, err
	}

	var envData EnvData
	if err := yaml.Unmarshal(data, &envData); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables file: %w", err)
	}

	if envData.Env == nil {
		envData.Env = make(map[string]string)
	}

	return &envData, nil
}

func (c *EnvCommand) saveEnvData(envData *EnvData) error {
	servoDir := c.projectManager.GetServoDir()
	envPath := filepath.Join(servoDir, "env.yaml")

	// Ensure directory exists
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		return fmt.Errorf("failed to create servo directory: %w", err)
	}

	data, err := yaml.Marshal(envData)
	if err != nil {
		return fmt.Errorf("failed to marshal environment variables: %w", err)
	}

	if err := os.WriteFile(envPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment variables file: %w", err)
	}

	return nil
}

func (c *EnvCommand) showHelp() error {
	fmt.Printf(`env - Manage project environment variables

USAGE:
    servo env [SUBCOMMAND] [OPTIONS]

SUBCOMMANDS:
    list                    List all environment variables
    set <key> <value>       Set an environment variable
    get <key>               Get an environment variable value
    delete <key>            Delete an environment variable
    export                  Export environment variables as shell commands
    import <file>           Import environment variables from YAML file
    help                    Show this help message

DESCRIPTION:
    Environment variables are non-sensitive configuration values like URLs,
    API endpoints, and other settings that your MCP servers need. They are
    stored in plain text in .servo/env.yaml and can be safely committed to git.
    
    For sensitive values like passwords and API keys, use 'servo secrets' instead.

EXAMPLES:
    servo env set API_BASE_URL https://api.example.com
    servo env set MAX_RETRIES 3
    servo env list
    servo env get API_BASE_URL
    servo env delete MAX_RETRIES
    servo env export > .env
`)
	return nil
}
