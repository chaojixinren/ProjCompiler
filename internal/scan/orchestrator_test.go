package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanOrchestratorCollectsRepoFactsForSmallGoProject(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Demo\nThis app must stay local-only.\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "app", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	orchestrator, err := NewOrchestrator(ScanRequest{RootPath: root})
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %v", err)
	}

	facts, err := orchestrator.Scan(context.Background(), root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if got, want := facts.Build.ModulePath, "example.com/demo"; got != want {
		t.Fatalf("module path = %q, want %q", got, want)
	}
	if len(facts.Docs) == 0 {
		t.Fatalf("expected documentation facts")
	}
	if len(facts.EntryPoints) == 0 {
		t.Fatalf("expected entrypoint facts")
	}
	if len(facts.Snippets) == 0 {
		t.Fatalf("expected snippet facts")
	}
	if len(facts.Constraints) == 0 {
		t.Fatalf("expected documentation constraints")
	}
}
