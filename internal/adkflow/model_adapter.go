package adkflow

import (
	"context"
	"errors"
	"iter"
	"strings"

	"google.golang.org/genai"

	adkmodel "google.golang.org/adk/model"
)

type OpenAICompatibleLLM struct {
	modelName string
	client    CompletionClient
}

var ErrNilLLMRequest = errors.New("llm request is nil")

func NewOpenAICompatibleLLM(modelName string, client CompletionClient) *OpenAICompatibleLLM {
	return &OpenAICompatibleLLM{
		modelName: strings.TrimSpace(modelName),
		client:    client,
	}
}

func (m *OpenAICompatibleLLM) Name() string {
	return m.modelName
}

func (m *OpenAICompatibleLLM) GenerateContent(ctx context.Context, req *adkmodel.LLMRequest, _ bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		if m.client == nil {
			yield(nil, errors.New("llm client is not configured"))
			return
		}
		if req == nil {
			yield(nil, ErrNilLLMRequest)
			return
		}

		messages := make([]Message, 0, len(req.Contents))
		for _, content := range req.Contents {
			text := flattenContentText(content)
			if text == "" {
				continue
			}
			role := strings.TrimSpace(content.Role)
			if role == "" {
				role = "user"
			}
			messages = append(messages, Message{
				Role:    role,
				Content: text,
			})
		}
		if len(messages) == 0 {
			yield(nil, errors.New("llm request has no text content"))
			return
		}

		temperature := float32(0.1)
		if req.Config != nil && req.Config.Temperature != nil {
			temperature = *req.Config.Temperature
		}

		result, err := m.client.Complete(ctx, messages, temperature)
		if err != nil {
			yield(nil, err)
			return
		}

		response := &adkmodel.LLMResponse{
			Content:      genai.NewContentFromText(result.Text, genai.RoleModel),
			ModelVersion: result.Model,
			UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
				PromptTokenCount:     result.PromptTokens,
				CandidatesTokenCount: result.CompletionTokens,
				TotalTokenCount:      result.TotalTokens,
			},
			TurnComplete: true,
		}
		yield(response, nil)
	}
}

func flattenContentText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	parts := make([]string, 0, len(content.Parts))
	for _, part := range content.Parts {
		if part == nil || strings.TrimSpace(part.Text) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(part.Text))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
