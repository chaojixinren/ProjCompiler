package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFileParsesCommentsAndQuotedValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\nBASEURL = \"https://example.com/v1\"\nAPIKEY='secret'\nmodel = gpt-5.4\n\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	values, err := parseEnvFile(path)
	if err != nil {
		t.Fatalf("parseEnvFile returned error: %v", err)
	}

	if got, want := values["BASEURL"], "https://example.com/v1"; got != want {
		t.Fatalf("BASEURL = %q, want %q", got, want)
	}
	if got, want := values["APIKEY"], "secret"; got != want {
		t.Fatalf("APIKEY = %q, want %q", got, want)
	}
	if got, want := values["model"], "gpt-5.4"; got != want {
		t.Fatalf("model = %q, want %q", got, want)
	}
}

func TestParseEnvFileRejectsMalformedLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("BROKEN_LINE\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if _, err := parseEnvFile(path); err == nil {
		t.Fatalf("expected malformed env line to fail")
	}
}

func TestLoadPrefersProcessEnvironmentAndNormalizesOutput(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "BASEURL=https://file.example.com\nAPIKEY=file-key\nMODEL=file-model\nOUTPUT_LANGUAGE=Chinese\nPROMPT_FILE=file-prompt.txt\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("BASEURL", "https://env.example.com")
	t.Setenv("APIKEY", "env-key")
	t.Setenv("MODEL", "env-model")
	t.Setenv("OUTPUT_LANGUAGE", " ENGLISH ")
	t.Setenv("PROMPT_FILE", "nested/out.txt")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := cfg.Paths.WorkingDir, dir; got != want {
		t.Fatalf("WorkingDir = %q, want %q", got, want)
	}
	if got, want := cfg.Model.BaseURL, "https://env.example.com"; got != want {
		t.Fatalf("BaseURL = %q, want %q", got, want)
	}
	if got, want := cfg.Model.APIKey, "env-key"; got != want {
		t.Fatalf("APIKey = %q, want %q", got, want)
	}
	if got, want := cfg.Model.Model, "env-model"; got != want {
		t.Fatalf("Model = %q, want %q", got, want)
	}
	if got, want := cfg.Output.Language, "english"; got != want {
		t.Fatalf("Language = %q, want %q", got, want)
	}
	if got, want := cfg.Output.PromptFile, "nested/out.txt"; got != want {
		t.Fatalf("PromptFile = %q, want %q", got, want)
	}
	if !cfg.HasModelConfig() {
		t.Fatalf("expected model config to be complete")
	}
}

func TestLoadUsesDefaultsWhenEnvIsMissing(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := cfg.Paths.EnvFile, filepath.Join(dir, ".env"); got != want {
		t.Fatalf("EnvFile = %q, want %q", got, want)
	}
	if got, want := cfg.Output.Language, "english"; got != want {
		t.Fatalf("Language = %q, want %q", got, want)
	}
	if got, want := cfg.Output.PromptFile, "prompt.txt"; got != want {
		t.Fatalf("PromptFile = %q, want %q", got, want)
	}
	if cfg.HasModelConfig() {
		t.Fatalf("expected model config to be incomplete without environment values")
	}
}

func TestLoadPrefersProjCompilerPrefixedVariables(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "BASEURL=https://file.example.com\nAPIKEY=file-key\nMODEL=file-model\nPROJCOMPILER_BASEURL=https://prefixed-file.example.com\nPROJCOMPILER_APIKEY=prefixed-file-key\nPROJCOMPILER_MODEL=prefixed-file-model\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("BASEURL", "https://generic-env.example.com")
	t.Setenv("APIKEY", "generic-env-key")
	t.Setenv("MODEL", "generic-env-model")
	t.Setenv("PROJCOMPILER_BASEURL", "https://prefixed-env.example.com")
	t.Setenv("PROJCOMPILER_APIKEY", "prefixed-env-key")
	t.Setenv("PROJCOMPILER_MODEL", "prefixed-env-model")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := cfg.Model.BaseURL, "https://prefixed-env.example.com"; got != want {
		t.Fatalf("BaseURL = %q, want %q", got, want)
	}
	if got, want := cfg.Model.APIKey, "prefixed-env-key"; got != want {
		t.Fatalf("APIKey = %q, want %q", got, want)
	}
	if got, want := cfg.Model.Model, "prefixed-env-model"; got != want {
		t.Fatalf("Model = %q, want %q", got, want)
	}
}

func TestParseEnvFileRejectsEmptyKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(" = value\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if _, err := parseEnvFile(path); err == nil {
		t.Fatalf("expected empty key to fail")
	}
}
