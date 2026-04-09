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
