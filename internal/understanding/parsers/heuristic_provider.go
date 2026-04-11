package parsers

import (
	"context"
	"path/filepath"

	"projcompiler/internal/understanding/types"
)

type HeuristicProvider struct{}

func NewHeuristicProvider() *HeuristicProvider {
	return &HeuristicProvider{}
}

func (p *HeuristicProvider) Name() string {
	return "heuristic"
}

func (p *HeuristicProvider) Supports(file types.File) bool {
	return file.Role == types.FileRoleSource || file.Role == types.FileRoleTest
}

func (p *HeuristicProvider) Parse(ctx context.Context, file types.File, _ []byte) (ParseResult, error) {
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}

	modName := filepath.Base(file.Path)
	span := types.Span{}
	ev := types.Evidence{
		ID:          types.NewID("ev", file.ID, "heuristic.module"),
		SourceKind:  "heuristic.module",
		Path:        file.Path,
		Span:        span,
		Extractor:   "heuristic-provider/v1",
		SnippetHash: file.ContentHash,
		Note:        "fallback module symbol",
	}
	conf := types.Confidence{
		ID:         types.NewID("conf", "inferred", "heuristic-module", file.ID),
		Kind:       "inferred",
		Score:      0.4,
		Band:       types.ConfidenceBandFor(0.4),
		ReasonCode: "heuristic-module",
	}
	sym := types.Symbol{
		ID:            types.NewID(file.ID, "module", modName),
		FileID:        file.ID,
		Kind:          types.SymbolKindModule,
		Name:          modName,
		QualifiedName: modName,
		Span:          span,
		SourceCapture: "@definition.module",
		EvidenceIDs:   []string{ev.ID},
	}
	return ParseResult{
		Symbols:     []types.Symbol{sym},
		Evidences:   []types.Evidence{ev},
		Confidences: []types.Confidence{conf},
	}, nil
}
