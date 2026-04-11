package parsers

import (
	"context"

	"projcompiler/internal/understanding/types"
)

type ParseResult struct {
	Symbols     []types.Symbol
	Relations   []types.Relation
	Evidences   []types.Evidence
	Confidences []types.Confidence
}

type Provider interface {
	Name() string
	Supports(file types.File) bool
	Parse(ctx context.Context, file types.File, src []byte) (ParseResult, error)
}
