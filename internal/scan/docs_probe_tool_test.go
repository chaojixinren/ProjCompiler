package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsProbeToolExtractsSummaryConstraintsAndMissingRootReadme(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}

	content := "# Usage\nThis tool scans a local project.\nIt must only emit an English prompt.\nDo not invent extra product scope.\n"
	if err := os.WriteFile(filepath.Join(root, "docs", "usage.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write docs file: %v", err)
	}

	output, err := NewDocsProbeTool().Execute(context.Background(), DocsProbeInput{
		RootPath:       root,
		CandidatePaths: []string{"docs/usage.md"},
		MaxDocs:        4,
		MaxBytes:       1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(output.Docs))
	}
	if got, want := output.Docs[0].Title, "Usage"; got != want {
		t.Fatalf("doc title = %q, want %q", got, want)
	}
	if !strings.Contains(output.Docs[0].Excerpt, "must only emit an English prompt") {
		t.Fatalf("expected excerpt to include body text, got %q", output.Docs[0].Excerpt)
	}
	if len(output.Constraints) < 2 {
		t.Fatalf("expected constraint lines to be extracted, got %d", len(output.Constraints))
	}
	if len(output.Signals) != 1 || output.Signals[0].Kind != "missing" {
		t.Fatalf("expected missing root README signal, got %#v", output.Signals)
	}
}

func TestDocsProbeToolPrioritizesRootReadmeWhenMaxDocsIsLimited(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "docs", "usage.md"), []byte("# Usage\nSecondary doc.\n"), 0o644); err != nil {
		t.Fatalf("write docs/usage.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Root\nPrimary doc.\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	output, err := NewDocsProbeTool().Execute(context.Background(), DocsProbeInput{
		RootPath:       root,
		CandidatePaths: []string{"docs/usage.md", "README.md"},
		MaxDocs:        1,
		MaxBytes:       1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(output.Docs))
	}
	if got, want := output.Docs[0].Path, "README.md"; got != want {
		t.Fatalf("doc path = %q, want %q", got, want)
	}
	if len(output.Signals) != 0 {
		t.Fatalf("expected no missing root README signal, got %#v", output.Signals)
	}
}
