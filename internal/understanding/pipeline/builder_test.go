package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/spec"
)

func TestBuildProducesUnifiedSpec(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), `package main
func main(){unknownCall()}
`)
	writeFile(t, filepath.Join(root, "README.md"), "# Demo\nProject summary.\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module demo\n\ngo 1.25.0\n")

	facts := spec.RepoFacts{
		RootPath: root,
		Tree: spec.FileTreeSummary{
			Nodes: []spec.FileNode{
				{Path: "main.go", Kind: spec.FileNodeFile, Size: 40},
				{Path: "README.md", Kind: spec.FileNodeFile, Size: 20},
				{Path: "go.mod", Kind: spec.FileNodeFile, Size: 24},
			},
		},
		Docs: []spec.DocumentFact{
			{Path: "README.md", Title: "Demo", Excerpt: "Project summary.", Signal: spec.ObservedSignal("doc")},
		},
		EntryPoints: []spec.EntryPoint{
			{Path: "main.go", Symbol: "main", Kind: "cli", Score: 1, Signal: spec.ObservedSignal("entry")},
		},
	}

	builder := NewBuilder()
	out, err := builder.Build(context.Background(), BuildRequest{
		RootPath: root,
		Facts:    &facts,
		RunID:    "run-1",
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if len(out.ProjectSpec.Understanding.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(out.ProjectSpec.Understanding.Files))
	}
	if out.ProjectSpec.Understanding.Stats.SymbolCount == 0 {
		t.Fatal("expected symbols to be extracted")
	}
	if len(out.ProjectSpec.OpenQuestions) == 0 {
		t.Fatal("expected at least one ambiguity question from unresolved call")
	}
	if len(out.ProjectSpec.ImplementationShape.CoreModules) == 0 {
		t.Fatal("expected unified spec to include core modules")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
