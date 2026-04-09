package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

func TestFileExporterExportPromptWritesResolvedFileWithTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(config.Config{
		Paths: config.Paths{WorkingDir: dir},
		Output: config.OutputConfig{
			PromptFile: "artifacts/prompt.txt",
		},
	})

	path, err := exporter.ExportPrompt(context.Background(), spec.PromptBundle{
		PromptText: "Build a small Go CLI",
	}, "")
	if err != nil {
		t.Fatalf("ExportPrompt returned error: %v", err)
	}

	if got, want := path, filepath.Join(dir, "artifacts", "prompt.txt"); got != want {
		t.Fatalf("export path = %q, want %q", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if got, want := string(data), "Build a small Go CLI\n"; got != want {
		t.Fatalf("exported content = %q, want %q", got, want)
	}
}

func TestFileExporterRejectsUnreadyBundle(t *testing.T) {
	exporter := NewFileExporter(config.Config{
		Paths: config.Paths{WorkingDir: t.TempDir()},
	})

	_, err := exporter.ExportPrompt(context.Background(), spec.PromptBundle{}, "")
	if !errors.Is(err, ErrPromptNotReady) {
		t.Fatalf("expected ErrPromptNotReady, got %v", err)
	}
}

func TestFileExporterResolveOutputPathKeepsAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(config.Config{
		Paths: config.Paths{WorkingDir: dir},
	})

	absPath := filepath.Join(dir, "absolute.txt")
	got, err := exporter.resolveOutputPath(absPath)
	if err != nil {
		t.Fatalf("resolveOutputPath returned error: %v", err)
	}
	if got != absPath {
		t.Fatalf("resolved path = %q, want %q", got, absPath)
	}
}

func TestFileExporterUsesDefaultPathWhenOutputPathIsBlank(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(config.Config{
		Paths: config.Paths{WorkingDir: dir},
		Output: config.OutputConfig{
			PromptFile: "prompt.txt",
		},
	})

	path, err := exporter.ExportPrompt(context.Background(), spec.PromptBundle{
		PromptText: "Build the project",
	}, "   ")
	if err != nil {
		t.Fatalf("ExportPrompt returned error: %v", err)
	}

	if got, want := path, filepath.Join(dir, "prompt.txt"); got != want {
		t.Fatalf("export path = %q, want %q", got, want)
	}
}

func TestFileExporterHonorsCancelledContext(t *testing.T) {
	exporter := NewFileExporter(config.Config{
		Paths: config.Paths{WorkingDir: t.TempDir()},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := exporter.ExportPrompt(ctx, spec.PromptBundle{PromptText: "Build the project"}, "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
