package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eliel-elie/ai-commits/internal/config"
	"github.com/eliel-elie/ai-commits/internal/prompt"
)

const defaultHTTPTimeout = 60 * time.Second

// Provider generates commit messages from a structured prompt.
type Provider interface {
	Name() string
	Model() string
	GenerateCommitMessage(ctx context.Context, p prompt.Prompt) (string, error)
}

// New instantiates the configured AI provider.
func New(cfg *config.Config) (Provider, error) {
	name := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if name == "" {
		name = config.DefaultProvider
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = config.DefaultMaxTokens
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	apiKey := cfg.APIKeyFor(name)

	switch name {
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("missing API key for gemini")
		}
		return newGeminiProvider(client, apiKey, cfg.Model, maxTokens), nil
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("missing API key for openai")
		}
		return newOpenAICompatProvider(client, "openai", firstNonEmpty(cfg.BaseURL, "https://api.openai.com/v1"), apiKey, firstNonEmpty(cfg.Model, "gpt-5-mini"), maxTokens), nil
	case "deepseek":
		if apiKey == "" {
			return nil, fmt.Errorf("missing API key for deepseek")
		}
		return newOpenAICompatProvider(client, "deepseek", firstNonEmpty(cfg.BaseURL, "https://api.deepseek.com/v1"), apiKey, firstNonEmpty(cfg.Model, "deepseek-chat"), maxTokens), nil
	case "groq":
		if apiKey == "" {
			return nil, fmt.Errorf("missing API key for groq")
		}
		return newOpenAICompatProvider(client, "groq", firstNonEmpty(cfg.BaseURL, "https://api.groq.com/openai/v1"), apiKey, firstNonEmpty(cfg.Model, "llama-3.3-70b-versatile"), maxTokens), nil
	case "openrouter":
		if apiKey == "" {
			return nil, fmt.Errorf("missing API key for openrouter")
		}
		return newOpenAICompatProvider(client, "openrouter", firstNonEmpty(cfg.BaseURL, "https://openrouter.ai/api/v1"), apiKey, firstNonEmpty(cfg.Model, "openai/gpt-4o-mini"), maxTokens), nil
	case "ollama":
		return newOpenAICompatProvider(client, "ollama", firstNonEmpty(cfg.BaseURL, "http://localhost:11434/v1"), apiKey, firstNonEmpty(cfg.Model, "llama3.2"), maxTokens), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.Provider)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// cleanMessage trims whitespace, accidental markdown fences, and matching surrounding quotes.
func cleanMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if strings.HasPrefix(msg, "```") {
		lines := strings.Split(msg, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
			if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
				lines = lines[:len(lines)-1]
			}
			msg = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	for _, q := range []string{"\"", "'", "`"} {
		if len(msg) >= 2 && strings.HasPrefix(msg, q) && strings.HasSuffix(msg, q) {
			msg = strings.TrimSpace(msg[1 : len(msg)-1])
		}
	}
	return msg
}
