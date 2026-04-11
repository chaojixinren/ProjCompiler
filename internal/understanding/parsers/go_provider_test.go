package parsers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/understanding/types"
)

func TestGoProviderExtractsSymbolsAndCalls(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	code := `package main

import "fmt"

func main() {
	hello()
}

func hello() {
	fmt.Println("hi")
}
`
	if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	file := types.File{
		ID:          "file1",
		Path:        "main.go",
		AbsPath:     path,
		Language:    "go",
		Role:        types.FileRoleSource,
		ContentHash: "hash",
	}

	provider := NewGoProvider()
	result, err := provider.Parse(context.Background(), file, []byte(code))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(result.Symbols) < 3 {
		t.Fatalf("expected at least 3 symbols, got %d", len(result.Symbols))
	}

	var hasMain bool
	var hasCall bool
	for _, symbol := range result.Symbols {
		if symbol.Name == "main" && symbol.IsEntrypoint {
			hasMain = true
		}
	}
	for _, relation := range result.Relations {
		if relation.Type == types.RelationCalls && relation.ToRef == "hello" {
			hasCall = true
		}
	}
	if !hasMain {
		t.Fatal("expected main symbol to be marked as entrypoint")
	}
	if !hasCall {
		t.Fatal("expected call relation to hello")
	}
}
