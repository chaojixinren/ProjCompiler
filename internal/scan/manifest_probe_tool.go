package scan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"projcompiler/internal/spec"
)

type ManifestProbeTool struct {
	toolInfo
}

func NewManifestProbeTool() ManifestProbeTool {
	return ManifestProbeTool{toolInfo{name: "ManifestProbeTool"}}
}

func (t ManifestProbeTool) Execute(ctx context.Context, input ManifestProbeInput) (ManifestProbeOutput, error) {
	output := ManifestProbeOutput{
		Build: spec.BuildSummary{},
	}

	if len(input.CandidatePaths) == 0 {
		output.Signals = append(output.Signals, spec.MissingSignal("no manifest files were observed during scan"))
		return output, nil
	}

	for _, relPath := range input.CandidatePaths {
		select {
		case <-ctx.Done():
			return ManifestProbeOutput{}, ctx.Err()
		default:
		}

		fullPath := filepath.Join(input.RootPath, relPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return ManifestProbeOutput{}, err
		}
		if int64(len(data)) > input.MaxBytes && input.MaxBytes > 0 {
			data = data[:input.MaxBytes]
		}

		output.Build.ManifestFiles = append(output.Build.ManifestFiles, relPath)
		base := strings.ToLower(filepath.Base(relPath))
		switch base {
		case "go.mod":
			modulePath, deps, hints := parseGoMod(relPath, data)
			if modulePath != "" {
				output.Build.ModulePath = modulePath
			}
			output.Build.Dependencies = append(output.Build.Dependencies, deps...)
			output.Build.Hints = append(output.Build.Hints, hints...)
		case "package.json":
			deps, hints := parsePackageJSON(relPath, data)
			output.Build.Dependencies = append(output.Build.Dependencies, deps...)
			output.Build.Hints = append(output.Build.Hints, hints...)
		case "pyproject.toml":
			output.Build.Hints = append(output.Build.Hints, spec.Hint{
				Key:   "build.manifest",
				Value: "pyproject.toml",
				Signal: spec.ObservedSignal("observed Python project manifest",
					spec.EvidenceRef{Path: relPath}),
			})
		case "makefile":
			output.Build.Hints = append(output.Build.Hints, manifestHint(relPath, "build.makefile", "present"))
		case "dockerfile":
			output.Build.Hints = append(output.Build.Hints, manifestHint(relPath, "build.dockerfile", "present"))
		}
	}

	sort.Strings(output.Build.ManifestFiles)
	sort.Slice(output.Build.Dependencies, func(i, j int) bool {
		if output.Build.Dependencies[i].Name == output.Build.Dependencies[j].Name {
			return output.Build.Dependencies[i].Version < output.Build.Dependencies[j].Version
		}
		return output.Build.Dependencies[i].Name < output.Build.Dependencies[j].Name
	})
	sort.Slice(output.Build.Hints, func(i, j int) bool {
		if output.Build.Hints[i].Key == output.Build.Hints[j].Key {
			return output.Build.Hints[i].Value < output.Build.Hints[j].Value
		}
		return output.Build.Hints[i].Key < output.Build.Hints[j].Key
	})

	if output.Build.ModulePath == "" {
		output.Signals = append(output.Signals, spec.MissingSignal("no module path was inferred from manifests"))
	}
	return output, nil
}

func parseGoMod(relPath string, data []byte) (string, []spec.Dependency, []spec.Hint) {
	var (
		modulePath string
		deps       []spec.Dependency
		hints      []spec.Hint
	)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	inRequireBlock := false
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "module "):
			modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		case line == "require (":
			inRequireBlock = true
		case inRequireBlock && line == ")":
			inRequireBlock = false
		case strings.HasPrefix(line, "require "):
			if dep := parseGoRequireLine(strings.TrimSpace(strings.TrimPrefix(line, "require ")), relPath, lineNo); dep.Name != "" {
				deps = append(deps, dep)
			}
		case inRequireBlock:
			if dep := parseGoRequireLine(line, relPath, lineNo); dep.Name != "" {
				deps = append(deps, dep)
			}
		case strings.HasPrefix(line, "go "):
			hints = append(hints, spec.Hint{
				Key:   "go.version",
				Value: strings.TrimSpace(strings.TrimPrefix(line, "go ")),
				Signal: spec.ObservedSignal("observed go language version",
					spec.EvidenceRef{Path: relPath, StartLine: lineNo, EndLine: lineNo}),
			})
		case strings.HasPrefix(line, "toolchain "):
			hints = append(hints, spec.Hint{
				Key:   "go.toolchain",
				Value: strings.TrimSpace(strings.TrimPrefix(line, "toolchain ")),
				Signal: spec.ObservedSignal("observed go toolchain pin",
					spec.EvidenceRef{Path: relPath, StartLine: lineNo, EndLine: lineNo}),
			})
		case strings.HasPrefix(line, "replace "):
			hints = append(hints, spec.Hint{
				Key:   "go.replace",
				Value: strings.TrimSpace(strings.TrimPrefix(line, "replace ")),
				Signal: spec.ObservedSignal("observed go replace directive",
					spec.EvidenceRef{Path: relPath, StartLine: lineNo, EndLine: lineNo}),
			})
		}
	}
	return modulePath, deps, hints
}

func parseGoRequireLine(line, relPath string, lineNo int) spec.Dependency {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return spec.Dependency{}
	}
	return spec.Dependency{
		Name:    fields[0],
		Version: fields[1],
		Signal: spec.ObservedSignal("observed go module dependency",
			spec.EvidenceRef{Path: relPath, StartLine: lineNo, EndLine: lineNo}),
	}
}

func parsePackageJSON(relPath string, data []byte) ([]spec.Dependency, []spec.Hint) {
	type packageJSON struct {
		Name            string            `json:"name"`
		Type            string            `json:"type"`
		Main            string            `json:"main"`
		PackageManager  string            `json:"packageManager"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, []spec.Hint{manifestHint(relPath, "manifest.parse_error", "package.json")}
	}

	var deps []spec.Dependency
	for name, version := range pkg.Dependencies {
		deps = append(deps, spec.Dependency{
			Name:    name,
			Version: version,
			Signal: spec.ObservedSignal("observed package.json dependency",
				spec.EvidenceRef{Path: relPath}),
		})
	}

	var hints []spec.Hint
	if pkg.PackageManager != "" {
		hints = append(hints, manifestHint(relPath, "package.manager", pkg.PackageManager))
	}
	if pkg.Main != "" {
		hints = append(hints, manifestHint(relPath, "node.main", pkg.Main))
	}
	if buildScript, ok := pkg.Scripts["build"]; ok {
		hints = append(hints, manifestHint(relPath, "node.build_script", buildScript))
	}
	if pkg.Type != "" {
		hints = append(hints, manifestHint(relPath, "node.type", pkg.Type))
	}
	return deps, hints
}

func manifestHint(relPath, key, value string) spec.Hint {
	return spec.Hint{
		Key:   key,
		Value: value,
		Signal: spec.ObservedSignal("observed manifest hint",
			spec.EvidenceRef{Path: relPath}),
	}
}
