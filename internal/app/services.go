package app

import (
	"context"

	adkagent "google.golang.org/adk/agent"

	"projcompiler/internal/config"
	"projcompiler/internal/scan"
	"projcompiler/internal/spec"
)

type Services struct {
	Scanner        Scanner
	SpecBuilder    SpecBuilder
	PromptCompiler PromptCompiler
	Exporter       Exporter
	AgentAssembler AgentAssembler
}

func withDefaults(cfg config.Config, services Services) Services {
	if services.Scanner == nil {
		services.Scanner = scan.NewDefaultOrchestrator()
	}
	if services.SpecBuilder == nil {
		services.SpecBuilder = unimplementedSpecBuilder{}
	}
	if services.PromptCompiler == nil {
		services.PromptCompiler = unimplementedPromptCompiler{}
	}
	if services.Exporter == nil {
		services.Exporter = NewFileExporter(cfg)
	}
	if services.AgentAssembler == nil {
		services.AgentAssembler = unimplementedAgentAssembler{}
	}
	return services
}

type unimplementedScanner struct{}

func (unimplementedScanner) Scan(context.Context, string) (spec.RepoFacts, error) {
	return spec.RepoFacts{}, serviceUnavailable("scanner")
}

type unimplementedSpecBuilder struct{}

func (unimplementedSpecBuilder) BuildProjectSpec(context.Context, spec.RepoFacts) (spec.ProjectSpec, error) {
	return spec.ProjectSpec{}, serviceUnavailable("spec builder")
}

type unimplementedPromptCompiler struct{}

func (unimplementedPromptCompiler) CompilePrompt(context.Context, spec.ProjectSpec) (spec.PromptBundle, error) {
	return spec.PromptBundle{}, serviceUnavailable("prompt compiler")
}

type unimplementedExporter struct{}

func (unimplementedExporter) ExportPrompt(context.Context, spec.PromptBundle, string) (string, error) {
	return "", serviceUnavailable("exporter")
}

func (unimplementedExporter) ExportSpecDebug(context.Context, spec.ProjectSpec, string) (string, error) {
	return "", serviceUnavailable("exporter")
}

type unimplementedAgentAssembler struct{}

func (unimplementedAgentAssembler) Build(context.Context) (adkagent.Agent, error) {
	return nil, serviceUnavailable("agent assembler")
}
