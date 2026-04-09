package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"projcompiler/internal/spec"
)

type uiStyles struct {
	page           lipgloss.Style
	headerBox      lipgloss.Style
	banner         lipgloss.Style
	subtitle       lipgloss.Style
	stageBar       lipgloss.Style
	stageActive    lipgloss.Style
	stageComplete  lipgloss.Style
	stageIdle      lipgloss.Style
	panel          lipgloss.Style
	panelEmphasis  lipgloss.Style
	panelDanger    lipgloss.Style
	panelSuccess   lipgloss.Style
	panelTitle     lipgloss.Style
	label          lipgloss.Style
	muted          lipgloss.Style
	text           lipgloss.Style
	statusInfo     lipgloss.Style
	statusSuccess  lipgloss.Style
	statusWarning  lipgloss.Style
	statusError    lipgloss.Style
	codeBox        lipgloss.Style
	codeLineNumber lipgloss.Style
	helpBar        lipgloss.Style
	logPanel       lipgloss.Style
	logLine        lipgloss.Style
	key            lipgloss.Style
	divider        lipgloss.Style
}

var styles = newUIStyles()

func newUIStyles() uiStyles {
	return uiStyles{
		page: lipgloss.NewStyle().Padding(0, 2),
		headerBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 2),
		banner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Bold(true),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("109")),
		stageBar: lipgloss.NewStyle(),
		stageActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("33")).
			Padding(0, 1),
		stageComplete: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("120")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		stageIdle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),
		panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 2),
		panelEmphasis: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 2),
		panelDanger: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(0, 2),
		panelSuccess: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")).
			Padding(0, 2),
		panelTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
		label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Bold(true),
		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		text: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		statusInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")).
			Padding(0, 1),
		statusSuccess: lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.Color("28")).
			Padding(0, 1),
		statusWarning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("16")).
			Background(lipgloss.Color("220")).
			Padding(0, 1),
		statusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.Color("160")).
			Padding(0, 1),
		codeBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		codeLineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		helpBar: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("238")).
			PaddingTop(0),
		logPanel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),
		logLine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")),
		key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

func appBanner(title string, width int) string {
	logo := []string{
		"██████╗ ██████╗  ██████╗      ██╗ ██████╗ ██████╗ ███╗   ███╗██████╗ ██╗██╗     ███████╗██████╗ ",
		"██╔══██╗██╔══██╗██╔═══██╗     ██║██╔════╝██╔═══██╗████╗ ████║██╔══██╗██║██║     ██╔════╝██╔══██╗",
		"██████╔╝██████╔╝██║   ██║     ██║██║     ██║   ██║██╔████╔██║██████╔╝██║██║     █████╗  ██████╔╝",
		"██╔═══╝ ██╔══██╗██║   ██║██   ██║██║     ██║   ██║██║╚██╔╝██║██╔═══╝ ██║██║     ██╔══╝  ██╔══██╗",
		"██║     ██║  ██║╚██████╔╝╚█████╔╝╚██████╗╚██████╔╝██║ ╚═╝ ██║██║     ██║███████╗███████╗██║  ██║",
		"╚═╝     ╚═╝  ╚═╝ ╚═════╝  ╚════╝  ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝",
	}
	plain := title
	logoText := strings.Join(logo, "\n")
	if width > 0 && width < lipgloss.Width(logo[0])+4 {
		return styles.banner.Render(plain)
	}
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Inherit(styles.banner).
		Render(logoText)
}

func stagePill(label string, active, complete bool) string {
	switch {
	case active:
		return styles.stageActive.Render(label)
	case complete:
		return styles.stageComplete.Render(label)
	default:
		return styles.stageIdle.Render(label)
	}
}

func stateBadge(state State) string {
	switch state {
	case StateScanning:
		return styles.statusInfo.Render("RUNNING")
	case StateSpecSummary:
		return styles.statusWarning.Render("REVIEW")
	case StatePromptPreview:
		return styles.statusInfo.Render("PROMPT")
	case StateDone:
		return styles.statusSuccess.Render("DONE")
	default:
		return styles.statusInfo.Render("READY")
	}
}

func signalBadge(kind spec.SignalKind) string {
	switch kind {
	case spec.SignalObserved:
		return styles.statusSuccess.Render("observed")
	case spec.SignalInferred:
		return styles.statusInfo.Render("inferred")
	case spec.SignalMissing:
		return styles.statusWarning.Render("missing")
	default:
		return styles.statusInfo.Render("unknown")
	}
}

func titledPanel(title, body string, variant lipgloss.Style) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.panelTitle.Render(title),
		body,
	)
	return variant.Render(content)
}

func metricLine(label, value string) string {
	return fmt.Sprintf("%s %s", styles.label.Render(label+":"), styles.text.Render(value))
}

func bulletLines(values []string) string {
	if len(values) == 0 {
		return styles.muted.Render("None")
	}
	lines := make([]string, 0, len(values))
	for _, value := range values {
		lines = append(lines, "• "+value)
	}
	return strings.Join(lines, "\n")
}

func joinKeyHints(values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		parts = append(parts, styles.key.Render(value))
	}
	return strings.Join(parts, " ")
}
