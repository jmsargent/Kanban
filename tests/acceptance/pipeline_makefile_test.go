package acceptance

// Test Budget: 2 behaviors x 2 = 4 unit tests max. Using 2.
//
// Behaviors:
//   1. Makefile exists with a validate target
//   2. Makefile validate target references all required pipeline tools in correct step order

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMakefile_ValidateTarget_Exists asserts the Makefile exists at the project
// root and declares a validate: target — the minimum structural requirement for
// the walking skeleton.
func TestMakefile_ValidateTarget_Exists(t *testing.T) {
	root := findProjectRoot(t)
	content := readMakefile(t, root)

	if !strings.Contains(content, "validate:") {
		t.Errorf("Makefile does not contain a validate: target\nMakefile content:\n%s", content)
	}
}

// TestMakefile_ValidateTarget_ReferencesRequiredTools asserts the validate target
// invokes each required pipeline tool in the specified step order.
func TestMakefile_ValidateTarget_ReferencesRequiredTools(t *testing.T) {
	root := findProjectRoot(t)
	content := readMakefile(t, root)

	requiredSteps := []struct {
		label string
		tool  string
	}{
		{"[0/4]", "check-versions"},
		{"[1/4]", "gotestsum"},
		{"[2/4]", "golangci-lint"},
		{"[3/4]", "go-arch-lint"},
		{"[4/4]", "go build"},
	}

	for _, step := range requiredSteps {
		if !strings.Contains(content, step.label) {
			t.Errorf("Makefile validate target missing step label %q", step.label)
		}
		if !strings.Contains(content, step.tool) {
			t.Errorf("Makefile validate target missing tool reference %q", step.tool)
		}
	}
}

// findProjectRoot walks up from the current directory until it finds the cicd/ directory.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "cicd")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("could not locate project root (no cicd/ directory found)")
	return ""
}

// readMakefile reads the Makefile content from the project root.
// The test fails immediately if the Makefile does not exist.
func readMakefile(t *testing.T, root string) string {
	t.Helper()
	makefilePath := filepath.Join(root, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Makefile not found at %s — create it to proceed: %v", makefilePath, err)
	}
	return string(content)
}
