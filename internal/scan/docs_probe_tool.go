package scan

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"projcompiler/internal/spec"
)

type DocsProbeTool struct {
	toolInfo
}

func NewDocsProbeTool() DocsProbeTool {
	return DocsProbeTool{toolInfo{name: "DocsProbeTool"}}
}

func (t DocsProbeTool) Execute(ctx context.Context, input DocsProbeInput) (DocsProbeOutput, error) {
	output := DocsProbeOutput{}
	paths := prioritizeDocs(input.CandidatePaths)
	if len(paths) > input.MaxDocs {
		paths = paths[:input.MaxDocs]
	}

	var sawRootReadme bool
	for _, relPath := range paths {
		select {
		case <-ctx.Done():
			return DocsProbeOutput{}, ctx.Err()
		default:
		}

		fullPath := filepath.Join(input.RootPath, relPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return DocsProbeOutput{}, err
		}
		if len(data) > input.MaxBytes && input.MaxBytes > 0 {
			data = data[:input.MaxBytes]
		}

		title, excerpt := extractDocSummary(string(data))
		doc := spec.DocumentFact{
			Path:    relPath,
			Title:   title,
			Excerpt: excerpt,
			Signal: spec.ObservedSignal("observed project documentation excerpt",
				spec.EvidenceRef{Path: relPath, StartLine: 1}),
		}
		output.Docs = append(output.Docs, doc)
		output.Constraints = append(output.Constraints, extractConstraints(relPath, string(data))...)

		base := strings.ToLower(filepath.Base(relPath))
		if strings.HasPrefix(base, "readme") && filepath.Dir(relPath) == "." {
			sawRootReadme = true
		}
	}

	sort.Slice(output.Docs, func(i, j int) bool {
		return output.Docs[i].Path < output.Docs[j].Path
	})
	sort.Slice(output.Constraints, func(i, j int) bool {
		return output.Constraints[i].Text < output.Constraints[j].Text
	})

	if !sawRootReadme {
		output.Signals = append(output.Signals, spec.MissingSignal("root README was not observed"))
	}
	return output, nil
}

func prioritizeDocs(paths []string) []string {
	ranked := append([]string(nil), paths...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return docRank(ranked[i]) < docRank(ranked[j])
	})
	return ranked
}

func docRank(relPath string) int {
	base := strings.ToLower(filepath.Base(relPath))
	switch {
	case strings.HasPrefix(base, "readme") && filepath.Dir(relPath) == ".":
		return 0
	case strings.HasPrefix(base, "readme"):
		return 1
	case strings.Contains(filepath.ToSlash(relPath), "docs/"):
		return 2
	default:
		return 3
	}
}

func extractDocSummary(text string) (string, string) {
	lines := strings.Split(text, "\n")
	title := ""
	var excerptLines []string
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if title == "" {
			switch {
			case strings.HasPrefix(line, "#"):
				title = strings.TrimSpace(strings.TrimLeft(line, "#"))
				continue
			case line != "":
				title = line
				continue
			}
		}
		if line == "" && len(excerptLines) > 0 {
			break
		}
		if line != "" && len(excerptLines) < 6 {
			excerptLines = append(excerptLines, line)
		}
	}
	return title, strings.Join(excerptLines, " ")
}

func extractConstraints(relPath, text string) []spec.Constraint {
	lines := strings.Split(text, "\n")
	var constraints []spec.Constraint
	for idx, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "must") &&
			!strings.Contains(lower, "required") &&
			!strings.Contains(lower, "only") &&
			!strings.Contains(lower, "do not") &&
			!strings.Contains(lower, "don't") {
			continue
		}
		constraints = append(constraints, spec.Constraint{
			Text: line,
			Signal: spec.ObservedSignal("observed explicit documentation constraint",
				spec.EvidenceRef{Path: relPath, StartLine: idx + 1, EndLine: idx + 1}),
		})
		if len(constraints) >= 8 {
			break
		}
	}
	return constraints
}
