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
