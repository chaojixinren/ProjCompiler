package adkflow

import (
	"context"
	"fmt"
	"strings"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"

	"projcompiler/internal/app"
	"projcompiler/internal/config"
)

type Assembler struct {
	cfg    config.Config
	client CompletionClient
}

var _ app.AgentAssembler = (*Assembler)(nil)

func NewAssembler(cfg config.Config) *Assembler {
	return &Assembler{
		cfg:    cfg,
		client: NewCompletionClient(cfg),
	}
}

func (a *Assembler) Build(ctx context.Context) (adkagent.Agent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a.client == nil {
		return nil, fmt.Errorf("model configuration is incomplete: expected baseurl, apikey, and model")
	}

	modelAdapter := NewOpenAICompatibleLLM(a.cfg.Model.Model, a.client)

	analyzer, err := llmagent.New(llmagent.Config{
		Name:        "spec_analyzer",
		Description: "Synthesizes a compact project specification from repository facts.",
		Model:       modelAdapter,
		Instruction: analyzerAgentInstruction(),
	})
	if err != nil {
		return nil, fmt.Errorf("build analyzer agent: %w", err)
	}

	compiler, err := llmagent.New(llmagent.Config{
		Name:        "prompt_compiler",
		Description: "Compiles a project specification into a tight implementation prompt.",
		Model:       modelAdapter,
		Instruction: compilerAgentInstruction(),
	})
	if err != nil {
		return nil, fmt.Errorf("build prompt compiler agent: %w", err)
	}

	workflow, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: adkagent.Config{
			Name:        "projcompiler_pipeline",
			Description: "Runs project-spec synthesis and prompt compilation in a fixed order.",
			SubAgents:   []adkagent.Agent{analyzer, compiler},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("build sequential workflow: %w", err)
	}

	return workflow, nil
}

func analyzerAgentInstruction() string {
	return strings.TrimSpace(`
You are the ProjCompiler analyzer agent.
Read repository facts and produce a compact project specification.
Only keep facts that materially change what a code-generation model should build.
Do not include generic engineering advice, file inventories, or default framework behavior.
`)
}

func compilerAgentInstruction() string {
	return strings.TrimSpace(`
You are the ProjCompiler prompt compiler agent.
Convert a project specification into a pure English implementation prompt.
Follow the ClaudeCode /init pruning rule:
if removing a detail would not make the model build the wrong thing, leave it out.
Never emit CLI wrappers or tool command prefixes.
`)
}
