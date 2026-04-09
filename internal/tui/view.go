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
	s := m.t()
	header := lipgloss.JoinVertical(
		lipgloss.Center,
		m.renderBanner(),
		styles.subtitle.Render(s.Subtitle),
		m.renderStageBar(),
	)

	meta := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stateBadge(m.state),
		"  ",
		styles.muted.Render(s.StateMeta),
		" ",
		styles.label.Render(s.StateTitle(m.state)),
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
	s := m.t()
	stages := []struct {
		state State
		label string
	}{
		{StatePathInput, s.StageInput},
		{StateScanning, s.StageScan},
		{StateSpecSummary, s.StageSpec},
		{StatePromptPreview, s.StagePrompt},
		{StateDone, s.StageDone},
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
	case StateConfigEdit:
		return m.renderConfigEdit()
	default:
		s := m.t()
		return titledPanel(s.UnknownPanelTitle, styles.text.Render(s.UnknownPanelBody), styles.panelDanger)
	}
}

func (m Model) renderFooter() string {
	s := m.t()
	parts := make([]string, 0, 2)
	if m.lastError != "" {
		parts = append(parts, titledPanel(s.ErrorPanelTitle, styles.text.Render(m.lastError), styles.panelDanger))
	}

	statusLines := []string{}
	if m.bundle.Ready() {
		statusLines = append(statusLines, metricLine(s.StatPrompt, fmt.Sprintf(s.FmtPromptChars, len(m.bundle.PromptText))))
	}
	if m.lastExportedPath != "" {
		statusLines = append(statusLines, metricLine(s.StatLastExport, m.lastExportedPath))
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
	s := m.t()
	lines := make([]string, 0, len(m.logs))
	for _, line := range m.logs {
		lines = append(lines, styles.logLine.Render(line))
	}
	return m.renderPanel(s.LogPanelTitle, strings.Join(lines, "\n"), styles.logPanel, m.singlePanelWidth())
}

func (m Model) renderPathInput() string {
	s := m.t()
	lw, rw := m.splitWidths()

	intro := []string{
		styles.text.Render(s.PathIntro),
		"",
		metricLine(s.PathExample, "/path/to/project"),
		metricLine(s.PathDefaultExp, valueOrFallback(m.outputPath, "prompt.txt")),
	}

	inputBody := []string{
		styles.muted.Render(s.PathFieldLabel),
		styles.codeBox.Width(m.embeddedBoxWidth(styles.panel, styles.codeBox, rw)).Render("> " + valueOrFallback(m.pathInput, "")),
	}

	left := m.renderPanel(s.PathPanelInfo, strings.Join(intro, "\n"), styles.panelEmphasis, lw)
	right := m.renderPanel(s.PathPanelInput, strings.Join(inputBody, "\n\n"), styles.panel, rw)
	return m.renderColumns(left, right)
}

func (m Model) renderScanning() string {
	s := m.t()
	status := s.ScanQueued
	if m.loading {
		status = s.ScanRunning
	}
	phase := s.ScanPhaseMain
	if len(m.projectSpec.Goal) > 0 {
		phase = s.ScanPhaseFinal
	}

	leftBody := strings.Join([]string{
		metricLine(s.ScanLabelPath, valueOrFallback(strings.TrimSpace(m.pathInput), s.NotSet)),
		metricLine(s.ScanLabelStatus, status),
		metricLine(s.ScanLabelPhase, phase),
		metricLine(s.ScanLabelCancel, s.ScanCancelHint),
	}, "\n")

	rightBody := bulletLines([]string{
		s.ScanNote1,
		s.ScanNote2,
		s.ScanNote3,
	})

	lw, rw := m.splitWidths()
	left := m.renderPanel(s.ScanPanelStatus, leftBody, styles.panelEmphasis, lw)
	right := m.renderPanel(s.ScanPanelNotes, rightBody, styles.panel, rw)
	return m.renderColumns(left, right)
}

func (m Model) renderSpecSummary() string {
	s := m.t()
	lw, rw := m.splitWidths()
	topLeft := m.renderPanel(s.SpecPanelGoal, styles.text.Render(valueOrFallback(m.projectSpec.Goal, s.SpecGoalFallback)), styles.panelEmphasis, lw)
	topRight := m.renderPanel(s.SpecPanelTech, formatTechProfile(s, m.projectSpec.TechProfile), styles.panel, rw)

	modules := m.renderPanel(s.SpecPanelModules, formatModules(s, m.projectSpec.ImplementationShape.CoreModules), styles.panel, lw)
	flows := m.renderPanel(s.SpecPanelFlows, formatFlows(s, m.projectSpec.ImplementationShape.KeyFlows), styles.panel, rw)
	constraints := m.renderPanel(s.SpecPanelConstraints, formatConstraints(s, m.projectSpec.CriticalConstraints), styles.panelEmphasis, lw)
	acceptance := m.renderPanel(s.SpecPanelChecks, formatAcceptanceChecks(s, m.projectSpec.AcceptanceChecks), styles.panel, rw)

	panels := []string{
		m.renderColumns(topLeft, topRight),
		m.renderColumns(modules, flows),
		m.renderColumns(constraints, acceptance),
	}

	if stats := m.renderFactStats(); stats != "" {
		panels = append(panels, m.renderPanel(s.SpecPanelEvidence, stats, styles.panel, m.singlePanelWidth()))
	}
	if questions := m.renderOpenQuestions(); questions != "" {
		panels = append(panels, m.renderPanel(s.SpecPanelQuestions, questions, styles.panelDanger, m.singlePanelWidth()))
	}

	return strings.Join(panels, "\n\n")
}

func (m Model) renderPromptPreview() string {
	s := m.t()
	if m.compilingPrompt {
		body := strings.Join([]string{
			metricLine(s.PromptStatusLabel, s.PromptCompiling),
			metricLine(s.SpecPanelGoal, valueOrFallback(m.projectSpec.Goal, s.SpecGoalFallback)),
			styles.muted.Render(s.PromptWaiting),
		}, "\n")
		return m.renderPanel(s.PromptPanelCompiler, body, styles.panelEmphasis, m.singlePanelWidth())
	}

	if !m.bundle.Ready() {
		lines := []string{styles.text.Render(s.PromptNotReady)}
		if len(m.bundle.BlockingQuestions) > 0 {
			lines = append(lines, "", styles.panelTitle.Render(s.PromptBlockingQTitle))
			for _, question := range m.bundle.BlockingQuestions {
				lines = append(lines, "• "+question.Question)
			}
		}
		return m.renderPanel(s.PromptPanelStatus, strings.Join(lines, "\n"), styles.panelDanger, m.singlePanelWidth())
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
		metricLine(s.PromptLangLabel, valueOrFallback(m.bundle.OutputLanguage, "English")),
		metricLine(s.PromptExportLabel, valueOrFallback(m.outputPath, "prompt.txt")),
		metricLine(s.PromptVisibleLabel, fmt.Sprintf(s.FmtVisibleLines, lineNumber(start), lineNumber(end-1), len(lines))),
		metricLine(s.PromptScrollLabel, fmt.Sprintf(s.FmtScrollPos, m.promptScroll, m.maxPromptScroll())),
	}
	if len(m.bundle.Sections) > 0 {
		sectionNames := make([]string, 0, len(m.bundle.Sections))
		for _, section := range m.bundle.Sections {
			if strings.TrimSpace(section.Title) != "" {
				sectionNames = append(sectionNames, section.Title)
			}
		}
		if len(sectionNames) > 0 {
			meta = append(meta, metricLine(s.PromptSectionsLabel, strings.Join(sectionNames, ", ")))
		}
	}

	lw, rw := m.splitWidths()
	preview := m.renderPanel(s.PromptPanelPreview, styles.codeBox.Width(m.embeddedBoxWidth(styles.panelEmphasis, styles.codeBox, rw)).Render(strings.Join(visible, "\n")), styles.panelEmphasis, rw)
	summary := m.renderPanel(s.PromptPanelMeta, strings.Join(meta, "\n"), styles.panel, lw)
	return m.renderColumns(summary, preview)
}

func (m Model) renderDone() string {
	s := m.t()
	details := []string{
		metricLine(s.DoneOutputPath, valueOrFallback(m.lastExportedPath, m.outputPath)),
		metricLine(s.DonePromptLen, fmt.Sprintf(s.FmtPromptLen, len(m.bundle.PromptText))),
		metricLine(s.DoneNextStep, s.DoneNextStepVal),
	}
	return m.renderPanel(s.DonePanelTitle, strings.Join(details, "\n"), styles.panelSuccess, m.singlePanelWidth())
}

func (m Model) renderConfigEdit() string {
	s := m.t()
	panelWidth := m.singlePanelWidth()
	boxWidth := m.embeddedBoxWidth(styles.panel, styles.codeBox, panelWidth)

	inputs := make([]string, 3)
	labels := []string{s.ConfigBaseURLLabel, s.ConfigAPIKeyLabel, s.ConfigModelLabel}
	for i, input := range m.configInputs {
		style := styles.codeBox
		if i == m.configFocusIndex {
			style = styles.codeBox.Copy().BorderForeground(lipgloss.Color("62"))
		}
		inputs[i] = fmt.Sprintf("%s\n%s",
			styles.muted.Render(labels[i]),
			style.Width(boxWidth).Render(input.View()),
		)
	}

	body := strings.Join([]string{
		styles.text.Render(s.ConfigPanelHint),
		"",
		strings.Join(inputs, "\n\n"),
	}, "\n")

	if m.configSaving {
		body = strings.Join([]string{
			body,
			"",
			styles.muted.Render("Saving..."),
		}, "\n")
	}

	return m.renderPanel(s.ConfigPanelTitle, body, styles.panelEmphasis, panelWidth)
}

func (m Model) renderKeyHints() string {
	s := m.t()
	langToggle := []string{
		joinKeyHints("ctrl+l"),
		styles.muted.Render(s.HintToggleLang),
	}
	switch m.state {
	case StatePathInput:
		parts := []string{
			joinKeyHints("Enter"),
			styles.muted.Render(s.HintStartScan),
			joinKeyHints("c"),
			styles.muted.Render(s.HintConfigEdit),
			joinKeyHints("q"),
			styles.muted.Render(s.HintQuit),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	case StateScanning:
		parts := []string{
			joinKeyHints("Esc"),
			styles.muted.Render(s.HintCancelRun),
			joinKeyHints("Ctrl+C"),
			styles.muted.Render(s.HintQuit),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	case StateSpecSummary:
		parts := []string{
			joinKeyHints("Enter", "g"),
			styles.muted.Render(s.HintCompilePrompt),
			joinKeyHints("Esc"),
			styles.muted.Render(s.HintRestart),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	case StatePromptPreview:
		parts := []string{
			joinKeyHints("j", "k", "PgUp", "PgDn"),
			styles.muted.Render(s.HintScroll),
			joinKeyHints("e"),
			styles.muted.Render(s.HintExport),
			joinKeyHints("Esc"),
			styles.muted.Render(s.HintBack),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	case StateDone:
		parts := []string{
			joinKeyHints("r"),
			styles.muted.Render(s.HintNewRun),
			joinKeyHints("Enter", "q"),
			styles.muted.Render(s.HintQuit),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	case StateConfigEdit:
		parts := []string{
			joinKeyHints("Tab", "Shift+Tab"),
			styles.muted.Render(s.HintConfigSwitch),
			joinKeyHints("Enter"),
			styles.muted.Render(s.HintConfigSave),
			joinKeyHints("Esc"),
			styles.muted.Render(s.HintConfigCancel),
		}
		return strings.Join(append(parts, langToggle...), "  ")
	default:
		return strings.Join(langToggle, "  ")
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
	s := m.t()
	observed, inferred, missing := summarizeSignals(m.facts.Signals)
	lines := []string{
		metricLine(s.StatRootPath, m.facts.RootPath),
		metricLine(s.StatDocuments, fmt.Sprintf("%d", len(m.facts.Docs))),
		metricLine(s.StatEntryPoints, fmt.Sprintf("%d", len(m.facts.EntryPoints))),
		metricLine(s.StatSnippets, fmt.Sprintf("%d", len(m.facts.Snippets))),
		metricLine(s.StatConstraints, fmt.Sprintf("%d", len(m.facts.Constraints))),
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
		lines = append(lines, "", styles.statusWarning.Render(s.StatTruncated))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderOpenQuestions() string {
	if len(m.projectSpec.OpenQuestions) == 0 {
		return ""
	}
	s := m.t()
	lines := make([]string, 0, len(m.projectSpec.OpenQuestions))
	for _, question := range m.projectSpec.OpenQuestions {
		prefix := "•"
		if question.Blocking {
			prefix = "• " + s.SpecBlockingPrefix
		}
		lines = append(lines, fmt.Sprintf("%s %s", prefix, question.Question))
	}
	return strings.Join(lines, "\n")
}

func formatTechProfile(s Strings, profile spec.TechProfile) string {
	lines := []string{
		metricLine(s.TechPlatform, valueOrFallback(profile.Platform, s.SpecUnknown)),
		metricLine(s.TechLanguages, joinOrFallback(profile.Languages, s.SpecNone)),
		metricLine(s.TechFrameworks, joinOrFallback(profile.Frameworks, s.SpecNone)),
		metricLine(s.TechBuildSystem, valueOrFallback(profile.BuildSystem, s.SpecUnknown)),
		metricLine(s.TechPackageManager, valueOrFallback(profile.PackageManager, s.SpecUnknown)),
		metricLine(s.TechIntegrations, joinOrFallback(profile.ExternalIntegrations, s.SpecNone)),
	}
	return strings.Join(lines, "\n")
}

func formatModules(s Strings, modules []spec.ModuleSpec) string {
	if len(modules) == 0 {
		return styles.muted.Render(s.SpecNone)
	}
	lines := make([]string, 0, len(modules))
	for _, module := range modules {
		lines = append(lines, fmt.Sprintf("• %s\n  %s",
			valueOrFallback(module.Name, s.SpecUnnamedModule),
			valueOrFallback(module.Responsibility, s.SpecNoSummary),
		))
	}
	return strings.Join(lines, "\n")
}

func formatFlows(s Strings, flows []spec.KeyFlow) string {
	if len(flows) == 0 {
		return styles.muted.Render(s.SpecNone)
	}
	lines := make([]string, 0, len(flows))
	for _, flow := range flows {
		lines = append(lines, fmt.Sprintf("• %s\n  %s",
			valueOrFallback(flow.Name, s.SpecUnnamedFlow),
			valueOrFallback(flow.Summary, s.SpecNoSummary),
		))
	}
	return strings.Join(lines, "\n")
}

func formatConstraints(s Strings, constraints []spec.Constraint) string {
	if len(constraints) == 0 {
		return styles.muted.Render(s.SpecNone)
	}
	lines := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		lines = append(lines, "• "+valueOrFallback(constraint.Text, s.SpecUnnamedConstraint))
	}
	return strings.Join(lines, "\n")
}

func formatAcceptanceChecks(s Strings, checks []spec.AcceptanceCheck) string {
	if len(checks) == 0 {
		return styles.muted.Render(s.SpecNone)
	}
	lines := make([]string, 0, len(checks))
	for _, check := range checks {
		lines = append(lines, "• "+valueOrFallback(check.Description, s.SpecUnnamedCheck))
	}
	return strings.Join(lines, "\n")
}

func valueOrFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func joinOrFallback(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
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

// splitWidths returns the left and right panel widths for a two-column layout.
// In wide mode the two values sum exactly to layoutWidth(); in narrow mode both
// equal singlePanelWidth() so callers can stack them without knowing the mode.
func (m Model) splitWidths() (left, right int) {
	total := m.layoutWidth()
	if total >= 100 {
		left = total / 2
		right = total - left // handles odd totals; left+right == total
		return
	}
	w := m.singlePanelWidth()
	return w, w
}

func (m Model) panelWidth() int {
	left, _ := m.splitWidths()
	return left
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
	case StateConfigEdit:
		return 0 // Show as if in path input stage
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
