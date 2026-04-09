package prompt

import (
	"context"
	"strings"

	"projcompiler/internal/app"
	"projcompiler/internal/config"
	"projcompiler/internal/spec"
)

type Compiler struct {
	outputLanguage string
}

var _ app.PromptCompiler = (*Compiler)(nil)

func NewCompiler(cfg config.Config) *Compiler {
	return &Compiler{
		outputLanguage: normalizedOutputLanguage(cfg.Output.Language),
	}
}

func (c *Compiler) CompilePrompt(ctx context.Context, projectSpec spec.ProjectSpec) (spec.PromptBundle, error) {
	if err := ctx.Err(); err != nil {
		return spec.PromptBundle{}, err
	}

	blocking := blockingQuestions(projectSpec.OpenQuestions)
	bundle := spec.PromptBundle{
		OutputLanguage:    c.outputLanguage,
		BlockingQuestions: blocking,
		Notes:             make([]string, 0, 2),
	}

	if len(blocking) > 0 {
		bundle.Notes = append(bundle.Notes, "Prompt generation stopped because there are blocking open questions.")
		return bundle, nil
	}

	bundle.Sections = filterPromptSections(projectSpec)
	if len(bundle.Sections) == 0 {
		bundle.Notes = append(bundle.Notes, "No prompt-worthy sections were found in the project spec.")
		return bundle, nil
	}

	bundle.PromptText = renderPrompt(bundle.Sections)
	bundle.Notes = append(bundle.Notes, "Prompt compiled with the /init-style pruning rule: only include details the model is likely to get wrong without explicit guidance.")
	return bundle, nil
}

func blockingQuestions(questions []spec.OpenQuestion) []spec.OpenQuestion {
	out := make([]spec.OpenQuestion, 0, len(questions))
	for _, question := range questions {
		if question.Blocking && strings.TrimSpace(question.Question) != "" {
			out = append(out, question)
		}
	}
	return out
}

func normalizedOutputLanguage(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "english"
	}
	return trimmed
}
