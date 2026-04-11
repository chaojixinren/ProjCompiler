package projection

import (
	"testing"
	"time"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/types"
)

func TestToProjectSpecProducesUnifiedUnderstanding(t *testing.T) {
	t.Parallel()

	out := ToProjectSpec(UnifiedInput{
		Repo: types.RepoIdentity{
			RootPath: "/tmp/repo",
			RepoName: "repo",
			VCSRef:   "abc123",
		},
		Snapshot: types.SnapshotMeta{
			RunID:          "run-1",
			BuiltAt:        time.Now(),
			ScannerVersion: "scan/v1",
			ParserVersion:  "parse/v1",
			Mode:           "quick",
		},
		Files: []types.File{
			{
				ID:       "f1",
				Path:     "main.go",
				Language: "go",
				Role:     types.FileRoleSource,
			},
		},
		Symbols: []types.Symbol{
			{
				ID:           "s1",
				FileID:       "f1",
				Name:         "main",
				Kind:         types.SymbolKindFunction,
				IsEntrypoint: true,
			},
		},
		Flows: []types.Flow{
			{
				ID:            "fl1",
				Name:          "entry:s1",
				Trigger:       "entrypoint",
				EntrySymbolID: "s1",
				StepSymbolIDs: []string{"s1"},
			},
		},
		Constraints: []types.Constraint{
			{
				ID:     "c1",
				Text:   "must run with go",
				Signal: spec.ObservedSignal("test"),
			},
		},
		Questions: []types.OpenQuestion{
			{
				ID:       "q1",
				Question: "where is config?",
				Blocking: false,
				Signal:   spec.InferredSignal(0.5, "test"),
			},
		},
	})

	if len(out.Understanding.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(out.Understanding.Files))
	}
	if len(out.ImplementationShape.CoreModules) == 0 {
		t.Fatal("expected at least one core module")
	}
	if len(out.ImplementationShape.KeyFlows) == 0 {
		t.Fatal("expected at least one key flow")
	}
	if len(out.OpenQuestions) != 1 {
		t.Fatalf("expected 1 open question, got %d", len(out.OpenQuestions))
	}
}
