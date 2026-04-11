package resolver

import (
	"context"
	"testing"

	"projcompiler/internal/understanding/types"
)

func TestResolveCallByName(t *testing.T) {
	t.Parallel()

	symbols := []types.Symbol{
		{ID: "a", Name: "main", PackageName: "main"},
		{ID: "b", Name: "hello", PackageName: "main"},
	}
	relations := []types.Relation{
		{
			ID:     "r1",
			Type:   types.RelationCalls,
			FromID: "a",
			ToRef:  "hello",
		},
	}

	r := NewResolver()
	out, err := r.Resolve(context.Background(), symbols, relations)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(out.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(out.Relations))
	}
	if out.Relations[0].ToID != "b" {
		t.Fatalf("expected relation resolved to b, got %q", out.Relations[0].ToID)
	}
}
