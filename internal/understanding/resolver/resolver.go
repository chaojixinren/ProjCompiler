package resolver

import (
	"context"
	"strings"

	"projcompiler/internal/understanding/types"
)

type Result struct {
	Relations   []types.Relation
	Confidences []types.Confidence
}

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(ctx context.Context, symbols []types.Symbol, relations []types.Relation) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	byID := make(map[string]types.Symbol, len(symbols))
	byName := make(map[string][]types.Symbol, len(symbols))
	for _, symbol := range symbols {
		byID[symbol.ID] = symbol
		byName[symbol.Name] = append(byName[symbol.Name], symbol)
		if symbol.QualifiedName != "" {
			byName[symbol.QualifiedName] = append(byName[symbol.QualifiedName], symbol)
		}
	}

	resolved := make([]types.Relation, 0, len(relations))
	confidences := make([]types.Confidence, 0, 16)
	for _, relation := range relations {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if relation.Type != types.RelationCalls || relation.ToID != "" || relation.ToRef == "" {
			resolved = append(resolved, relation)
			continue
		}
		fromSym, ok := byID[relation.FromID]
		target := resolveTarget(fromSym, relation.ToRef, byName)
		if !ok || target.ID == "" {
			conf := types.Confidence{
				ID:         types.NewID("conf", relation.ID, "unresolved"),
				Kind:       "missing",
				Score:      0,
				Band:       types.ConfidenceBandLow,
				ReasonCode: "call-unresolved",
			}
			confidences = append(confidences, conf)
			relation.Resolver = "name-matching"
			relation.ConfidenceID = conf.ID
			resolved = append(resolved, relation)
			continue
		}

		score := 0.7
		reason := "name-match"
		if fromSym.PackageName != "" && target.PackageName != "" && fromSym.PackageName == target.PackageName {
			score = 0.9
			reason = "same-package-name-match"
		}
		conf := types.Confidence{
			ID:         types.NewID("conf", relation.ID, reason),
			Kind:       "inferred",
			Score:      score,
			Band:       types.ConfidenceBandFor(score),
			ReasonCode: reason,
		}
		confidences = append(confidences, conf)
		relation.ToID = target.ID
		relation.Resolver = "name-matching"
		relation.ConfidenceID = conf.ID
		resolved = append(resolved, relation)
	}

	return Result{
		Relations:   resolved,
		Confidences: confidences,
	}, nil
}

func resolveTarget(from types.Symbol, toRef string, byName map[string][]types.Symbol) types.Symbol {
	candidates := byName[toRef]
	if len(candidates) == 0 {
		parts := strings.Split(toRef, ".")
		candidates = byName[parts[len(parts)-1]]
	}
	if len(candidates) == 0 {
		return types.Symbol{}
	}
	if from.PackageName == "" {
		return candidates[0]
	}
	for _, candidate := range candidates {
		if candidate.PackageName == from.PackageName {
			return candidate
		}
	}
	return candidates[0]
}
