package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/servo/servo/internal/project"
	"gopkg.in/yaml.v3"
)

// SecretsCommand handles project-based secrets management
type SecretsCommand struct {
	projectManager *project.Manager
}

// NewSecretsCommand creates a new project secrets command
func NewSecretsCommand(projectManager *project.Manager) *SecretsCommand {
	return &SecretsCommand{
		projectManager: projectManager,
	}
}

// SecretsData represents the structure of base64-encoded secrets file
type SecretsData struct {
	Version string            `yaml:"version"`
	Secrets map[string]string `yaml:"secrets"`
}

// Execute runs the secrets command
func (c *SecretsCommand) Execute(args []string) error {
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
		return c.listSecrets(subArgs)
	case "set":
		return c.setSecret(subArgs)
	case "get":
		return c.getSecret(subArgs)
	case "delete":
		return c.deleteSecret(subArgs)
	case "export":
		return c.exportSecrets(subArgs)
	case "import":
		return c.importSecrets(subArgs)
	default:
		return fmt.Errorf("unknown secrets subcommand: %s", subcommand)
	}
}

func (c *SecretsCommand) listSecrets(args []string) error {
	data, err := c.loadSecretsData()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	if len(data.Secrets) == 0 {
		fmt.Println("No secrets configured for this project")
		return nil
	}

	fmt.Println("Project secrets:")
	for key := range data.Secrets {
		fmt.Printf("  • %s\n", key)
	}

	return nil
}

func (c *SecretsCommand) setSecret(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: servo secrets set <key> <value>")
	}

	key := args[0]
	value := args[1]

	data, err := c.loadSecretsData()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	if data.Secrets == nil {
		data.Secrets = make(map[string]string)
	}

	data.Secrets[key] = value

	if err := c.saveSecretsData(data); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	fmt.Printf("✅ Secret '%s' set successfully\n", key)
	return nil
}

func (c *SecretsCommand) getSecret(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo secrets get <key>")
	}

	key := args[0]

	data, err := c.loadSecretsData()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	if data.Secrets == nil {
		return fmt.Errorf("secret '%s' not found", key)
	}

	value, exists := data.Secrets[key]
	if !exists {
		return fmt.Errorf("secret '%s' not found", key)
	}

	fmt.Println(value)
	return nil
}

func (c *SecretsCommand) deleteSecret(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo secrets delete <key>")
	}

	key := args[0]

	data, err := c.loadSecretsData()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	if data.Secrets != nil {
		delete(data.Secrets, key)
	}

	if err := c.saveSecretsData(data); err != nil {
		return fmt.Errorf("failed to save secrets: %w", err)
	}

	fmt.Printf("✅ Secret '%s' deleted successfully\n", key)
	return nil
}

func (c *SecretsCommand) exportSecrets(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo secrets export <output-file>")
	}

	outputPath := args[0]

	data, err := c.loadSecretsData()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	// Export the YAML format
	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	fmt.Printf("✅ Secrets exported to %s\n", outputPath)
	return nil
}

func (c *SecretsCommand) importSecrets(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: servo secrets import <input-file>")
	}

	inputPath := args[0]

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var importData SecretsData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse import data: %w", err)
	}

	secretsPath := c.getSecretsPath()

	// Backup existing secrets
	if _, err := os.Stat(secretsPath); err == nil {
		backupPath := secretsPath + ".backup"
		if err := os.Rename(secretsPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing secrets: %w", err)
		}
	}

	// Save the imported data using our standard save method
	if err := c.saveSecretsData(&importData); err != nil {
		return fmt.Errorf("failed to save imported secrets: %w", err)
	}

	fmt.Printf("✅ Secrets imported from %s\n", inputPath)
	return nil
}

func (c *SecretsCommand) showHelp() error {
	fmt.Printf(`Project secrets management

USAGE:
    servo secrets <SUBCOMMAND> [OPTIONS]

SUBCOMMANDS:
    list                       List all secrets in current project
    set <key> <value>          Set a secret value
    get <key>                  Get a secret value
    delete <key>               Delete a secret
    export <file>              Export secrets to file
    import <file>              Import secrets from file

EXAMPLES:
    servo secrets set database_url "postgresql://user:pass@localhost/db"
    servo secrets set api_key "your-secret-api-key"
    servo secrets get database_url
    servo secrets list
    servo secrets delete api_key
    
    # Backup and restore
    servo secrets export backup.yaml
    servo secrets import backup.yaml

SECURITY:
    • All secrets are base64 encoded for basic obscurity
    • Secrets are stored in .servo/secrets.yaml (project-local)
    • File permissions are set to 0600 (owner read/write only)
    • Export/import maintains the same format

ENVIRONMENT VARIABLES:
    SERVO_NON_INTERACTIVE    Set to prevent interactive prompts in scripts
`)
	return nil
}

// Helper methods for direct secrets management

func (c *SecretsCommand) getSecretsPath() string {
	return filepath.Join(c.projectManager.GetServoDir(), "secrets.yaml")
}

func (c *SecretsCommand) loadSecretsData() (*SecretsData, error) {
	secretsPath := c.getSecretsPath()

	// If secrets file doesn't exist, return empty data
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return &SecretsData{
			Version: "1.0",
			Secrets: make(map[string]string),
		}, nil
	}

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	var secretsData SecretsData
	if err := yaml.Unmarshal(data, &secretsData); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	// Decode base64 secrets
	decodedSecrets := make(map[string]string)
	for key, encodedValue := range secretsData.Secrets {
		decodedValue, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			// If decoding fails, assume it's plain text (for migration)
			decodedSecrets[key] = encodedValue
		} else {
			decodedSecrets[key] = string(decodedValue)
		}
	}

	secretsData.Secrets = decodedSecrets
	return &secretsData, nil
}

func (c *SecretsCommand) saveSecretsData(data *SecretsData) error {
	// Ensure directory exists
	servoDir := c.projectManager.GetServoDir()
	if err := os.MkdirAll(servoDir, 0755); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	secretsPath := c.getSecretsPath()

	// Encode secrets with base64 for basic obscurity
	encodedSecrets := make(map[string]string)
	for key, value := range data.Secrets {
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		encodedSecrets[key] = encodedValue
	}

	// Prepare final data structure for file
	fileData := SecretsData{
		Version: "1.0",
		Secrets: encodedSecrets,
	}

	output, err := yaml.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("failed to marshal file data: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(secretsPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	return nil
}
