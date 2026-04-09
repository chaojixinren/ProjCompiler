package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIgnoreMatcherHonorsBuiltinsAndCustomNegation(t *testing.T) {
	root := t.TempDir()
	content := "generated/\n!important.go\n"
	if err := os.WriteFile(filepath.Join(root, ".projcompilerignore"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	matcher, err := LoadIgnoreMatcher(root, false)
	if err != nil {
		t.Fatalf("LoadIgnoreMatcher returned error: %v", err)
	}

	if ignored, _ := matcher.ShouldIgnore("vendor/pkg.go", true); !ignored {
		t.Fatalf("expected vendor/ to be ignored by builtin rules")
	}
	if ignored, _ := matcher.ShouldIgnore("generated/file.go", true); !ignored {
		t.Fatalf("expected generated/ to be ignored by custom rule")
	}
	if ignored, _ := matcher.ShouldIgnore("important.go", false); ignored {
		t.Fatalf("expected negated path to remain visible")
	}
}

func TestLoadIgnoreMatcherCanKeepVendorWhenRequested(t *testing.T) {
	root := t.TempDir()
	matcher, err := LoadIgnoreMatcher(root, true)
	if err != nil {
		t.Fatalf("LoadIgnoreMatcher returned error: %v", err)
	}

	if ignored, _ := matcher.ShouldIgnore("vendor/pkg.go", true); ignored {
		t.Fatalf("expected vendor/ not to be ignored when IncludeVendor is true")
	}
}
