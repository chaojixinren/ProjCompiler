package adkflow

import (
	"context"
	"strings"
	"testing"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

func TestBuildUnderstandingJSONEmitsSnapshot(t *testing.T) {
	facts := spec.RepoFacts{
		RootPath: "/tmp/repo",
		Build: spec.BuildSummary{
			ModulePath:    "example.com/repo",
			ManifestFiles: []string{"go.mod"},
			Dependencies: []spec.Dependency{
				{Name: "google.golang.org/adk"},
			},
		},
		EntryPoints: []spec.EntryPoint{
			{Path: "cmd/projcompiler/main.go", Kind: "cli", Score: 90},
		},
	}

	payload, err := buildUnderstandingJSON(facts)
	if err != nil {
		t.Fatalf("buildUnderstandingJSON returned error: %v", err)
	}
	if !strings.Contains(payload, `"schema_version": "v0"`) {
		t.Fatalf("missing schema_version in payload: %q", payload)
	}
	if !strings.Contains(payload, `"root_path": "/tmp/repo"`) {
		t.Fatalf("missing root_path in payload: %q", payload)
	}
}

func TestSpecBuilderEmbedsInternalUnderstandingEvidence(t *testing.T) {
	builder := NewSpecBuilder(config.Config{})
	projectSpec, err := builder.BuildProjectSpec(context.Background(), spec.RepoFacts{})
	if err != nil {
		t.Fatalf("BuildProjectSpec returned error: %v", err)
	}
	// The spec builder now calls understanding pipeline and emits understanding_pipeline_embedded
	// or understanding_snapshot_debug when the pipeline succeeds or fails
	hasUnderstandingEvidence := containsEvidence(projectSpec.EvidenceLog, "understanding_pipeline_embedded") ||
		containsEvidence(projectSpec.EvidenceLog, "understanding_snapshot_debug")
	if !hasUnderstandingEvidence {
		t.Fatalf("expected understanding pipeline evidence entry, got labels: %v", evidenceLabels(projectSpec.EvidenceLog))
	}
}

func evidenceLabels(items []spec.EvidenceItem) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

func containsEvidence(items []spec.EvidenceItem, label string) bool {
	for _, item := range items {
		if item.Label == label {
			return true
		}
	}
	return false
}
