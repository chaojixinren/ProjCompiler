package adkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"projcompiler/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResult struct {
	Text             string
	Model            string
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

type CompletionClient interface {
	Complete(ctx context.Context, messages []Message, temperature float32) (CompletionResult, error)
}

type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	modelName  string
	httpClient *http.Client
}

func NewCompletionClient(cfg config.Config) CompletionClient {
	if !cfg.HasModelConfig() {
		return nil
	}
	return &OpenAICompatibleClient{
		baseURL:   chatCompletionsURL(cfg.Model.BaseURL),
		apiKey:    cfg.Model.APIKey,
		modelName: cfg.Model.Model,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *OpenAICompatibleClient) Complete(ctx context.Context, messages []Message, temperature float32) (CompletionResult, error) {
	if len(messages) == 0 {
		return CompletionResult{}, errors.New("completion request requires at least one message")
	}

	payload := chatCompletionsRequest{
		Model:       c.modelName,
		Messages:    messages,
		Temperature: temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("marshal completion payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return CompletionResult{}, fmt.Errorf("create completion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("invoke completion API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("read completion response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CompletionResult{}, fmt.Errorf("completion API returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	var decoded chatCompletionsResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return CompletionResult{}, fmt.Errorf("decode completion response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return CompletionResult{}, errors.New("completion API returned no choices")
	}

	return CompletionResult{
		Text:             extractChoiceText(decoded.Choices[0].Message.Content),
		Model:            firstNonEmpty(decoded.Model, c.modelName),
		PromptTokens:     decoded.Usage.PromptTokens,
		CompletionTokens: decoded.Usage.CompletionTokens,
		TotalTokens:      decoded.Usage.TotalTokens,
	}, nil
}

func chatCompletionsURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, "/chat/completions") {
		return trimmed
	}
	return trimmed + "/chat/completions"
}

func extractChoiceText(content any) string {
	switch value := content.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		var builder strings.Builder
		for _, item := range value {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, _ := part["text"].(string)
			if strings.TrimSpace(text) == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(strings.TrimSpace(text))
		}
		return strings.TrimSpace(builder.String())
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type chatCompletionsRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
}

type chatCompletionsResponse struct {
	Model   string                 `json:"model"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   chatCompletionUsage    `json:"usage"`
}

type chatCompletionChoice struct {
	Message chatCompletionMessage `json:"message"`
}

type chatCompletionMessage struct {
	Content any `json:"content"`
}

type chatCompletionUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}
