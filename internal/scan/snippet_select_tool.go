package scan

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"projcompiler/internal/spec"
)

type SnippetSelectTool struct {
	toolInfo
}

func NewSnippetSelectTool() SnippetSelectTool {
	return SnippetSelectTool{toolInfo{name: "SnippetSelectTool"}}
}

func (t SnippetSelectTool) Execute(ctx context.Context, input SnippetSelectInput) (SnippetSelectOutput, error) {
	output := SnippetSelectOutput{}
	seen := map[string]bool{}

	for _, entry := range input.EntryPoints {
		if len(output.Snippets) >= input.MaxSnippets {
			break
		}
		select {
		case <-ctx.Done():
			return SnippetSelectOutput{}, ctx.Err()
		default:
		}

		snippet, ok, err := snippetForEntry(input.RootPath, entry, input.MaxSnippetLines, input.MaxBytes)
		if err != nil {
			return SnippetSelectOutput{}, err
		}
		if ok {
			key := snippet.Path + ":" + snippet.Purpose
			if !seen[key] {
				output.Snippets = append(output.Snippets, snippet)
				seen[key] = true
			}
		}
	}

	if len(output.Snippets) < input.MaxSnippets {
		for _, relPath := range input.Tree.CandidateSources {
			if len(output.Snippets) >= input.MaxSnippets {
				break
			}
			if !looksLikeConfigSource(relPath) {
				continue
			}
			snippet, ok, err := snippetForFile(input.RootPath, relPath, "config_or_bootstrap", input.MaxSnippetLines, input.MaxBytes)
			if err != nil {
				return SnippetSelectOutput{}, err
			}
			if ok {
				key := snippet.Path + ":" + snippet.Purpose
				if !seen[key] {
					output.Snippets = append(output.Snippets, snippet)
					seen[key] = true
				}
			}
		}
	}

	sort.Slice(output.Snippets, func(i, j int) bool {
		if output.Snippets[i].Path == output.Snippets[j].Path {
			return output.Snippets[i].StartLine < output.Snippets[j].StartLine
		}
		return output.Snippets[i].Path < output.Snippets[j].Path
	})
	if len(output.Snippets) == 0 {
		output.Signals = append(output.Signals, spec.MissingSignal("no key snippets were selected"))
	}
	return output, nil
}

func snippetForEntry(rootPath string, entry spec.EntryPoint, maxLines int, maxBytes int64) (spec.CodeSnippet, bool, error) {
	return snippetForFile(rootPath, entry.Path, "entrypoint", maxLines, maxBytes)
}

func snippetForFile(rootPath, relPath, purpose string, maxLines int, maxBytes int64) (spec.CodeSnippet, bool, error) {
	fullPath := filepath.Join(rootPath, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return spec.CodeSnippet{}, false, err
	}
	if int64(len(data)) > maxBytes && maxBytes > 0 {
		data = data[:maxBytes]
	}

	lines := readLines(data)
	if len(lines) == 0 {
		return spec.CodeSnippet{}, false, nil
	}

	start := 1
	for idx, line := range lines {
		if strings.Contains(line, "func main(") || strings.Contains(line, "__main__") || strings.Contains(line, "@main") {
			start = idx + 1
			break
		}
	}
	end := start + maxLines - 1
	if end > len(lines) {
		end = len(lines)
	}
	content := strings.Join(lines[start-1:end], "\n")
	return spec.CodeSnippet{
		Path:      relPath,
		StartLine: start,
		EndLine:   end,
		Purpose:   purpose,
		Content:   content,
		Signal: spec.ObservedSignal("observed key code snippet",
			spec.EvidenceRef{Path: relPath, StartLine: start, EndLine: end}),
	}, true, nil
}

func readLines(data []byte) []string {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func looksLikeConfigSource(relPath string) bool {
	lower := strings.ToLower(relPath)
	return strings.Contains(lower, "config") || strings.Contains(lower, "bootstrap") || strings.Contains(lower, "setup")
}
