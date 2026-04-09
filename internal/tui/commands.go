package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"projcompiler/internal/spec"
)

func startScanCmd(runID int, path string, scanner Scanner) tea.Cmd {
	path = strings.TrimSpace(path)
	return tea.Batch(
		func() tea.Msg { return scanStartedMsg{RunID: runID, Path: path} },
		func() tea.Msg {
			if scanner == nil {
				return scanFinishedMsg{RunID: runID, Err: missingServiceError("scanner")}
			}
			facts, err := scanner.Scan(context.Background(), path)
			return scanFinishedMsg{RunID: runID, Facts: facts, Err: err}
		},
	)
}

func startBuildSpecCmd(runID int, facts spec.RepoFacts, builder SpecBuilder) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return specStartedMsg{RunID: runID} },
		func() tea.Msg {
			if builder == nil {
				return specFinishedMsg{RunID: runID, Err: missingServiceError("spec builder")}
			}
			projectSpec, err := builder.BuildProjectSpec(context.Background(), facts)
			return specFinishedMsg{RunID: runID, ProjectSpec: projectSpec, Err: err}
		},
	)
}

func startCompilePromptCmd(runID int, projectSpec spec.ProjectSpec, compiler PromptCompiler) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return promptStartedMsg{RunID: runID} },
		func() tea.Msg {
			if compiler == nil {
				return promptFinishedMsg{RunID: runID, Err: missingServiceError("prompt compiler")}
			}
			bundle, err := compiler.CompilePrompt(context.Background(), projectSpec)
			return promptFinishedMsg{RunID: runID, Bundle: bundle, Err: err}
		},
	)
}

func startExportCmd(runID int, bundle spec.PromptBundle, outputPath string, exporter Exporter) tea.Cmd {
	outputPath = strings.TrimSpace(outputPath)
	return tea.Batch(
		func() tea.Msg { return exportStartedMsg{RunID: runID, OutputPath: outputPath} },
		func() tea.Msg {
			if exporter == nil {
				return exportFinishedMsg{RunID: runID, OutputPath: outputPath, Err: missingServiceError("exporter")}
			}
			resolvedPath, err := exporter.ExportPrompt(context.Background(), bundle, outputPath)
			if resolvedPath != "" {
				outputPath = resolvedPath
			}
			return exportFinishedMsg{RunID: runID, OutputPath: outputPath, Err: err}
		},
	)
}
