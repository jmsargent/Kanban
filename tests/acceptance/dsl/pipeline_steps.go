package dsl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// projectRoot returns the absolute path to the project root directory.
// It resolves up from the tests/acceptance/dsl package location.
func projectRoot() (string, error) {
	// Walk up from cwd until we find a Makefile or cicd/ directory.
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	// Acceptance tests run with cwd = tests/acceptance, so resolve from there.
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "cicd")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("could not locate project root (no cicd/ directory found)")
}

// shellCmd runs a shell command in dir with the provided env additions.
// It captures combined output and returns exit code + output.
func shellCmd(dir string, extraEnv []string, timeout time.Duration, name string, args ...string) (string, int, error) {
	cmdCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	output := strings.TrimSpace(buf.String())
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		exitCode = 1
	}
	return output, exitCode, nil
}

// PipelineContext holds state for a pipeline-level test scenario.
// It is separate from the kanban binary Context — pipeline tests
// invoke real shell scripts and make targets, not the kanban binary.
type PipelineContext struct {
	t          interface {
		Helper()
		Fatal(...interface{})
		Fatalf(string, ...interface{})
		Cleanup(func())
		Logf(string, ...interface{})
		Skip(...interface{})
		Skipf(string, ...interface{})
	}
	projectRoot string
	lastOutput  string
	lastExit    int
}

// NewPipelineContext creates a PipelineContext rooted at the real project root.
func NewPipelineContext(t interface {
	Helper()
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Cleanup(func())
	Logf(string, ...interface{})
	Skip(...interface{})
	Skipf(string, ...interface{})
}) *PipelineContext {
	t.Helper()
	root, err := projectRoot()
	if err != nil {
		t.Fatalf("pipeline test setup: %v", err)
	}
	return &PipelineContext{t: t, projectRoot: root}
}

// PipelineOutput returns the combined stdout+stderr from the most recent pipeline command.
func (pc *PipelineContext) PipelineOutput() string { return pc.lastOutput }

// PipelineExitCode returns the exit code from the most recent pipeline command.
func (pc *PipelineContext) PipelineExitCode() int { return pc.lastExit }

// PipelineStep pairs a human-readable description with a pipeline-level action.
type PipelineStep struct {
	Description string
	Run         func(*PipelineContext) error
}

// PipelineGiven executes a pipeline setup step.
func PipelineGiven(pc *PipelineContext, step PipelineStep) {
	pc.t.Helper()
	if err := step.Run(pc); err != nil {
		pc.t.Fatalf("Given: %s: %v", step.Description, err)
	}
}

// PipelineWhen executes a pipeline action step.
func PipelineWhen(pc *PipelineContext, step PipelineStep) {
	pc.t.Helper()
	if err := step.Run(pc); err != nil {
		pc.t.Fatalf("When: %s: %v", step.Description, err)
	}
}

// PipelineThen executes a pipeline assertion step.
func PipelineThen(pc *PipelineContext, step PipelineStep) {
	pc.t.Helper()
	if err := step.Run(pc); err != nil {
		pc.t.Fatalf("Then: %s: %v", step.Description, err)
	}
}

// PipelineAnd is an alias for PipelineThen for readability.
func PipelineAnd(pc *PipelineContext, step PipelineStep) {
	pc.t.Helper()
	if err := step.Run(pc); err != nil {
		pc.t.Fatalf("And: %s: %v", step.Description, err)
	}
}

// ---- Setup steps ----

// TheProjectMakefile asserts the Makefile exists at the project root.
func TheProjectMakefile() PipelineStep {
	return PipelineStep{
		Description: "the project Makefile exists at the project root",
		Run: func(pc *PipelineContext) error {
			makefilePath := filepath.Join(pc.projectRoot, "Makefile")
			if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
				return fmt.Errorf("makefile not found at %s", makefilePath)
			}
			return nil
		},
	}
}

// ThePreCommitHookScript asserts cicd/pre-commit exists.
func ThePreCommitHookScript() PipelineStep {
	return PipelineStep{
		Description: "the pre-commit hook script exists at cicd/pre-commit",
		Run: func(pc *PipelineContext) error {
			p := filepath.Join(pc.projectRoot, "cicd", "pre-commit")
			if _, err := os.Stat(p); os.IsNotExist(err) {
				return fmt.Errorf("cicd/pre-commit not found at %s", p)
			}
			return nil
		},
	}
}

// TheCheckVersionsScript asserts cicd/check-versions.sh exists.
func TheCheckVersionsScript() PipelineStep {
	return PipelineStep{
		Description: "cicd/check-versions.sh exists",
		Run: func(pc *PipelineContext) error {
			p := filepath.Join(pc.projectRoot, "cicd", "check-versions.sh")
			if _, err := os.Stat(p); os.IsNotExist(err) {
				return fmt.Errorf("cicd/check-versions.sh not found at %s", p)
			}
			return nil
		},
	}
}

// TheKanbanBinaryIsBuilt asserts the ./kanban binary exists at the project root.
func TheKanbanBinaryIsBuilt() PipelineStep {
	return PipelineStep{
		Description: "the kanban binary has been built",
		Run: func(pc *PipelineContext) error {
			p := filepath.Join(pc.projectRoot, "kanban")
			if _, err := os.Stat(p); os.IsNotExist(err) {
				return fmt.Errorf("kanban binary not found at %s — run 'make validate' first", p)
			}
			return nil
		},
	}
}

// TheKanbanBinaryIsAbsent removes the kanban binary if it exists, so tests for
// "binary not built" scenarios start from a clean state.
func TheKanbanBinaryIsAbsent() PipelineStep {
	return PipelineStep{
		Description: "the kanban binary has not been built",
		Run: func(pc *PipelineContext) error {
			p := filepath.Join(pc.projectRoot, "kanban")
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove kanban binary: %w", err)
			}
			return nil
		},
	}
}

// ---- Action steps ----

// DeveloperRunsMakeTarget invokes `make <target>` at the project root and
// captures exit code and combined output into the PipelineContext.
func DeveloperRunsMakeTarget(target string) PipelineStep {
	return PipelineStep{
		Description: fmt.Sprintf("developer runs make %s", target),
		Run: func(pc *PipelineContext) error {
			output, exit, _ := shellCmd(pc.projectRoot, nil, 5*time.Minute, "make", target)
			pc.lastOutput = output
			pc.lastExit = exit
			return nil
		},
	}
}

// DeveloperRunsCheckVersions invokes `cicd/check-versions.sh` at the project root.
func DeveloperRunsCheckVersions() PipelineStep {
	return PipelineStep{
		Description: "developer runs cicd/check-versions.sh",
		Run: func(pc *PipelineContext) error {
			script := filepath.Join(pc.projectRoot, "cicd", "check-versions.sh")
			output, exit, _ := shellCmd(pc.projectRoot, nil, 30*time.Second, "bash", script)
			pc.lastOutput = output
			pc.lastExit = exit
			return nil
		},
	}
}

// DeveloperRunsPreCommitHook invokes the pre-commit script from cicd/pre-commit.
// This simulates what `git commit` would trigger.
func DeveloperRunsPreCommitHook() PipelineStep {
	return PipelineStep{
		Description: "developer triggers the pre-commit quality gate",
		Run: func(pc *PipelineContext) error {
			script := filepath.Join(pc.projectRoot, "cicd", "pre-commit")
			output, exit, _ := shellCmd(pc.projectRoot, nil, 5*time.Minute, "sh", script)
			pc.lastOutput = output
			pc.lastExit = exit
			return nil
		},
	}
}

// ---- Assertion steps ----

// PipelineExitsSuccessfully asserts the last pipeline command exited 0.
func PipelineExitsSuccessfully() PipelineStep {
	return PipelineStep{
		Description: "the command completes successfully",
		Run: func(pc *PipelineContext) error {
			if pc.lastExit != 0 {
				return fmt.Errorf("expected exit 0, got %d\nOutput:\n%s", pc.lastExit, pc.lastOutput)
			}
			return nil
		},
	}
}

// PipelineExitsWithFailure asserts the last pipeline command exited non-zero.
func PipelineExitsWithFailure() PipelineStep {
	return PipelineStep{
		Description: "the command exits with a failure code",
		Run: func(pc *PipelineContext) error {
			if pc.lastExit == 0 {
				return fmt.Errorf("expected non-zero exit, got 0\nOutput:\n%s", pc.lastOutput)
			}
			return nil
		},
	}
}

// PipelineOutputContains asserts the last output contains the given text.
func PipelineOutputContains(text string) PipelineStep {
	return PipelineStep{
		Description: fmt.Sprintf("output contains %q", text),
		Run: func(pc *PipelineContext) error {
			if !strings.Contains(pc.lastOutput, text) {
				return fmt.Errorf("expected output to contain %q\nOutput:\n%s", text, pc.lastOutput)
			}
			return nil
		},
	}
}

// PipelineOutputDoesNotContain asserts the last output does not contain text.
func PipelineOutputDoesNotContain(text string) PipelineStep {
	return PipelineStep{
		Description: fmt.Sprintf("output does not contain %q", text),
		Run: func(pc *PipelineContext) error {
			if strings.Contains(pc.lastOutput, text) {
				return fmt.Errorf("expected output NOT to contain %q\nOutput:\n%s", text, pc.lastOutput)
			}
			return nil
		},
	}
}

// MakefileContainsTarget asserts the Makefile has a target named target.
func MakefileContainsTarget(target string) PipelineStep {
	return PipelineStep{
		Description: fmt.Sprintf("Makefile contains target %q", target),
		Run: func(pc *PipelineContext) error {
			content, err := os.ReadFile(filepath.Join(pc.projectRoot, "Makefile"))
			if err != nil {
				return fmt.Errorf("read Makefile: %w", err)
			}
			// A Makefile target begins at column 0 followed by ':'
			needle := target + ":"
			if !strings.Contains(string(content), needle) {
				return fmt.Errorf("makefile does not contain target %q", target)
			}
			return nil
		},
	}
}

// PreCommitHookContainsStep asserts the pre-commit script contains text.
func PreCommitHookContainsStep(text string) PipelineStep {
	return PipelineStep{
		Description: fmt.Sprintf("pre-commit hook contains %q", text),
		Run: func(pc *PipelineContext) error {
			content, err := os.ReadFile(filepath.Join(pc.projectRoot, "cicd", "pre-commit"))
			if err != nil {
				return fmt.Errorf("read cicd/pre-commit: %w", err)
			}
			if !strings.Contains(string(content), text) {
				return fmt.Errorf("cicd/pre-commit does not contain %q", text)
			}
			return nil
		},
	}
}

// PreCommitHookDoesNotContainCommentedArchLint asserts go-arch-lint is not commented out.
func PreCommitHookDoesNotContainCommentedArchLint() PipelineStep {
	return PipelineStep{
		Description: "go-arch-lint step in pre-commit hook is not commented out",
		Run: func(pc *PipelineContext) error {
			content, err := os.ReadFile(filepath.Join(pc.projectRoot, "cicd", "pre-commit"))
			if err != nil {
				return fmt.Errorf("read cicd/pre-commit: %w", err)
			}
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				// A commented-out arch-lint invocation looks like: #if ! go-arch-lint
				if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "go-arch-lint") {
					return fmt.Errorf("go-arch-lint step is still commented out in cicd/pre-commit:\n  %s", line)
				}
			}
			return nil
		},
	}
}

// CheckVersionsIsFirstStepInPreCommit asserts check-versions.sh appears before
// go test in cicd/pre-commit.
func CheckVersionsIsFirstStepInPreCommit() PipelineStep {
	return PipelineStep{
		Description: "check-versions.sh runs as step 0 before go test in the pre-commit hook",
		Run: func(pc *PipelineContext) error {
			content, err := os.ReadFile(filepath.Join(pc.projectRoot, "cicd", "pre-commit"))
			if err != nil {
				return fmt.Errorf("read cicd/pre-commit: %w", err)
			}
			checkIdx := strings.Index(string(content), "check-versions")
			goTestIdx := strings.Index(string(content), "go test")
			if checkIdx == -1 {
				return fmt.Errorf("cicd/pre-commit does not invoke check-versions.sh")
			}
			if goTestIdx == -1 {
				return fmt.Errorf("cicd/pre-commit does not invoke go test")
			}
			if checkIdx > goTestIdx {
				return fmt.Errorf("check-versions.sh appears after go test in cicd/pre-commit (check at byte %d, go test at byte %d)", checkIdx, goTestIdx)
			}
			return nil
		},
	}
}

// AllToolVersionsMatchPipeline skips the test when check-versions.sh exits
// non-zero, meaning one or more local tools do not match the versions declared
// in cicd/config.yml. The test is skipped (not failed) because tool installation
// is a precondition outside the test's control — not a product defect.
func AllToolVersionsMatchPipeline() PipelineStep {
	return PipelineStep{
		Description: "all local tool versions match the pipeline configuration",
		Run: func(pc *PipelineContext) error {
			script := filepath.Join(pc.projectRoot, "cicd", "check-versions.sh")
			output, exit, _ := shellCmd(pc.projectRoot, nil, 30*time.Second, "bash", script)
			if exit != 0 {
				pc.t.Skipf("skipping: local tool versions do not match pipeline — run cicd/check-versions.sh to see mismatches:\n%s", output)
			}
			return nil
		},
	}
}
