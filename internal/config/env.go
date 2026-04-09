package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open env file %q: %w", path, err)
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("parse env file %q: line %d is missing '='", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("parse env file %q: line %d has empty key", path, lineNumber)
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file %q: %w", path, err)
	}

	return values, nil
}

func envValue(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// writeEnvFile writes key-value pairs to the specified .env file.
// If the file exists, it updates only the keys provided while preserving comments and other keys.
// If the file does not exist, it creates it with the provided values.
func writeEnvFile(path string, values map[string]string) error {
	existingContent := ""
	if file, err := os.Open(path); err == nil {
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("read existing env file %q: %w", path, err)
		}
		existingContent = string(content)
	}

	var output strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(existingContent))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			output.WriteString(line + "\n")
			continue
		}

		key, _, ok := strings.Cut(trimmed, "=")
		if !ok {
			output.WriteString(line + "\n")
			continue
		}
		key = strings.TrimSpace(key)

		if newVal, hasNew := values[key]; hasNew {
			output.WriteString(key + "=" + newVal + "\n")
		} else {
			output.WriteString(line + "\n")
		}
	}

	for key, value := range values {
		if !containsKey(existingContent, key) {
			output.WriteString(key + "=" + value + "\n")
		}
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("write temp env file %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp env file: %w", err)
	}
	return nil
}

func containsKey(content, key string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		existingKey, _, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(existingKey) == key {
			return true
		}
	}
	return false
}

// ensureEnvFile checks if .env exists in workingDir; if not, copies from .env.example.
// Returns the absolute path to the .env file.
func ensureEnvFile(workingDir string) (string, error) {
	absDir, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	envPath := filepath.Join(absDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		return envPath, nil
	}

	examplePath := filepath.Join(absDir, ".env.example")
	exampleContent, err := os.ReadFile(examplePath)
	if err != nil {
		// Create a default .env if .env.example doesn't exist
		defaultContent := `# ProjCompiler configuration file

# Required. The program will append /chat/completions automatically.
PROJCOMPILER_BASEURL=https://your-model-endpoint.example/v1

# Required. Bearer token for the model provider.
PROJCOMPILER_APIKEY=your-api-key

# Required. Model name sent to the provider.
PROJCOMPILER_MODEL=your-model-name

# Optional. Defaults to english.
PROJCOMPILER_OUTPUT_LANGUAGE=english

# Optional. Defaults to prompt.txt in the working directory.
PROJCOMPILER_PROMPT_FILE=prompt.txt
`
		if err := os.WriteFile(envPath, []byte(defaultContent), 0644); err != nil {
			return "", fmt.Errorf("create default env file %q: %w", envPath, err)
		}
		return envPath, nil
	}

	if err := os.WriteFile(envPath, exampleContent, 0644); err != nil {
		return "", fmt.Errorf("copy env example to %q: %w", envPath, err)
	}
	return envPath, nil
}
