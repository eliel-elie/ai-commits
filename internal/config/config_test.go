package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLayerPrecedence(t *testing.T) {
	homeDir := t.TempDir()
	repoDir := t.TempDir()

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_KEY", "")
	t.Setenv("AIC_GEMINI_KEY", "")

	// 1. Generic env acts as fallback when no config files define the key.
	t.Setenv("GEMINI_API_KEY", "generic-env-key")
	cfg, err := Load(repoDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiKey != "generic-env-key" {
		t.Fatalf("expected generic-env-key, got %q", cfg.GeminiKey)
	}
	if got := cfg.KeySourceFor("gemini"); got != "env:GEMINI_API_KEY" {
		t.Fatalf("expected source env:GEMINI_API_KEY, got %q", got)
	}

	// 2. Global ~/.ai-commits overrides generic env.
	globalPath := filepath.Join(homeDir, GlobalConfigName)
	if err := os.WriteFile(globalPath, []byte(`{"GEMINI_KEY":"global-key"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(repoDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiKey != "global-key" {
		t.Fatalf("expected global-key to override generic env, got %q", cfg.GeminiKey)
	}
	if got := cfg.KeySourceFor("gemini"); got != "global" {
		t.Fatalf("expected source global, got %q", got)
	}

	// 3. Local .aicommits.json overrides global ~/.ai-commits.
	localPath := filepath.Join(repoDir, LocalConfigName)
	if err := os.WriteFile(localPath, []byte(`{"GEMINI_KEY":"local-key"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(repoDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiKey != "local-key" {
		t.Fatalf("expected local-key to override global, got %q", cfg.GeminiKey)
	}
	if got := cfg.KeySourceFor("gemini"); got != "local" {
		t.Fatalf("expected source local, got %q", got)
	}

	// 4. Scoped AIC_GEMINI_KEY overrides local, global, and generic env.
	t.Setenv("AIC_GEMINI_KEY", "scoped-aic-key")
	cfg, err = Load(repoDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GeminiKey != "scoped-aic-key" {
		t.Fatalf("expected scoped-aic-key to override all files, got %q", cfg.GeminiKey)
	}
	if got := cfg.KeySourceFor("gemini"); got != "env:AIC_GEMINI_KEY" {
		t.Fatalf("expected source env:AIC_GEMINI_KEY, got %q", got)
	}
}
