package client

import (
	"fmt"
	"sync"

	"github.com/servo/servo/pkg"
)

// Registry implements ClientRegistry interface
type Registry struct {
	clients map[string]pkg.Client
	mutex   sync.RWMutex
}

// NewRegistry creates a new client registry
func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]pkg.Client),
	}
}

// Register adds a client to the registry
func (r *Registry) Register(client pkg.Client) error {
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}

	name := client.Name()
	if name == "" {
		return fmt.Errorf("client name cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.clients[name]; exists {
		return fmt.Errorf("client %q is already registered", name)
	}

	r.clients[name] = client
	return nil
}

// Unregister removes a client from the registry
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.clients[name]; !exists {
		return fmt.Errorf("client %q is not registered", name)
	}

	delete(r.clients, name)
	return nil
}

// Get retrieves a client by name
func (r *Registry) Get(name string) (pkg.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	client, exists := r.clients[name]
	if !exists {
		return nil, fmt.Errorf("client %q is not registered", name)
	}

	return client, nil
}

// List returns all registered clients
func (r *Registry) List() []pkg.Client {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	clients := make([]pkg.Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

// Detect discovers installed MCP clients automatically
func (r *Registry) Detect() ([]pkg.Client, error) {
	var detected []pkg.Client

	for _, client := range r.List() {
		if client.IsInstalled() {
			detected = append(detected, client)
		}
	}

	return detected, nil
}

// GetByNames returns clients by their names
func (r *Registry) GetByNames(names []string) ([]pkg.Client, error) {
	var clients []pkg.Client
	var errors []error

	for _, name := range names {
		client, err := r.Get(name)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		clients = append(clients, client)
	}

	if len(errors) > 0 {
		return clients, fmt.Errorf("failed to get some clients: %v", errors)
	}

	return clients, nil
}

// GetSupportingScope returns clients that support the given scope
func (r *Registry) GetSupportingScope(scope string) []pkg.Client {
	var supporting []pkg.Client

	for _, client := range r.List() {
		if client.ValidateScope(scope) == nil {
			supporting = append(supporting, client)
		}
	}

	return supporting
}

// GetInstalledClients returns only clients that are actually installed
func (r *Registry) GetInstalledClients() []pkg.Client {
	var installed []pkg.Client

	for _, client := range r.List() {
		if client.IsInstalled() {
			installed = append(installed, client)
		}
	}

	return installed
}
