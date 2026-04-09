package prompt

import (
	"slices"
	"strings"

	"projcompiler/internal/spec"
)

var genericConstraintFragments = []string{
	"best practice",
	"clean code",
	"maintainable",
	"scalable",
	"proper error handling",
	"handle errors gracefully",
	"user-friendly",
	"add logging",
	"good documentation",
	"write unit tests",
	"high quality",
}

func filterPromptSections(projectSpec spec.ProjectSpec) []spec.PromptSection {
	sections := make([]spec.PromptSection, 0, 4)

	if goal := buildGoalSection(projectSpec); goal != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Goal",
			Content: goal,
		})
	}

	if constraints := buildConstraintSection(projectSpec); constraints != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Must-get-right constraints",
			Content: constraints,
		})
	}

	if shape := buildImplementationSection(projectSpec); shape != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Non-obvious implementation shape",
			Content: shape,
		})
	}

	if acceptance := buildAcceptanceSection(projectSpec); acceptance != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Acceptance bar",
			Content: acceptance,
		})
	}

	return sections
}

func buildGoalSection(projectSpec spec.ProjectSpec) string {
	lines := make([]string, 0, 2)
	if goal := strings.TrimSpace(projectSpec.Goal); goal != "" {
		lines = append(lines, bullet(goal))
	}

	if techSummary := summarizeTechProfile(projectSpec.TechProfile); techSummary != "" {
		lines = append(lines, bullet(techSummary))
	}

	return strings.Join(lines, "\n")
}

func buildConstraintSection(projectSpec spec.ProjectSpec) string {
	lines := make([]string, 0, len(projectSpec.CriticalConstraints)+6)
	lines = append(lines, techConstraintLines(projectSpec.TechProfile)...)

	for _, constraint := range projectSpec.CriticalConstraints {
		text := normalizeSentence(constraint.Text)
		if text == "" || !isPromptWorthyConstraint(text) {
			continue
		}
		lines = append(lines, bullet(text))
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func buildImplementationSection(projectSpec spec.ProjectSpec) string {
	lines := make([]string, 0, len(projectSpec.ImplementationShape.CoreModules)+len(projectSpec.ImplementationShape.KeyFlows))

	for _, module := range projectSpec.ImplementationShape.CoreModules {
		name := strings.TrimSpace(module.Name)
		responsibility := normalizeSentence(module.Responsibility)
		if name == "" || responsibility == "" {
			continue
		}
		lines = append(lines, bullet(name+": "+responsibility))
	}

	for _, flow := range projectSpec.ImplementationShape.KeyFlows {
		name := strings.TrimSpace(flow.Name)
		summary := normalizeSentence(flow.Summary)
		if name == "" || summary == "" {
			continue
		}
		lines = append(lines, bullet(name+": "+summary))
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func buildAcceptanceSection(projectSpec spec.ProjectSpec) string {
	lines := make([]string, 0, len(projectSpec.AcceptanceChecks))
	for _, check := range projectSpec.AcceptanceChecks {
		text := normalizeSentence(check.Description)
		if text == "" {
			continue
		}
		lines = append(lines, bullet(text))
	}
	return strings.Join(uniqueLines(lines), "\n")
}

func summarizeTechProfile(profile spec.TechProfile) string {
	parts := make([]string, 0, 5)
	if len(profile.Languages) > 0 {
		parts = append(parts, "Use "+joinNatural(profile.Languages)+" as the primary implementation language")
	}
	if len(profile.Frameworks) > 0 {
		parts = append(parts, "lean on "+joinNatural(profile.Frameworks)+" where it materially shapes the project")
	}
	if profile.Platform != "" {
		parts = append(parts, "target "+strings.TrimSpace(profile.Platform))
	}
	if len(parts) == 0 {
		return ""
	}
	return normalizeSentence(strings.Join(parts, "; "))
}

func techConstraintLines(profile spec.TechProfile) []string {
	lines := make([]string, 0, 6)
	if profile.Platform != "" {
		lines = append(lines, bullet("Target platform: "+strings.TrimSpace(profile.Platform)))
	}
	if len(profile.Languages) > 0 {
		lines = append(lines, bullet("Primary languages: "+joinNatural(profile.Languages)))
	}
	if len(profile.Frameworks) > 0 {
		lines = append(lines, bullet("Required frameworks/libraries: "+joinNatural(profile.Frameworks)))
	}
	if profile.BuildSystem != "" {
		lines = append(lines, bullet("Build system: "+strings.TrimSpace(profile.BuildSystem)))
	}
	if profile.PackageManager != "" && !strings.EqualFold(profile.PackageManager, profile.BuildSystem) {
		lines = append(lines, bullet("Package/dependency manager: "+strings.TrimSpace(profile.PackageManager)))
	}
	if len(profile.ExternalIntegrations) > 0 {
		lines = append(lines, bullet("External integrations that matter to the design: "+joinNatural(profile.ExternalIntegrations)))
	}
	return lines
}

func isPromptWorthyConstraint(text string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(text))
	if trimmed == "" {
		return false
	}
	for _, fragment := range genericConstraintFragments {
		if strings.Contains(trimmed, fragment) {
			return false
		}
	}
	return true
}

func uniqueLines(lines []string) []string {
	seen := make(map[string]struct{}, len(lines))
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func bullet(text string) string {
	return "- " + normalizeSentence(text)
}

func joinNatural(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	slices.Sort(parts)
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}

func normalizeSentence(text string) string {
	trimmed := strings.TrimSpace(text)
	trimmed = strings.TrimSuffix(trimmed, ".")
	if trimmed == "" {
		return ""
	}
	return trimmed + "."
}
