package filesystem_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmsargent/kanban/internal/adapters/filesystem"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

func readDir(root string) ([]string, error) {
	kanbanDir := filepath.Join(root, ".kanban")
	entries, err := os.ReadDir(kanbanDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

func defaultConfig() ports.Config {
	return ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
		CITaskPattern: "TASK-[0-9]+",
	}
}

// Behavior 1: Write then Read round-trips all config fields
func TestConfigRepository_WriteAndRead_RoundTrips(t *testing.T) {
	repoRoot := t.TempDir()
	repo := filesystem.NewConfigRepository()
	cfg := defaultConfig()

	if err := repo.Write(repoRoot, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := repo.Read(repoRoot)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.CITaskPattern != cfg.CITaskPattern {
		t.Errorf("CITaskPattern: got %q, want %q", got.CITaskPattern, cfg.CITaskPattern)
	}
	if len(got.Columns) != len(cfg.Columns) {
		t.Fatalf("Columns length: got %d, want %d", len(got.Columns), len(cfg.Columns))
	}
	for i, col := range cfg.Columns {
		if got.Columns[i].Name != col.Name {
			t.Errorf("Column[%d].Name: got %q, want %q", i, got.Columns[i].Name, col.Name)
		}
		if got.Columns[i].Label != col.Label {
			t.Errorf("Column[%d].Label: got %q, want %q", i, got.Columns[i].Label, col.Label)
		}
	}
}

// Behavior 2: Read returns ErrNotInitialised when config file is absent
func TestConfigRepository_Read_ReturnsErrNotInitialised_WhenMissing(t *testing.T) {
	repoRoot := t.TempDir()
	repo := filesystem.NewConfigRepository()

	_, err := repo.Read(repoRoot)
	if !errors.Is(err, ports.ErrNotInitialised) {
		t.Errorf("expected ErrNotInitialised, got %v", err)
	}
}

// Behavior 1b: Write is atomic — no temp files remain
func TestConfigRepository_Write_NoTempFilesRemain(t *testing.T) {
	repoRoot := t.TempDir()
	repo := filesystem.NewConfigRepository()

	if err := repo.Write(repoRoot, defaultConfig()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := readDir(repoRoot)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, name := range entries {
		if len(name) > 4 && name[len(name)-4:] == ".tmp" {
			t.Errorf("temp file remains: %s", name)
		}
	}
}
