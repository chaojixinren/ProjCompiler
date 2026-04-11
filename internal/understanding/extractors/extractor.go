package extractors

import (
	"context"
	"path/filepath"
	"slices"
	"strings"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/parsers"
	"projcompiler/internal/understanding/types"
)

type Input struct {
	RepoRoot string
	Facts    *spec.RepoFacts
	Files    []types.File
	Parsed   map[string]parsers.ParseResult
}

type Output struct {
	Documents   []types.Document
	Symbols     []types.Symbol
	Relations   []types.Relation
	Flows       []types.Flow
	Constraints []types.Constraint
	Evidences   []types.Evidence
	Confidences []types.Confidence
}

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Build(ctx context.Context, input Input) (Output, error) {
	if err := ctx.Err(); err != nil {
		return Output{}, err
	}

	out := Output{
		Documents:   make([]types.Document, 0, 8),
		Symbols:     make([]types.Symbol, 0, 64),
		Relations:   make([]types.Relation, 0, 64),
		Flows:       make([]types.Flow, 0, 8),
		Constraints: make([]types.Constraint, 0, 16),
		Evidences:   make([]types.Evidence, 0, 64),
		Confidences: make([]types.Confidence, 0, 64),
	}

	for _, file := range input.Files {
		if err := ctx.Err(); err != nil {
			return Output{}, err
		}
		result, ok := input.Parsed[file.ID]
		if !ok {
			continue
		}
		out.Symbols = append(out.Symbols, result.Symbols...)
		out.Relations = append(out.Relations, result.Relations...)
		out.Evidences = append(out.Evidences, result.Evidences...)
		out.Confidences = append(out.Confidences, result.Confidences...)
	}

	if input.Facts != nil {
		out.Documents = append(out.Documents, b.documentsFromFacts(input.Facts, input.Files, &out)...)
		constraints, confidences := b.constraintsFromFacts(input.Facts)
		out.Constraints = append(out.Constraints, constraints...)
		out.Confidences = append(out.Confidences, confidences...)
	}

	entrySymbolIDs := b.markEntrypoints(input, &out)
	out.Flows = append(out.Flows, b.inferEntryFlows(entrySymbolIDs, out.Relations)...)

	return out, nil
}

func (b *Builder) documentsFromFacts(facts *spec.RepoFacts, files []types.File, out *Output) []types.Document {
	byPath := make(map[string]types.File, len(files))
	for _, file := range files {
		byPath[file.Path] = file
	}

	docs := make([]types.Document, 0, len(facts.Docs))
	for _, doc := range facts.Docs {
		path := types.NormalizePath(doc.Path)
		file, ok := byPath[path]
		if !ok {
			continue
		}
		ev := types.Evidence{
			ID:          types.NewID("ev", "doc", path, doc.Title),
			SourceKind:  "doc.excerpt",
			Path:        path,
			Span:        types.Span{StartLine: 1, EndLine: 1},
			Extractor:   "facts-docs/v1",
			SnippetHash: file.ContentHash,
			Note:        doc.Title,
		}
		conf := types.Confidence{
			ID:         types.NewID("conf", "observed", "doc-fact", path),
			Kind:       "observed",
			Score:      doc.Signal.Confidence,
			Band:       types.ConfidenceBandFor(doc.Signal.Confidence),
			ReasonCode: "doc-fact",
		}
		out.Evidences = append(out.Evidences, ev)
		out.Confidences = append(out.Confidences, conf)
		docs = append(docs, types.Document{
			ID:            types.NewID("doc", path),
			FileID:        file.ID,
			Title:         strings.TrimSpace(doc.Title),
			Summary:       strings.TrimSpace(doc.Excerpt),
			ConstraintIDs: nil,
			EvidenceIDs:   []string{ev.ID},
		})
	}
	return docs
}

func (b *Builder) constraintsFromFacts(facts *spec.RepoFacts) ([]types.Constraint, []types.Confidence) {
	constraints := make([]types.Constraint, 0, len(facts.Constraints))
	confidences := make([]types.Confidence, 0, len(facts.Constraints))
	for _, constraint := range facts.Constraints {
		text := strings.TrimSpace(constraint.Text)
		if text == "" {
			continue
		}
		score := constraint.Signal.Confidence
		conf := types.Confidence{
			ID:         types.NewID("conf", "constraint", text),
			Kind:       string(constraint.Signal.Kind),
			Score:      score,
			Band:       types.ConfidenceBandFor(score),
			ReasonCode: "facts-constraint",
		}
		confidences = append(confidences, conf)
		constraints = append(constraints, types.Constraint{
			ID:           types.NewID("constraint", text),
			Text:         text,
			Category:     "runtime",
			Scope:        "repo",
			ConfidenceID: conf.ID,
			Signal:       constraint.Signal,
		})
	}
	return constraints, confidences
}

func (b *Builder) markEntrypoints(input Input, out *Output) []string {
	entrySymbolIDs := make([]string, 0, 8)
	if input.Facts == nil {
		for i, symbol := range out.Symbols {
			if symbol.Name != "main" {
				continue
			}
			out.Symbols[i].IsEntrypoint = true
			entrySymbolIDs = append(entrySymbolIDs, symbol.ID)
		}
		return entrySymbolIDs
	}

	entryPaths := make(map[string]spec.EntryPoint, len(input.Facts.EntryPoints))
	for _, entry := range input.Facts.EntryPoints {
		entryPaths[types.NormalizePath(entry.Path)] = entry
	}
	fileIDByPath := map[string]string{}
	for _, file := range input.Files {
		fileIDByPath[file.Path] = file.ID
	}

	for i, symbol := range out.Symbols {
		entry, ok := entryPaths[pathForFileID(symbol.FileID, input.Files)]
		if !ok {
			continue
		}
		if entry.Symbol != "" && !strings.EqualFold(entry.Symbol, symbol.Name) && !strings.EqualFold(entry.Symbol, symbol.QualifiedName) {
			continue
		}
		out.Symbols[i].IsEntrypoint = true
		entrySymbolIDs = append(entrySymbolIDs, symbol.ID)
	}

	for path, entry := range entryPaths {
		fileID, ok := fileIDByPath[path]
		if !ok {
			continue
		}
		found := false
		for _, symbolID := range entrySymbolIDs {
			if symbolBelongsToFile(symbolID, fileID, out.Symbols) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		span := types.Span{}
		ev := types.Evidence{
			ID:          types.NewID("ev", "entry-fallback", path, entry.Symbol),
			SourceKind:  "entrypoint.fact",
			Path:        path,
			Span:        span,
			Extractor:   "entrypoint-fallback/v1",
			SnippetHash: "",
			Note:        "entrypoint derived from facts",
		}
		out.Evidences = append(out.Evidences, ev)
		sym := types.Symbol{
			ID:            types.NewID("sym", fileID, "entry", entry.Symbol),
			FileID:        fileID,
			Kind:          types.SymbolKindFunction,
			Name:          fallbackEntryName(entry),
			QualifiedName: fallbackEntryName(entry),
			IsEntrypoint:  true,
			Span:          span,
			SourceCapture: "@definition.entrypoint",
			EvidenceIDs:   []string{ev.ID},
		}
		out.Symbols = append(out.Symbols, sym)
		entrySymbolIDs = append(entrySymbolIDs, sym.ID)
	}

	return slices.Compact(entrySymbolIDs)
}

func (b *Builder) inferEntryFlows(entrySymbolIDs []string, relations []types.Relation) []types.Flow {
	flows := make([]types.Flow, 0, len(entrySymbolIDs))
	for _, entryID := range entrySymbolIDs {
		flow := types.Flow{
			ID:            types.NewID("flow", entryID),
			Name:          "entry:" + entryID,
			Trigger:       "entrypoint",
			EntrySymbolID: entryID,
			StepSymbolIDs: []string{entryID},
			EvidenceIDs:   nil,
		}
		for _, relation := range relations {
			if relation.Type != types.RelationCalls || relation.FromID != entryID {
				continue
			}
			if relation.ToID != "" {
				flow.StepSymbolIDs = append(flow.StepSymbolIDs, relation.ToID)
			}
			flow.EvidenceIDs = append(flow.EvidenceIDs, relation.EvidenceIDs...)
		}
		flows = append(flows, flow)
	}
	return flows
}

func pathForFileID(fileID string, files []types.File) string {
	for _, file := range files {
		if file.ID == fileID {
			return file.Path
		}
	}
	return ""
}

func symbolBelongsToFile(symbolID, fileID string, symbols []types.Symbol) bool {
	for _, symbol := range symbols {
		if symbol.ID == symbolID && symbol.FileID == fileID {
			return true
		}
	}
	return false
}

func fallbackEntryName(entry spec.EntryPoint) string {
	if trimmed := strings.TrimSpace(entry.Symbol); trimmed != "" {
		return trimmed
	}
	base := filepath.Base(entry.Path)
	if base == "" {
		return "entrypoint"
	}
	return base
}
