package app

import (
	"projcompiler/internal/config"
)

type App struct {
	config   config.Config
	services Services
}

func New(cfg config.Config, services Services) *App {
	return &App{
		config:   cfg,
		services: withDefaults(cfg, services),
	}
}

func (a *App) Run() error {
	_, err := newProgram(a).Run()
	return err
}

func (a *App) ServiceStatuses() []ServiceStatus {
	return []ServiceStatus{
		{Name: "Scanner", Implemented: !isUnavailableScanner(a.services.Scanner)},
		{Name: "SpecBuilder", Implemented: !isUnavailableSpecBuilder(a.services.SpecBuilder)},
		{Name: "PromptCompiler", Implemented: !isUnavailablePromptCompiler(a.services.PromptCompiler)},
		{Name: "Exporter", Implemented: !isUnavailableExporter(a.services.Exporter)},
		{Name: "AgentAssembler", Implemented: !isUnavailableAgentAssembler(a.services.AgentAssembler)},
	}
}

func isUnavailableScanner(scanner Scanner) bool {
	_, ok := scanner.(unimplementedScanner)
	return ok
}

func isUnavailableSpecBuilder(builder SpecBuilder) bool {
	_, ok := builder.(unimplementedSpecBuilder)
	return ok
}

func isUnavailablePromptCompiler(compiler PromptCompiler) bool {
	_, ok := compiler.(unimplementedPromptCompiler)
	return ok
}

func isUnavailableExporter(exporter Exporter) bool {
	_, ok := exporter.(unimplementedExporter)
	return ok
}

func isUnavailableAgentAssembler(assembler AgentAssembler) bool {
	_, ok := assembler.(unimplementedAgentAssembler)
	return ok
}
