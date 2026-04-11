package app

import (
	"context"
	"encoding/json"
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
	workingDir           string
	defaultOutputPath    string
	defaultSpecDebugPath string
}

var _ Exporter = (*FileExporter)(nil)

func NewFileExporter(cfg config.Config) *FileExporter {
	promptPath := defaultPromptPath(cfg)
	return &FileExporter{
		workingDir:           cfg.Paths.WorkingDir,
		defaultOutputPath:    promptPath,
		defaultSpecDebugPath: filepath.Join(filepath.Dir(promptPath), "spec.debug.json"),
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
	return e.resolvePath(outputPath, e.defaultOutputPath)
}

func (e *FileExporter) ExportSpecDebug(ctx context.Context, projectSpec spec.ProjectSpec, outputPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	resolvedPath, err := e.resolvePath(outputPath, e.defaultSpecDebugPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return "", fmt.Errorf("create spec debug export directory: %w", err)
	}

	payload := map[string]any{
		"schema_version": "project_spec_debug_v1",
		"project_spec":   projectSpec,
	}
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode spec debug payload: %w", err)
	}
	content = append(content, '\n')

	if err := os.WriteFile(resolvedPath, content, 0o644); err != nil {
		return "", fmt.Errorf("write spec debug file: %w", err)
	}
	return resolvedPath, nil
}

func (e *FileExporter) resolvePath(outputPath, defaultPath string) (string, error) {
	path := strings.TrimSpace(outputPath)
	if path == "" {
		path = defaultPath
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.workingDir, path)
	}
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve export path: %w", err)
	}
	return resolved, nil
}
