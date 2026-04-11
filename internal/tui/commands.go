package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"projcompiler/internal/config"
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

func startUnderstandingCmd(runID int, facts spec.RepoFacts, builder UnderstandingBuilder) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return understandingStartedMsg{RunID: runID} },
		func() tea.Msg {
			if builder == nil {
				return understandingFinishedMsg{
					RunID:        runID,
					Summary:      fallbackUnderstandingFromFacts(facts),
					UsedFallback: true,
				}
			}
			summary, err := builder.BuildUnderstanding(context.Background(), facts)
			if err != nil {
				return understandingFinishedMsg{
					RunID:        runID,
					Summary:      fallbackUnderstandingFromFacts(facts),
					Err:          err,
					UsedFallback: true,
				}
			}
			return understandingFinishedMsg{
				RunID:   runID,
				Summary: normalizeUnderstandingSummary(summary, facts),
			}
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

func startSaveConfigCmd(baseURL, apiKey, model string, cfg *config.Config) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return configSaveStartedMsg{} },
		func() tea.Msg {
			if cfg == nil {
				return configSaveFinishedMsg{Err: missingServiceError("config")}
			}
			cfg.Model.BaseURL = strings.TrimSpace(baseURL)
			cfg.Model.APIKey = strings.TrimSpace(apiKey)
			cfg.Model.Model = strings.TrimSpace(model)
			err := cfg.Save()
			return configSaveFinishedMsg{Err: err}
		},
	)
}

func normalizeUnderstandingSummary(summary UnderstandingSummary, facts spec.RepoFacts) UnderstandingSummary {
	if strings.TrimSpace(summary.Status) == "" {
		summary.Status = "ready"
	}
	if strings.TrimSpace(summary.Source) == "" {
		summary.Source = "understanding_builder"
	}
	if strings.TrimSpace(summary.GeneratedAt) == "" {
		summary.GeneratedAt = time.Now().Format(time.RFC3339)
	}
	if summary.Stats.Modules == 0 {
		summary.Stats.Modules = len(facts.EntryPoints)
	}
	if summary.Stats.Flows == 0 {
		summary.Stats.Flows = len(facts.Snippets)
	}
	if summary.Stats.OpenQuestions == 0 {
		summary.Stats.OpenQuestions = len(summary.OpenQuestions)
	}
	if summary.Stats.Evidence == 0 {
		summary.Stats.Evidence = len(summary.Evidence)
	}
	if len(summary.Overview) == 0 {
		summary.Overview = []string{
			"Understanding output is available, but no overview entries were returned.",
		}
	}
	if len(summary.ModuleMap) == 0 {
		summary.ModuleMap = []string{"No module-map edges were returned by the understanding builder."}
	}
	return summary
}

func fallbackUnderstandingFromFacts(facts spec.RepoFacts) UnderstandingSummary {
	summary := UnderstandingSummary{
		Status:      "fallback",
		Source:      "fallback:facts",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Stats: UnderstandingStats{
			Modules: len(facts.EntryPoints),
			Flows:   len(facts.Snippets),
		},
	}

	summary.Overview = []string{
		fmt.Sprintf("Heuristic summary synthesized from scan facts under %s.", valueOrFallback(facts.RootPath, "unknown root")),
		fmt.Sprintf("Detected %d docs, %d entry points, %d snippets, %d constraints.",
			len(facts.Docs), len(facts.EntryPoints), len(facts.Snippets), len(facts.Constraints)),
	}

	for _, doc := range facts.Docs {
		if strings.TrimSpace(doc.Path) == "" {
			continue
		}
		summary.Evidence = append(summary.Evidence, "doc:"+doc.Path)
		if len(summary.Evidence) >= 6 {
			break
		}
	}
	for _, entry := range facts.EntryPoints {
		if len(summary.Evidence) >= 12 {
			break
		}
		path := strings.TrimSpace(entry.Path)
		if path == "" {
			continue
		}
		if strings.TrimSpace(entry.Symbol) != "" {
			summary.Evidence = append(summary.Evidence, fmt.Sprintf("entrypoint:%s#%s", path, entry.Symbol))
		} else {
			summary.Evidence = append(summary.Evidence, "entrypoint:"+path)
		}
	}
	for index, entry := range facts.EntryPoints {
		target := "runtime"
		if len(facts.Snippets) > 0 {
			target = strings.TrimSpace(facts.Snippets[index%len(facts.Snippets)].Path)
			if target == "" {
				target = "runtime"
			}
		}
		name := strings.TrimSpace(entry.Path)
		if name == "" {
			name = valueOrFallback(strings.TrimSpace(entry.Symbol), "unknown_entrypoint")
		}
		summary.ModuleMap = append(summary.ModuleMap, fmt.Sprintf("%s -> %s", name, target))
		if len(summary.ModuleMap) >= 8 {
			break
		}
	}
	if len(summary.ModuleMap) == 0 {
		summary.ModuleMap = []string{"No module map edges available from fallback facts."}
	}

	if len(facts.Docs) == 0 {
		summary.OpenQuestions = append(summary.OpenQuestions, "No project documentation was detected. Confirm scope and target runtime.")
	}
	if len(facts.EntryPoints) == 0 {
		summary.OpenQuestions = append(summary.OpenQuestions, "No entry point was identified. Clarify startup path or command target.")
	}
	if len(summary.OpenQuestions) == 0 {
		summary.OpenQuestions = append(summary.OpenQuestions, "Confirm deployment target, environment assumptions, and primary execution flow.")
	}

	summary.Stats.OpenQuestions = len(summary.OpenQuestions)
	summary.Stats.Evidence = len(summary.Evidence)
	return summary
}
