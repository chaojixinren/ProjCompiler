package scan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"projcompiler/internal/spec"
)

type ScanOrchestrator struct {
	request       ScanRequest
	repoTree      RepoTreeTool
	manifestProbe ManifestProbeTool
	docsProbe     DocsProbeTool
	entryProbe    EntrypointProbeTool
	snippetSelect SnippetSelectTool
}

func NewOrchestrator(request ScanRequest) (*ScanOrchestrator, error) {
	normalized, err := request.normalize()
	if err != nil {
		return nil, err
	}
	return &ScanOrchestrator{
		request:       normalized,
		repoTree:      NewRepoTreeTool(),
		manifestProbe: NewManifestProbeTool(),
		docsProbe:     NewDocsProbeTool(),
		entryProbe:    NewEntrypointProbeTool(),
		snippetSelect: NewSnippetSelectTool(),
	}, nil
}

func NewDefaultOrchestrator() *ScanOrchestrator {
	orchestrator, err := NewOrchestrator(ScanRequest{RootPath: "."})
	if err != nil {
		return &ScanOrchestrator{
			request:       ScanRequest{},
			repoTree:      NewRepoTreeTool(),
			manifestProbe: NewManifestProbeTool(),
			docsProbe:     NewDocsProbeTool(),
			entryProbe:    NewEntrypointProbeTool(),
			snippetSelect: NewSnippetSelectTool(),
		}
	}
	return orchestrator
}

func (o *ScanOrchestrator) Scan(ctx context.Context, projectPath string) (spec.RepoFacts, error) {
	request := o.request
	request.RootPath = projectPath
	normalized, err := request.normalize()
	if err != nil {
		return spec.RepoFacts{}, err
	}
	info, err := os.Stat(normalized.RootPath)
	if err != nil {
		return spec.RepoFacts{}, err
	}
	if !info.IsDir() {
		return spec.RepoFacts{}, fmt.Errorf("project path must be a directory: %s", normalized.RootPath)
	}

	ignoreMatcher, err := LoadIgnoreMatcher(normalized.RootPath, normalized.IncludeVendor)
	if err != nil {
		return spec.RepoFacts{}, err
	}

	treeOutput, err := o.repoTree.Execute(ctx, RepoTreeInput{
		Request: normalized,
		Ignore:  ignoreMatcher,
	})
	if err != nil {
		return spec.RepoFacts{}, err
	}

	var (
		manifestOutput ManifestProbeOutput
		docsOutput     DocsProbeOutput
		manifestErr    error
		docsErr        error
		wg             sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		manifestOutput, manifestErr = o.manifestProbe.Execute(ctx, ManifestProbeInput{
			RootPath:       normalized.RootPath,
			CandidatePaths: treeOutput.Tree.CandidateManifests,
			MaxBytes:       normalized.MaxFileSize,
		})
	}()
	go func() {
		defer wg.Done()
		docsOutput, docsErr = o.docsProbe.Execute(ctx, DocsProbeInput{
			RootPath:       normalized.RootPath,
			CandidatePaths: treeOutput.Tree.CandidateDocs,
			MaxDocs:        normalized.MaxDocFiles,
			MaxBytes:       normalized.MaxDocBytes,
		})
	}()
	wg.Wait()
	if manifestErr != nil {
		return spec.RepoFacts{}, manifestErr
	}
	if docsErr != nil {
		return spec.RepoFacts{}, docsErr
	}

	entryOutput, err := o.entryProbe.Execute(ctx, EntrypointProbeInput{
		RootPath: normalized.RootPath,
		Tree:     treeOutput.Tree,
		Build:    manifestOutput.Build,
		Docs:     docsOutput.Docs,
	})
	if err != nil {
		return spec.RepoFacts{}, err
	}

	snippetOutput, err := o.snippetSelect.Execute(ctx, SnippetSelectInput{
		RootPath:        normalized.RootPath,
		Tree:            treeOutput.Tree,
		EntryPoints:     entryOutput.EntryPoints,
		MaxSnippets:     normalized.MaxSnippets,
		MaxSnippetLines: normalized.MaxSnippetLines,
		MaxBytes:        normalized.MaxFileSize,
	})
	if err != nil {
		return spec.RepoFacts{}, err
	}

	facts := spec.RepoFacts{
		RootPath:    normalized.RootPath,
		ScanMeta:    treeOutput.ScanMeta,
		Tree:        treeOutput.Tree,
		Docs:        docsOutput.Docs,
		Build:       manifestOutput.Build,
		EntryPoints: entryOutput.EntryPoints,
		Snippets:    snippetOutput.Snippets,
		Constraints: mergeConstraints(docsOutput.Constraints, manifestOutput.Constraints),
		Signals: mergeSignals(
			manifestOutput.Signals,
			docsOutput.Signals,
			entryOutput.Signals,
			snippetOutput.Signals,
			scanSignals(treeOutput.ScanMeta),
		),
	}
	return facts, nil
}

func scanSignals(meta spec.ScanMeta) []spec.FactSignal {
	var signals []spec.FactSignal
	if meta.Truncated {
		signals = append(signals, spec.MissingSignal("scan was truncated by file budget"))
	}
	if len(meta.IgnoredPaths) == 0 {
		return signals
	}
	ignored := meta.IgnoredPaths
	if len(ignored) > 3 {
		ignored = ignored[:3]
	}
	evidence := make([]spec.EvidenceRef, 0, len(ignored))
	for _, item := range ignored {
		evidence = append(evidence, spec.EvidenceRef{
			Path: item.Path,
			Note: item.Reason,
		})
	}
	signals = append(signals, spec.ObservedSignal("recorded ignored paths during scan", evidence...))
	return signals
}

func mergeConstraints(groups ...[]spec.Constraint) []spec.Constraint {
	seen := map[string]bool{}
	var merged []spec.Constraint
	for _, group := range groups {
		for _, item := range group {
			key := item.Text + ":" + item.Signal.Reason
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, item)
		}
	}
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Text < merged[j].Text
	})
	return merged
}

func mergeSignals(groups ...[]spec.FactSignal) []spec.FactSignal {
	var merged []spec.FactSignal
	for _, group := range groups {
		merged = append(merged, group...)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].Kind == merged[j].Kind {
			return merged[i].Reason < merged[j].Reason
		}
		return merged[i].Kind < merged[j].Kind
	})
	return merged
}

func NormalizePath(rootPath, relPath string) string {
	if relPath == "" {
		return rootPath
	}
	return filepath.Join(rootPath, relPath)
}
