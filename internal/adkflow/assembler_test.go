package adkflow

import (
	"context"
	"testing"

	"projcompiler/internal/config"
)

func TestAssemblerBuildIncludesUnifiedSpecPipeline(t *testing.T) {
	assembler := NewAssembler(config.Config{
		Model: config.ModelConfig{
			BaseURL: "https://example.invalid/v1",
			APIKey:  "test-key",
			Model:   "test-model",
		},
	})

	workflow, err := assembler.Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	subAgents := workflow.SubAgents()
	if got, want := len(subAgents), 2; got != want {
		t.Fatalf("sub-agent count = %d, want %d", got, want)
	}
	if got, want := subAgents[0].Name(), "spec_analyzer"; got != want {
		t.Fatalf("first sub-agent = %q, want %q", got, want)
	}
	if got, want := subAgents[1].Name(), "prompt_compiler"; got != want {
		t.Fatalf("second sub-agent = %q, want %q", got, want)
	}
}
