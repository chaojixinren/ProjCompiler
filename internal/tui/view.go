package tui

import (
	"fmt"
	"strings"

	"projcompiler/internal/spec"
)

func (m Model) View() string {
	var sections []string
	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderBody())
	if footer := m.renderFooter(); footer != "" {
		sections = append(sections, footer)
	}
	if logs := m.renderLogs(); logs != "" {
		sections = append(sections, logs)
	}
	return strings.Join(sections, "\n\n")
}

func (m Model) renderHeader() string {
	return fmt.Sprintf("%s\n%s", m.config.Title, strings.Repeat("=", max(14, len(m.config.Title))))
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
		return "Unknown state."
	}
}

func (m Model) renderFooter() string {
	var lines []string
	if m.lastError != "" {
		lines = append(lines, "Error: "+m.lastError)
	}
	switch m.state {
	case StatePathInput:
		lines = append(lines, "Keys: type a local path, Enter to start, q to quit")
	case StateScanning:
		lines = append(lines, "Keys: esc to cancel, ctrl+c to quit")
	case StateSpecSummary:
		lines = append(lines, "Keys: Enter or g to generate prompt, esc to restart")
	case StatePromptPreview:
		lines = append(lines, "Keys: j/k or up/down to scroll, e to export, esc to return")
	case StateDone:
		lines = append(lines, "Keys: r to start again, Enter or q to quit")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderLogs() string {
	if len(m.logs) == 0 {
		return ""
	}
	return "Recent Log\n" + strings.Join(m.logs, "\n")
}

func (m Model) renderPathInput() string {
	var builder strings.Builder
	builder.WriteString("State: Path Input\n")
	builder.WriteString("Enter a local repository path to scan and turn into a build prompt.\n\n")
	builder.WriteString("Project Path\n")
	builder.WriteString("> ")
	builder.WriteString(m.pathInput)
	builder.WriteString("\n")
	if m.outputPath != "" {
		builder.WriteString(fmt.Sprintf("Default export path: %s\n", m.outputPath))
	}
	return builder.String()
}

func (m Model) renderScanning() string {
	var builder strings.Builder
	builder.WriteString("State: Scanning\n")
	builder.WriteString("The UI is waiting on background commands from the scan/spec pipeline.\n\n")
	builder.WriteString(fmt.Sprintf("Project path: %s\n", strings.TrimSpace(m.pathInput)))
	switch {
	case m.loading:
		builder.WriteString("Status: working\n")
	default:
		builder.WriteString("Status: queued\n")
	}
	builder.WriteString("Current phase: ")
	if len(m.projectSpec.Goal) == 0 {
		builder.WriteString("repo scan / spec extraction")
	} else {
		builder.WriteString("finishing")
	}
	builder.WriteString("\n")
	return builder.String()
}

func (m Model) renderSpecSummary() string {
	var builder strings.Builder
	builder.WriteString("State: Spec Summary\n")
	builder.WriteString("Review the extracted project shape before prompt compilation.\n\n")
	builder.WriteString("Goal\n")
	builder.WriteString(valueOrFallback(m.projectSpec.Goal, "Not available"))
	builder.WriteString("\n\n")
	builder.WriteString("Tech Profile\n")
	builder.WriteString(formatTechProfile(m.projectSpec.TechProfile))
	builder.WriteString("\n\n")
	builder.WriteString("Core Modules\n")
	builder.WriteString(formatModules(m.projectSpec.ImplementationShape.CoreModules))
	builder.WriteString("\n\n")
	builder.WriteString("Key Flows\n")
	builder.WriteString(formatFlows(m.projectSpec.ImplementationShape.KeyFlows))
	builder.WriteString("\n\n")
	builder.WriteString("Critical Constraints\n")
	builder.WriteString(formatConstraints(m.projectSpec.CriticalConstraints))
	builder.WriteString("\n\n")
	builder.WriteString("Acceptance Checks\n")
	builder.WriteString(formatAcceptanceChecks(m.projectSpec.AcceptanceChecks))
	return builder.String()
}

func (m Model) renderPromptPreview() string {
	var builder strings.Builder
	builder.WriteString("State: Prompt Preview\n")
	if m.compilingPrompt {
		builder.WriteString("Generating the final prompt in the background.\n")
		return builder.String()
	}
	if !m.bundle.Ready() {
		builder.WriteString("Prompt is not ready yet.\n")
		if len(m.bundle.BlockingQuestions) > 0 {
			builder.WriteString("\nBlocking Questions\n")
			for _, question := range m.bundle.BlockingQuestions {
				builder.WriteString("- ")
				builder.WriteString(question.Question)
				builder.WriteString("\n")
			}
		}
		return builder.String()
	}

	lines := strings.Split(m.bundle.PromptText, "\n")
	start := clamp(m.promptScroll, 0, max(0, len(lines)-1))
	end := min(len(lines), start+m.promptPageSize())

	builder.WriteString(fmt.Sprintf("Output language: %s\n", valueOrFallback(m.bundle.OutputLanguage, "English")))
	builder.WriteString(fmt.Sprintf("Export path: %s\n", valueOrFallback(m.outputPath, "prompt.txt")))
	builder.WriteString(fmt.Sprintf("Showing lines %d-%d of %d\n\n", lineNumber(start), lineNumber(end-1), len(lines)))
	builder.WriteString(strings.Join(lines[start:end], "\n"))
	return builder.String()
}

func (m Model) renderDone() string {
	var builder strings.Builder
	builder.WriteString("State: Done\n")
	builder.WriteString("The prompt has been exported successfully.\n\n")
	builder.WriteString(fmt.Sprintf("Output path: %s\n", valueOrFallback(m.lastExportedPath, m.outputPath)))
	builder.WriteString(fmt.Sprintf("Prompt length: %d characters\n", len(m.bundle.PromptText)))
	return builder.String()
}

func formatTechProfile(profile spec.TechProfile) string {
	lines := []string{
		fmt.Sprintf("- Platform: %s", valueOrFallback(profile.Platform, "Unknown")),
		fmt.Sprintf("- Languages: %s", joinOrFallback(profile.Languages)),
		fmt.Sprintf("- Frameworks: %s", joinOrFallback(profile.Frameworks)),
		fmt.Sprintf("- Build system: %s", valueOrFallback(profile.BuildSystem, "Unknown")),
		fmt.Sprintf("- Package manager: %s", valueOrFallback(profile.PackageManager, "Unknown")),
		fmt.Sprintf("- External integrations: %s", joinOrFallback(profile.ExternalIntegrations)),
	}
	return strings.Join(lines, "\n")
}

func formatModules(modules []spec.ModuleSpec) string {
	if len(modules) == 0 {
		return "- None"
	}
	lines := make([]string, 0, len(modules))
	for _, module := range modules {
		lines = append(lines, fmt.Sprintf("- %s: %s", valueOrFallback(module.Name, "Unnamed module"), valueOrFallback(module.Responsibility, "No summary")))
	}
	return strings.Join(lines, "\n")
}

func formatFlows(flows []spec.KeyFlow) string {
	if len(flows) == 0 {
		return "- None"
	}
	lines := make([]string, 0, len(flows))
	for _, flow := range flows {
		lines = append(lines, fmt.Sprintf("- %s: %s", valueOrFallback(flow.Name, "Unnamed flow"), valueOrFallback(flow.Summary, "No summary")))
	}
	return strings.Join(lines, "\n")
}

func formatConstraints(constraints []spec.Constraint) string {
	if len(constraints) == 0 {
		return "- None"
	}
	lines := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		lines = append(lines, "- "+valueOrFallback(constraint.Text, "Unnamed constraint"))
	}
	return strings.Join(lines, "\n")
}

func formatAcceptanceChecks(checks []spec.AcceptanceCheck) string {
	if len(checks) == 0 {
		return "- None"
	}
	lines := make([]string, 0, len(checks))
	for _, check := range checks {
		lines = append(lines, "- "+valueOrFallback(check.Description, "Unnamed acceptance check"))
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
	if m.height >= 20 {
		return m.height - 12
	}
	return 16
}

func (m Model) maxPromptScroll() int {
	if m.bundle.PromptText == "" {
		return 0
	}
	lines := strings.Split(m.bundle.PromptText, "\n")
	return max(0, len(lines)-m.promptPageSize())
}

func lineNumber(index int) int {
	if index < 0 {
		return 0
	}
	return index + 1
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
