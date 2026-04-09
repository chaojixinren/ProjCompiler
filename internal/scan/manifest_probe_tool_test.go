package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestProbeToolParsesGoModDependenciesAndHints(t *testing.T) {
	root := t.TempDir()
	content := "module example.com/proj\n\ngo 1.25.0\ntoolchain go1.25.1\nrequire (\n\tgithub.com/charmbracelet/bubbletea v1.3.4\n)\nreplace example.com/old => example.com/new v1.0.0\n"
	path := filepath.Join(root, "go.mod")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	output, err := NewManifestProbeTool().Execute(context.Background(), ManifestProbeInput{
		RootPath:       root,
		CandidatePaths: []string{"go.mod"},
		MaxBytes:       1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got, want := output.Build.ModulePath, "example.com/proj"; got != want {
		t.Fatalf("module path = %q, want %q", got, want)
	}
	if len(output.Build.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(output.Build.Dependencies))
	}
	if got, want := output.Build.Dependencies[0].Name, "github.com/charmbracelet/bubbletea"; got != want {
		t.Fatalf("dependency name = %q, want %q", got, want)
	}
	if len(output.Build.Hints) < 3 {
		t.Fatalf("expected go version, toolchain, and replace hints, got %d", len(output.Build.Hints))
	}
}
