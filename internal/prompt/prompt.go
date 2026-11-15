package prompt

import (
	"fmt"
	"strings"

	"github.com/eliel-elie/ai-commits/internal/config"
)

const DefaultMaxDiffBytes = 60_000

// Prompt holds separated system and user instructions for AI providers.
type Prompt struct {
	System    string
	User      string
	Truncated bool
}

// Options configures commit prompt generation.
type Options struct {
	Locale       string
	MaxLength    int
	Hint         string
	Stat         string
	Diff         string
	MaxDiffBytes int
}

// Build constructs a structured System + User prompt from the staged diff.
func Build(opts Options) Prompt {
	langName, ok := config.SupportedLocales[opts.Locale]
	if !ok {
		langName = config.SupportedLocales[config.DefaultLocale]
	}

	maxLen := opts.MaxLength
	if maxLen <= 0 {
		maxLen = config.DefaultMaxLength
	}

	maxDiff := opts.MaxDiffBytes
	if maxDiff <= 0 {
		maxDiff = DefaultMaxDiffBytes
	}

	diff := opts.Diff
	truncated := false
	if len(diff) > maxDiff {
		diff = diff[:maxDiff] + "\n\n[diff truncated due to size limit]"
		truncated = true
	}

	system := fmt.Sprintf(`You are an expert software engineer writing git commit messages.
Generate a Conventional Commits message written in present tense.
Language: %s
Target subject line length: up to %d characters (add a concise bulleted body if multiple changes warrant it).

Commit Types:
- feat: A new feature
- fix: A bug fix
- docs: Documentation only changes
- style: Formatting, missing semi-colons, white-space, etc.
- refactor: Code change that neither fixes a bug nor adds a feature
- perf: A code change that improves performance
- test: Adding or correcting tests
- build: Build system or external dependencies
- ci: CI configuration files and scripts
- chore: Other changes that do not modify src or test files

Rules:
- Write in present tense
- Be descriptive yet concise
- Output ONLY the raw commit message (no markdown code fences, no surrounding quotes, no commentary)`, langName, maxLen)

	var user strings.Builder
	if strings.TrimSpace(opts.Hint) != "" {
		user.WriteString("Additional user context/hint:\n")
		user.WriteString(strings.TrimSpace(opts.Hint))
		user.WriteString("\n\n")
	}
	if strings.TrimSpace(opts.Stat) != "" {
		user.WriteString("Diff summary (--stat):\n")
		user.WriteString(strings.TrimSpace(opts.Stat))
		user.WriteString("\n\n")
	}
	user.WriteString("Analyze the following code diff and generate the commit message:\n")
	user.WriteString(diff)

	return Prompt{
		System:    system,
		User:      user.String(),
		Truncated: truncated,
	}
}
