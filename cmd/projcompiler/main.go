package main

import (
	"errors"
	"fmt"
	"os"

	"projcompiler/internal/adkflow"
	"projcompiler/internal/app"
	"projcompiler/internal/config"
	"projcompiler/internal/prompt"
)

func main() {
	// Ensure .env file exists before loading config
	if _, err := config.EnsureEnvFile("."); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ensure env file: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	application := app.New(cfg, app.Services{
		SpecBuilder:    adkflow.NewSpecBuilder(cfg),
		PromptCompiler: prompt.NewCompiler(cfg),
		AgentAssembler: adkflow.NewAssembler(cfg),
	})
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ProjCompiler: %v\n", err)
		if errors.Is(err, app.ErrServiceUnavailable) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
