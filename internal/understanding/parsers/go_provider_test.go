package parsers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"projcompiler/internal/understanding/types"
)

func TestGoProviderExtractsStructsAndInterfaces(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "types.go")
	code := `package config

type Config struct {
	AccessKey string
	Strategy  string
	Upstreams []Upstream
}

type Upstream struct {
	Name    string
	BaseURL string
}

type Handler interface {
	ServeHTTP()
}
`
	if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	file := types.File{
		ID:          "file-types",
		Path:        "types.go",
		AbsPath:     path,
		Language:    "go",
		Role:        types.FileRoleSource,
		ContentHash: "hash-types",
	}

	provider := NewGoProvider()
	result, err := provider.Parse(context.Background(), file, []byte(code))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	var hasConfigStruct, hasUpstreamStruct, hasHandlerIface bool
	for _, sym := range result.Symbols {
		switch {
		case sym.Kind == types.SymbolKindStruct && sym.Name == "Config":
			hasConfigStruct = true
		case sym.Kind == types.SymbolKindStruct && sym.Name == "Upstream":
			hasUpstreamStruct = true
		case sym.Kind == types.SymbolKindInterface && sym.Name == "Handler":
			hasHandlerIface = true
		}
	}
	if !hasConfigStruct {
		t.Fatal("expected Config struct symbol")
	}
	if !hasUpstreamStruct {
		t.Fatal("expected Upstream struct symbol")
	}
	if !hasHandlerIface {
		t.Fatal("expected Handler interface symbol")
	}

	var fieldCount int
	for _, rel := range result.Relations {
		if rel.Type == types.RelationDefinesField {
			fieldCount++
		}
	}
	if fieldCount < 5 {
		t.Fatalf("expected at least 5 field relations (3 Config + 2 Upstream), got %d", fieldCount)
	}
}

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
