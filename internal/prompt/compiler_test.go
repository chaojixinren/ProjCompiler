package prompt

import (
	"context"
	"strings"
	"testing"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

func TestCompilerPrunesGenericConstraints(t *testing.T) {
	compiler := NewCompiler(config.Config{})
	projectSpec := spec.ProjectSpec{
		Goal: "Create a terminal application that analyzes a repository and emits an implementation prompt",
		TechProfile: spec.TechProfile{
			Languages:      []string{"Go"},
			Frameworks:     []string{"Bubble Tea", "Google ADK"},
			BuildSystem:    "Go toolchain",
			PackageManager: "Go Modules",
		},
		CriticalConstraints: []spec.Constraint{
			{Text: "Use Bubble Tea for the terminal UI", Signal: spec.ObservedSignal("required by repo design")},
			{Text: "Follow best practices and write clean code", Signal: spec.InferredSignal(0.4, "generic suggestion")},
		},
		AcceptanceChecks: []spec.AcceptanceCheck{
			{Description: "The CLI should emit a pure English prompt with no tool wrapper", Signal: spec.ObservedSignal("explicit requirement")},
		},
	}

	bundle, err := compiler.CompilePrompt(context.Background(), projectSpec)
	if err != nil {
		t.Fatalf("CompilePrompt returned error: %v", err)
	}
	if !bundle.Ready() {
		t.Fatalf("expected ready prompt bundle")
	}
	if strings.Contains(strings.ToLower(bundle.PromptText), "best practices") {
		t.Fatalf("expected generic advice to be pruned, got %q", bundle.PromptText)
	}
	if !strings.Contains(bundle.PromptText, "Bubble Tea") {
		t.Fatalf("expected prompt to retain non-obvious constraint, got %q", bundle.PromptText)
	}
}

func TestCompilerStopsOnBlockingQuestions(t *testing.T) {
	compiler := NewCompiler(config.Config{})
	projectSpec := spec.ProjectSpec{
		OpenQuestions: []spec.OpenQuestion{
			{Question: "Which subproject in the monorepo is the real target?", Blocking: true, Signal: spec.MissingSignal("not discoverable from code alone")},
		},
	}

	bundle, err := compiler.CompilePrompt(context.Background(), projectSpec)
	if err != nil {
		t.Fatalf("CompilePrompt returned error: %v", err)
	}
	if bundle.PromptText != "" {
		t.Fatalf("expected no prompt text when blocking questions exist, got %q", bundle.PromptText)
	}
	if len(bundle.BlockingQuestions) != 1 {
		t.Fatalf("expected blocking questions to be preserved")
	}
}
