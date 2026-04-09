package scan

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"projcompiler/internal/spec"
)

type RepoTreeTool struct {
	toolInfo
}

func NewRepoTreeTool() RepoTreeTool {
	return RepoTreeTool{toolInfo{name: "RepoTreeTool"}}
}

func (t RepoTreeTool) Execute(ctx context.Context, input RepoTreeInput) (RepoTreeOutput, error) {
	request := input.Request
	output := RepoTreeOutput{
		Tree: spec.FileTreeSummary{},
		ScanMeta: spec.ScanMeta{
			IgnoreSources: input.Ignore.Sources(),
			MaxDepth:      request.MaxDepth,
			MaxFiles:      request.MaxFiles,
		},
	}

	fileCount := 0
	err := filepath.WalkDir(request.RootPath, func(fullPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(request.RootPath, fullPath)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		relPath = filepath.ToSlash(relPath)
		depth := strings.Count(relPath, "/") + 1
		if depth > request.MaxDepth {
			output.ScanMeta.IgnoredPaths = append(output.ScanMeta.IgnoredPaths, spec.IgnoredPath{
				Path:   relPath,
				Reason: "depth_limit",
			})
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ignored, reason := input.Ignore.ShouldIgnore(relPath, entry.IsDir())
		if ignored {
			output.ScanMeta.IgnoredPaths = append(output.ScanMeta.IgnoredPaths, spec.IgnoredPath{
				Path:   relPath,
				Reason: reason,
			})
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !entry.IsDir() && info.Size() > request.MaxFileSize {
			output.ScanMeta.IgnoredPaths = append(output.ScanMeta.IgnoredPaths, spec.IgnoredPath{
				Path:   relPath,
				Reason: "file_too_large",
			})
			return nil
		}

		if fileCount >= request.MaxFiles {
			output.ScanMeta.Truncated = true
			output.ScanMeta.IgnoredPaths = append(output.ScanMeta.IgnoredPaths, spec.IgnoredPath{
				Path:   relPath,
				Reason: "file_budget_exceeded",
			})
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		kind := spec.FileNodeFile
		if entry.IsDir() {
			kind = spec.FileNodeDirectory
		} else {
			fileCount++
		}
		output.Tree.Nodes = append(output.Tree.Nodes, spec.FileNode{
			Path: relPath,
			Kind: kind,
			Size: info.Size(),
		})

		if entry.IsDir() {
			return nil
		}
		classifyCandidate(relPath, &output.Tree)
		return nil
	})
	if err != nil {
		return RepoTreeOutput{}, err
	}

	sort.Slice(output.Tree.Nodes, func(i, j int) bool {
		return output.Tree.Nodes[i].Path < output.Tree.Nodes[j].Path
	})
	sort.Strings(output.Tree.CandidateDocs)
	sort.Strings(output.Tree.CandidateManifests)
	sort.Strings(output.Tree.CandidateSources)
	sort.Slice(output.ScanMeta.IgnoredPaths, func(i, j int) bool {
		return output.ScanMeta.IgnoredPaths[i].Path < output.ScanMeta.IgnoredPaths[j].Path
	})

	return output, nil
}

func classifyCandidate(relPath string, tree *spec.FileTreeSummary) {
	base := strings.ToLower(filepath.Base(relPath))
	ext := strings.ToLower(filepath.Ext(relPath))

	switch {
	case isDocCandidate(base, ext):
		tree.CandidateDocs = append(tree.CandidateDocs, relPath)
	case isManifestCandidate(base):
		tree.CandidateManifests = append(tree.CandidateManifests, relPath)
	case isSourceCandidate(ext):
		tree.CandidateSources = append(tree.CandidateSources, relPath)
	}
}

func isDocCandidate(base, ext string) bool {
	if strings.HasPrefix(base, "readme") {
		return true
	}
	return ext == ".md" || ext == ".mdx" || ext == ".txt"
}

func isManifestCandidate(base string) bool {
	switch base {
	case "go.mod", "go.sum", "package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
		"pyproject.toml", "cargo.toml", "makefile", "dockerfile", ".env.example":
		return true
	default:
		return false
	}
}

func isSourceCandidate(ext string) bool {
	switch ext {
	case ".go", ".js", ".jsx", ".ts", ".tsx", ".py", ".swift", ".rs", ".java", ".kt":
		return true
	default:
		return false
	}
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
