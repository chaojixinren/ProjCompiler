package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"projcompiler/internal/spec"
)

func TestSnippetSelectToolSelectsEntrypointSnippetAroundMain(t *testing.T) {
	root := t.TempDir()
	content := strings.Join([]string{
		"package main",
		"",
		"import \"fmt\"",
		"",
		"func helper() {}",
		"",
		"func main() {",
		"\tfmt.Println(\"hello\")",
		"}",
		"",
		"func tail() {}",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	output, err := NewSnippetSelectTool().Execute(context.Background(), SnippetSelectInput{
		RootPath: root,
		EntryPoints: []spec.EntryPoint{
			{Path: "main.go", Score: 90},
		},
		MaxSnippets:     2,
		MaxSnippetLines: 3,
		MaxBytes:        1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Snippets) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(output.Snippets))
	}
	if got, want := output.Snippets[0].StartLine, 7; got != want {
		t.Fatalf("StartLine = %d, want %d", got, want)
	}
	if !strings.Contains(output.Snippets[0].Content, "fmt.Println") {
		t.Fatalf("expected snippet content to include main body, got %q", output.Snippets[0].Content)
	}
}

func TestSnippetSelectToolFallsBackToConfigSources(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config_setup.go"), []byte("package main\n\nfunc loadConfig() {}\n"), 0o644); err != nil {
		t.Fatalf("write config source: %v", err)
	}

	output, err := NewSnippetSelectTool().Execute(context.Background(), SnippetSelectInput{
		RootPath: root,
		Tree: spec.FileTreeSummary{
			CandidateSources: []string{"config_setup.go"},
		},
		MaxSnippets:     1,
		MaxSnippetLines: 10,
		MaxBytes:        1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Snippets) != 1 {
		t.Fatalf("expected fallback config snippet, got %#v", output.Snippets)
	}
	if got, want := output.Snippets[0].Purpose, "config_or_bootstrap"; got != want {
		t.Fatalf("Purpose = %q, want %q", got, want)
	}
}

func TestReadLinesNormalizesCRLFAndTrailingNewline(t *testing.T) {
	lines := readLines([]byte("first\r\nsecond\r\n"))
	if got, want := len(lines), 2; got != want {
		t.Fatalf("len(lines) = %d, want %d", got, want)
	}
	if lines[0] != "first" || lines[1] != "second" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestSnippetSelectToolEmitsMissingSignalWhenNoSnippetCandidateExists(t *testing.T) {
	output, err := NewSnippetSelectTool().Execute(context.Background(), SnippetSelectInput{
		RootPath:        t.TempDir(),
		MaxSnippets:     1,
		MaxSnippetLines: 10,
		MaxBytes:        1 << 20,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(output.Snippets) != 0 {
		t.Fatalf("expected no snippets, got %#v", output.Snippets)
	}
	if len(output.Signals) != 1 || output.Signals[0].Kind != "missing" {
		t.Fatalf("expected missing snippet signal, got %#v", output.Signals)
	}
}
