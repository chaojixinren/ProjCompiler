package prompt

import (
	"strings"
	"testing"

	"projcompiler/internal/spec"
)

func TestFilterPromptSectionsProducesOrderedPromptBrief(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		Goal: "Recreate a local CLI that scans a repository and emits an implementation prompt",
		TechProfile: spec.TechProfile{
			Platform:       "local terminal",
			Languages:      []string{"Go"},
			Frameworks:     []string{"Bubble Tea", "Google ADK"},
			BuildSystem:    "Go toolchain",
			PackageManager: "Go Modules",
		},
		CriticalConstraints: []spec.Constraint{
			{Text: "Only emit an English prompt and do not wrap it in tool-specific CLI commands"},
			{Text: "Write unit tests and keep the code maintainable"},
		},
		ImplementationShape: spec.ImplementationShape{
			CoreModules: []spec.ModuleSpec{
				{Name: "scanner", Responsibility: "Collect repository facts without calling the model"},
			},
			KeyFlows: []spec.KeyFlow{
				{Name: "prompt compilation", Summary: "Turn the distilled spec into a single English build brief"},
			},
		},
		AcceptanceChecks: []spec.AcceptanceCheck{
			{Description: "The output should be a pure English prompt"},
		},
	}

	sections := filterPromptSections(projectSpec)
	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(sections))
	}

	titles := []string{sections[0].Title, sections[1].Title, sections[2].Title, sections[3].Title}
	wantTitles := []string{"Goal", "Must-get-right constraints", "Non-obvious implementation shape", "Acceptance bar"}
	for i, want := range wantTitles {
		if titles[i] != want {
			t.Fatalf("section %d title = %q, want %q", i, titles[i], want)
		}
	}

	if strings.Contains(strings.ToLower(sections[1].Content), "maintainable") {
		t.Fatalf("expected generic constraint to be pruned, got %q", sections[1].Content)
	}
	if !strings.Contains(sections[2].Content, "scanner") {
		t.Fatalf("expected implementation section to include module summary, got %q", sections[2].Content)
	}
}

func TestRenderPromptWrapsSectionsWithStableBoilerplate(t *testing.T) {
	text := renderPrompt([]spec.PromptSection{
		{Title: "Goal", Content: "- Build a small CLI."},
		{Title: "Acceptance bar", Content: "- Emit a prompt.txt file."},
	})

	if !strings.HasPrefix(text, "Build this project from scratch.") {
		t.Fatalf("prompt is missing opening brief: %q", text)
	}
	if !strings.Contains(text, "Goal\n- Build a small CLI.") {
		t.Fatalf("prompt is missing goal section: %q", text)
	}
	if !strings.Contains(text, "Acceptance bar\n- Emit a prompt.txt file.") {
		t.Fatalf("prompt is missing acceptance section: %q", text)
	}
	if !strings.HasSuffix(text, "choose the narrowest sensible default instead of inventing extra product scope.") {
		t.Fatalf("prompt is missing pruning reminder: %q", text)
	}
}

func TestFilterPromptSectionsDeduplicatesConstraintAndShapeLines(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		Goal: "Build a deterministic CLI.",
		CriticalConstraints: []spec.Constraint{
			{Text: "Do not emit tool-specific command wrappers"},
			{Text: "Do not emit tool-specific command wrappers."},
		},
		ImplementationShape: spec.ImplementationShape{
			CoreModules: []spec.ModuleSpec{
				{Name: "scanner", Responsibility: "Collect repository facts"},
				{Name: "scanner", Responsibility: "Collect repository facts."},
			},
		},
	}

	sections := filterPromptSections(projectSpec)
	if len(sections) < 3 {
		t.Fatalf("expected goal, constraint, and implementation sections, got %#v", sections)
	}

	if got := strings.Count(sections[1].Content, "Do not emit tool-specific command wrappers."); got != 1 {
		t.Fatalf("expected deduplicated constraint line, got %q", sections[1].Content)
	}
	if got := strings.Count(sections[2].Content, "scanner: Collect repository facts."); got != 1 {
		t.Fatalf("expected deduplicated implementation line, got %q", sections[2].Content)
	}
}

func TestBuildProjectStructureSectionGroupsByDirectory(t *testing.T) {
	understanding := spec.UnderstandingSpec{
		Files: []spec.UnderstandingFile{
			{Path: "internal/config/config.go", Language: "go", Role: spec.UnderstandingFileRoleSource},
			{Path: "internal/config/config_test.go", Language: "go", Role: spec.UnderstandingFileRoleTest},
			{Path: "internal/gateway/handler.go", Language: "go", Role: spec.UnderstandingFileRoleSource},
			{Path: "internal/gateway/router.go", Language: "go", Role: spec.UnderstandingFileRoleSource},
			{Path: "frontend/src/App.vue", Language: "vue", Role: spec.UnderstandingFileRoleSource},
			{Path: "generated/proto.go", Language: "go", Role: spec.UnderstandingFileRoleGenerated},
		},
	}

	result := buildProjectStructureSection(understanding)
	if result == "" {
		t.Fatal("expected non-empty project structure section")
	}
	if !strings.Contains(result, "internal/config/") {
		t.Fatalf("expected config directory in output, got %q", result)
	}
	if !strings.Contains(result, "internal/gateway/") {
		t.Fatalf("expected gateway directory in output, got %q", result)
	}
	if !strings.Contains(result, "frontend/src/") {
		t.Fatalf("expected frontend/src directory in output, got %q", result)
	}
	if strings.Contains(result, "generated") {
		t.Fatalf("expected generated files to be excluded, got %q", result)
	}
}

func TestBuildModuleDependenciesSectionBuildsPackageTopology(t *testing.T) {
	understanding := spec.UnderstandingSpec{
		Symbols: []spec.UnderstandingSymbol{
			{ID: "pkg-gateway", Kind: spec.UnderstandingSymbolKindPackage, Name: "gateway", PackageName: "gateway"},
			{ID: "pkg-config", Kind: spec.UnderstandingSymbolKindPackage, Name: "config", PackageName: "config"},
		},
		Relations: []spec.UnderstandingRelation{
			{Type: spec.UnderstandingRelationImports, FromID: "pkg-gateway", ToRef: "internal/config"},
			{Type: spec.UnderstandingRelationImports, FromID: "pkg-gateway", ToRef: "net/http"},
		},
	}

	result := buildModuleDependenciesSection(understanding)
	if result == "" {
		t.Fatal("expected non-empty module dependencies section")
	}
	if !strings.Contains(result, "gateway imports internal/config") {
		t.Fatalf("expected gateway->config import, got %q", result)
	}
}

func TestBuildDataStructuresSectionRendersStructs(t *testing.T) {
	understanding := spec.UnderstandingSpec{
		Symbols: []spec.UnderstandingSymbol{
			{ID: "sym-config", Kind: spec.UnderstandingSymbolKindStruct, Name: "Config", QualifiedName: "config.Config"},
		},
		Relations: []spec.UnderstandingRelation{
			{Type: spec.UnderstandingRelationDefinesField, FromID: "sym-config", ToRef: "AccessKey string"},
			{Type: spec.UnderstandingRelationDefinesField, FromID: "sym-config", ToRef: "Strategy string"},
		},
	}

	result := buildDataStructuresSection(understanding)
	if result == "" {
		t.Fatal("expected non-empty data structures section")
	}
	if !strings.Contains(result, "config.Config (struct)") {
		t.Fatalf("expected Config struct in output, got %q", result)
	}
	if !strings.Contains(result, "AccessKey string") {
		t.Fatalf("expected AccessKey field in output, got %q", result)
	}
}

func TestBuildAPIRoutesSectionRendersRoutes(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		ImplementationShape: spec.ImplementationShape{
			APIRoutes: []spec.APIRoute{
				{Method: "POST", Path: "/v1/chat/completions", Handler: "gateway.HandleChat", Signal: spec.InferredSignal(0.7, "test")},
				{Method: "GET", Path: "/health", Handler: "gateway.HandleHealth", Middleware: []string{"auth"}, Signal: spec.InferredSignal(0.7, "test")},
			},
		},
	}

	result := buildAPIRoutesSection(projectSpec)
	if result == "" {
		t.Fatal("expected non-empty API routes section")
	}
	if !strings.Contains(result, "/v1/chat/completions") {
		t.Fatalf("expected chat completions route, got %q", result)
	}
	if !strings.Contains(result, "[auth]") {
		t.Fatalf("expected middleware in output, got %q", result)
	}
}

func TestBuildDataFlowSectionRendersSteps(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		ImplementationShape: spec.ImplementationShape{
			DataFlows: []spec.DataFlow{
				{Name: "Request proxy", Steps: []string{"parse request", "transform to Anthropic", "forward upstream", "transform response"}, Signal: spec.InferredSignal(0.7, "test")},
			},
		},
	}

	result := buildDataFlowSection(projectSpec)
	if result == "" {
		t.Fatal("expected non-empty data flow section")
	}
	if !strings.Contains(result, "parse request -> transform to Anthropic") {
		t.Fatalf("expected step chain in output, got %q", result)
	}
}

func TestBuildUnderstandingFlowsSectionExpandsSteps(t *testing.T) {
	understanding := spec.UnderstandingSpec{
		Symbols: []spec.UnderstandingSymbol{
			{ID: "s1", Name: "main", QualifiedName: "main.main"},
			{ID: "s2", Name: "Load", QualifiedName: "config.Load"},
			{ID: "s3", Name: "NewServer", QualifiedName: "gateway.NewServer"},
		},
		Flows: []spec.UnderstandingFlow{
			{ID: "f1", Name: "main", Trigger: "entrypoint", EntrySymbolID: "s1", StepSymbolIDs: []string{"s1", "s2", "s3"}},
		},
	}

	result := buildUnderstandingFlowsSection(understanding)
	if result == "" {
		t.Fatal("expected non-empty flows section")
	}
	if !strings.Contains(result, "main.main -> config.Load -> gateway.NewServer") {
		t.Fatalf("expected expanded step chain, got %q", result)
	}
}

func TestFilterPromptSectionsSkipsLowConfidenceInferredItems(t *testing.T) {
	projectSpec := spec.ProjectSpec{
		Goal: "Build a deterministic CLI.",
		CriticalConstraints: []spec.Constraint{
			{Text: "Do not emit tool-specific command wrappers", Signal: spec.InferredSignal(0.9, "high confidence")},
			{Text: "Expose a web dashboard", Signal: spec.InferredSignal(0.2, "weak guess")},
		},
		ImplementationShape: spec.ImplementationShape{
			CoreModules: []spec.ModuleSpec{
				{Name: "scanner", Responsibility: "Collect repository facts", Signal: spec.ObservedSignal("direct code evidence")},
				{Name: "dashboard", Responsibility: "Serve browser UI", Signal: spec.InferredSignal(0.2, "weak guess")},
			},
		},
		AcceptanceChecks: []spec.AcceptanceCheck{
			{Description: "Emit prompt.txt", Signal: spec.ObservedSignal("deterministic requirement")},
			{Description: "Render dashboard", Signal: spec.MissingSignal("no supporting evidence")},
		},
	}

	sections := filterPromptSections(projectSpec)
	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(sections))
	}
	if strings.Contains(sections[1].Content, "Expose a web dashboard.") {
		t.Fatalf("expected low-confidence inferred constraint to be pruned, got %q", sections[1].Content)
	}
	if strings.Contains(sections[2].Content, "dashboard: Serve browser UI.") {
		t.Fatalf("expected low-confidence inferred module to be pruned, got %q", sections[2].Content)
	}
	if strings.Contains(sections[3].Content, "Render dashboard.") {
		t.Fatalf("expected missing-signal acceptance check to be pruned, got %q", sections[3].Content)
	}
}
