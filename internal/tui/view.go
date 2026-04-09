package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"projcompiler/internal/spec"
)

func (m Model) View() string {
	content := []string{
		m.renderHeader(),
		m.renderBody(),
		m.renderFooter(),
	}
	if logs := m.renderLogs(); logs != "" {
		content = append(content, logs)
	}
	return styles.page.Render(strings.Join(compact(content), "\n\n"))
}

func (m Model) renderHeader() string {
	subtitle := "Local repository analysis -> project specification -> final build prompt"
	header := lipgloss.JoinVertical(
		lipgloss.Center,
		m.renderBanner(),
		styles.subtitle.Render(subtitle),
		m.renderStageBar(),
	)

	meta := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stateBadge(m.state),
		"  ",
		styles.muted.Render("State"),
		" ",
		styles.label.Render(m.state.Title()),
	)

	return m.renderPanelWithStyle(
		lipgloss.JoinVertical(lipgloss.Center, header, meta),
		styles.headerBox,
		m.singlePanelWidth(),
	)
}

func (m Model) renderBanner() string {
	// Large ASCII art needs both width and height headroom; otherwise it pushes
	// the first rows out of view on smaller terminals.
	if m.height > 0 && m.height < 36 {
		return styles.banner.Render(m.config.Title)
	}
	return appBanner(m.config.Title, m.layoutWidth())
}

func (m Model) renderStageBar() string {
	stages := []struct {
		state State
		label string
	}{
		{StatePathInput, "1 Input"},
		{StateScanning, "2 Scan"},
		{StateSpecSummary, "3 Spec"},
		{StatePromptPreview, "4 Prompt"},
		{StateDone, "5 Done"},
	}

	current := m.stageIndex()
	parts := make([]string, 0, len(stages)*2)
	for index, stage := range stages {
		parts = append(parts, stagePill(stage.label, index == current, index < current))
		if index < len(stages)-1 {
			parts = append(parts, styles.divider.Render("->"))
		}
	}
	return styles.stageBar.Render(strings.Join(parts, " "))
}

func (m Model) renderBody() string {
	switch m.state {
	case StatePathInput:
		return m.renderPathInput()
	case StateScanning:
		return m.renderScanning()
	case StateSpecSummary:
		return m.renderSpecSummary()
	case StatePromptPreview:
		return m.renderPromptPreview()
	case StateDone:
		return m.renderDone()
	default:
		return titledPanel("Unknown State", styles.text.Render("The TUI entered an unknown state."), styles.panelDanger)
	}
}

func (m Model) renderFooter() string {
	parts := make([]string, 0, 2)
	if m.lastError != "" {
		parts = append(parts, titledPanel("Error", styles.text.Render(m.lastError), styles.panelDanger))
	}

	statusLines := []string{}
	if m.bundle.Ready() {
		statusLines = append(statusLines, metricLine("Prompt", fmt.Sprintf("%d chars", len(m.bundle.PromptText))))
	}
	if m.lastExportedPath != "" {
		statusLines = append(statusLines, metricLine("Last export", m.lastExportedPath))
	}

	helpWidth := max(24, m.singlePanelWidth()-styles.helpBar.GetHorizontalBorderSize())
	parts = append(parts, styles.helpBar.Width(helpWidth).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderKeyHints(),
			styles.muted.Render(strings.Join(statusLines, "\n")),
		),
	))

	return strings.Join(parts, "\n\n")
}

func (m Model) renderLogs() string {
	if len(m.logs) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.logs))
	for _, line := range m.logs {
		lines = append(lines, styles.logLine.Render(line))
	}
	return m.renderPanel("Recent Activity", strings.Join(lines, "\n"), styles.logPanel, m.singlePanelWidth())
}

func (m Model) renderPathInput() string {
	intro := []string{
		styles.text.Render("Enter a local repository path. The CLI will scan the repo, extract a compact project spec, and compile a production-oriented prompt."),
		"",
		metricLine("Example", "/path/to/project"),
		metricLine("Default export", valueOrFallback(m.outputPath, "prompt.txt")),
	}

	inputBody := []string{
		styles.muted.Render("Repository path"),
		styles.codeBox.Width(m.embeddedBoxWidth(styles.panel, styles.codeBox, m.panelWidth())).Render("> " + valueOrFallback(m.pathInput, "")),
		styles.muted.Render("The path is submitted exactly as typed after trimming leading and trailing whitespace."),
	}

	left := m.renderPanel("What This Run Does", strings.Join(intro, "\n"), styles.panelEmphasis, m.panelWidth())
	right := m.renderPanel("Input", strings.Join(inputBody, "\n\n"), styles.panel, m.panelWidth())
	return m.renderColumns(left, right)
}

func (m Model) renderScanning() string {
	status := "Queued"
	if m.loading {
		status = "Running repository scan and spec extraction"
	}
	phase := "Repository scan -> document probe -> spec build"
	if len(m.projectSpec.Goal) > 0 {
		phase = "Finalizing project specification"
	}

	leftBody := strings.Join([]string{
		metricLine("Project path", valueOrFallback(strings.TrimSpace(m.pathInput), "not set")),
		metricLine("Status", status),
		metricLine("Current phase", phase),
		metricLine("Cancellation", "Press Esc to abandon this run and return to input"),
	}, "\n")

	rightBody := bulletLines([]string{
		"Scan results are assembled from deterministic local reads.",
		"Prompt compilation starts only after the project spec is ready.",
		"No code is generated during this stage.",
	})

	left := m.renderPanel("Pipeline Status", leftBody, styles.panelEmphasis, m.panelWidth())
	right := m.renderPanel("Notes", rightBody, styles.panel, m.panelWidth())
	return m.renderColumns(left, right)
}

func (m Model) renderSpecSummary() string {
	topLeft := m.renderPanel("Goal", styles.text.Render(valueOrFallback(m.projectSpec.Goal, "Not available")), styles.panelEmphasis, m.panelWidth())
	topRight := m.renderPanel("Tech Profile", formatTechProfile(m.projectSpec.TechProfile), styles.panel, m.panelWidth())

	modules := m.renderPanel("Core Modules", formatModules(m.projectSpec.ImplementationShape.CoreModules), styles.panel, m.panelWidth())
	flows := m.renderPanel("Key Flows", formatFlows(m.projectSpec.ImplementationShape.KeyFlows), styles.panel, m.panelWidth())
	constraints := m.renderPanel("Critical Constraints", formatConstraints(m.projectSpec.CriticalConstraints), styles.panelEmphasis, m.panelWidth())
	acceptance := m.renderPanel("Acceptance Checks", formatAcceptanceChecks(m.projectSpec.AcceptanceChecks), styles.panel, m.panelWidth())

	panels := []string{
		m.renderColumns(topLeft, topRight),
		m.renderColumns(modules, flows),
		m.renderColumns(constraints, acceptance),
	}

	if stats := m.renderFactStats(); stats != "" {
		panels = append(panels, m.renderPanel("Scan Evidence", stats, styles.panel, m.singlePanelWidth()))
	}
	if questions := m.renderOpenQuestions(); questions != "" {
		panels = append(panels, m.renderPanel("Open Questions", questions, styles.panelDanger, m.singlePanelWidth()))
	}

	return strings.Join(panels, "\n\n")
}

func (m Model) renderPromptPreview() string {
	if m.compilingPrompt {
		body := strings.Join([]string{
			metricLine("Status", "Compiling the final prompt"),
			metricLine("Goal", valueOrFallback(m.projectSpec.Goal, "Not available")),
			styles.muted.Render("The prompt will appear here as soon as compilation finishes."),
		}, "\n")
		return m.renderPanel("Prompt Compiler", body, styles.panelEmphasis, m.singlePanelWidth())
	}

	if !m.bundle.Ready() {
		lines := []string{styles.text.Render("Prompt output is not ready yet.")}
		if len(m.bundle.BlockingQuestions) > 0 {
			lines = append(lines, "", styles.panelTitle.Render("Blocking Questions"))
			for _, question := range m.bundle.BlockingQuestions {
				lines = append(lines, "• "+question.Question)
			}
		}
		return m.renderPanel("Prompt Status", strings.Join(lines, "\n"), styles.panelDanger, m.singlePanelWidth())
	}

	lines := strings.Split(m.bundle.PromptText, "\n")
	start := clamp(m.promptScroll, 0, max(0, len(lines)-1))
	end := min(len(lines), start+m.promptPageSize())
	visible := make([]string, 0, max(0, end-start))
	for idx := start; idx < end; idx++ {
		line := lines[idx]
		if line == "" {
			line = " "
		}
		visible = append(visible, fmt.Sprintf("%s %s",
			styles.codeLineNumber.Render(fmt.Sprintf("%4d │", lineNumber(idx))),
			line,
		))
	}

	meta := []string{
		metricLine("Output language", valueOrFallback(m.bundle.OutputLanguage, "English")),
		metricLine("Export path", valueOrFallback(m.outputPath, "prompt.txt")),
		metricLine("Visible lines", fmt.Sprintf("%d-%d of %d", lineNumber(start), lineNumber(end-1), len(lines))),
		metricLine("Scroll", fmt.Sprintf("%d / %d", m.promptScroll, m.maxPromptScroll())),
	}
	if len(m.bundle.Sections) > 0 {
		sectionNames := make([]string, 0, len(m.bundle.Sections))
		for _, section := range m.bundle.Sections {
			if strings.TrimSpace(section.Title) != "" {
				sectionNames = append(sectionNames, section.Title)
			}
		}
		if len(sectionNames) > 0 {
			meta = append(meta, metricLine("Sections", strings.Join(sectionNames, ", ")))
		}
	}

	preview := m.renderPanel("Prompt Preview", styles.codeBox.Width(m.embeddedBoxWidth(styles.panelEmphasis, styles.codeBox, m.panelWidth())).Render(strings.Join(visible, "\n")), styles.panelEmphasis, m.panelWidth())
	summary := m.renderPanel("Preview Metadata", strings.Join(meta, "\n"), styles.panel, m.panelWidth())
	return m.renderColumns(summary, preview)
}

func (m Model) renderDone() string {
	details := []string{
		metricLine("Output path", valueOrFallback(m.lastExportedPath, m.outputPath)),
		metricLine("Prompt length", fmt.Sprintf("%d characters", len(m.bundle.PromptText))),
		metricLine("Next step", "Use the exported prompt as the build brief for your code generation workflow"),
	}
	return m.renderPanel("Export Complete", strings.Join(details, "\n"), styles.panelSuccess, m.singlePanelWidth())
}

func (m Model) renderKeyHints() string {
	switch m.state {
	case StatePathInput:
		return strings.Join([]string{
			joinKeyHints("Enter"),
			styles.muted.Render("start scan"),
			joinKeyHints("q"),
			styles.muted.Render("quit"),
		}, "  ")
	case StateScanning:
		return strings.Join([]string{
			joinKeyHints("Esc"),
			styles.muted.Render("cancel run"),
			joinKeyHints("Ctrl+C"),
			styles.muted.Render("quit"),
		}, "  ")
	case StateSpecSummary:
		return strings.Join([]string{
			joinKeyHints("Enter", "g"),
			styles.muted.Render("compile prompt"),
			joinKeyHints("Esc"),
			styles.muted.Render("restart"),
		}, "  ")
	case StatePromptPreview:
		return strings.Join([]string{
			joinKeyHints("j", "k", "PgUp", "PgDn"),
			styles.muted.Render("scroll"),
			joinKeyHints("e"),
			styles.muted.Render("export"),
			joinKeyHints("Esc"),
			styles.muted.Render("back"),
		}, "  ")
	case StateDone:
		return strings.Join([]string{
			joinKeyHints("r"),
			styles.muted.Render("new run"),
			joinKeyHints("Enter", "q"),
			styles.muted.Render("quit"),
		}, "  ")
	default:
		return ""
	}
}

func (m Model) renderColumns(left, right string) string {
	if m.layoutWidth() >= 100 {
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	return lipgloss.JoinVertical(lipgloss.Left, left, "", right)
}

func (m Model) renderFactStats() string {
	if m.facts.RootPath == "" {
		return ""
	}
	observed, inferred, missing := summarizeSignals(m.facts.Signals)
	lines := []string{
		metricLine("Root path", m.facts.RootPath),
		metricLine("Documents", fmt.Sprintf("%d", len(m.facts.Docs))),
		metricLine("Entry points", fmt.Sprintf("%d", len(m.facts.EntryPoints))),
		metricLine("Snippets", fmt.Sprintf("%d", len(m.facts.Snippets))),
		metricLine("Constraints", fmt.Sprintf("%d", len(m.facts.Constraints))),
		"",
		strings.Join([]string{
			signalBadge(spec.SignalObserved), styles.muted.Render(fmt.Sprintf("%d", observed)),
			"  ",
			signalBadge(spec.SignalInferred), styles.muted.Render(fmt.Sprintf("%d", inferred)),
			"  ",
			signalBadge(spec.SignalMissing), styles.muted.Render(fmt.Sprintf("%d", missing)),
		}, ""),
	}
	if m.facts.ScanMeta.Truncated {
		lines = append(lines, "", styles.statusWarning.Render("Scan output was truncated to stay within the current budget."))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderOpenQuestions() string {
	if len(m.projectSpec.OpenQuestions) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.projectSpec.OpenQuestions))
	for _, question := range m.projectSpec.OpenQuestions {
		prefix := "•"
		if question.Blocking {
			prefix = "• [blocking]"
		}
		lines = append(lines, fmt.Sprintf("%s %s", prefix, question.Question))
	}
	return strings.Join(lines, "\n")
}

func formatTechProfile(profile spec.TechProfile) string {
	lines := []string{
		metricLine("Platform", valueOrFallback(profile.Platform, "Unknown")),
		metricLine("Languages", joinOrFallback(profile.Languages)),
		metricLine("Frameworks", joinOrFallback(profile.Frameworks)),
		metricLine("Build system", valueOrFallback(profile.BuildSystem, "Unknown")),
		metricLine("Package manager", valueOrFallback(profile.PackageManager, "Unknown")),
		metricLine("Integrations", joinOrFallback(profile.ExternalIntegrations)),
	}
	return strings.Join(lines, "\n")
}

func formatModules(modules []spec.ModuleSpec) string {
	if len(modules) == 0 {
		return styles.muted.Render("None")
	}
	lines := make([]string, 0, len(modules))
	for _, module := range modules {
		lines = append(lines, fmt.Sprintf("• %s\n  %s", valueOrFallback(module.Name, "Unnamed module"), valueOrFallback(module.Responsibility, "No summary")))
	}
	return strings.Join(lines, "\n")
}

func formatFlows(flows []spec.KeyFlow) string {
	if len(flows) == 0 {
		return styles.muted.Render("None")
	}
	lines := make([]string, 0, len(flows))
	for _, flow := range flows {
		lines = append(lines, fmt.Sprintf("• %s\n  %s", valueOrFallback(flow.Name, "Unnamed flow"), valueOrFallback(flow.Summary, "No summary")))
	}
	return strings.Join(lines, "\n")
}

func formatConstraints(constraints []spec.Constraint) string {
	if len(constraints) == 0 {
		return styles.muted.Render("None")
	}
	lines := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		lines = append(lines, "• "+valueOrFallback(constraint.Text, "Unnamed constraint"))
	}
	return strings.Join(lines, "\n")
}

func formatAcceptanceChecks(checks []spec.AcceptanceCheck) string {
	if len(checks) == 0 {
		return styles.muted.Render("None")
	}
	lines := make([]string, 0, len(checks))
	for _, check := range checks {
		lines = append(lines, "• "+valueOrFallback(check.Description, "Unnamed acceptance check"))
	}
	return strings.Join(lines, "\n")
}

func valueOrFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func joinOrFallback(values []string) string {
	if len(values) == 0 {
		return "None"
	}
	return strings.Join(values, ", ")
}

func (m Model) promptPageSize() int {
	if m.height >= 24 {
		return m.height - 18
	}
	return 14
}

func (m Model) maxPromptScroll() int {
	if m.bundle.PromptText == "" {
		return 0
	}
	lines := strings.Split(m.bundle.PromptText, "\n")
	return max(0, len(lines)-m.promptPageSize())
}

func (m Model) layoutWidth() int {
	if m.width <= 0 {
		return 88
	}
	safe := m.width - styles.page.GetHorizontalFrameSize() - 4
	return max(60, safe)
}

func (m Model) panelWidth() int {
	total := m.layoutWidth()
	if total >= 100 {
		return total / 2
	}
	return m.singlePanelWidth()
}

func (m Model) singlePanelWidth() int {
	return max(48, m.layoutWidth())
}

func (m Model) stageIndex() int {
	switch m.state {
	case StatePathInput:
		return 0
	case StateScanning:
		return 1
	case StateSpecSummary:
		return 2
	case StatePromptPreview:
		return 3
	case StateDone:
		return 4
	default:
		return 0
	}
}

func lineNumber(index int) int {
	if index < 0 {
		return 0
	}
	return index + 1
}

func summarizeSignals(signals []spec.FactSignal) (observed, inferred, missing int) {
	for _, signal := range signals {
		switch signal.Kind {
		case spec.SignalObserved:
			observed++
		case spec.SignalInferred:
			inferred++
		case spec.SignalMissing:
			missing++
		}
	}
	return observed, inferred, missing
}

func compact(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func (m Model) renderPanel(title, body string, variant lipgloss.Style, width int) string {
	bodyWidth := m.panelBodyWidth(variant, width)
	wrappedBody := lipgloss.NewStyle().Width(bodyWidth).Render(body)
	return m.renderPanelWithStyle(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.panelTitle.Render(title),
			wrappedBody,
		),
		variant,
		width,
	)
}

func (m Model) renderPanelWithStyle(body string, variant lipgloss.Style, totalWidth int) string {
	boxWidth := max(20, totalWidth-variant.GetHorizontalBorderSize()-variant.GetHorizontalMargins())
	return variant.Width(boxWidth).Render(body)
}

func (m Model) panelBodyWidth(variant lipgloss.Style, totalWidth int) int {
	return max(20, totalWidth-variant.GetHorizontalBorderSize()-variant.GetHorizontalPadding())
}

func (m Model) embeddedBoxWidth(container, child lipgloss.Style, containerTotalWidth int) int {
	return max(12, m.panelBodyWidth(container, containerTotalWidth)-child.GetHorizontalBorderSize())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, lower, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
