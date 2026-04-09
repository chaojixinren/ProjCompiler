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

func TestManifestProbeToolReturnsMissingSignalWithoutCandidates(t *testing.T) {
	output, err := NewManifestProbeTool().Execute(context.Background(), ManifestProbeInput{
		RootPath:       t.TempDir(),
		CandidatePaths: nil,
		MaxBytes:       1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Signals) != 1 || output.Signals[0].Kind != "missing" {
		t.Fatalf("expected missing manifest signal, got %#v", output.Signals)
	}
}

func TestManifestProbeToolReportsPackageJSONParseErrorAsHint(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	output, err := NewManifestProbeTool().Execute(context.Background(), ManifestProbeInput{
		RootPath:       root,
		CandidatePaths: []string{"package.json"},
		MaxBytes:       1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Build.Hints) == 0 {
		t.Fatalf("expected manifest parse error hint")
	}
	found := false
	for _, hint := range output.Build.Hints {
		if hint.Key == "manifest.parse_error" && hint.Value == "package.json" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected package.json parse error hint, got %#v", output.Build.Hints)
	}
}
