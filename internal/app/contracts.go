package app

import (
	"context"
	"errors"
	"fmt"

	adkagent "google.golang.org/adk/agent"

	"projcompiler/internal/spec"
)

var ErrServiceUnavailable = errors.New("service unavailable")

type Scanner interface {
	Scan(ctx context.Context, projectPath string) (spec.RepoFacts, error)
}

type SpecBuilder interface {
	BuildProjectSpec(ctx context.Context, facts spec.RepoFacts) (spec.ProjectSpec, error)
}

type PromptCompiler interface {
	CompilePrompt(ctx context.Context, projectSpec spec.ProjectSpec) (spec.PromptBundle, error)
}

type Exporter interface {
	ExportPrompt(ctx context.Context, bundle spec.PromptBundle, outputPath string) (string, error)
}

type AgentAssembler interface {
	Build(ctx context.Context) (adkagent.Agent, error)
}

type ServiceStatus struct {
	Name        string
	Implemented bool
}

func serviceUnavailable(name string) error {
	return fmt.Errorf("%w: %s is not connected yet", ErrServiceUnavailable, name)
}
