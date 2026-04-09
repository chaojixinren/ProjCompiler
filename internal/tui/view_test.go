package tui

import (
	"strings"
	"testing"

	"projcompiler/internal/spec"
)

func TestModelViewIncludesHeaderBodyFooterAndLogs(t *testing.T) {
	model := NewModel(Services{}, WithTitle("ProjCompiler UI"))
	model.pathInput = "/tmp/demo"

	view := model.View()

	if !strings.Contains(view, "ProjCompiler UI") {
		t.Fatalf("expected title in view, got %q", view)
	}
	if !strings.Contains(view, "What This Run Does") {
		t.Fatalf("expected path input body, got %q", view)
	}
	if !strings.Contains(view, "start scan") {
		t.Fatalf("expected footer help, got %q", view)
	}
	if !strings.Contains(view, "Recent Activity") {
		t.Fatalf("expected recent log section, got %q", view)
	}
}

func TestRenderPromptPreviewShowsBlockingQuestions(t *testing.T) {
	model := NewModel(Services{})
	model.state = StatePromptPreview
	model.bundle = spec.PromptBundle{
		BlockingQuestions: []spec.OpenQuestion{
			{Question: "Which subproject should be targeted?"},
		},
	}

	view := model.renderPromptPreview()

	if !strings.Contains(view, "Prompt output is not ready yet.") {
		t.Fatalf("expected not-ready banner, got %q", view)
	}
	if !strings.Contains(view, "Blocking Questions") {
		t.Fatalf("expected blocking questions heading, got %q", view)
	}
	if !strings.Contains(view, "Which subproject should be targeted?") {
		t.Fatalf("expected blocking question content, got %q", view)
	}
}

func TestRenderPromptPreviewShowsPromptWindowAndMetadata(t *testing.T) {
	model := NewModel(Services{}, WithDefaultPromptPath("build/prompt.txt"))
	model.state = StatePromptPreview
	model.height = 20
	model.bundle = spec.PromptBundle{
		PromptText:     "line1\nline2\nline3",
		OutputLanguage: "english",
	}

	view := model.renderPromptPreview()

	if !strings.Contains(view, "Output language:") || !strings.Contains(view, "english") {
		t.Fatalf("expected output language, got %q", view)
	}
	if !strings.Contains(view, "Export path:") || !strings.Contains(view, "build/prompt.txt") {
		t.Fatalf("expected export path, got %q", view)
	}
	if !strings.Contains(view, "Visible lines:") || !strings.Contains(view, "1-3 of 3") {
		t.Fatalf("expected line window summary, got %q", view)
	}
}

func TestRenderDonePrefersExportedPath(t *testing.T) {
	model := NewModel(Services{}, WithDefaultPromptPath("prompt.txt"))
	model.state = StateDone
	model.lastExportedPath = "/tmp/out.txt"
	model.bundle = spec.PromptBundle{PromptText: "hello"}

	view := model.renderDone()

	if !strings.Contains(view, "Output path:") || !strings.Contains(view, "/tmp/out.txt") {
		t.Fatalf("expected exported path, got %q", view)
	}
	if !strings.Contains(view, "Prompt length:") || !strings.Contains(view, "5 characters") {
		t.Fatalf("expected prompt length, got %q", view)
	}
}

func TestFormatHelpersReturnFallbacksForEmptyValues(t *testing.T) {
	s := stringsEN()
	if got := formatModules(s, nil); !strings.Contains(got, "None") {
		t.Fatalf("formatModules(nil) = %q", got)
	}
	if got := formatFlows(s, nil); !strings.Contains(got, "None") {
		t.Fatalf("formatFlows(nil) = %q", got)
	}
	if got := formatConstraints(s, nil); !strings.Contains(got, "None") {
		t.Fatalf("formatConstraints(nil) = %q", got)
	}
	if got := formatAcceptanceChecks(s, nil); !strings.Contains(got, "None") {
		t.Fatalf("formatAcceptanceChecks(nil) = %q", got)
	}
	if got := joinOrFallback(nil, "None"); got != "None" {
		t.Fatalf("joinOrFallback(nil) = %q", got)
	}
	if got := valueOrFallback("   ", "fallback"); got != "fallback" {
		t.Fatalf("valueOrFallback returned %q", got)
	}
}
