package acceptance

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestPipelineDriver_ReadCIConfig_ReturnsContent(t *testing.T) {
	driver := NewPipelineDriver(t, testdataPath("example-circleci-config.yml"))
	config := driver.ReadCIConfig()

	if !strings.Contains(config, "version: 2.1") {
		t.Error("expected config to contain 'version: 2.1'")
	}
}

func TestPipelineDriver_ReadCIConfig_ContainsInstallToolsCommand(t *testing.T) {
	driver := NewPipelineDriver(t, testdataPath("example-circleci-config.yml"))
	config := driver.ReadCIConfig()

	if !strings.Contains(config, "install-tools:") {
		t.Error("expected config to contain 'install-tools:' command")
	}

	if !strings.Contains(config, "make install-tools") {
		t.Error("expected install-tools command to delegate to 'make install-tools'")
	}
}

func TestPipelineDriver_ReadCIConfig_ShouldFailOnMissingFile(t *testing.T) {
	var ft fakeTesting
	driver := &PipelineDriver{t: &ft, root: ".", ciConfigPath: "/nonexistent/path/config.yml"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		driver.ReadCIConfig()
	}()
	<-done

	if !ft.failed {
		t.Error("expected ReadCIConfig to fatal on missing file")
	}
}

func TestPipelineDriver_ReadCommandsShouldReturnSliceOfStringSlices(t *testing.T) {
	driver := NewPipelineDriver(t, testdataPath("example-circleci-config.yml"))
	commands := driver.ReadCommands()

	if len(commands) == 0 {
		t.Fatal("expected at least one command, got none")
	}

	// The example config job has run steps: "make test" and "make build".
	found := false
	for _, cmd := range commands {
		for _, command := range cmd {
			if command == "make test" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected commands to contain 'make test', got %v", commands)
	}
}

// fakeTesting captures Fatalf calls without requiring a real *testing.T.
type fakeTesting struct {
	failed bool
}

func (f *fakeTesting) Fatalf(format string, args ...any) {
	f.failed = true
	runtime.Goexit()
}

func (f *fakeTesting) Helper() {}
