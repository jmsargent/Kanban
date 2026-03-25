package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
	"gopkg.in/yaml.v3"
)

type TestingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

type PipelineDriver struct {
	t              TestingT
	root           string
	ciConfigPath   string
	ciConfigCached *string
}

func NewPipelineDriver(t *testing.T, ciConfigPath ...string) *PipelineDriver {
	t.Helper()
	root, err := dsl.ProjectRoot()
	if err != nil {
		t.Fatalf("locate project root: %v", err)
	}
	path := filepath.Join(root, "cicd", "config.yml")
	if len(ciConfigPath) > 0 && ciConfigPath[0] != "" {
		path = ciConfigPath[0]
	}
	return &PipelineDriver{t: t, root: root, ciConfigPath: path}
}

func (d *PipelineDriver) ReadCIConfig() string {
	d.t.Helper()
	if d.ciConfigCached != nil {
		return *d.ciConfigCached
	}
	content, err := os.ReadFile(d.ciConfigPath)
	if err != nil {
		d.t.Fatalf("read CI config %s: %v", d.ciConfigPath, err)
	}
	s := string(content)
	d.ciConfigCached = &s
	return s
}

type CircleCIConfigCommand = map[string]string

func (d *PipelineDriver) ReadCommands() []CircleCIConfigCommand {
	d.t.Helper()
	raw := d.ReadCIConfig()

	var config struct {
		Jobs map[string]struct {
			Steps []any `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(raw), &config); err != nil {
		d.t.Fatalf("parse CI config: %v", err)
	}

	var commands []CircleCIConfigCommand
	for _, job := range config.Jobs {
		for _, step := range job.Steps {
			m, ok := step.(map[string]any)
			if !ok {
				continue
			}
			run, ok := m["run"]
			if !ok {
				continue
			}
			r, ok := run.(map[string]any)
			if !ok {
				continue
			}
			name, _ := r["name"].(string)
			command, _ := r["command"].(string)
			if command != "" {
				commands = append(commands, CircleCIConfigCommand{name: command})
			}
		}
	}
	return commands
}

func (d *PipelineDriver) ReadPipelineMakeTargets(excludeJobs []string, excludeTargets []string) []string {
	d.t.Helper()
	raw := d.ReadCIConfig()

	var config struct {
		Jobs      map[string]struct {
			Steps []any `yaml:"steps"`
		} `yaml:"jobs"`
		Workflows map[string]struct {
			Jobs []any `yaml:"jobs"`
		} `yaml:"workflows"`
	}
	if err := yaml.Unmarshal([]byte(raw), &config); err != nil {
		d.t.Fatalf("parse CI config: %v", err)
	}

	excluded := make(map[string]bool)
	for _, j := range excludeJobs {
		excluded[j] = true
	}
	excludedTargets := make(map[string]bool)
	for _, t := range excludeTargets {
		excludedTargets[t] = true
	}

	// Extract job order from the first workflow.
	var jobOrder []string
	for _, wf := range config.Workflows {
		for _, entry := range wf.Jobs {
			switch e := entry.(type) {
			case string:
				jobOrder = append(jobOrder, e)
			case map[string]any:
				for name := range e {
					jobOrder = append(jobOrder, name)
				}
			}
		}
		break
	}

	// Collect make targets in job order.
	var targets []string
	for _, jobName := range jobOrder {
		if excluded[jobName] {
			continue
		}
		job, ok := config.Jobs[jobName]
		if !ok {
			continue
		}
		for _, step := range job.Steps {
			m, ok := step.(map[string]any)
			if !ok {
				continue
			}
			run, ok := m["run"]
			if !ok {
				continue
			}
			r, ok := run.(map[string]any)
			if !ok {
				continue
			}
			command, _ := r["command"].(string)
			if len(command) > 5 && command[:5] == "make " {
				target := command[5:]
				if !excludedTargets[target] {
					targets = append(targets, target)
				}
			}
		}
	}
	return targets
}

func (d *PipelineDriver) ReadMakeTargetSteps(target string) []string {
	d.t.Helper()
	makefile := d.ReadMakefile()

	var steps []string
	inTarget := false
	for _, line := range strings.Split(makefile, "\n") {
		if !inTarget {
			if strings.HasPrefix(line, target+":") {
				inTarget = true
			}
			continue
		}
		if line == "" || (len(line) > 0 && line[0] != '\t' && line[0] != ' ') {
			break
		}
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "@")
		if len(trimmed) > 5 && trimmed[:5] == "make " {
			steps = append(steps, trimmed[5:])
		}
	}
	return steps
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
