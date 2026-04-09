package app

import (
	"context"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
	"projcompiler/internal/tui"
)

func newProgram(application *App) *tea.Program {
	return tui.NewProgram(
		newTUIServices(application.services),
		tui.WithTitle("ProjCompiler"),
		tui.WithDefaultPromptPath(defaultPromptPath(application.config)),
	)
}

func defaultPromptPath(cfg config.Config) string {
	if filepath.IsAbs(cfg.Output.PromptFile) {
		return cfg.Output.PromptFile
	}
	if cfg.Output.PromptFile == "" {
		return filepath.Join(cfg.Paths.WorkingDir, "prompt.txt")
	}
	return filepath.Join(cfg.Paths.WorkingDir, cfg.Output.PromptFile)
}

func newTUIServices(services Services) tui.Services {
	return tui.Services{
		Scanner:        scannerAdapter{inner: services.Scanner},
		SpecBuilder:    specBuilderAdapter{inner: services.SpecBuilder},
		PromptCompiler: promptCompilerAdapter{inner: services.PromptCompiler},
		Exporter:       exporterAdapter{inner: services.Exporter},
	}
}

type scannerAdapter struct {
	inner Scanner
}

func (a scannerAdapter) Scan(ctx context.Context, projectPath string) (spec.RepoFacts, error) {
	return a.inner.Scan(ctx, projectPath)
}

type specBuilderAdapter struct {
	inner SpecBuilder
}

func (a specBuilderAdapter) BuildProjectSpec(ctx context.Context, facts spec.RepoFacts) (spec.ProjectSpec, error) {
	return a.inner.BuildProjectSpec(ctx, facts)
}

type promptCompilerAdapter struct {
	inner PromptCompiler
}

func (a promptCompilerAdapter) CompilePrompt(ctx context.Context, projectSpec spec.ProjectSpec) (spec.PromptBundle, error) {
	return a.inner.CompilePrompt(ctx, projectSpec)
}

type exporterAdapter struct {
	inner Exporter
}

func (a exporterAdapter) ExportPrompt(ctx context.Context, bundle spec.PromptBundle, outputPath string) (string, error) {
	return a.inner.ExportPrompt(ctx, bundle, outputPath)
}
