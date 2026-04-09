package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

var ErrPromptNotReady = errors.New("prompt bundle is not ready for export")

type FileExporter struct {
	workingDir        string
	defaultOutputPath string
}

var _ Exporter = (*FileExporter)(nil)

func NewFileExporter(cfg config.Config) *FileExporter {
	return &FileExporter{
		workingDir:        cfg.Paths.WorkingDir,
		defaultOutputPath: defaultPromptPath(cfg),
	}
}

func (e *FileExporter) ExportPrompt(ctx context.Context, bundle spec.PromptBundle, outputPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !bundle.Ready() {
		return "", ErrPromptNotReady
	}

	resolvedPath, err := e.resolveOutputPath(outputPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}

	content := bundle.PromptText
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(resolvedPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write prompt file: %w", err)
	}
	return resolvedPath, nil
}

func (e *FileExporter) resolveOutputPath(outputPath string) (string, error) {
	path := strings.TrimSpace(outputPath)
	if path == "" {
		path = e.defaultOutputPath
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.workingDir, path)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve export path: %w", err)
	}
	return absPath, nil
}
