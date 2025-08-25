package config

// Generator defines the interface for configuration generators
type Generator interface {
	Generate() error
}

// ConfigGeneratorManager coordinates infrastructure configuration generators
type ConfigGeneratorManager struct {
	devcontainerGen  *DevcontainerGenerator
	dockerComposeGen *DockerComposeGenerator
}

// NewConfigGeneratorManager creates a new configuration generator manager
func NewConfigGeneratorManager(servoDir string) *ConfigGeneratorManager {
	return &ConfigGeneratorManager{
		devcontainerGen:  NewDevcontainerGenerator(servoDir),
		dockerComposeGen: NewDockerComposeGenerator(servoDir),
	}
}

// GenerateDevcontainer generates devcontainer configuration
func (m *ConfigGeneratorManager) GenerateDevcontainer() error {
	return m.devcontainerGen.Generate()
}

// GenerateDockerCompose generates docker-compose configuration
func (m *ConfigGeneratorManager) GenerateDockerCompose() error {
	return m.dockerComposeGen.Generate()
}

// GenerateAll generates all infrastructure configuration files
func (m *ConfigGeneratorManager) GenerateAll() error {
	if err := m.GenerateDevcontainer(); err != nil {
		return err
	}

	if err := m.GenerateDockerCompose(); err != nil {
		return err
	}

	return nil
}
