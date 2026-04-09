package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/spec"
)

func TestEntrypointProbeToolFindsGoMainEntrypoint(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/app: %v", err)
	}
	mainPath := filepath.Join(root, "cmd", "app", "main.go")
	content := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(mainPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	output, err := NewEntrypointProbeTool().Execute(context.Background(), EntrypointProbeInput{
		RootPath: root,
		Tree: spec.FileTreeSummary{
			CandidateSources: []string{"cmd/app/main.go"},
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.EntryPoints) != 1 {
		t.Fatalf("expected 1 entrypoint, got %d", len(output.EntryPoints))
	}
	if got, want := output.EntryPoints[0].Path, "cmd/app/main.go"; got != want {
		t.Fatalf("entrypoint path = %q, want %q", got, want)
	}
	if output.EntryPoints[0].Score <= 0 {
		t.Fatalf("expected positive entrypoint score, got %d", output.EntryPoints[0].Score)
	}
}

func TestEntrypointProbeToolSkipsNonMainGoFiles(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "service.go")
	content := "package service\n\nfunc Run() {}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write service.go: %v", err)
	}

	output, err := NewEntrypointProbeTool().Execute(context.Background(), EntrypointProbeInput{
		RootPath: root,
		Tree: spec.FileTreeSummary{
			CandidateSources: []string{"service.go"},
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.EntryPoints) != 0 {
		t.Fatalf("expected non-main Go files to be skipped, got %#v", output.EntryPoints)
	}
	if len(output.Signals) == 0 {
		t.Fatalf("expected a missing-entrypoint signal when nothing qualifies")
	}
}
