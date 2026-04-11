package projection

import (
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/types"
)

type UnifiedInput struct {
	BaseFacts   *spec.RepoFacts
	Repo        types.RepoIdentity
	Snapshot    types.SnapshotMeta
	Files       []types.File
	Documents   []types.Document
	Symbols     []types.Symbol
	Relations   []types.Relation
	Flows       []types.Flow
	Constraints []types.Constraint
	Questions   []types.OpenQuestion
	Evidences   []types.Evidence
	Confidences []types.Confidence
}

func ToProjectSpec(input UnifiedInput) spec.ProjectSpec {
	specOut := spec.ProjectSpec{
		Goal:                deriveGoal(input),
		TechProfile:         deriveTechProfile(input.Files),
		ImplementationShape: deriveShape(input.Symbols, input.Flows, input.Confidences),
		CriticalConstraints: cloneConstraints(input.Constraints),
		OpenQuestions:       cloneQuestions(input.Questions),
		EvidenceLog:         deriveEvidenceLog(input.Evidences, input.Confidences),
		AcceptanceChecks:    deriveAcceptanceChecks(input.Symbols),
		Understanding: spec.UnderstandingSpec{
			SchemaVersion: types.SchemaVersion,
			Repo:          input.Repo,
			Snapshot:      input.Snapshot,
			Files:         slices.Clone(input.Files),
			Documents:     slices.Clone(input.Documents),
			Symbols:       slices.Clone(input.Symbols),
			Relations:     slices.Clone(input.Relations),
			Flows:         slices.Clone(input.Flows),
			Evidences:     slices.Clone(input.Evidences),
			Confidences:   slices.Clone(input.Confidences),
		},
	}

	specOut.Understanding.Stats = spec.UnderstandingStats{
		FileCount:       len(specOut.Understanding.Files),
		SymbolCount:     len(specOut.Understanding.Symbols),
		RelationCount:   len(specOut.Understanding.Relations),
		FlowCount:       len(specOut.Understanding.Flows),
		QuestionCount:   len(specOut.OpenQuestions),
		ConstraintCount: len(specOut.CriticalConstraints),
	}

	return specOut
}

func deriveGoal(input UnifiedInput) string {
	if input.BaseFacts != nil {
		for _, doc := range input.BaseFacts.Docs {
			if strings.TrimSpace(doc.Excerpt) != "" {
				return strings.TrimSpace(doc.Excerpt)
			}
			if strings.TrimSpace(doc.Title) != "" {
				return strings.TrimSpace(doc.Title)
			}
		}
	}
	repoName := strings.TrimSpace(input.Repo.RepoName)
	if repoName == "" {
		return "Reconstruct and operate the repository behavior with minimal deviation."
	}
	return "Reconstruct and operate " + repoName + " with minimal deviation."
}

func deriveTechProfile(files []types.File) spec.TechProfile {
	languages := make([]string, 0, 8)
	langSet := map[string]struct{}{}
	var buildSystem string
	var packageManager string

	for _, file := range files {
		if file.Language != "" && file.Language != "unknown" {
			if _, ok := langSet[file.Language]; !ok {
				langSet[file.Language] = struct{}{}
				languages = append(languages, file.Language)
			}
		}
		base := strings.ToLower(filepath.Base(file.Path))
		switch base {
		case "go.mod":
			buildSystem = "go"
			packageManager = "go modules"
		case "package.json":
			if buildSystem == "" {
				buildSystem = "node"
			}
			packageManager = "npm"
		case "cargo.toml":
			if buildSystem == "" {
				buildSystem = "cargo"
			}
			packageManager = "cargo"
		}
	}
	sort.Strings(languages)
	return spec.TechProfile{
		Languages:      languages,
		BuildSystem:    buildSystem,
		PackageManager: packageManager,
		Platform:       inferPlatform(languages),
	}
}

func inferPlatform(languages []string) string {
	for _, lang := range languages {
		if lang == "go" {
			return "backend/service"
		}
	}
	if len(languages) > 0 {
		return "software"
	}
	return ""
}

func deriveShape(symbols []types.Symbol, flows []types.Flow, confidences []types.Confidence) spec.ImplementationShape {
	modules := make([]spec.ModuleSpec, 0, 12)
	moduleSeen := map[string]struct{}{}
	confidenceByID := make(map[string]types.Confidence, len(confidences))
	for _, confidence := range confidences {
		confidenceByID[confidence.ID] = confidence
	}

	for _, symbol := range symbols {
		if !symbol.IsEntrypoint && symbol.Kind != types.SymbolKindPackage {
			continue
		}
		moduleID := symbol.ID
		if _, seen := moduleSeen[moduleID]; seen {
			continue
		}
		moduleSeen[moduleID] = struct{}{}
		responsibility := "Core module"
		if symbol.IsEntrypoint {
			responsibility = "Primary entrypoint module"
		}
		modules = append(modules, spec.ModuleSpec{
			ID:             moduleID,
			Name:           choose(symbol.QualifiedName, symbol.Name),
			Responsibility: responsibility,
			SymbolIDs:      []string{symbol.ID},
			EvidenceIDs:    slices.Clone(symbol.EvidenceIDs),
			Signal:         signalFromConfidenceID("", confidenceByID),
		})
	}

	keyFlows := make([]spec.KeyFlow, 0, len(flows))
	for _, flow := range flows {
		keyFlows = append(keyFlows, spec.KeyFlow{
			ID:            flow.ID,
			Name:          flow.Name,
			Summary:       "Flow from " + flow.EntrySymbolID + " across " + countSteps(flow.StepSymbolIDs),
			Trigger:       flow.Trigger,
			EntrySymbolID: flow.EntrySymbolID,
			StepSymbolIDs: slices.Clone(flow.StepSymbolIDs),
			EvidenceIDs:   slices.Clone(flow.EvidenceIDs),
			Signal:        spec.InferredSignal(0.7, "Derived from relation traversal"),
		})
	}

	return spec.ImplementationShape{
		CoreModules: modules,
		KeyFlows:    keyFlows,
	}
}

func deriveAcceptanceChecks(symbols []types.Symbol) []spec.AcceptanceCheck {
	checks := make([]spec.AcceptanceCheck, 0, 4)
	for _, symbol := range symbols {
		if !symbol.IsEntrypoint {
			continue
		}
		checks = append(checks, spec.AcceptanceCheck{
			Description: "Entrypoint `" + symbol.Name + "` remains executable after generation.",
			Signal:      spec.InferredSignal(0.7, "Generated from entrypoint symbol"),
		})
	}
	if len(checks) == 0 {
		checks = append(checks, spec.AcceptanceCheck{
			Description: "Primary runtime behavior is preserved.",
			Signal:      spec.InferredSignal(0.5, "Fallback acceptance check"),
		})
	}
	return checks
}

func deriveEvidenceLog(evidences []types.Evidence, confidences []types.Confidence) []spec.EvidenceItem {
	confidenceByID := make(map[string]types.Confidence, len(confidences))
	for _, confidence := range confidences {
		confidenceByID[confidence.ID] = confidence
	}

	items := make([]spec.EvidenceItem, 0, len(evidences))
	for _, evidence := range evidences {
		items = append(items, spec.EvidenceItem{
			ID:          evidence.ID,
			Label:       evidence.SourceKind,
			Source:      evidence.Path,
			EvidenceIDs: []string{evidence.ID},
			Signal:      signalFromConfidenceID("", confidenceByID),
		})
	}
	return items
}

func cloneConstraints(items []types.Constraint) []spec.Constraint {
	out := make([]spec.Constraint, 0, len(items))
	for _, item := range items {
		constraint := item
		if constraint.Signal.Reason == "" {
			constraint.Signal = spec.InferredSignal(0.6, "constraint from understanding extraction")
		}
		out = append(out, constraint)
	}
	return out
}

func cloneQuestions(items []types.OpenQuestion) []spec.OpenQuestion {
	out := make([]spec.OpenQuestion, 0, len(items))
	for _, item := range items {
		question := item
		if question.Signal.Reason == "" {
			question.Signal = spec.InferredSignal(0.5, "question from ambiguity analysis")
		}
		out = append(out, question)
	}
	return out
}

func signalFromConfidenceID(confidenceID string, confidenceByID map[string]types.Confidence) spec.FactSignal {
	confidence, ok := confidenceByID[confidenceID]
	if !ok {
		return spec.InferredSignal(0.6, "projection-default")
	}

	switch confidence.Kind {
	case string(spec.SignalObserved):
		return spec.ObservedSignal(confidence.ReasonCode)
	case string(spec.SignalMissing):
		return spec.MissingSignal(confidence.ReasonCode)
	default:
		return spec.InferredSignal(confidence.Score, confidence.ReasonCode)
	}
}

func choose(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func countSteps(stepSymbolIDs []string) string {
	count := len(stepSymbolIDs)
	if count <= 0 {
		return "0 steps"
	}
	if count == 1 {
		return "1 step"
	}
	return strconv.Itoa(count) + " steps"
}
