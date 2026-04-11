package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

var ErrMissingService = errors.New("missing tui service")

type State string

const (
	StatePathInput     State = "path_input"
	StateScanning      State = "scanning"
	StateSpecSummary   State = "spec_summary"
	StatePromptPreview State = "prompt_preview"
	StateDone          State = "done"
	StateConfigEdit    State = "config_edit"
)

func (s State) Title() string {
	switch s {
	case StatePathInput:
		return "Path Input"
	case StateScanning:
		return "Scanning"
	case StateSpecSummary:
		return "Spec Summary"
	case StatePromptPreview:
		return "Prompt Preview"
	case StateDone:
		return "Done"
	case StateConfigEdit:
		return "Config Edit"
	default:
		return "Unknown"
	}
}

type Scanner interface {
	Scan(ctx context.Context, projectPath string) (spec.RepoFacts, error)
}

type SpecBuilder interface {
	BuildProjectSpec(ctx context.Context, facts spec.RepoFacts) (spec.ProjectSpec, error)
}

type UnderstandingBuilder interface {
	BuildUnderstanding(ctx context.Context, facts spec.RepoFacts) (UnderstandingSummary, error)
}

type PromptCompiler interface {
	CompilePrompt(ctx context.Context, projectSpec spec.ProjectSpec) (spec.PromptBundle, error)
}

type Exporter interface {
	ExportPrompt(ctx context.Context, bundle spec.PromptBundle, outputPath string) (string, error)
}

type Services struct {
	Scanner              Scanner
	UnderstandingBuilder UnderstandingBuilder
	SpecBuilder          SpecBuilder
	PromptCompiler       PromptCompiler
	Exporter             Exporter
	Config               *config.Config
}

type SpecSection string

const (
	SpecSectionOverview  SpecSection = "overview"
	SpecSectionModules   SpecSection = "modules"
	SpecSectionFlows     SpecSection = "flows"
	SpecSectionEvidence  SpecSection = "evidence"
	SpecSectionQuestions SpecSection = "questions"
	SpecSectionModuleMap SpecSection = "module_map"
)

type ScrollMode string

const (
	ScrollModeWrap       ScrollMode = "wrap"
	ScrollModeHorizontal ScrollMode = "horizontal"
)

type UnderstandingStats struct {
	Modules       int
	Flows         int
	OpenQuestions int
	Evidence      int
}

type UnderstandingSummary struct {
	Status        string
	Source        string
	GeneratedAt   string
	Overview      []string
	OpenQuestions []string
	Evidence      []string
	ModuleMap     []string
	Stats         UnderstandingStats
}

type Config struct {
	Title             string
	DefaultPromptPath string
	MaxLogLines       int
}

func DefaultConfig() Config {
	return Config{
		Title:             "ProjCompiler",
		DefaultPromptPath: "prompt.txt",
		MaxLogLines:       12,
	}
}

type Option func(*Config)

func WithTitle(title string) Option {
	return func(cfg *Config) {
		if strings.TrimSpace(title) != "" {
			cfg.Title = title
		}
	}
}

func WithDefaultPromptPath(path string) Option {
	return func(cfg *Config) {
		if strings.TrimSpace(path) != "" {
			cfg.DefaultPromptPath = path
		}
	}
}

func WithMaxLogLines(n int) Option {
	return func(cfg *Config) {
		if n > 0 {
			cfg.MaxLogLines = n
		}
	}
}

func missingServiceError(name string) error {
	return fmt.Errorf("%w: %s", ErrMissingService, name)
}
