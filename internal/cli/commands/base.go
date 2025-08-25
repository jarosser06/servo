package commands

import (
	"os"

	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/internal/config"
	"github.com/servo/servo/internal/mcp"
	"github.com/servo/servo/internal/project"
	"github.com/servo/servo/internal/registry"
	"github.com/servo/servo/internal/session"
)

// BaseCommandDependencies holds common dependencies needed by commands
type BaseCommandDependencies struct {
	ProjectManager *project.Manager
	SessionManager *session.Manager
	ConfigManager  *config.ConfigGeneratorManager
	ClientRegistry *client.Registry
	Parser         *mcp.Parser
	Validator      *mcp.Validator
	ServoDir       string
}

// NewBaseCommandDependencies creates a new set of base command dependencies
func NewBaseCommandDependencies() *BaseCommandDependencies {
	servoDir := os.Getenv("SERVO_DIR")
	if servoDir == "" {
		servoDir = ".servo" // Use project-local servo directory
	}

	return &BaseCommandDependencies{
		ProjectManager: project.NewManager(),
		SessionManager: session.NewManager(servoDir),
		ConfigManager:  config.NewConfigGeneratorManager(servoDir),
		ClientRegistry: registry.GetDefaultRegistry(),
		Parser:         mcp.NewParser(),
		Validator:      mcp.NewValidator(),
		ServoDir:       servoDir,
	}
}

// NewBaseDependenciesWithParsers creates base dependencies with custom parser and validator
func NewBaseDependenciesWithParsers(parser *mcp.Parser, validator *mcp.Validator) *BaseCommandDependencies {
	servoDir := os.Getenv("SERVO_DIR")
	if servoDir == "" {
		servoDir = ".servo" // Use project-local servo directory
	}

	return &BaseCommandDependencies{
		ProjectManager: project.NewManager(),
		SessionManager: session.NewManager(servoDir),
		ConfigManager:  config.NewConfigGeneratorManager(servoDir),
		ClientRegistry: registry.GetDefaultRegistry(),
		Parser:         parser,
		Validator:      validator,
		ServoDir:       servoDir,
	}
}