package tui

import "projcompiler/internal/spec"

type scanStartedMsg struct {
	RunID int
	Path  string
}

type scanFinishedMsg struct {
	RunID int
	Facts spec.RepoFacts
	Err   error
}

type specStartedMsg struct {
	RunID int
}

type understandingStartedMsg struct {
	RunID int
}

type understandingFinishedMsg struct {
	RunID        int
	Summary      UnderstandingSummary
	Err          error
	UsedFallback bool
}

type specFinishedMsg struct {
	RunID       int
	ProjectSpec spec.ProjectSpec
	Err         error
}

type promptStartedMsg struct {
	RunID int
}

type promptFinishedMsg struct {
	RunID  int
	Bundle spec.PromptBundle
	Err    error
}

type exportStartedMsg struct {
	RunID      int
	OutputPath string
}

type exportFinishedMsg struct {
	RunID      int
	OutputPath string
	Err        error
}

type configSaveStartedMsg struct{}

type configSaveFinishedMsg struct {
	Err error
}
