package adkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"projcompiler/internal/app"
	"projcompiler/internal/config"
	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/pipeline"
)

type SpecBuilder struct {
	client               CompletionClient
	understandingBuilder *pipeline.Builder
}

var _ app.SpecBuilder = (*SpecBuilder)(nil)

func NewSpecBuilder(cfg config.Config) *SpecBuilder {
	return &SpecBuilder{
		client:               NewCompletionClient(cfg),
		understandingBuilder: pipeline.NewBuilder(),
	}
}

func (b *SpecBuilder) BuildProjectSpec(ctx context.Context, facts spec.RepoFacts) (spec.ProjectSpec, error) {
	if err := ctx.Err(); err != nil {
		return spec.ProjectSpec{}, err
	}

	fallback := buildFallbackProjectSpec(facts)
	baseline := fallback
	understandingJSON := ""

	if b.understandingBuilder != nil {
		out, buildErr := b.understandingBuilder.Build(ctx, pipeline.BuildRequest{
			RootPath: facts.RootPath,
			Facts:    &facts,
		})
		if buildErr != nil {
			baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
				Label:  "understanding_pipeline_fallback",
				Source: "Unified understanding pipeline unavailable: " + buildErr.Error(),
				Signal: spec.InferredSignal(0.4, "Spec synthesis continued with deterministic repository facts"),
			})
		} else {
			baseline = mergeProjectSpec(fallback, out.ProjectSpec)
			baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
				Label:  "understanding_pipeline_embedded",
				Source: "Unified understanding pipeline baseline was merged before analyzer refinement",
				Signal: spec.ObservedSignal("SpecBuilder executed internal understanding pipeline before analyzer refinement"),
			})
			understandingJSON, _ = buildUnderstandingJSONFromSpec(baseline.Understanding)
		}
	}

	if understandingJSON == "" {
		snapshotJSON, err := buildUnderstandingJSON(facts)
		if err != nil {
			baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
				Label:  "understanding_snapshot_unavailable",
				Source: "Understanding snapshot unavailable: " + err.Error(),
				Signal: spec.InferredSignal(0.4, "Analyzer context continued with repository facts only"),
			})
		} else {
			understandingJSON = snapshotJSON
			baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
				Label:  "understanding_snapshot_debug",
				Source: "Fallback understanding snapshot was included for analyzer context (debug only)",
				Signal: spec.InferredSignal(0.5, "Analyzer used a lightweight understanding snapshot as a debug fallback"),
			})
		}
	}

	if b.client == nil {
		return baseline, nil
	}

	result, err := b.client.Complete(ctx, []Message{
		{Role: "system", Content: analyzerSystemPrompt()},
		{Role: "user", Content: analyzerUserPrompt(facts, understandingJSON)},
	}, 0.1)
	if err != nil {
		baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
			Label:  "analyzer_fallback",
			Source: "LLM synthesis unavailable: " + err.Error(),
			Signal: spec.InferredSignal(0.4, "Fell back to deterministic spec synthesis"),
		})
		return baseline, nil
	}

	llmSpec, err := parseAnalyzerResponse(result.Text)
	if err != nil {
		baseline.EvidenceLog = append(baseline.EvidenceLog, spec.EvidenceItem{
			Label:  "analyzer_parse_fallback",
			Source: "LLM response could not be parsed as structured JSON",
			Signal: spec.InferredSignal(0.4, "Fell back to deterministic spec synthesis"),
		})
		return baseline, nil
	}

	return mergeProjectSpec(baseline, llmSpec), nil
}

func analyzerSystemPrompt() string {
	return strings.TrimSpace(`
You are the Analyzer stage of ProjCompiler.

Turn repository facts into a compact project specification.
Only include facts that are grounded in the provided repository summary.
Do not invent team process, hidden infrastructure, or product scope that is not supported by the repository facts.
When the evidence is incomplete, leave fields empty or emit an open question instead of guessing.

Return JSON only with this shape:
{
  "goal": "string",
  "tech_profile": {
    "platform": "string",
    "languages": ["string"],
    "frameworks": ["string"],
    "build_system": "string",
    "package_manager": "string",
    "external_integrations": ["string"]
  },
  "implementation_shape": {
    "core_modules": [{"name": "string", "responsibility": "string"}],
    "key_flows": [{"name": "string", "summary": "string"}]
  },
  "critical_constraints": ["string"],
  "acceptance_checks": ["string"],
  "open_questions": [{"question": "string", "blocking": true}]
}
`)
}

func analyzerUserPrompt(facts spec.RepoFacts, understandingJSON string) string {
	data, err := json.MarshalIndent(facts, "", "  ")
	if err != nil {
		return fmt.Sprintf("Repository facts could not be JSON-encoded cleanly: %v", err)
	}

	if understandingJSON == "" {
		return strings.TrimSpace(`
Project facts:
` + string(data) + `

Write the smallest useful project spec.
Use the same pruning rule as ClaudeCode /init:
- keep only details the model is likely to get wrong without help
- drop generic engineering advice
- prefer short, concrete statements over cataloging the whole repository
`)
	}

	return strings.TrimSpace(`
Project facts:
` + string(data) + `

Understanding snapshot:
` + understandingJSON + `

Write the smallest useful project spec.
Use the same pruning rule as ClaudeCode /init:
- keep only details the model is likely to get wrong without help
- drop generic engineering advice
- prefer short, concrete statements over cataloging the whole repository
`)
}

func parseAnalyzerResponse(raw string) (spec.ProjectSpec, error) {
	cleaned := extractJSONObject(raw)
	if cleaned == "" {
		return spec.ProjectSpec{}, fmt.Errorf("no JSON object found in analyzer response")
	}

	var decoded analyzerResponse
	if err := json.Unmarshal([]byte(cleaned), &decoded); err != nil {
		return spec.ProjectSpec{}, fmt.Errorf("decode analyzer response: %w", err)
	}

	projectSpec := spec.ProjectSpec{
		Goal: strings.TrimSpace(decoded.Goal),
		TechProfile: spec.TechProfile{
			Platform:             strings.TrimSpace(decoded.TechProfile.Platform),
			Languages:            nonEmptyStrings(decoded.TechProfile.Languages),
			Frameworks:           nonEmptyStrings(decoded.TechProfile.Frameworks),
			BuildSystem:          strings.TrimSpace(decoded.TechProfile.BuildSystem),
			PackageManager:       strings.TrimSpace(decoded.TechProfile.PackageManager),
			ExternalIntegrations: nonEmptyStrings(decoded.TechProfile.ExternalIntegrations),
		},
		CriticalConstraints: make([]spec.Constraint, 0, len(decoded.CriticalConstraints)),
		AcceptanceChecks:    make([]spec.AcceptanceCheck, 0, len(decoded.AcceptanceChecks)),
		OpenQuestions:       make([]spec.OpenQuestion, 0, len(decoded.OpenQuestions)),
		EvidenceLog: []spec.EvidenceItem{
			{
				Label:  "llm_synthesis",
				Source: "OpenAI-compatible analyzer synthesis",
				Signal: spec.InferredSignal(0.7, "Project spec was refined by the analyzer model"),
			},
		},
	}

	for _, module := range decoded.ImplementationShape.CoreModules {
		if strings.TrimSpace(module.Name) == "" || strings.TrimSpace(module.Responsibility) == "" {
			continue
		}
		projectSpec.ImplementationShape.CoreModules = append(projectSpec.ImplementationShape.CoreModules, spec.ModuleSpec{
			Name:           strings.TrimSpace(module.Name),
			Responsibility: strings.TrimSpace(module.Responsibility),
			Signal:         spec.InferredSignal(0.7, "Synthesized by analyzer model from repository facts"),
		})
	}

	for _, flow := range decoded.ImplementationShape.KeyFlows {
		if strings.TrimSpace(flow.Name) == "" || strings.TrimSpace(flow.Summary) == "" {
			continue
		}
		projectSpec.ImplementationShape.KeyFlows = append(projectSpec.ImplementationShape.KeyFlows, spec.KeyFlow{
			Name:    strings.TrimSpace(flow.Name),
			Summary: strings.TrimSpace(flow.Summary),
			Signal:  spec.InferredSignal(0.7, "Synthesized by analyzer model from repository facts"),
		})
	}

	for _, constraint := range decoded.CriticalConstraints {
		if trimmed := strings.TrimSpace(constraint); trimmed != "" {
			projectSpec.CriticalConstraints = append(projectSpec.CriticalConstraints, spec.Constraint{
				Text:   trimmed,
				Signal: spec.InferredSignal(0.7, "Synthesized by analyzer model from repository facts"),
			})
		}
	}

	for _, check := range decoded.AcceptanceChecks {
		if trimmed := strings.TrimSpace(check); trimmed != "" {
			projectSpec.AcceptanceChecks = append(projectSpec.AcceptanceChecks, spec.AcceptanceCheck{
				Description: trimmed,
				Signal:      spec.InferredSignal(0.7, "Synthesized by analyzer model from repository facts"),
			})
		}
	}

	for _, question := range decoded.OpenQuestions {
		if strings.TrimSpace(question.Question) == "" {
			continue
		}
		projectSpec.OpenQuestions = append(projectSpec.OpenQuestions, spec.OpenQuestion{
			Question: strings.TrimSpace(question.Question),
			Blocking: question.Blocking,
			Signal:   spec.InferredSignal(0.6, "Analyzer identified a gap that code alone may not answer"),
		})
	}

	return projectSpec, nil
}

func buildFallbackProjectSpec(facts spec.RepoFacts) spec.ProjectSpec {
	projectSpec := spec.ProjectSpec{
		Goal:                deriveGoal(facts),
		TechProfile:         deriveTechProfile(facts),
		ImplementationShape: deriveImplementationShape(facts),
		CriticalConstraints: slices.Clone(facts.Constraints),
		AcceptanceChecks:    deriveAcceptanceChecks(facts),
		OpenQuestions:       deriveOpenQuestions(facts),
		EvidenceLog:         deriveEvidenceLog(facts),
	}

	return projectSpec
}

func deriveGoal(facts spec.RepoFacts) string {
	for _, doc := range facts.Docs {
		title := strings.TrimSpace(doc.Title)
		excerpt := firstSentence(doc.Excerpt)
		switch {
		case title != "" && excerpt != "":
			return title + ": " + excerpt
		case title != "":
			return "Recreate the project described as " + title + "."
		case excerpt != "":
			return excerpt
		}
	}

	if facts.Build.ModulePath != "" {
		return "Recreate a project modeled on the repository at " + facts.Build.ModulePath + "."
	}

	if len(facts.EntryPoints) > 0 {
		return "Recreate the primary behavior of the repository's observed entry points."
	}

	base := filepath.Base(strings.TrimSpace(facts.RootPath))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "the provided repository"
	}
	return "Recreate a medium-sized project based on " + base + " without adding unrelated scope."
}

func deriveTechProfile(facts spec.RepoFacts) spec.TechProfile {
	return spec.TechProfile{
		Platform:             derivePlatform(facts),
		Languages:            deriveLanguages(facts),
		Frameworks:           deriveFrameworks(facts),
		BuildSystem:          deriveBuildSystem(facts),
		PackageManager:       derivePackageManager(facts),
		ExternalIntegrations: deriveExternalIntegrations(facts),
	}
}

func deriveImplementationShape(facts spec.RepoFacts) spec.ImplementationShape {
	shape := spec.ImplementationShape{
		CoreModules: make([]spec.ModuleSpec, 0, len(facts.EntryPoints)+2),
		KeyFlows:    make([]spec.KeyFlow, 0, len(facts.Snippets)),
	}

	for _, entry := range facts.EntryPoints {
		name := filepath.Base(entry.Path)
		if strings.TrimSpace(name) == "" {
			name = strings.TrimSpace(entry.Symbol)
		}
		responsibility := "Acts as a likely entry point"
		if entry.Kind != "" {
			responsibility = "Acts as a likely " + entry.Kind + " entry point"
		}
		shape.CoreModules = append(shape.CoreModules, spec.ModuleSpec{
			Name:           name,
			Responsibility: responsibility,
			Signal:         entry.Signal,
		})
	}

	for _, snippet := range facts.Snippets {
		if strings.TrimSpace(snippet.Purpose) == "" {
			continue
		}
		shape.KeyFlows = append(shape.KeyFlows, spec.KeyFlow{
			Name:    filepath.Base(snippet.Path),
			Summary: snippet.Purpose,
			Signal:  snippet.Signal,
		})
	}

	return shape
}

func deriveAcceptanceChecks(facts spec.RepoFacts) []spec.AcceptanceCheck {
	checks := make([]spec.AcceptanceCheck, 0, len(facts.EntryPoints)+2)

	if buildSystem := deriveBuildSystem(facts); buildSystem != "" {
		checks = append(checks, spec.AcceptanceCheck{
			Description: "The generated project should build successfully with " + buildSystem,
			Signal:      spec.InferredSignal(0.7, "Derived from manifest and build signals"),
		})
	}

	for _, entry := range facts.EntryPoints {
		description := "The primary behavior represented by " + entry.Path + " should be preserved."
		checks = append(checks, spec.AcceptanceCheck{
			Description: description,
			Signal:      entry.Signal,
		})
	}

	if len(checks) == 0 {
		checks = append(checks, spec.AcceptanceCheck{
			Description: "The generated project should preserve the core documented behavior of the repository without adding unrelated scope.",
			Signal:      spec.InferredSignal(0.5, "Fallback acceptance rule from repository summary"),
		})
	}

	return checks
}

func deriveOpenQuestions(facts spec.RepoFacts) []spec.OpenQuestion {
	questions := make([]spec.OpenQuestion, 0, 2)
	if len(facts.Docs) == 0 {
		questions = append(questions, spec.OpenQuestion{
			Question: "Should the generated project aim for a high-fidelity reproduction, or is a same-category abstraction acceptable?",
			Blocking: false,
			Signal:   spec.MissingSignal("No README or high-signal project documentation was observed"),
		})
	}
	if len(facts.EntryPoints) == 0 {
		questions = append(questions, spec.OpenQuestion{
			Question: "Which executable entry point or subproject should be treated as the canonical target?",
			Blocking: true,
			Signal:   spec.MissingSignal("No clear entry point was observed from the scanned repository facts"),
		})
	}
	return questions
}

func deriveEvidenceLog(facts spec.RepoFacts) []spec.EvidenceItem {
	out := make([]spec.EvidenceItem, 0, len(facts.Docs)+len(facts.EntryPoints)+2)
	if facts.Build.ModulePath != "" {
		out = append(out, spec.EvidenceItem{
			Label:  "module_path",
			Source: facts.Build.ModulePath,
			Signal: spec.ObservedSignal("Captured from repository manifest"),
		})
	}
	for _, doc := range facts.Docs {
		out = append(out, spec.EvidenceItem{
			Label:  "doc",
			Source: doc.Path,
			Signal: doc.Signal,
		})
	}
	for _, entry := range facts.EntryPoints {
		out = append(out, spec.EvidenceItem{
			Label:  "entry_point",
			Source: entry.Path,
			Signal: entry.Signal,
		})
	}
	return out
}

func derivePlatform(facts spec.RepoFacts) string {
	for _, constraint := range facts.Constraints {
		text := strings.ToLower(constraint.Text)
		switch {
		case strings.Contains(text, "macos"):
			return "macOS"
		case strings.Contains(text, "linux"):
			return "Linux"
		case strings.Contains(text, "windows"):
			return "Windows"
		}
	}
	return ""
}

func deriveLanguages(facts spec.RepoFacts) []string {
	seen := map[string]struct{}{}
	for _, node := range facts.Tree.Nodes {
		switch strings.ToLower(filepath.Ext(node.Path)) {
		case ".go":
			seen["Go"] = struct{}{}
		case ".ts", ".tsx":
			seen["TypeScript"] = struct{}{}
		case ".js", ".jsx":
			seen["JavaScript"] = struct{}{}
		case ".py":
			seen["Python"] = struct{}{}
		case ".swift":
			seen["Swift"] = struct{}{}
		case ".rs":
			seen["Rust"] = struct{}{}
		}
	}
	return mapKeys(seen)
}

func deriveFrameworks(facts spec.RepoFacts) []string {
	frameworks := map[string]struct{}{}
	for _, dep := range facts.Build.Dependencies {
		name := strings.ToLower(dep.Name)
		switch {
		case strings.Contains(name, "bubbletea"):
			frameworks["Bubble Tea"] = struct{}{}
		case strings.Contains(name, "google.golang.org/adk"):
			frameworks["Google ADK"] = struct{}{}
		case strings.Contains(name, "cobra"):
			frameworks["Cobra"] = struct{}{}
		case strings.Contains(name, "gin"):
			frameworks["Gin"] = struct{}{}
		case strings.Contains(name, "echo"):
			frameworks["Echo"] = struct{}{}
		case strings.Contains(name, "react"):
			frameworks["React"] = struct{}{}
		case strings.Contains(name, "next"):
			frameworks["Next.js"] = struct{}{}
		}
	}
	return mapKeys(frameworks)
}

func deriveBuildSystem(facts spec.RepoFacts) string {
	for _, manifest := range facts.Build.ManifestFiles {
		switch filepath.Base(manifest) {
		case "go.mod":
			return "the Go toolchain"
		case "package.json":
			return "the Node.js toolchain"
		case "Cargo.toml":
			return "Cargo"
		case "Package.swift":
			return "Swift Package Manager"
		}
	}
	return ""
}

func derivePackageManager(facts spec.RepoFacts) string {
	for _, manifest := range facts.Build.ManifestFiles {
		switch filepath.Base(manifest) {
		case "go.mod":
			return "Go Modules"
		case "package.json":
			return "an npm-compatible package manager"
		case "Cargo.toml":
			return "Cargo"
		case "Package.swift":
			return "Swift Package Manager"
		}
	}
	return ""
}

func deriveExternalIntegrations(facts spec.RepoFacts) []string {
	integrations := map[string]struct{}{}
	for _, dep := range facts.Build.Dependencies {
		name := strings.ToLower(dep.Name)
		switch {
		case strings.Contains(name, "openai"):
			integrations["OpenAI-compatible LLM API"] = struct{}{}
		case strings.Contains(name, "postgres"):
			integrations["PostgreSQL"] = struct{}{}
		case strings.Contains(name, "mysql"):
			integrations["MySQL"] = struct{}{}
		case strings.Contains(name, "redis"):
			integrations["Redis"] = struct{}{}
		case strings.Contains(name, "aws"):
			integrations["AWS services"] = struct{}{}
		}
	}
	return mapKeys(integrations)
}

func mergeProjectSpec(baseline, llmSpec spec.ProjectSpec) spec.ProjectSpec {
	merged := baseline
	if strings.TrimSpace(llmSpec.Goal) != "" {
		merged.Goal = strings.TrimSpace(llmSpec.Goal)
	}

	merged.TechProfile = mergeTechProfile(baseline.TechProfile, llmSpec.TechProfile)
	if len(llmSpec.ImplementationShape.CoreModules) > 0 {
		merged.ImplementationShape.CoreModules = llmSpec.ImplementationShape.CoreModules
	}
	if len(llmSpec.ImplementationShape.KeyFlows) > 0 {
		merged.ImplementationShape.KeyFlows = llmSpec.ImplementationShape.KeyFlows
	}
	if len(llmSpec.CriticalConstraints) > 0 {
		merged.CriticalConstraints = llmSpec.CriticalConstraints
	}
	if len(llmSpec.AcceptanceChecks) > 0 {
		merged.AcceptanceChecks = llmSpec.AcceptanceChecks
	}
	if len(llmSpec.OpenQuestions) > 0 {
		merged.OpenQuestions = llmSpec.OpenQuestions
	}
	if len(llmSpec.EvidenceLog) > 0 {
		merged.EvidenceLog = append(merged.EvidenceLog, llmSpec.EvidenceLog...)
	}
	if !isUnderstandingSpecEmpty(merged.Understanding) && isUnderstandingSpecEmpty(llmSpec.Understanding) {
		// Keep baseline's rich understanding
	} else if !isUnderstandingSpecEmpty(llmSpec.Understanding) {
		merged.Understanding = llmSpec.Understanding
	}
	return merged
}

func isUnderstandingSpecEmpty(value spec.UnderstandingSpec) bool {
	return strings.TrimSpace(value.SchemaVersion) == "" &&
		len(value.Files) == 0 &&
		len(value.Documents) == 0 &&
		len(value.Symbols) == 0 &&
		len(value.Relations) == 0 &&
		len(value.Flows) == 0 &&
		len(value.Evidences) == 0 &&
		len(value.Confidences) == 0
}

func mergeTechProfile(base, override spec.TechProfile) spec.TechProfile {
	merged := base
	merged.Platform = firstNonEmpty(override.Platform, base.Platform)
	merged.Languages = chooseStrings(override.Languages, base.Languages)
	merged.Frameworks = chooseStrings(override.Frameworks, base.Frameworks)
	merged.BuildSystem = firstNonEmpty(override.BuildSystem, base.BuildSystem)
	merged.PackageManager = firstNonEmpty(override.PackageManager, base.PackageManager)
	merged.ExternalIntegrations = chooseStrings(override.ExternalIntegrations, base.ExternalIntegrations)
	return merged
}

func chooseStrings(preferred, fallback []string) []string {
	if len(preferred) == 0 {
		return fallback
	}
	return nonEmptyStrings(preferred)
}

func nonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func mapKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	slices.Sort(out)
	return out
}

func firstSentence(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	if idx := strings.Index(trimmed, "."); idx > 0 {
		return strings.TrimSpace(trimmed[:idx+1])
	}
	return trimmed
}

func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return trimmed[start : end+1]
}

type analyzerResponse struct {
	Goal        string `json:"goal"`
	TechProfile struct {
		Platform             string   `json:"platform"`
		Languages            []string `json:"languages"`
		Frameworks           []string `json:"frameworks"`
		BuildSystem          string   `json:"build_system"`
		PackageManager       string   `json:"package_manager"`
		ExternalIntegrations []string `json:"external_integrations"`
	} `json:"tech_profile"`
	ImplementationShape struct {
		CoreModules []struct {
			Name           string `json:"name"`
			Responsibility string `json:"responsibility"`
		} `json:"core_modules"`
		KeyFlows []struct {
			Name    string `json:"name"`
			Summary string `json:"summary"`
		} `json:"key_flows"`
	} `json:"implementation_shape"`
	CriticalConstraints []string `json:"critical_constraints"`
	AcceptanceChecks    []string `json:"acceptance_checks"`
	OpenQuestions       []struct {
		Question string `json:"question"`
		Blocking bool   `json:"blocking"`
	} `json:"open_questions"`
}
