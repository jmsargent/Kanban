package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// ConfigRepository implements ports.ConfigRepository using a YAML file at
// {repoRoot}/.kanban/config.
type ConfigRepository struct{}

// NewConfigRepository constructs a ConfigRepository.
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// configFilePath returns the path to the config file.
func configFilePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".kanban", "config")
}

// configYAML is the on-disk serialisation shape for Config.
type configYAML struct {
	Columns []struct {
		Name  string `yaml:"name"`
		Label string `yaml:"label"`
	} `yaml:"columns"`
	CITaskPattern string `yaml:"ci_task_pattern"`
}

// Read loads the board configuration from {repoRoot}/.kanban/config.
func (r *ConfigRepository) Read(repoRoot string) (ports.Config, error) {
	path := configFilePath(repoRoot)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ports.Config{}, ports.ErrNotInitialised
	}
	if err != nil {
		return ports.Config{}, fmt.Errorf("read config file: %w", err)
	}

	var raw configYAML
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return ports.Config{}, fmt.Errorf("parse config: %w", err)
	}

	columns := make([]domain.Column, 0, len(raw.Columns))
	for _, c := range raw.Columns {
		columns = append(columns, domain.Column{Name: c.Name, Label: c.Label})
	}

	return ports.Config{
		Columns:       columns,
		CITaskPattern: raw.CITaskPattern,
	}, nil
}

// Write stores the board configuration atomically to {repoRoot}/.kanban/config.
func (r *ConfigRepository) Write(repoRoot string, config ports.Config) error {
	kanbanDir := filepath.Join(repoRoot, ".kanban")
	if err := os.MkdirAll(kanbanDir, 0o755); err != nil {
		return fmt.Errorf("create kanban dir: %w", err)
	}

	raw := configYAML{CITaskPattern: config.CITaskPattern}
	raw.Columns = make([]struct {
		Name  string `yaml:"name"`
		Label string `yaml:"label"`
	}, len(config.Columns))
	for i, col := range config.Columns {
		raw.Columns[i].Name = col.Name
		raw.Columns[i].Label = col.Label
	}

	content, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	finalPath := configFilePath(repoRoot)
	return atomicOverwrite(finalPath, content)
}

// ensure compile-time interface compliance
var _ ports.ConfigRepository = (*ConfigRepository)(nil)
