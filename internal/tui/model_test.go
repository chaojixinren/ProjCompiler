package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"projcompiler/internal/spec"
)

func TestNewModelStartsInPathInputState(t *testing.T) {
	model := NewModel(Services{}, WithTitle("Test UI"), WithDefaultPromptPath("out.txt"), WithMaxLogLines(2))

	if model.state != StatePathInput {
		t.Fatalf("state = %q, want %q", model.state, StatePathInput)
	}
	if model.outputPath != "out.txt" {
		t.Fatalf("outputPath = %q, want %q", model.outputPath, "out.txt")
	}
	if len(model.logs) != 1 || !strings.Contains(model.logs[0], "Ready for a local project path.") {
		t.Fatalf("expected initial readiness log, got %#v", model.logs)
	}
}

func TestModelReturnsToInputOnScanFailure(t *testing.T) {
	model := NewModel(Services{})
	model.state = StateScanning

	updated, _ := model.Update(scanFinishedMsg{Err: errors.New("boom")})
	next := updated.(Model)

	if next.state != StatePathInput {
		t.Fatalf("state = %q, want %q", next.state, StatePathInput)
	}
	if !strings.Contains(next.lastError, "Scan failed: boom") {
		t.Fatalf("lastError = %q, want scan failure message", next.lastError)
	}
}

func TestModelPromptPreviewScrollRespectsBounds(t *testing.T) {
	model := NewModel(Services{})
	model.state = StatePromptPreview
	model.height = 16
	model.bundle = spec.PromptBundle{
		PromptText: strings.Repeat("line\n", 30),
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	next := updated.(Model)
	if next.promptScroll == 0 {
		t.Fatalf("expected pgdown to advance scroll")
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	next = updated.(Model)
	if next.promptScroll != 0 {
		t.Fatalf("expected pgup to clamp scroll back to zero, got %d", next.promptScroll)
	}
}

func TestModelIgnoresMessagesFromStaleRun(t *testing.T) {
	model := NewModel(Services{})
	model.activeRunID = 3
	model.state = StateScanning
	model.loading = true

	updated, _ := model.Update(scanFinishedMsg{
		RunID: 2,
		Facts: spec.RepoFacts{
			Docs: []spec.DocumentFact{{Path: "README.md"}},
		},
	})
	next := updated.(Model)

	if next.state != StateScanning {
		t.Fatalf("state = %q, want %q", next.state, StateScanning)
	}
	if !next.loading {
		t.Fatalf("expected stale message not to mutate loading state")
	}
	if len(next.facts.Docs) != 0 {
		t.Fatalf("expected stale message not to replace facts, got %#v", next.facts)
	}
}

func TestModelCancelsScanningBackToInputState(t *testing.T) {
	model := NewModel(Services{})
	model.state = StateScanning
	model.activeRunID = 4

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	next := updated.(Model)

	if next.state != StatePathInput {
		t.Fatalf("state = %q, want %q", next.state, StatePathInput)
	}
	if next.activeRunID != 5 {
		t.Fatalf("activeRunID = %d, want %d", next.activeRunID, 5)
	}
	if len(next.logs) == 0 || !strings.Contains(next.logs[len(next.logs)-1], "Cancelled current run.") {
		t.Fatalf("expected cancel log, got %#v", next.logs)
	}
}

func TestModelSpecSummaryEnterTransitionsToPromptPreviewStart(t *testing.T) {
	model := NewModel(Services{})
	model.state = StateSpecSummary
	model.activeRunID = 2

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(Model)
	if cmd == nil {
		t.Fatalf("expected compile command")
	}
	if next.state != StateSpecSummary {
		t.Fatalf("expected state to remain spec summary until start message is processed, got %q", next.state)
	}
}

func TestModelUnderstandingMessagesTransitionStateAndTriggerSpecBuild(t *testing.T) {
	model := NewModel(Services{})
	model.activeRunID = 7
	model.state = StateScanning
	model.facts = spec.RepoFacts{
		RootPath: "/tmp/repo",
		Docs:     []spec.DocumentFact{{Path: "README.md"}},
	}

	updated, _ := model.Update(understandingStartedMsg{RunID: 7})
	started := updated.(Model)
	if started.state != StateScanning {
		t.Fatalf("state = %q, want %q", started.state, StateScanning)
	}
	if !started.understandingLoading {
		t.Fatalf("expected understandingLoading=true after start")
	}

	updated, cmd := started.Update(understandingFinishedMsg{
		RunID: 7,
		Summary: UnderstandingSummary{
			Status:   "ready",
			Source:   "builder:test",
			Overview: []string{"summary"},
		},
	})
	finished := updated.(Model)
	if cmd == nil {
		t.Fatalf("expected understanding finish to trigger spec build command")
	}
	if finished.state != StateScanning {
		t.Fatalf("state = %q, want %q", finished.state, StateScanning)
	}
	if finished.understandingLoading {
		t.Fatalf("expected understandingLoading=false after finish")
	}
	if got := finished.understanding.Source; got != "builder:test" {
		t.Fatalf("understanding.Source = %q, want %q", got, "builder:test")
	}
}

func TestModelSpecSummaryCanSwitchSpecSections(t *testing.T) {
	model := NewModel(Services{})
	model.state = StateSpecSummary
	model.understanding = UnderstandingSummary{
		Overview:      []string{"overview"},
		OpenQuestions: []string{"q1"},
		Evidence:      []string{"e1"},
		ModuleMap:     []string{"a -> b"},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	next := updated.(Model)
	if next.state != StateSpecSummary {
		t.Fatalf("state = %q, want %q", next.state, StateSpecSummary)
	}
	if next.specSection != SpecSectionModules {
		t.Fatalf("section = %q, want %q", next.specSection, SpecSectionModules)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	next = updated.(Model)
	if next.state != StateSpecSummary {
		t.Fatalf("state = %q, want %q", next.state, StateSpecSummary)
	}
	if next.specSection != SpecSectionOverview {
		t.Fatalf("section = %q, want %q", next.specSection, SpecSectionOverview)
	}
}
