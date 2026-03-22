package repofiles_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// repoRoot walks up from this file's location until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine source file path")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

func TestLicenseExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "LICENSE")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("LICENSE not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("LICENSE is empty")
	}
}

func TestLicenseContainsCurrentYear(t *testing.T) {
	path := filepath.Join(repoRoot(t), "LICENSE")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("LICENSE not found: %v", err)
	}
	currentYear := fmt.Sprintf("%d", time.Now().Year())
	if !strings.Contains(string(content), currentYear) {
		t.Fatalf("LICENSE does not contain current year %s", currentYear)
	}
}

func TestReadmeExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), "README.md")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("README.md not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("README.md is empty")
	}
}

// TestNoBarGoTest asserts that no code or script file in the repository
// invokes bare "go test" — gotestsum must be used everywhere.
//
// Files scanned: .go, .sh, .yml, .yaml, Makefile, and extensionless files
// under cicd/ (shell scripts without a .sh suffix).
func TestNoBarGoTest(t *testing.T) {
	root := repoRoot(t)

	codeExts := map[string]bool{
		".go":   true,
		".sh":   true,
		".yml":  true,
		".yaml": true,
	}

	skipDirs := map[string]bool{
		".git":    true,
		"vendor":  true,
		".kanban": true,
	}

	var violations []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(root, path)

		// Exclude this file — it contains "go test" as a search term.
		if rel == filepath.Join("tests", "repofiles", "repofiles_test.go") {
			return nil
		}

		name := info.Name()
		ext := strings.ToLower(filepath.Ext(name))

		isCode := codeExts[ext] || name == "Makefile"
		// Include extensionless files under cicd/ (scripts without .sh suffix).
		if !isCode && ext == "" && strings.HasPrefix(rel, "cicd"+string(filepath.Separator)) {
			isCode = true
		}
		if !isCode {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.Contains(string(content), "go test") {
			violations = append(violations, rel)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("found 'go test' in %d file(s) — use gotestsum instead:\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}
