package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eliel-elie/ai-commits/internal/prompt"
)

type openAICompatProvider struct {
	client    *http.Client
	name      string
	baseURL   string
	apiKey    string
	model     string
	maxTokens int
}

func newOpenAICompatProvider(client *http.Client, name, baseURL, apiKey, model string, maxTokens int) *openAICompatProvider {
	return &openAICompatProvider{
		client:    client,
		name:      name,
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
	}
}

func (o *openAICompatProvider) Name() string  { return o.name }
func (o *openAICompatProvider) Model() string { return o.model }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []chatMessage `json:"messages"`
	Temperature         *float64      `json:"temperature,omitempty"`
	MaxTokens           *int          `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (o *openAICompatProvider) GenerateCommitMessage(ctx context.Context, p prompt.Prompt) (string, error) {
	lowerModel := strings.ToLower(strings.TrimSpace(o.model))
	isReasoningModel := o.name == "openai" && (strings.HasPrefix(lowerModel, "o1") || strings.HasPrefix(lowerModel, "o3") || strings.HasPrefix(lowerModel, "gpt-5"))

	var messages []chatMessage
	if isReasoningModel {
		messages = []chatMessage{
			{Role: "user", Content: p.System + "\n\n" + p.User},
		}
	} else {
		messages = []chatMessage{
			{Role: "system", Content: p.System},
			{Role: "user", Content: p.User},
		}
	}

	reqBody := chatCompletionRequest{
		Model:    o.model,
		Messages: messages,
	}
	if isReasoningModel {
		tokens := o.maxTokens
		if tokens < 4096 {
			tokens = 4096
		}
		reqBody.MaxCompletionTokens = &tokens
	} else {
		temp := 0.7
		tokens := o.maxTokens
		reqBody.Temperature = &temp
		reqBody.MaxTokens = &tokens
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal %s request: %w", o.name, err)
	}

	endpoint := o.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create %s request: %w", o.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s request failed: %w", o.name, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s response: %w", o.name, err)
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("%s HTTP %d: invalid JSON response: %w", o.name, resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("%s API error: %s", o.name, parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s HTTP %d: %s", o.name, resp.StatusCode, string(body))
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("%s returned no commit message", o.name)
	}

	return cleanMessage(parsed.Choices[0].Message.Content), nil
}
