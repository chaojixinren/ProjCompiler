package scan

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type IgnoreMatcher struct {
	rootPath string
	rules    []ignoreRule
	sources  []string
}

type ignoreRule struct {
	pattern  string
	negated  bool
	dirOnly  bool
	anchored bool
	source   string
}

func LoadIgnoreMatcher(rootPath string, includeVendor bool) (IgnoreMatcher, error) {
	matcher := IgnoreMatcher{
		rootPath: rootPath,
	}
	matcher.addBuiltins(includeVendor)

	for _, name := range []string{".projcompilerignore", ".gitignore"} {
		fullPath := filepath.Join(rootPath, name)
		file, err := os.Open(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return IgnoreMatcher{}, err
		}
		if err := matcher.loadFile(file, name); err != nil {
			file.Close()
			return IgnoreMatcher{}, err
		}
		file.Close()
	}

	return matcher, nil
}

func (m IgnoreMatcher) Sources() []string {
	return append([]string(nil), m.sources...)
}

func (m IgnoreMatcher) ShouldIgnore(relPath string, isDir bool) (bool, string) {
	normalized := filepath.ToSlash(strings.TrimPrefix(relPath, "./"))
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" || normalized == "." {
		return false, ""
	}

	matched := false
	reason := ""
	for _, rule := range m.rules {
		if rule.matches(normalized, isDir) {
			matched = !rule.negated
			reason = rule.source + ":" + rule.pattern
		}
	}
	return matched, reason
}

func (m *IgnoreMatcher) addBuiltins(includeVendor bool) {
	patterns := []string{
		".git/",
		".gocache/",
		".gopath/",
		"node_modules/",
		"dist/",
		"build/",
		"bin/",
		".DS_Store",
	}
	if !includeVendor {
		patterns = append(patterns, "vendor/")
	}
	for _, pattern := range patterns {
		m.rules = append(m.rules, parseIgnoreRule(pattern, "builtin"))
	}
	m.sources = append(m.sources, "builtin")
}

func (m *IgnoreMatcher) loadFile(file *os.File, source string) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rule := parseIgnoreRule(line, source)
		if rule.pattern == "" {
			continue
		}
		m.rules = append(m.rules, rule)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	m.sources = append(m.sources, source)
	return nil
}

func parseIgnoreRule(line, source string) ignoreRule {
	rule := ignoreRule{source: source}
	if strings.HasPrefix(line, "!") {
		rule.negated = true
		line = strings.TrimPrefix(line, "!")
	}
	if strings.HasPrefix(line, "/") {
		rule.anchored = true
		line = strings.TrimPrefix(line, "/")
	}
	if strings.HasSuffix(line, "/") {
		rule.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	rule.pattern = filepath.ToSlash(strings.TrimSpace(line))
	return rule
}

func (r ignoreRule) matches(relPath string, isDir bool) bool {
	if r.pattern == "" {
		return false
	}
	if r.dirOnly && !isDir && relPath != r.pattern && !strings.HasPrefix(relPath, r.pattern+"/") {
		return false
	}

	pattern := strings.TrimPrefix(r.pattern, "**/")
	if r.anchored {
		return matchPath(pattern, relPath, isDir, true)
	}
	if strings.Contains(pattern, "/") {
		if matchPath(pattern, relPath, isDir, false) {
			return true
		}
		parts := strings.Split(relPath, "/")
		for i := 1; i < len(parts); i++ {
			if matchPath(pattern, strings.Join(parts[i:], "/"), isDir, false) {
				return true
			}
		}
		return false
	}

	segments := strings.Split(relPath, "/")
	for _, segment := range segments {
		if simpleMatch(pattern, segment) {
			return true
		}
	}
	return false
}

func matchPath(pattern, relPath string, isDir bool, anchored bool) bool {
	if simpleMatch(pattern, relPath) {
		return true
	}
	if !strings.ContainsAny(pattern, "*?[") {
		if relPath == pattern {
			return true
		}
		if isDir && strings.HasPrefix(relPath, pattern+"/") {
			return true
		}
		if !anchored && strings.HasSuffix(relPath, "/"+pattern) {
			return true
		}
	}
	return false
}

func simpleMatch(pattern, target string) bool {
	ok, err := path.Match(pattern, target)
	if err == nil && ok {
		return true
	}
	return pattern == target
}
