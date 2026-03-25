package acceptance

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestPipelineDriver_ReadCIConfig_ReturnsContent(t *testing.T) {
	driver := NewPipelineDriver(t)
	config := driver.ReadCIConfig(testdataPath("example-circleci-config.yml"))

	if !strings.Contains(config, "version: 2.1") {
		t.Error("expected config to contain 'version: 2.1'")
	}
}

func TestPipelineDriver_ReadCIConfig_ContainsInstallToolsCommand(t *testing.T) {
	driver := NewPipelineDriver(t)
	config := driver.ReadCIConfig(testdataPath("example-circleci-config.yml"))

	if !strings.Contains(config, "install-tools:") {
		t.Error("expected config to contain 'install-tools:' command")
	}

	if !strings.Contains(config, "make install-tools") {
		t.Error("expected install-tools command to delegate to 'make install-tools'")
	}
}

func TestPipelineDriver_ReadCIConfig_ShouldFailOnMissingFile(t *testing.T) {
	// Verify the driver fatals when given a nonexistent path.
	var ft fakeTesting
	driver := &PipelineDriver{t: &ft, root: "."}

	// Run in a goroutine so runtime.Goexit from Fatalf doesn't kill this test.
	done := make(chan struct{})
	go func() {
		defer close(done)
		driver.ReadCIConfig("/nonexistent/path/config.yml")
	}()
	<-done

	if !ft.failed {
		t.Error("expected ReadCIConfig to fatal on missing file")
	}
}

func TestPipelineDriver_ReadCommandsShouldReturnArrayOfCommands(t *testing.T) {
	driver := NewPipelineDriver(t)
	commands := driver.ReadCommands(testdataPath("example-circleci-config.yml"))

	if len(commands) == 0 {
		t.Fatal("expected at least one command, got none")
	}

	if !slices.Contains(commands, "install-tools") {
		t.Errorf("expected commands to contain 'install-tools', got %v", commands)
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
