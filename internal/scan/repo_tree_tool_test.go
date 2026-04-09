package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/spec"
)

func TestRepoTreeToolAppliesDepthFileBudgetAndClassification(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/app: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "deep", "nested", "too", "far"), 0o755); err != nil {
		t.Fatalf("mkdir deep path: %v", err)
	}

	files := map[string]string{
		"README.md":                "# Demo\n",
		"go.mod":                   "module example.com/demo\n",
		"cmd/app/main.go":          "package main\nfunc main() {}\n",
		"deep/nested/too/far/x.go": "package far\n",
	}
	for relPath, content := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir parent for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", relPath, err)
		}
	}

	matcher, err := LoadIgnoreMatcher(root, false)
	if err != nil {
		t.Fatalf("LoadIgnoreMatcher returned error: %v", err)
	}

	output, err := NewRepoTreeTool().Execute(context.Background(), RepoTreeInput{
		Request: ScanRequest{
			RootPath:    root,
			MaxDepth:    3,
			MaxFiles:    3,
			MaxFileSize: 1 << 20,
		},
		Ignore: matcher,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got, want := output.Tree.CandidateDocs, []string{"README.md"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("CandidateDocs = %#v, want %#v", got, want)
	}
	if got, want := output.Tree.CandidateManifests, []string{"go.mod"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("CandidateManifests = %#v, want %#v", got, want)
	}
	if got, want := output.Tree.CandidateSources, []string{"cmd/app/main.go"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("CandidateSources = %#v, want %#v", got, want)
	}

	if !containsIgnoredReason(output.ScanMeta.IgnoredPaths, "depth_limit") {
		t.Fatalf("expected depth-limited path in ignored list, got %#v", output.ScanMeta.IgnoredPaths)
	}
	if !containsFileNode(output.Tree.Nodes, "cmd/app/main.go", spec.FileNodeFile) {
		t.Fatalf("expected file node for cmd/app/main.go, got %#v", output.Tree.Nodes)
	}
}

func TestRepoTreeToolMarksTruncationWhenFileBudgetIsExceeded(t *testing.T) {
	root := t.TempDir()
	files := []string{"README.md", "go.mod", "main.go"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	matcher, err := LoadIgnoreMatcher(root, false)
	if err != nil {
		t.Fatalf("LoadIgnoreMatcher returned error: %v", err)
	}

	output, err := NewRepoTreeTool().Execute(context.Background(), RepoTreeInput{
		Request: ScanRequest{
			RootPath:    root,
			MaxDepth:    4,
			MaxFiles:    1,
			MaxFileSize: 1 << 20,
		},
		Ignore: matcher,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !output.ScanMeta.Truncated {
		t.Fatalf("expected scan truncation when file budget is exceeded")
	}
	if !containsIgnoredReason(output.ScanMeta.IgnoredPaths, "file_budget_exceeded") {
		t.Fatalf("expected file budget ignored reason, got %#v", output.ScanMeta.IgnoredPaths)
	}
}

func containsIgnoredReason(paths []spec.IgnoredPath, reason string) bool {
	for _, item := range paths {
		if item.Reason == reason {
			return true
		}
	}
	return false
}

func containsFileNode(nodes []spec.FileNode, path string, kind spec.FileNodeKind) bool {
	for _, node := range nodes {
		if node.Path == path && node.Kind == kind {
			return true
		}
	}
	return false
}
