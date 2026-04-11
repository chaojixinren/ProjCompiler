package prompt

import (
	"slices"
	"strconv"
	"strings"

	"projcompiler/internal/spec"
	"projcompiler/internal/util"
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
	sections := make([]spec.PromptSection, 0, 8)

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

	if flows := buildUnderstandingFlowsSection(projectSpec.Understanding); flows != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Key execution flows",
			Content: flows,
		})
	}

	if symbols := buildUnderstandingSymbolsSection(projectSpec.Understanding); symbols != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Core symbols",
			Content: symbols,
		})
	}

	if relations := buildUnderstandingRelationsSection(projectSpec.Understanding); relations != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Call graph",
			Content: relations,
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
		if !isPromptWorthySignal(constraint.Signal) {
			continue
		}
		lines = append(lines, bullet(text))
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func buildImplementationSection(projectSpec spec.ProjectSpec) string {
	lines := make([]string, 0, len(projectSpec.ImplementationShape.CoreModules)+len(projectSpec.ImplementationShape.KeyFlows))

	for _, module := range projectSpec.ImplementationShape.CoreModules {
		if !isPromptWorthySignal(module.Signal) {
			continue
		}
		name := strings.TrimSpace(module.Name)
		responsibility := normalizeSentence(module.Responsibility)
		if name == "" || responsibility == "" {
			continue
		}
		lines = append(lines, bullet(name+": "+responsibility))
	}

	for _, flow := range projectSpec.ImplementationShape.KeyFlows {
		if !isPromptWorthySignal(flow.Signal) {
			continue
		}
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
		if !isPromptWorthySignal(check.Signal) {
			continue
		}
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

func isPromptWorthySignal(signal spec.FactSignal) bool {
	switch signal.Kind {
	case spec.SignalMissing:
		return false
	case spec.SignalInferred:
		return signal.Confidence >= 0.45
	default:
		return true
	}
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

func buildUnderstandingFlowsSection(understanding spec.UnderstandingSpec) string {
	if len(understanding.Flows) == 0 {
		return ""
	}

	lines := make([]string, 0, util.Min(5, len(understanding.Flows)))
	for _, flow := range understanding.Flows {
		name := strings.TrimSpace(flow.Name)
		if name == "" {
			continue
		}
		trigger := strings.TrimSpace(flow.Trigger)
		entry := strings.TrimSpace(flow.EntrySymbolID)
		stepCount := len(flow.StepSymbolIDs)

		var desc string
		if trigger != "" && entry != "" {
			desc = name + " triggered by " + trigger + ", starting at " + entry
		} else if entry != "" {
			desc = name + " starts at " + entry
		} else {
			desc = name
		}

		if stepCount > 0 {
			desc += " (" + pluralize(stepCount, "step") + ")"
		}

		lines = append(lines, bullet(desc))
		if len(lines) >= 5 {
			break
		}
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func buildUnderstandingSymbolsSection(understanding spec.UnderstandingSpec) string {
	entrypoints := make([]string, 0, 4)
	packages := make([]string, 0, 4)

	for _, symbol := range understanding.Symbols {
		name := strings.TrimSpace(symbol.Name)
		if name == "" {
			continue
		}

		switch symbol.Kind {
		case spec.UnderstandingSymbolKindFunction, spec.UnderstandingSymbolKindMethod:
			if symbol.IsEntrypoint {
				qualified := strings.TrimSpace(symbol.QualifiedName)
				if qualified != "" {
					entrypoints = append(entrypoints, bullet("Entrypoint: "+qualified))
				} else {
					entrypoints = append(entrypoints, bullet("Entrypoint: "+name))
				}
			}
		case spec.UnderstandingSymbolKindPackage:
			packages = append(packages, bullet("Package: "+name))
		}

		if len(entrypoints) >= 4 && len(packages) >= 4 {
			break
		}
	}

	lines := append(entrypoints, packages...)
	return strings.Join(uniqueLines(lines), "\n")
}

func buildUnderstandingRelationsSection(understanding spec.UnderstandingSpec) string {
	callLines := make([]string, 0, util.Min(6, len(understanding.Relations)))
	importLines := make([]string, 0, util.Min(3, len(understanding.Relations)))

	for _, relation := range understanding.Relations {
		from := strings.TrimSpace(relation.FromID)
		to := util.FirstNonEmpty(strings.TrimSpace(relation.ToID), strings.TrimSpace(relation.ToRef))
		if from == "" || to == "" {
			continue
		}

		switch relation.Type {
		case spec.UnderstandingRelationCalls, spec.UnderstandingRelationEntryFlow:
			callLines = append(callLines, bullet(from+" calls "+to))
		case spec.UnderstandingRelationImports:
			importLines = append(importLines, bullet(from+" imports "+to))
		}

		if len(callLines) >= 6 && len(importLines) >= 3 {
			break
		}
	}

	lines := append(callLines, importLines...)
	return strings.Join(uniqueLines(lines), "\n")
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strings.Join([]string{strconv.Itoa(count), singular + "s"}, " ")
}
