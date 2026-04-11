package pipeline

import (
	"context"
	"path/filepath"
	"time"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/ambiguity"
	"projcompiler/internal/understanding/cache"
	"projcompiler/internal/understanding/extractors"
	"projcompiler/internal/understanding/inventory"
	"projcompiler/internal/understanding/parsers"
	"projcompiler/internal/understanding/projection"
	"projcompiler/internal/understanding/resolver"
	"projcompiler/internal/understanding/types"
)

type BuildRequest struct {
	RootPath       string
	Facts          *spec.RepoFacts
	RunID          string
	Mode           string
	ParserVersion  string
	ScannerVersion string
	VCSRef         string
}

type BuildOutput struct {
	ProjectSpec spec.ProjectSpec
}

type Builder struct {
	inventoryBuilder *inventory.Builder
	parserRegistry   *parsers.Registry
	extractor        *extractors.Builder
	resolver         *resolver.Resolver
	ambiguity        *ambiguity.Analyzer
	cache            cache.Store
}

func NewBuilder() *Builder {
	return &Builder{
		inventoryBuilder: inventory.NewBuilder(),
		parserRegistry:   parsers.NewRegistry(),
		extractor:        extractors.NewBuilder(),
		resolver:         resolver.NewResolver(),
		ambiguity:        ambiguity.NewAnalyzer(),
		cache:            cache.NewInMemoryStore(),
	}
}

func (b *Builder) Build(ctx context.Context, req BuildRequest) (BuildOutput, error) {
	if err := ctx.Err(); err != nil {
		return BuildOutput{}, err
	}
	rootPath := req.RootPath
	if rootPath == "" && req.Facts != nil {
		rootPath = req.Facts.RootPath
	}

	files, err := b.inventoryBuilder.Build(ctx, inventory.BuildInput{
		RootPath: rootPath,
		Facts:    req.Facts,
	})
	if err != nil {
		return BuildOutput{}, err
	}

	parsedByFileID := make(map[string]parsers.ParseResult, len(files))
	for i := range files {
		if err := ctx.Err(); err != nil {
			return BuildOutput{}, err
		}
		file := files[i]
		if file.Role != types.FileRoleSource && file.Role != types.FileRoleTest {
			files[i].ParseStatus = types.ParseStatusSkipped
			continue
		}

		key := cache.Key(file, req.ParserVersion)
		if cached, ok := b.cache.Get(key); ok {
			parsedByFileID[file.ID] = cached
			files[i].ParseStatus = types.ParseStatusParsed
			continue
		}

		result, _, parseErr := b.parserRegistry.ParseFile(ctx, file)
		if parseErr != nil {
			files[i].ParseStatus = types.ParseStatusFailed
			files[i].SkipReason = parseErr.Error()
			continue
		}

		b.cache.Put(key, result)
		parsedByFileID[file.ID] = result
		files[i].ParseStatus = types.ParseStatusParsed
	}

	extracted, err := b.extractor.Build(ctx, extractors.Input{
		RepoRoot: rootPath,
		Facts:    req.Facts,
		Files:    files,
		Parsed:   parsedByFileID,
	})
	if err != nil {
		return BuildOutput{}, err
	}

	resolved, err := b.resolver.Resolve(ctx, extracted.Symbols, extracted.Relations)
	if err != nil {
		return BuildOutput{}, err
	}
	extracted.Relations = resolved.Relations
	extracted.Confidences = append(extracted.Confidences, resolved.Confidences...)

	questions, err := b.ambiguity.Analyze(ctx, ambiguity.Input{
		Symbols:   extracted.Symbols,
		Relations: extracted.Relations,
		Documents: extracted.Documents,
	})
	if err != nil {
		return BuildOutput{}, err
	}

	projectSpec := projection.ToProjectSpec(projection.UnifiedInput{
		BaseFacts: req.Facts,
		Repo: types.RepoIdentity{
			RootPath: rootPath,
			RepoName: filepath.Base(rootPath),
			VCSRef:   req.VCSRef,
		},
		Snapshot: types.SnapshotMeta{
			RunID:          req.RunID,
			BuiltAt:        time.Now().UTC(),
			ScannerVersion: defaultValue(req.ScannerVersion, "scan/v1"),
			ParserVersion:  defaultValue(req.ParserVersion, "parse/v1"),
			Mode:           defaultValue(req.Mode, "quick"),
		},
		Files:       files,
		Documents:   extracted.Documents,
		Symbols:     extracted.Symbols,
		Relations:   extracted.Relations,
		Flows:       extracted.Flows,
		Constraints: extracted.Constraints,
		Questions:   questions,
		Evidences:   extracted.Evidences,
		Confidences: extracted.Confidences,
	})

	return BuildOutput{
		ProjectSpec: projectSpec,
	}, nil
}

func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
