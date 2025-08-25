package registry

import (
	claude_code "github.com/servo/servo/clients/claude_code"
	cursor "github.com/servo/servo/clients/cursor"
	vscode "github.com/servo/servo/clients/vscode"
	"github.com/servo/servo/internal/client"
)

// GetDefaultRegistry creates a new registry with all built-in clients pre-registered
func GetDefaultRegistry() *client.Registry {
	registry := client.NewRegistry()
	
	// Register all built-in clients
	registry.Register(claude_code.New())
	registry.Register(vscode.New())
	registry.Register(cursor.New())
	
	return registry
}