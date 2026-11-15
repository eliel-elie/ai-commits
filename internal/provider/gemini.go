package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/eliel-elie/ai-commits/internal/prompt"
)

const defaultGeminiModel = "gemini-2.5-flash"

type geminiProvider struct {
	client    *http.Client
	apiKey    string
	model     string
	maxTokens int
}

func newGeminiProvider(client *http.Client, apiKey, model string, maxTokens int) *geminiProvider {
	if model == "" {
		model = defaultGeminiModel
	}
	return &geminiProvider{
		client:    client,
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
	}
}

func (g *geminiProvider) Name() string  { return "gemini" }
func (g *geminiProvider) Model() string { return g.model }

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (g *geminiProvider) GenerateCommitMessage(ctx context.Context, p prompt.Prompt) (string, error) {
	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: p.System}},
		},
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: p.User}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: g.maxTokens,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", url.PathEscape(g.model))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("gemini HTTP %d: invalid JSON response: %w", resp.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini API error (%d): %s", parsed.Error.Code, parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini HTTP %d: %s", resp.StatusCode, string(body))
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no commit message candidates")
	}

	return cleanMessage(parsed.Candidates[0].Content.Parts[0].Text), nil
}
