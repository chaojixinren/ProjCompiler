package parsers

import (
	"context"
	"os"

	"projcompiler/internal/understanding/types"
)

type Registry struct {
	providers []Provider
}

func NewRegistry(providers ...Provider) *Registry {
	if len(providers) == 0 {
		providers = []Provider{
			NewGoProvider(),
			NewHeuristicProvider(),
		}
	}
	return &Registry{providers: providers}
}

func (r *Registry) ParseFile(ctx context.Context, file types.File) (ParseResult, string, error) {
	src, err := os.ReadFile(file.AbsPath)
	if err != nil {
		return ParseResult{}, "", err
	}
	for _, provider := range r.providers {
		if !provider.Supports(file) {
			continue
		}
		result, parseErr := provider.Parse(ctx, file, src)
		return result, provider.Name(), parseErr
	}
	return ParseResult{}, "", nil
}
