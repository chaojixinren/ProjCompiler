package inventory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/types"
)

func TestBuildFromFactsNodes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\nfunc main(){}\n")
	writeFile(t, filepath.Join(root, "README.md"), "# title\n")

	facts := spec.RepoFacts{
		RootPath: root,
		Tree: spec.FileTreeSummary{
			Nodes: []spec.FileNode{
				{Path: "main.go", Kind: spec.FileNodeFile, Size: 24},
				{Path: "README.md", Kind: spec.FileNodeFile, Size: 8},
			},
		},
	}

	builder := NewBuilder()
	files, err := builder.Build(context.Background(), BuildInput{Facts: &facts})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	mainFile := findByPath(files, "main.go")
	if mainFile.ID == "" {
		t.Fatal("expected main.go to exist")
	}
	if mainFile.Role != types.FileRoleSource {
		t.Fatalf("expected source role, got %q", mainFile.Role)
	}

	readme := findByPath(files, "README.md")
	if readme.Role != types.FileRoleDoc {
		t.Fatalf("expected doc role, got %q", readme.Role)
	}
}

func findByPath(files []types.File, path string) types.File {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return types.File{}
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
