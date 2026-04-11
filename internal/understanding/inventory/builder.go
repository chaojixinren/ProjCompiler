package inventory

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"projcompiler/internal/spec"
	"projcompiler/internal/understanding/types"
)

type BuildInput struct {
	RootPath string
	Facts    *spec.RepoFacts
}

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Build(ctx context.Context, input BuildInput) ([]types.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	root := input.RootPath
	if root == "" && input.Facts != nil {
		root = input.Facts.RootPath
	}

	nodes := make([]spec.FileNode, 0)
	if input.Facts != nil {
		nodes = slices.Clone(input.Facts.Tree.Nodes)
	}

	if len(nodes) == 0 {
		walked, err := walkNodes(root)
		if err != nil {
			return nil, err
		}
		nodes = walked
	}

	files := make([]types.File, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind != spec.FileNodeFile {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		relPath := types.NormalizePath(node.Path)
		absPath := filepath.Join(root, relPath)
		language := detectLanguage(relPath)
		role := detectRole(relPath)
		hash := hashFile(absPath)
		fileID := types.NewID("file", relPath)
		files = append(files, types.File{
			ID:          fileID,
			Path:        relPath,
			AbsPath:     absPath,
			Language:    language,
			Role:        role,
			SizeBytes:   node.Size,
			ContentHash: hash,
			ParseStatus: types.ParseStatusSkipped,
		})
	}
	return files, nil
}

func walkNodes(root string) ([]spec.FileNode, error) {
	out := make([]spec.FileNode, 0, 64)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		out = append(out, spec.FileNode{
			Path: types.NormalizePath(relPath),
			Kind: spec.FileNodeFile,
			Size: info.Size(),
		})
		return nil
	})
	return out, err
}

func detectLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".md":
		return "markdown"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return "unknown"
	}
}

func detectRole(path string) types.FileRole {
	lower := strings.ToLower(path)
	base := strings.ToLower(filepath.Base(path))
	switch {
	case strings.HasSuffix(lower, "_test.go") || strings.Contains(lower, "/test/"):
		return types.FileRoleTest
	case base == "go.mod" || base == "go.sum" || strings.HasPrefix(base, "package.") || base == "pom.xml" || base == "cargo.toml":
		return types.FileRoleManifest
	case strings.HasPrefix(base, "readme") || strings.HasSuffix(base, ".md"):
		return types.FileRoleDoc
	case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".toml") || strings.Contains(base, ".env"):
		return types.FileRoleConfig
	case strings.HasSuffix(lower, ".pb.go") || strings.Contains(lower, "/vendor/"):
		return types.FileRoleGenerated
	case strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, ".ts") || strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".py"):
		return types.FileRoleSource
	default:
		return types.FileRoleOther
	}
}

func hashFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha1.Sum(content)
	return hex.EncodeToString(sum[:])
}
