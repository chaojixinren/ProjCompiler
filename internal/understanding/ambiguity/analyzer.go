package ambiguity

import (
	"context"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/types"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

type Input struct {
	Symbols   []types.Symbol
	Relations []types.Relation
	Documents []types.Document
}

func (a *Analyzer) Analyze(ctx context.Context, input Input) ([]types.OpenQuestion, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	questions := make([]types.OpenQuestion, 0, 4)
	if countEntrypoints(input.Symbols) == 0 {
		questions = append(questions, types.OpenQuestion{
			ID:          types.NewID("question", "entrypoint-missing"),
			Question:    "No clear entrypoint was detected. Which command/service should be treated as the primary runtime path?",
			Category:    "entrypoint",
			Blocking:    true,
			Reason:      "entrypoint-missing",
			Signal:      spec.MissingSignal("No entrypoint symbol detected"),
			EvidenceIDs: nil,
		})
	}

	for _, relation := range input.Relations {
		if relation.Type != types.RelationCalls {
			continue
		}
		if relation.ToID != "" {
			continue
		}
		questions = append(questions, types.OpenQuestion{
			ID:               types.NewID("question", "call-unresolved", relation.ID),
			Question:         "Call target `" + relation.ToRef + "` could not be resolved. Is this function provided by generated code or another package?",
			Category:         "dependency",
			Blocking:         false,
			Reason:           "call-unresolved",
			MissingEntityIDs: []string{relation.ID},
			EvidenceIDs:      relation.EvidenceIDs,
			Signal:           spec.MissingSignal("Call target unresolved"),
		})
	}

	if len(input.Documents) == 0 {
		questions = append(questions, types.OpenQuestion{
			ID:       types.NewID("question", "docs-missing"),
			Question: "No project documentation was detected. Is there a design/spec file that should be included?",
			Category: "scope",
			Blocking: false,
			Reason:   "docs-missing",
			Signal:   spec.InferredSignal(0.3, "Documentation sources not detected"),
		})
	}

	return dedupe(questions), nil
}

func countEntrypoints(symbols []types.Symbol) int {
	count := 0
	for _, symbol := range symbols {
		if symbol.IsEntrypoint {
			count++
		}
	}
	return count
}

func dedupe(items []types.OpenQuestion) []types.OpenQuestion {
	seen := make(map[string]struct{}, len(items))
	out := make([]types.OpenQuestion, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item)
	}
	return out
}
