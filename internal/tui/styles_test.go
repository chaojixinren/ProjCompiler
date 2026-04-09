package tui

import (
	"strings"
	"testing"

	"projcompiler/internal/spec"
)

func TestAppBannerSwitchesBetweenCompactArtAndPlainTitle(t *testing.T) {
	wide := appBanner("ProjCompiler", 120)
	narrow := appBanner("ProjCompiler", 10)

	if !strings.Contains(wide, "██████╗") {
		t.Fatalf("expected wide banner art, got %q", wide)
	}
	if strings.Contains(narrow, "██████╗") {
		t.Fatalf("expected narrow banner to fall back to plain title, got %q", narrow)
	}
	if !strings.Contains(narrow, "ProjCompiler") {
		t.Fatalf("expected plain title in narrow banner, got %q", narrow)
	}
}

func TestStageAndStateBadgesIncludeStableLabels(t *testing.T) {
	if got := stagePill("Scan", true, false); !strings.Contains(got, "Scan") {
		t.Fatalf("expected active stage pill label, got %q", got)
	}
	if got := stagePill("Spec", false, true); !strings.Contains(got, "Spec") {
		t.Fatalf("expected completed stage pill label, got %q", got)
	}
	if got := stateBadge(StatePathInput); !strings.Contains(got, "READY") {
		t.Fatalf("expected READY badge, got %q", got)
	}
	if got := stateBadge(StateDone); !strings.Contains(got, "DONE") {
		t.Fatalf("expected DONE badge, got %q", got)
	}
}

func TestSignalBadgeIncludesSignalText(t *testing.T) {
	if got := signalBadge(spec.SignalObserved); !strings.Contains(got, "observed") {
		t.Fatalf("expected observed badge, got %q", got)
	}
	if got := signalBadge(spec.SignalMissing); !strings.Contains(got, "missing") {
		t.Fatalf("expected missing badge, got %q", got)
	}
}

func TestTextHelpersHandleEmptyAndTrimmedValues(t *testing.T) {
	if got := bulletLines(nil); !strings.Contains(got, "None") {
		t.Fatalf("expected None fallback, got %q", got)
	}
	if got := bulletLines([]string{"one", "two"}); !strings.Contains(got, "• one") || !strings.Contains(got, "• two") {
		t.Fatalf("expected bullet content, got %q", got)
	}
	if got := joinKeyHints(" enter ", "", "q"); !strings.Contains(got, " enter ") && !strings.Contains(got, "enter") {
		t.Fatalf("expected trimmed enter hint, got %q", got)
	}
	if got := metricLine("State", "Ready"); !strings.Contains(got, "State:") || !strings.Contains(got, "Ready") {
		t.Fatalf("expected metric line content, got %q", got)
	}
}

func TestTitledPanelIncludesTitleAndBody(t *testing.T) {
	rendered := titledPanel("Summary", "Body text", styles.panel)
	if !strings.Contains(rendered, "Summary") {
		t.Fatalf("expected panel title, got %q", rendered)
	}
	if !strings.Contains(rendered, "Body text") {
		t.Fatalf("expected panel body, got %q", rendered)
	}
}
