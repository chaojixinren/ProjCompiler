package prompt

import (
	"strings"
	"testing"

	"projcompiler/internal/spec"
)

func TestFilterPromptSectionsProducesOrderedPromptBrief(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		Goal: "Recreate a local CLI that scans a repository and emits an implementation prompt",
		TechProfile: spec.TechProfile{
			Platform:       "local terminal",
			Languages:      []string{"Go"},
			Frameworks:     []string{"Bubble Tea", "Google ADK"},
			BuildSystem:    "Go toolchain",
			PackageManager: "Go Modules",
		},
		CriticalConstraints: []spec.Constraint{
			{Text: "Only emit an English prompt and do not wrap it in tool-specific CLI commands"},
			{Text: "Write unit tests and keep the code maintainable"},
		},
		ImplementationShape: spec.ImplementationShape{
			CoreModules: []spec.ModuleSpec{
				{Name: "scanner", Responsibility: "Collect repository facts without calling the model"},
			},
			KeyFlows: []spec.KeyFlow{
				{Name: "prompt compilation", Summary: "Turn the distilled spec into a single English build brief"},
			},
		},
		AcceptanceChecks: []spec.AcceptanceCheck{
			{Description: "The output should be a pure English prompt"},
		},
	}

	sections := filterPromptSections(projectSpec)
	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(sections))
	}

	titles := []string{sections[0].Title, sections[1].Title, sections[2].Title, sections[3].Title}
	wantTitles := []string{"Goal", "Must-get-right constraints", "Non-obvious implementation shape", "Acceptance bar"}
	for i, want := range wantTitles {
		if titles[i] != want {
			t.Fatalf("section %d title = %q, want %q", i, titles[i], want)
		}
	}

	if strings.Contains(strings.ToLower(sections[1].Content), "maintainable") {
		t.Fatalf("expected generic constraint to be pruned, got %q", sections[1].Content)
	}
	if !strings.Contains(sections[2].Content, "scanner") {
		t.Fatalf("expected implementation section to include module summary, got %q", sections[2].Content)
	}
}

func TestRenderPromptWrapsSectionsWithStableBoilerplate(t *testing.T) {
	text := renderPrompt([]spec.PromptSection{
		{Title: "Goal", Content: "- Build a small CLI."},
		{Title: "Acceptance bar", Content: "- Emit a prompt.txt file."},
	})

	if !strings.HasPrefix(text, "Build this project from scratch.") {
		t.Fatalf("prompt is missing opening brief: %q", text)
	}
	if !strings.Contains(text, "Goal\n- Build a small CLI.") {
		t.Fatalf("prompt is missing goal section: %q", text)
	}
	if !strings.Contains(text, "Acceptance bar\n- Emit a prompt.txt file.") {
		t.Fatalf("prompt is missing acceptance section: %q", text)
	}
	if !strings.HasSuffix(text, "choose the narrowest sensible default instead of inventing extra product scope.") {
		t.Fatalf("prompt is missing pruning reminder: %q", text)
	}
}

func TestFilterPromptSectionsDeduplicatesConstraintAndShapeLines(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		Goal: "Build a deterministic CLI.",
		CriticalConstraints: []spec.Constraint{
			{Text: "Do not emit tool-specific command wrappers"},
			{Text: "Do not emit tool-specific command wrappers."},
		},
		ImplementationShape: spec.ImplementationShape{
			CoreModules: []spec.ModuleSpec{
				{Name: "scanner", Responsibility: "Collect repository facts"},
				{Name: "scanner", Responsibility: "Collect repository facts."},
			},
		},
	}

	sections := filterPromptSections(projectSpec)
	if len(sections) < 3 {
		t.Fatalf("expected goal, constraint, and implementation sections, got %#v", sections)
	}

	if got := strings.Count(sections[1].Content, "Do not emit tool-specific command wrappers."); got != 1 {
		t.Fatalf("expected deduplicated constraint line, got %q", sections[1].Content)
	}
	if got := strings.Count(sections[2].Content, "scanner: Collect repository facts."); got != 1 {
		t.Fatalf("expected deduplicated implementation line, got %q", sections[2].Content)
	}
}
