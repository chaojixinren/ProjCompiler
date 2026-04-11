package adkflow

import (
	"encoding/json"
	"strings"
	"time"

	"projcompiler/internal/spec"
)

type understandingSnapshot struct {
	SchemaVersion string                    `json:"schema_version"`
	GeneratedAt   string                    `json:"generated_at"`
	RootPath      string                    `json:"root_path"`
	ModulePath    string                    `json:"module_path,omitempty"`
	Summary       understandingSummary      `json:"summary"`
	ManifestFiles []string                  `json:"manifest_files,omitempty"`
	Dependencies  []string                  `json:"dependencies,omitempty"`
	EntryPoints   []understandingEntryPoint `json:"entry_points,omitempty"`
	Questions     []understandingQuestion   `json:"open_questions,omitempty"`
}

type understandingSummary struct {
	Documents   int `json:"documents"`
	EntryPoints int `json:"entry_points"`
	Snippets    int `json:"snippets"`
	Constraints int `json:"constraints"`
	Signals     int `json:"signals"`
}

type understandingEntryPoint struct {
	Path   string `json:"path"`
	Symbol string `json:"symbol,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Score  int    `json:"score,omitempty"`
}

type understandingQuestion struct {
	Question string `json:"question"`
	Blocking bool   `json:"blocking"`
}

func buildUnderstandingJSON(facts spec.RepoFacts) (string, error) {
	doc := understandingSnapshot{
		SchemaVersion: "v0",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		RootPath:      strings.TrimSpace(facts.RootPath),
		ModulePath:    strings.TrimSpace(facts.Build.ModulePath),
		Summary: understandingSummary{
			Documents:   len(facts.Docs),
			EntryPoints: len(facts.EntryPoints),
			Snippets:    len(facts.Snippets),
			Constraints: len(facts.Constraints),
			Signals:     len(facts.Signals),
		},
		ManifestFiles: append([]string(nil), facts.Build.ManifestFiles...),
		Dependencies:  uniqueDependencyNames(facts.Build.Dependencies),
		EntryPoints:   make([]understandingEntryPoint, 0, len(facts.EntryPoints)),
		Questions:     make([]understandingQuestion, 0),
	}

	for _, entry := range facts.EntryPoints {
		doc.EntryPoints = append(doc.EntryPoints, understandingEntryPoint{
			Path:   entry.Path,
			Symbol: entry.Symbol,
			Kind:   entry.Kind,
			Score:  entry.Score,
		})
	}

	for _, question := range deriveOpenQuestions(facts) {
		doc.Questions = append(doc.Questions, understandingQuestion{
			Question: question.Question,
			Blocking: question.Blocking,
		})
	}

	content, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func uniqueDependencyNames(deps []spec.Dependency) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(deps))
	for _, dep := range deps {
		name := strings.TrimSpace(dep.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// buildUnderstandingJSONFromSpec generates a debug JSON snapshot from the UnderstandingSpec.
// This is used as a fallback when the rich understanding pipeline is available.
func buildUnderstandingJSONFromSpec(understanding spec.UnderstandingSpec) (string, error) {
	if strings.TrimSpace(understanding.SchemaVersion) == "" {
		return "", nil
	}

	doc := understandingUnifiedSnapshot{
		SchemaVersion: understanding.SchemaVersion,
		RootPath:      strings.TrimSpace(understanding.Repo.RootPath),
		RepoName:      strings.TrimSpace(understanding.Repo.RepoName),
		Stats:         understanding.Stats,
		Flows:         make([]understandingFlowSnapshot, 0, min(5, len(understanding.Flows))),
		Relations:     make([]understandingRelationSnapshot, 0, min(8, len(understanding.Relations))),
		Evidences:     make([]understandingEvidenceSnapshot, 0, min(8, len(understanding.Evidences))),
	}

	for _, flow := range understanding.Flows {
		if len(doc.Flows) >= 5 {
			break
		}
		doc.Flows = append(doc.Flows, understandingFlowSnapshot{
			Name:      strings.TrimSpace(flow.Name),
			Trigger:   strings.TrimSpace(flow.Trigger),
			Entry:     strings.TrimSpace(flow.EntrySymbolID),
			StepCount: len(flow.StepSymbolIDs),
			Evidence:  append([]string(nil), flow.EvidenceIDs...),
		})
	}

	for _, relation := range understanding.Relations {
		if len(doc.Relations) >= 8 {
			break
		}
		doc.Relations = append(doc.Relations, understandingRelationSnapshot{
			Type: strings.TrimSpace(string(relation.Type)),
			From: strings.TrimSpace(relation.FromID),
			To:   firstNonEmpty(strings.TrimSpace(relation.ToID), strings.TrimSpace(relation.ToRef)),
		})
	}

	for _, evidence := range understanding.Evidences {
		if len(doc.Evidences) >= 8 {
			break
		}
		doc.Evidences = append(doc.Evidences, understandingEvidenceSnapshot{
			Path:       strings.TrimSpace(evidence.Path),
			SourceKind: strings.TrimSpace(evidence.SourceKind),
			Note:       strings.TrimSpace(evidence.Note),
		})
	}

	content, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

type understandingUnifiedSnapshot struct {
	SchemaVersion string                          `json:"schema_version"`
	RootPath      string                          `json:"root_path"`
	RepoName      string                          `json:"repo_name,omitempty"`
	Stats         spec.UnderstandingStats         `json:"stats"`
	Flows         []understandingFlowSnapshot     `json:"flows,omitempty"`
	Relations     []understandingRelationSnapshot `json:"relations,omitempty"`
	Evidences     []understandingEvidenceSnapshot `json:"evidences,omitempty"`
}

type understandingFlowSnapshot struct {
	Name      string   `json:"name"`
	Trigger   string   `json:"trigger,omitempty"`
	Entry     string   `json:"entry,omitempty"`
	StepCount int      `json:"step_count"`
	Evidence  []string `json:"evidence_ids,omitempty"`
}

type understandingRelationSnapshot struct {
	Type string `json:"type"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type understandingEvidenceSnapshot struct {
	Path       string `json:"path,omitempty"`
	SourceKind string `json:"source_kind,omitempty"`
	Note       string `json:"note,omitempty"`
}
