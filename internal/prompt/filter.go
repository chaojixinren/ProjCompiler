package prompt

import (
	"path/filepath"
	"slices"
	"sort"
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
	sections := make([]spec.PromptSection, 0, 12)

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

	if structure := buildProjectStructureSection(projectSpec.Understanding); structure != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Project structure",
			Content: structure,
		})
	}

	if shape := buildImplementationSection(projectSpec); shape != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Non-obvious implementation shape",
			Content: shape,
		})
	}

	if dataStructs := buildDataStructuresSection(projectSpec.Understanding); dataStructs != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Key data structures",
			Content: dataStructs,
		})
	}

	if routes := buildAPIRoutesSection(projectSpec); routes != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "API routes",
			Content: routes,
		})
	}

	if dataFlow := buildDataFlowSection(projectSpec); dataFlow != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Data flow",
			Content: dataFlow,
		})
	}

	if deps := buildModuleDependenciesSection(projectSpec.Understanding); deps != "" {
		sections = append(sections, spec.PromptSection{
			Title:   "Module dependencies",
			Content: deps,
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

	symbolNames := makeSymbolNameLookup(understanding.Symbols)

	lines := make([]string, 0, util.Min(5, len(understanding.Flows)))
	for _, flow := range understanding.Flows {
		if len(lines) >= 5 {
			break
		}
		const maxSteps = 8
		stepNames := make([]string, 0, util.Min(maxSteps, len(flow.StepSymbolIDs)))
		for _, stepID := range flow.StepSymbolIDs {
			name := symbolNames[stepID]
			if name == "" {
				continue
			}
			stepNames = append(stepNames, name)
			if len(stepNames) >= maxSteps {
				break
			}
		}
		if len(stepNames) > 1 {
			lines = append(lines, bullet(strings.Join(stepNames, " -> ")))
		} else if len(stepNames) == 1 {
			trigger := strings.TrimSpace(flow.Trigger)
			if trigger == "" {
				trigger = "entrypoint"
			}
			lines = append(lines, bullet(stepNames[0]+" (triggered by "+trigger+")"))
		}
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func makeSymbolNameLookup(symbols []spec.UnderstandingSymbol) map[string]string {
	lookup := make(map[string]string, len(symbols))
	for _, symbol := range symbols {
		name := strings.TrimSpace(symbol.QualifiedName)
		if name == "" {
			name = strings.TrimSpace(symbol.Name)
		}
		if name != "" {
			lookup[symbol.ID] = name
		}
	}
	return lookup
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

func buildProjectStructureSection(understanding spec.UnderstandingSpec) string {
	if len(understanding.Files) == 0 {
		return ""
	}

	type dirInfo struct {
		files     []string
		languages map[string]int
	}
	dirs := make(map[string]*dirInfo)
	var dirOrder []string

	for _, file := range understanding.Files {
		switch file.Role {
		case spec.UnderstandingFileRoleGenerated, spec.UnderstandingFileRoleOther:
			continue
		}
		dir := filepath.Dir(file.Path)
		if dir == "" || dir == "." {
			dir = "."
		}
		info, ok := dirs[dir]
		if !ok {
			info = &dirInfo{languages: make(map[string]int)}
			dirs[dir] = info
			dirOrder = append(dirOrder, dir)
		}
		base := filepath.Base(file.Path)
		info.files = append(info.files, base)
		lang := file.Language
		if (lang == "" || lang == "unknown") && base != "" {
			lang = langFromExt(filepath.Ext(base))
		}
		if lang != "" && lang != "unknown" {
			info.languages[lang]++
		}
	}

	sort.Strings(dirOrder)

	lines := make([]string, 0, len(dirOrder))
	for _, dir := range dirOrder {
		info := dirs[dir]
		lang := dominantLanguage(info.languages)
		if len(info.files) <= 6 {
			label := dir + "/: " + strings.Join(info.files, ", ")
			lines = append(lines, bullet(label))
		} else {
			label := dir + "/: " + pluralize(len(info.files), lang+" file")
			lines = append(lines, bullet(label))
		}
	}
	return strings.Join(lines, "\n")
}

func langFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".mod", ".sum":
		return "go"
	default:
		return ""
	}
}

func dominantLanguage(languages map[string]int) string {
	if len(languages) == 0 {
		return ""
	}
	best := ""
	bestCount := 0
	for lang, count := range languages {
		if count > bestCount {
			best = lang
			bestCount = count
		}
	}
	return best
}

func buildModuleDependenciesSection(understanding spec.UnderstandingSpec) string {
	if len(understanding.Relations) == 0 {
		return ""
	}

	projectPkgs := make(map[string]bool, len(understanding.Symbols))
	symbolPkg := make(map[string]string, len(understanding.Symbols))
	for _, sym := range understanding.Symbols {
		if sym.PackageName != "" {
			symbolPkg[sym.ID] = sym.PackageName
			projectPkgs[sym.PackageName] = true
		}
	}

	importEdges := make(map[depEdge]bool)
	callEdges := make(map[depEdge]bool)

	for _, rel := range understanding.Relations {
		fromPkg := symbolPkg[rel.FromID]
		if fromPkg == "" {
			continue
		}
		switch rel.Type {
		case spec.UnderstandingRelationImports:
			toRef := strings.TrimSpace(rel.ToRef)
			if toRef == "" {
				continue
			}
			if isStdlibImport(toRef) {
				continue
			}
			importEdges[depEdge{fromPkg, toRef}] = true
		case spec.UnderstandingRelationCalls, spec.UnderstandingRelationEntryFlow:
			toPkg := symbolPkg[rel.ToID]
			if toPkg == "" || toPkg == fromPkg {
				continue
			}
			callEdges[depEdge{fromPkg, toPkg}] = true
		}
	}

	lines := make([]string, 0, len(importEdges)+len(callEdges))

	sortedCalls := sortEdges(callEdges)
	for _, e := range sortedCalls {
		lines = append(lines, bullet(e.from+" -> "+e.to+" (cross-package call)"))
	}
	sortedImports := sortEdges(importEdges)
	for _, e := range sortedImports {
		lines = append(lines, bullet(e.from+" imports "+e.to))
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func isStdlibImport(path string) bool {
	if strings.Contains(path, ".") {
		return false
	}
	parts := strings.SplitN(path, "/", 2)
	first := parts[0]
	switch first {
	case "internal", "cmd":
		return false
	}
	return true
}

type depEdge struct{ from, to string }

func sortEdges(edges map[depEdge]bool) []depEdge {
	out := make([]depEdge, 0, len(edges))
	for e := range edges {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].from != out[j].from {
			return out[i].from < out[j].from
		}
		return out[i].to < out[j].to
	})
	return out
}

func buildDataStructuresSection(understanding spec.UnderstandingSpec) string {
	if len(understanding.Symbols) == 0 {
		return ""
	}

	symbolNames := makeSymbolNameLookup(understanding.Symbols)

	fieldsByOwner := make(map[string][]string)
	for _, rel := range understanding.Relations {
		if rel.Type != spec.UnderstandingRelationDefinesField {
			continue
		}
		ownerName := symbolNames[rel.FromID]
		if ownerName == "" {
			continue
		}
		fieldDesc := strings.TrimSpace(rel.ToRef)
		if fieldDesc == "" || fieldDesc == "unknown" {
			continue
		}
		fieldsByOwner[ownerName] = append(fieldsByOwner[ownerName], fieldDesc)
	}

	refCount := make(map[string]int)
	for _, rel := range understanding.Relations {
		if rel.ToID != "" {
			refCount[rel.ToID]++
		}
	}

	type candidate struct {
		sym   spec.UnderstandingSymbol
		name  string
		score int
	}
	candidates := make([]candidate, 0, 16)
	for _, sym := range understanding.Symbols {
		switch sym.Kind {
		case spec.UnderstandingSymbolKindStruct, spec.UnderstandingSymbolKindInterface:
		default:
			continue
		}
		name := strings.TrimSpace(sym.QualifiedName)
		if name == "" {
			name = strings.TrimSpace(sym.Name)
		}
		if name == "" {
			continue
		}
		shortName := strings.TrimSpace(sym.Name)
		if shortName == "" || !isExported(shortName) {
			continue
		}
		score := len(fieldsByOwner[name]) + refCount[sym.ID]*2
		if sym.Kind == spec.UnderstandingSymbolKindInterface {
			score += 3
		}
		candidates = append(candidates, candidate{sym, name, score})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	const maxTypes = 10
	lines := make([]string, 0, maxTypes)
	for _, c := range candidates {
		if len(lines) >= maxTypes {
			break
		}
		kind := "struct"
		if c.sym.Kind == spec.UnderstandingSymbolKindInterface {
			kind = "interface"
		}
		fields := fieldsByOwner[c.name]
		if len(fields) > 6 {
			fields = fields[:6]
			fields = append(fields, "...")
		}
		if len(fields) > 0 {
			lines = append(lines, bullet(c.name+" ("+kind+"): "+strings.Join(fields, ", ")))
		} else {
			lines = append(lines, bullet(c.name+" ("+kind+")"))
		}
	}

	return strings.Join(uniqueLines(lines), "\n")
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func buildAPIRoutesSection(projectSpec spec.ProjectSpec) string {
	routes := projectSpec.ImplementationShape.APIRoutes
	if len(routes) == 0 {
		return ""
	}
	lines := make([]string, 0, len(routes))
	for _, route := range routes {
		if !isPromptWorthySignal(route.Signal) {
			continue
		}
		method := strings.TrimSpace(route.Method)
		path := strings.TrimSpace(route.Path)
		handler := strings.TrimSpace(route.Handler)
		if path == "" {
			continue
		}
		desc := method + " " + path
		if handler != "" {
			desc += " -> " + handler
		}
		if len(route.Middleware) > 0 {
			desc += " [" + strings.Join(route.Middleware, ", ") + "]"
		}
		lines = append(lines, bullet(strings.TrimSpace(desc)))
	}
	return strings.Join(uniqueLines(lines), "\n")
}

func buildDataFlowSection(projectSpec spec.ProjectSpec) string {
	flows := projectSpec.ImplementationShape.DataFlows
	if len(flows) == 0 {
		return ""
	}
	lines := make([]string, 0, len(flows))
	for _, flow := range flows {
		if !isPromptWorthySignal(flow.Signal) {
			continue
		}
		name := strings.TrimSpace(flow.Name)
		if name == "" || len(flow.Steps) == 0 {
			continue
		}
		lines = append(lines, bullet(name+": "+strings.Join(flow.Steps, " -> ")))
	}
	return strings.Join(uniqueLines(lines), "\n")
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strings.Join([]string{strconv.Itoa(count), singular + "s"}, " ")
}
