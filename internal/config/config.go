package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultProvider  = "gemini"
	DefaultLocale    = "en"
	DefaultMaxLength = 100
	DefaultMaxTokens = 2048
	LocalConfigName  = ".aicommits.json"
	GlobalConfigName = ".ai-commits"
)

// SupportedLocales maps CLI locale codes to language names used in prompts.
var SupportedLocales = map[string]string{
	"en":    "english",
	"es":    "spanish",
	"pt-BR": "portuguese",
}

// SupportedProviders lists built-in provider identifiers.
var SupportedProviders = []string{
	"gemini",
	"openai",
	"deepseek",
	"ollama",
	"groq",
	"openrouter",
}

// Config holds the resolved application configuration.
type Config struct {
	Provider  string   `json:"AI_PROVIDER,omitempty"`
	Model     string   `json:"model,omitempty"`
	Locale    string   `json:"locale,omitempty"`
	MaxLength int      `json:"-"`
	MaxTokens int      `json:"max_tokens,omitempty"`
	BaseURL   string   `json:"base_url,omitempty"`
	Exclude   []string `json:"exclude,omitempty"`

	GeminiKey   string `json:"GEMINI_KEY,omitempty"`
	OpenAIKey   string `json:"OPENAI_KEY,omitempty"`
	DeepSeekKey string `json:"DEEPSEEK_KEY,omitempty"`
	GroqKey     string `json:"GROQ_KEY,omitempty"`
	OpenRouterKey string `json:"OPENROUTER_KEY,omitempty"`
}

// rawConfig supports legacy Node.js values where MAX_LENGTH could be stored as string or number.
type rawConfig struct {
	Provider      string          `json:"AI_PROVIDER,omitempty"`
	Model         string          `json:"model,omitempty"`
	Locale        string          `json:"locale,omitempty"`
	MaxLength     json.RawMessage `json:"MAX_LENGTH,omitempty"`
	MaxTokens     int             `json:"max_tokens,omitempty"`
	BaseURL       string          `json:"base_url,omitempty"`
	Exclude       []string        `json:"exclude,omitempty"`
	GeminiKey     string          `json:"GEMINI_KEY,omitempty"`
	OpenAIKey     string          `json:"OPENAI_KEY,omitempty"`
	DeepSeekKey   string          `json:"DEEPSEEK_KEY,omitempty"`
	GroqKey       string          `json:"GROQ_KEY,omitempty"`
	OpenRouterKey string          `json:"OPENROUTER_KEY,omitempty"`
}

// GlobalConfigPath returns the path to ~/.ai-commits.
func GlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, GlobalConfigName), nil
}

// Load merges global (~/.ai-commits), local (<repoDir>/.aicommits.json), and environment variables.
func Load(repoDir string) (*Config, error) {
	cfg := &Config{
		Provider:  DefaultProvider,
		Locale:    DefaultLocale,
		MaxLength: DefaultMaxLength,
		MaxTokens: DefaultMaxTokens,
	}

	globalPath, err := GlobalConfigPath()
	if err == nil {
		if err := mergeFile(cfg, globalPath); err != nil {
			return nil, err
		}
	}

	if repoDir != "" {
		localPath := filepath.Join(repoDir, LocalConfigName)
		if err := mergeFile(cfg, localPath); err != nil {
			return nil, err
		}
	}

	applyEnv(cfg)
	return cfg, nil
}

func mergeFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}

	if raw.Provider != "" {
		cfg.Provider = raw.Provider
	}
	if raw.Model != "" {
		cfg.Model = raw.Model
	}
	if raw.Locale != "" {
		cfg.Locale = raw.Locale
	}
	if raw.MaxTokens > 0 {
		cfg.MaxTokens = raw.MaxTokens
	}
	if raw.BaseURL != "" {
		cfg.BaseURL = raw.BaseURL
	}
	if len(raw.Exclude) > 0 {
		cfg.Exclude = raw.Exclude
	}
	if raw.GeminiKey != "" {
		cfg.GeminiKey = raw.GeminiKey
	}
	if raw.OpenAIKey != "" {
		cfg.OpenAIKey = raw.OpenAIKey
	}
	if raw.DeepSeekKey != "" {
		cfg.DeepSeekKey = raw.DeepSeekKey
	}
	if raw.GroqKey != "" {
		cfg.GroqKey = raw.GroqKey
	}
	if raw.OpenRouterKey != "" {
		cfg.OpenRouterKey = raw.OpenRouterKey
	}

	if len(raw.MaxLength) > 0 {
		if parsed := parseFlexibleInt(raw.MaxLength); parsed > 0 {
			cfg.MaxLength = parsed
		}
	}

	return nil
}

func parseFlexibleInt(raw json.RawMessage) int {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if parsed, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return parsed
		}
	}
	return 0
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("AI_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("AI_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := firstNonEmpty(os.Getenv("GEMINI_API_KEY"), os.Getenv("GEMINI_KEY")); v != "" {
		cfg.GeminiKey = v
	}
	if v := firstNonEmpty(os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENAI_KEY")); v != "" {
		cfg.OpenAIKey = v
	}
	if v := firstNonEmpty(os.Getenv("DEEPSEEK_API_KEY"), os.Getenv("DEEPSEEK_KEY")); v != "" {
		cfg.DeepSeekKey = v
	}
	if v := firstNonEmpty(os.Getenv("GROQ_API_KEY"), os.Getenv("GROQ_KEY")); v != "" {
		cfg.GroqKey = v
	}
	if v := firstNonEmpty(os.Getenv("OPENROUTER_API_KEY"), os.Getenv("OPENROUTER_KEY")); v != "" {
		cfg.OpenRouterKey = v
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// APIKeyFor returns the configured API key for the given provider.
func (c *Config) APIKeyFor(provider string) string {
	switch strings.ToLower(provider) {
	case "gemini":
		return c.GeminiKey
	case "openai":
		return c.OpenAIKey
	case "deepseek":
		return c.DeepSeekKey
	case "groq":
		return c.GroqKey
	case "openrouter":
		return c.OpenRouterKey
	default:
		return ""
	}
}

// SetGlobal updates a key in ~/.ai-commits, preserving unknown fields.
func SetGlobal(key, value string) (string, error) {
	canonical, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	if err := validateValue(canonical, value); err != nil {
		return "", err
	}

	path, err := GlobalConfigPath()
	if err != nil {
		return "", err
	}

	dataMap := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(raw))) > 0 {
		if err := json.Unmarshal(raw, &dataMap); err != nil {
			return "", fmt.Errorf("parse existing config %s: %w", path, err)
		}
	}

	// Reset model when switching provider to avoid invalid model/provider combinations.
	if canonical == "AI_PROVIDER" {
		delete(dataMap, "model")
	}

	switch canonical {
	case "MAX_LENGTH", "max_tokens":
		n, _ := strconv.Atoi(value)
		dataMap[canonical] = n
	default:
		dataMap[canonical] = value
	}

	encoded, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write config %s: %w", path, err)
	}
	return canonical, nil
}

// GetGlobalRaw reads the raw map from ~/.ai-commits.
func GetGlobalRaw() (map[string]any, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	dataMap := map[string]any{}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return dataMap, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return dataMap, nil
	}
	if err := json.Unmarshal(raw, &dataMap); err != nil {
		return nil, err
	}
	return dataMap, nil
}

// MaskSecret masks an API key for safe display.
func MaskSecret(s string) string {
	if len(s) <= 8 {
		if s == "" {
			return ""
		}
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func normalizeKey(key string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "provider", "ai_provider":
		return "AI_PROVIDER", nil
	case "model":
		return "model", nil
	case "locale":
		return "locale", nil
	case "max-length", "max_length":
		return "MAX_LENGTH", nil
	case "max-tokens", "max_tokens":
		return "max_tokens", nil
	case "base-url", "base_url":
		return "base_url", nil
	case "gemini-key", "gemini_key":
		return "GEMINI_KEY", nil
	case "openai-key", "openai_key":
		return "OPENAI_KEY", nil
	case "deepseek-key", "deepseek_key":
		return "DEEPSEEK_KEY", nil
	case "groq-key", "groq_key":
		return "GROQ_KEY", nil
	case "openrouter-key", "openrouter_key":
		return "OPENROUTER_KEY", nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func validateValue(canonical, value string) error {
	switch canonical {
	case "locale":
		if _, ok := SupportedLocales[value]; !ok {
			return fmt.Errorf("invalid locale %q (supported: en, es, pt-BR)", value)
		}
	case "AI_PROVIDER":
		v := strings.ToLower(value)
		for _, p := range SupportedProviders {
			if p == v {
				return nil
			}
		}
		return fmt.Errorf("invalid provider %q (supported: %s)", value, strings.Join(SupportedProviders, ", "))
	case "MAX_LENGTH", "max_tokens":
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return fmt.Errorf("%s must be a positive integer", canonical)
		}
	}
	return nil
}
