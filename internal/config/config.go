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

	GeminiKey     string `json:"GEMINI_KEY,omitempty"`
	OpenAIKey     string `json:"OPENAI_KEY,omitempty"`
	DeepSeekKey   string `json:"DEEPSEEK_KEY,omitempty"`
	GroqKey       string `json:"GROQ_KEY,omitempty"`
	OpenRouterKey string `json:"OPENROUTER_KEY,omitempty"`

	Sources map[string]string `json:"-"`
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

// Load merges defaults, generic env fallbacks, global (~/.ai-commits), local (<repoDir>/.aicommits.json), and AIC_* env overrides.
func Load(repoDir string) (*Config, error) {
	cfg := &Config{
		Provider:  DefaultProvider,
		Locale:    DefaultLocale,
		MaxLength: DefaultMaxLength,
		MaxTokens: DefaultMaxTokens,
		Sources: map[string]string{
			"AI_PROVIDER": "default",
			"locale":      "default",
			"MAX_LENGTH":  "default",
			"max_tokens":  "default",
		},
	}

	applyFallbackEnv(cfg)

	globalPath, err := GlobalConfigPath()
	if err == nil {
		if err := mergeFile(cfg, globalPath, "global"); err != nil {
			return nil, err
		}
	}

	if repoDir != "" {
		localPath := filepath.Join(repoDir, LocalConfigName)
		if err := mergeFile(cfg, localPath, "local"); err != nil {
			return nil, err
		}
	}

	applyScopedEnv(cfg)
	return cfg, nil
}

func mergeFile(cfg *Config, path, source string) error {
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
		cfg.Sources["AI_PROVIDER"] = source
	}
	if raw.Model != "" {
		cfg.Model = raw.Model
		cfg.Sources["model"] = source
	}
	if raw.Locale != "" {
		cfg.Locale = raw.Locale
		cfg.Sources["locale"] = source
	}
	if raw.MaxTokens > 0 {
		cfg.MaxTokens = raw.MaxTokens
		cfg.Sources["max_tokens"] = source
	}
	if raw.BaseURL != "" {
		cfg.BaseURL = raw.BaseURL
		cfg.Sources["base_url"] = source
	}
	if len(raw.Exclude) > 0 {
		cfg.Exclude = raw.Exclude
		cfg.Sources["exclude"] = source
	}
	if v := cleanSecret(raw.GeminiKey); v != "" {
		cfg.GeminiKey = v
		cfg.Sources["GEMINI_KEY"] = source
	}
	if v := cleanSecret(raw.OpenAIKey); v != "" {
		cfg.OpenAIKey = v
		cfg.Sources["OPENAI_KEY"] = source
	}
	if v := cleanSecret(raw.DeepSeekKey); v != "" {
		cfg.DeepSeekKey = v
		cfg.Sources["DEEPSEEK_KEY"] = source
	}
	if v := cleanSecret(raw.GroqKey); v != "" {
		cfg.GroqKey = v
		cfg.Sources["GROQ_KEY"] = source
	}
	if v := cleanSecret(raw.OpenRouterKey); v != "" {
		cfg.OpenRouterKey = v
		cfg.Sources["OPENROUTER_KEY"] = source
	}

	if len(raw.MaxLength) > 0 {
		if parsed := parseFlexibleInt(raw.MaxLength); parsed > 0 {
			cfg.MaxLength = parsed
			cfg.Sources["MAX_LENGTH"] = source
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

func applyFallbackEnv(cfg *Config) {
	if v, name := firstEnv("GEMINI_API_KEY", "GEMINI_KEY"); v != "" {
		cfg.GeminiKey = v
		cfg.Sources["GEMINI_KEY"] = "env:" + name
	}
	if v, name := firstEnv("OPENAI_API_KEY", "OPENAI_KEY"); v != "" {
		cfg.OpenAIKey = v
		cfg.Sources["OPENAI_KEY"] = "env:" + name
	}
	if v, name := firstEnv("DEEPSEEK_API_KEY", "DEEPSEEK_KEY"); v != "" {
		cfg.DeepSeekKey = v
		cfg.Sources["DEEPSEEK_KEY"] = "env:" + name
	}
	if v, name := firstEnv("GROQ_API_KEY", "GROQ_KEY"); v != "" {
		cfg.GroqKey = v
		cfg.Sources["GROQ_KEY"] = "env:" + name
	}
	if v, name := firstEnv("OPENROUTER_API_KEY", "OPENROUTER_KEY"); v != "" {
		cfg.OpenRouterKey = v
		cfg.Sources["OPENROUTER_KEY"] = "env:" + name
	}
}

func applyScopedEnv(cfg *Config) {
	if v, name := firstEnv("AIC_PROVIDER", "AI_PROVIDER"); v != "" {
		cfg.Provider = v
		cfg.Sources["AI_PROVIDER"] = "env:" + name
	}
	if v, name := firstEnv("AIC_MODEL", "AI_MODEL"); v != "" {
		cfg.Model = v
		cfg.Sources["model"] = "env:" + name
	}
	if v, name := firstEnv("AIC_LOCALE"); v != "" {
		cfg.Locale = v
		cfg.Sources["locale"] = "env:" + name
	}
	if v, name := firstEnv("AIC_GEMINI_KEY"); v != "" {
		cfg.GeminiKey = v
		cfg.Sources["GEMINI_KEY"] = "env:" + name
	}
	if v, name := firstEnv("AIC_OPENAI_KEY"); v != "" {
		cfg.OpenAIKey = v
		cfg.Sources["OPENAI_KEY"] = "env:" + name
	}
	if v, name := firstEnv("AIC_DEEPSEEK_KEY"); v != "" {
		cfg.DeepSeekKey = v
		cfg.Sources["DEEPSEEK_KEY"] = "env:" + name
	}
	if v, name := firstEnv("AIC_GROQ_KEY"); v != "" {
		cfg.GroqKey = v
		cfg.Sources["GROQ_KEY"] = "env:" + name
	}
	if v, name := firstEnv("AIC_OPENROUTER_KEY"); v != "" {
		cfg.OpenRouterKey = v
		cfg.Sources["OPENROUTER_KEY"] = "env:" + name
	}
}

func firstEnv(names ...string) (string, string) {
	for _, name := range names {
		if v := cleanSecret(os.Getenv(name)); v != "" {
			return v, name
		}
	}
	return "", ""
}

func cleanSecret(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, `"'`)
	return strings.TrimSpace(v)
}

// SourceFor returns the origin of a canonical config key (e.g. "global", "local", "env:GEMINI_API_KEY", "default").
func (c *Config) SourceFor(canonical string) string {
	if c.Sources == nil {
		return ""
	}
	return c.Sources[canonical]
}

// KeySourceFor returns the origin of the API key for the given provider.
func (c *Config) KeySourceFor(provider string) string {
	switch strings.ToLower(provider) {
	case "gemini":
		return c.SourceFor("GEMINI_KEY")
	case "openai":
		return c.SourceFor("OPENAI_KEY")
	case "deepseek":
		return c.SourceFor("DEEPSEEK_KEY")
	case "groq":
		return c.SourceFor("GROQ_KEY")
	case "openrouter":
		return c.SourceFor("OPENROUTER_KEY")
	default:
		return ""
	}
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

func DefaultModelFor(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "gemini", "":
		return "gemini-3.1-flash-lite"
	case "openai":
		return "gpt-5-mini"
	case "deepseek":
		return "deepseek-chat"
	case "groq":
		return "llama-3.3-70b-versatile"
	case "openrouter":
		return "openai/gpt-4o-mini"
	case "ollama":
		return "llama3.2"
	default:
		return ""
	}
}

func NormalizeKey(key string) (string, error) {
	return normalizeKey(key)
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
