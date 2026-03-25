package acceptance

import (
	"os"
	"path/filepath"
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
	"gopkg.in/yaml.v3"
)

type TestingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

type PipelineDriver struct {
	t    TestingT
	root string
}

func NewPipelineDriver(t *testing.T) *PipelineDriver {
	t.Helper()
	root, err := dsl.ProjectRoot()
	if err != nil {
		t.Fatalf("locate project root: %v", err)
	}
	return &PipelineDriver{t: t, root: root}
}

func (d *PipelineDriver) ReadCIConfig(path string) string {
	d.t.Helper()
	if path == "" {
		path = filepath.Join(d.root, "cicd", "config.yml")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		d.t.Fatalf("read CI config %s: %v", path, err)
	}
	return string(content)
}

func (d *PipelineDriver) ReadCommands(path string) []string {
	d.t.Helper()
	raw := d.ReadCIConfig(path)

	var config struct {
		Commands map[string]any `yaml:"commands"`
	}
	if err := yaml.Unmarshal([]byte(raw), &config); err != nil {
		d.t.Fatalf("parse CI config: %v", err)
	}

	names := make([]string, 0, len(config.Commands))
	for name := range config.Commands {
		names = append(names, name)
	}
	return names
}

func (d *PipelineDriver) ReadMakefile() string {
	d.t.Helper()
	makefilePath := filepath.Join(d.root, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		d.t.Fatalf("Makefile not found at %s — create it to proceed: %v", makefilePath, err)
	}
	return string(content)
}
