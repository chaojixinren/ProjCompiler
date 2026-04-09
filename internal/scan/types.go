package scan

import (
	"errors"
	"path/filepath"

	"projcompiler/internal/spec"
)

const (
	defaultMaxDepth        = 4
	defaultMaxFiles        = 400
	defaultMaxFileSize     = 2 << 20
	defaultMaxDocFiles     = 4
	defaultMaxDocBytes     = 64 << 10
	defaultMaxSnippets     = 6
	defaultSnippetLines    = 120
	defaultEntrypointFiles = 80
)

var ErrProjectPathRequired = errors.New("project path is required")

type ScanRequest struct {
	RootPath        string
	MaxDepth        int
	MaxFiles        int
	MaxFileSize     int64
	IncludeVendor   bool
	MaxDocFiles     int
	MaxDocBytes     int
	MaxSnippets     int
	MaxSnippetLines int
}

func (r ScanRequest) normalize() (ScanRequest, error) {
	if r.RootPath == "" {
		return ScanRequest{}, ErrProjectPathRequired
	}

	abs, err := filepath.Abs(r.RootPath)
	if err != nil {
		return ScanRequest{}, err
	}

	r.RootPath = abs
	if r.MaxDepth <= 0 {
		r.MaxDepth = defaultMaxDepth
	}
	if r.MaxFiles <= 0 {
		r.MaxFiles = defaultMaxFiles
	}
	if r.MaxFileSize <= 0 {
		r.MaxFileSize = defaultMaxFileSize
	}
	if r.MaxDocFiles <= 0 {
		r.MaxDocFiles = defaultMaxDocFiles
	}
	if r.MaxDocBytes <= 0 {
		r.MaxDocBytes = defaultMaxDocBytes
	}
	if r.MaxSnippets <= 0 {
		r.MaxSnippets = defaultMaxSnippets
	}
	if r.MaxSnippetLines <= 0 {
		r.MaxSnippetLines = defaultSnippetLines
	}
	return r, nil
}

type RepoTreeInput struct {
	Request ScanRequest
	Ignore  IgnoreMatcher
}

type RepoTreeOutput struct {
	Tree     spec.FileTreeSummary
	ScanMeta spec.ScanMeta
}

type ManifestProbeInput struct {
	RootPath       string
	CandidatePaths []string
	MaxBytes       int64
}

type ManifestProbeOutput struct {
	Build       spec.BuildSummary
	Constraints []spec.Constraint
	Signals     []spec.FactSignal
}

type DocsProbeInput struct {
	RootPath       string
	CandidatePaths []string
	MaxDocs        int
	MaxBytes       int
}

type DocsProbeOutput struct {
	Docs        []spec.DocumentFact
	Constraints []spec.Constraint
	Signals     []spec.FactSignal
}

type EntrypointProbeInput struct {
	RootPath string
	Tree     spec.FileTreeSummary
	Build    spec.BuildSummary
	Docs     []spec.DocumentFact
}

type EntrypointProbeOutput struct {
	EntryPoints []spec.EntryPoint
	Signals     []spec.FactSignal
}

type SnippetSelectInput struct {
	RootPath        string
	Tree            spec.FileTreeSummary
	EntryPoints     []spec.EntryPoint
	MaxSnippets     int
	MaxSnippetLines int
	MaxBytes        int64
}

type SnippetSelectOutput struct {
	Snippets []spec.CodeSnippet
	Signals  []spec.FactSignal
}
