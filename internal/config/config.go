package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	defaultOutputLanguage = "english"
	defaultPromptFilename = "prompt.txt"
)

type Config struct {
	Paths  Paths
	Model  ModelConfig
	Output OutputConfig
}

type Paths struct {
	WorkingDir string
	EnvFile    string
}

type ModelConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type OutputConfig struct {
	Language   string
	PromptFile string
}

func Load(workingDir string) (Config, error) {
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return Config{}, fmt.Errorf("resolve working directory: %w", err)
	}

	envFile := filepath.Join(absWorkingDir, ".env")
	fileValues, err := parseEnvFile(envFile)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Paths: Paths{
			WorkingDir: absWorkingDir,
			EnvFile:    envFile,
		},
		Model: ModelConfig{
			BaseURL: firstNonEmpty(
				envValue("PROJCOMPILER_BASEURL"),
				envValue("BASEURL"),
				fileValues["PROJCOMPILER_BASEURL"],
				fileValues["BASEURL"],
				fileValues["baseurl"],
			),
			APIKey: firstNonEmpty(
				envValue("PROJCOMPILER_APIKEY"),
				envValue("APIKEY"),
				fileValues["PROJCOMPILER_APIKEY"],
				fileValues["APIKEY"],
				fileValues["apikey"],
			),
			Model: firstNonEmpty(
				envValue("PROJCOMPILER_MODEL"),
				envValue("MODEL"),
				fileValues["PROJCOMPILER_MODEL"],
				fileValues["MODEL"],
				fileValues["model"],
			),
		},
		Output: OutputConfig{
			Language: normalizeLanguage(firstNonEmpty(
				envValue("PROJCOMPILER_OUTPUT_LANGUAGE"),
				envValue("OUTPUT_LANGUAGE"),
				fileValues["PROJCOMPILER_OUTPUT_LANGUAGE"],
				fileValues["OUTPUT_LANGUAGE"],
				fileValues["output_language"],
				defaultOutputLanguage,
			)),
			PromptFile: firstNonEmpty(
				envValue("PROJCOMPILER_PROMPT_FILE"),
				envValue("PROMPT_FILE"),
				fileValues["PROJCOMPILER_PROMPT_FILE"],
				fileValues["PROMPT_FILE"],
				defaultPromptFilename,
			),
		},
	}

	return cfg, nil
}

func (c Config) HasModelConfig() bool {
	return c.Model.BaseURL != "" && c.Model.APIKey != "" && c.Model.Model != ""
}

func normalizeLanguage(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return defaultOutputLanguage
	}
	return trimmed
}

// EnsureEnvFile ensures .env exists in workingDir; if not, copies from .env.example.
// Returns the absolute path to the .env file.
func EnsureEnvFile(workingDir string) (string, error) {
	return ensureEnvFile(workingDir)
}

// Save writes the ModelConfig values to the .env file.
func (c *Config) Save() error {
	values := map[string]string{
		"PROJCOMPILER_BASEURL": c.Model.BaseURL,
		"PROJCOMPILER_APIKEY":  c.Model.APIKey,
		"PROJCOMPILER_MODEL":   c.Model.Model,
	}
	return writeEnvFile(c.Paths.EnvFile, values)
}
