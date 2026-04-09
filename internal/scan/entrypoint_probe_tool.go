package scan

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"projcompiler/internal/spec"
)

type EntrypointProbeTool struct {
	toolInfo
}

func NewEntrypointProbeTool() EntrypointProbeTool {
	return EntrypointProbeTool{toolInfo{name: "EntrypointProbeTool"}}
}

func (t EntrypointProbeTool) Execute(ctx context.Context, input EntrypointProbeInput) (EntrypointProbeOutput, error) {
	output := EntrypointProbeOutput{}

	candidates := input.Tree.CandidateSources
	if len(candidates) > defaultEntrypointFiles {
		candidates = candidates[:defaultEntrypointFiles]
	}

	for _, relPath := range candidates {
		select {
		case <-ctx.Done():
			return EntrypointProbeOutput{}, ctx.Err()
		default:
		}

		ep, ok, err := inferEntrypoint(input.RootPath, relPath, input.Build)
		if err != nil {
			return EntrypointProbeOutput{}, err
		}
		if ok {
			output.EntryPoints = append(output.EntryPoints, ep)
		}
	}

	if len(output.EntryPoints) == 0 {
		output.Signals = append(output.Signals, spec.MissingSignal("no entrypoint candidates were inferred"))
		return output, nil
	}

	sort.Slice(output.EntryPoints, func(i, j int) bool {
		if output.EntryPoints[i].Score == output.EntryPoints[j].Score {
			return output.EntryPoints[i].Path < output.EntryPoints[j].Path
		}
		return output.EntryPoints[i].Score > output.EntryPoints[j].Score
	})
	if len(output.EntryPoints) > 8 {
		output.EntryPoints = output.EntryPoints[:8]
	}
	return output, nil
}

func inferEntrypoint(rootPath, relPath string, build spec.BuildSummary) (spec.EntryPoint, bool, error) {
	fullPath := filepath.Join(rootPath, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return spec.EntryPoint{}, false, err
	}

	content := string(data)
	ext := strings.ToLower(filepath.Ext(relPath))
	switch ext {
	case ".go":
		ep := inferGoEntrypoint(relPath, content, build)
		return ep, ep.Score > 0, nil
	case ".py":
		return inferScriptEntrypoint(relPath, content, "__main__", "script"), strings.Contains(content, "__main__"), nil
	case ".js", ".ts":
		return inferNodeEntrypoint(relPath, content, build), hasNodeEntrypointSignal(relPath, content, build), nil
	case ".swift":
		return inferSwiftEntrypoint(relPath, content), strings.Contains(content, "@main") || strings.Contains(content, "NSApplicationMain"), nil
	default:
		return spec.EntryPoint{}, false, nil
	}
}

func inferGoEntrypoint(relPath, content string, build spec.BuildSummary) spec.EntryPoint {
	score := 0
	reasons := []string{}
	if strings.Contains(content, "package main") {
		score += 45
		reasons = append(reasons, "contains package main")
	}
	if strings.Contains(content, "func main()") {
		score += 45
		reasons = append(reasons, "contains func main")
	}
	if strings.HasPrefix(relPath, "cmd/") {
		score += 10
		reasons = append(reasons, "lives under cmd/")
	}
	if filepath.Base(relPath) == "main.go" {
		score += 10
		reasons = append(reasons, "named main.go")
	}
	if score == 0 {
		return spec.EntryPoint{}
	}
	return spec.EntryPoint{
		Path:   relPath,
		Symbol: "main",
		Kind:   "go_main",
		Score:  score,
		Signal: spec.InferredSignal(scoreToConfidence(score), strings.Join(reasons, "; "),
			spec.EvidenceRef{Path: relPath, Note: build.ModulePath}),
	}
}

func inferScriptEntrypoint(relPath, content, marker, kind string) spec.EntryPoint {
	score := 0
	reasons := []string{}
	if strings.Contains(content, marker) {
		score += 70
		reasons = append(reasons, "contains runtime entry guard")
	}
	if strings.HasPrefix(relPath, "cmd/") || filepath.Base(relPath) == "main.py" {
		score += 20
		reasons = append(reasons, "looks like a command entry")
	}
	return spec.EntryPoint{
		Path:   relPath,
		Symbol: "main",
		Kind:   kind,
		Score:  score,
		Signal: spec.InferredSignal(scoreToConfidence(score), strings.Join(reasons, "; "),
			spec.EvidenceRef{Path: relPath}),
	}
}

func inferNodeEntrypoint(relPath, content string, build spec.BuildSummary) spec.EntryPoint {
	score := 0
	reasons := []string{}
	base := filepath.Base(relPath)
	if base == "index.ts" || base == "index.js" || base == "main.ts" || base == "main.js" {
		score += 35
		reasons = append(reasons, "conventional node entry filename")
	}
	if strings.Contains(content, "process.argv") || strings.Contains(content, "createServer(") || strings.Contains(content, "express(") {
		score += 35
		reasons = append(reasons, "contains process or server startup signal")
	}
	for _, hint := range build.Hints {
		if hint.Key == "node.main" && hint.Value == relPath {
			score += 30
			reasons = append(reasons, "matches package.json main")
		}
	}
	return spec.EntryPoint{
		Path:   relPath,
		Symbol: "default",
		Kind:   "node_entry",
		Score:  score,
		Signal: spec.InferredSignal(scoreToConfidence(score), strings.Join(reasons, "; "),
			spec.EvidenceRef{Path: relPath}),
	}
}

func inferSwiftEntrypoint(relPath, content string) spec.EntryPoint {
	score := 0
	reasons := []string{}
	if strings.Contains(content, "@main") {
		score += 70
		reasons = append(reasons, "contains @main")
	}
	if strings.Contains(content, "NSApplicationMain") || strings.Contains(content, "MenuBarExtra") {
		score += 20
		reasons = append(reasons, "contains application bootstrap")
	}
	return spec.EntryPoint{
		Path:   relPath,
		Symbol: "main",
		Kind:   "swift_entry",
		Score:  score,
		Signal: spec.InferredSignal(scoreToConfidence(score), strings.Join(reasons, "; "),
			spec.EvidenceRef{Path: relPath}),
	}
}

func hasNodeEntrypointSignal(relPath, content string, build spec.BuildSummary) bool {
	ep := inferNodeEntrypoint(relPath, content, build)
	return ep.Score > 0
}

func scoreToConfidence(score int) float64 {
	if score <= 0 {
		return 0.1
	}
	if score >= 90 {
		return 0.95
	}
	return float64(score) / 100
}
