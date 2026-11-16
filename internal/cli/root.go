package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eliel-elie/ai-commits/internal/config"
	"github.com/eliel-elie/ai-commits/internal/git"
	"github.com/eliel-elie/ai-commits/internal/i18n"
	"github.com/eliel-elie/ai-commits/internal/prompt"
	"github.com/eliel-elie/ai-commits/internal/provider"
	"github.com/eliel-elie/ai-commits/internal/ui"
	"github.com/eliel-elie/ai-commits/internal/updater"
	"github.com/spf13/cobra"
)

// Version is overridden at build time via -ldflags when creating releases.
var Version = "1.0.0"

type rootOptions struct {
	dir       string
	files     []string
	all       bool
	dry       bool
	yes       bool
	noVerify  bool
	hint      string
	provider  string
	model     string
	locale    string
	maxLength int
	excludes  []string
}

// NewRootCmd builds the root `aic` Cobra command.
func NewRootCmd() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:           "aic",
		Short:         "Generate git commit messages using AI",
		Version:       Version,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if len(opts.files) == 0 {
					return fmt.Errorf("unexpected arguments %v (use -f/--file to specify files)", args)
				}
				opts.files = append(opts.files, args...)
			}
			return runCommitFlow(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.dir, "dir", "C", ".", "Target git repository directory")
	f.StringArrayVarP(&opts.files, "file", "f", nil, "Specific file(s) to stage and commit (e.g. -f file1.go file2.go)")
	f.BoolVarP(&opts.all, "all", "a", false, "Stage all changes before generating the commit message")
	f.BoolVarP(&opts.dry, "dry", "d", false, "Generate and print the commit message without committing")
	f.BoolVarP(&opts.yes, "yes", "y", false, "Skip interactive confirmation and commit immediately")
	f.BoolVar(&opts.noVerify, "no-verify", false, "Bypass git pre-commit and commit-msg hooks")
	f.StringVar(&opts.hint, "hint", "", "Extra context or instructions for the AI")
	f.StringVarP(&opts.provider, "provider", "p", "", "Override AI provider (gemini, openai, deepseek, groq, openrouter, ollama)")
	f.StringVarP(&opts.model, "model", "m", "", "Override AI model")
	f.StringVarP(&opts.locale, "locale", "l", "", "Override locale (en, pt-BR, es)")
	f.IntVar(&opts.maxLength, "max-length", 0, "Override target subject line length")
	f.StringSliceVar(&opts.excludes, "exclude", nil, "Additional file patterns to exclude from diff")

	cmd.MarkFlagsMutuallyExclusive("all", "file")
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newUpdateCmd())

	return cmd
}

func runCommitFlow(ctx context.Context, opts *rootOptions) error {
	updater.CleanupOldBinary()
	updateCh := updater.StartBackgroundCheck(ctx, Version)
	defer func() {
		select {
		case newerTag, ok := <-updateCh:
			if ok && newerTag != "" {
				ui.PrintWarn(fmt.Sprintf("Update available: v%s -> %s (run \"aic update\")", Version, newerTag))
			}
		default:
		}
	}()

	repo, err := git.Open(opts.dir)
	if err != nil {
		return err
	}

	cfg, err := config.Load(repo.Root)
	if err != nil {
		return err
	}
	applyFlagOverrides(cfg, opts)

	msgBundle := i18n.Get(cfg.Locale)

	prov, err := provider.New(cfg)
	if err != nil {
		return fmt.Errorf(msgBundle.MissingAPIKey+" (%w)", cfg.Provider, cfg.Provider, err)
	}

	ui.PrintHeader(msgBundle.Intro, msgBundle.ProviderLabel, prov.Name(), prov.Model(), repo.WorkDir)

	if len(opts.files) > 0 {
		ui.PrintInfo(fmt.Sprintf(msgBundle.StagingFiles, len(opts.files)))
		if err := repo.StageFiles(opts.files); err != nil {
			return err
		}
	} else if opts.all {
		ui.PrintInfo(msgBundle.StagingAll)
		if err := repo.StageAll(); err != nil {
			return err
		}
	}

	excludes := mergeExcludes(cfg.Exclude, opts.excludes)

	stagedFiles, diff, stat, err := detectStaged(msgBundle.DetectingFiles, repo, opts.files, excludes)
	if err != nil {
		return err
	}

	if len(stagedFiles) == 0 || strings.TrimSpace(diff) == "" {
		if len(opts.files) == 0 && !opts.all {
			if unstaged, uErr := repo.UnstagedFiles(); uErr == nil && len(unstaged) > 0 {
				ui.PrintUnstagedFiles(msgBundle.UnstagedFound(len(unstaged)), unstaged)
				if opts.yes || opts.dry {
					ui.PrintWarn(msgBundle.NoChanges)
					return nil
				}
				ok, cErr := ui.ConfirmStage(msgBundle.ConfirmStageAll)
				if cErr != nil {
					return cErr
				}
				if !ok {
					ui.PrintWarn(msgBundle.CommitCancelled)
					return nil
				}
				ui.PrintInfo(msgBundle.StagingAll)
				if err := repo.StageAll(); err != nil {
					return err
				}
				stagedFiles, diff, stat, err = detectStaged(msgBundle.DetectingFiles, repo, opts.files, excludes)
				if err != nil {
					return err
				}
			}
		}
		if len(stagedFiles) == 0 || strings.TrimSpace(diff) == "" {
			ui.PrintWarn(msgBundle.NoChanges)
			return nil
		}
	}

	ui.PrintFiles(msgBundle.FilesDetected(len(stagedFiles)), stagedFiles)

	currentHint := opts.hint
	for {
		p := prompt.Build(prompt.Options{
			Locale:    cfg.Locale,
			MaxLength: cfg.MaxLength,
			Hint:      currentHint,
			Stat:      stat,
			Diff:      diff,
		})
		if p.Truncated {
			ui.PrintWarn(fmt.Sprintf(msgBundle.TruncatedDiff, prompt.DefaultMaxDiffBytes))
		}

		var commitMsg string
		err = ui.RunWithSpinner(msgBundle.Analyzing, func() error {
			var genErr error
			commitMsg, genErr = prov.GenerateCommitMessage(ctx, p)
			return genErr
		})
		if err != nil {
			return err
		}

		if opts.dry {
			ui.PrintSuccess(msgBundle.DryRunHeader)
			ui.PrintMessageBox(commitMsg)
			return nil
		}

		if opts.yes {
			ui.PrintMessageBox(commitMsg)
			if err := repo.Commit(commitMsg, git.CommitOptions{
				NoVerify: opts.noVerify,
				Files:    opts.files,
			}); err != nil {
				return err
			}
			ui.PrintSuccess(msgBundle.CommitSuccess)
			return nil
		}

		ui.PrintMessageBox(commitMsg)
		action, err := ui.ChooseAction(msgBundle)
		if err != nil {
			return err
		}

		switch action {
		case ui.ActionConfirm:
			if err := repo.Commit(commitMsg, git.CommitOptions{
				NoVerify: opts.noVerify,
				Files:    opts.files,
			}); err != nil {
				return err
			}
			ui.PrintSuccess(msgBundle.CommitSuccess)
			return nil

		case ui.ActionEdit:
			edited, ok, err := ui.EditMessage(msgBundle, commitMsg)
			if err != nil {
				return err
			}
			if !ok || strings.TrimSpace(edited) == "" {
				ui.PrintWarn(msgBundle.CommitCancelled)
				return nil
			}
			if err := repo.Commit(strings.TrimSpace(edited), git.CommitOptions{
				NoVerify: opts.noVerify,
				Files:    opts.files,
			}); err != nil {
				return err
			}
			ui.PrintSuccess(msgBundle.CommitSuccess)
			return nil

		case ui.ActionRegenerate:
			newHint, ok, err := ui.AskHint(msgBundle)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarn(msgBundle.CommitCancelled)
				return nil
			}
			if strings.TrimSpace(newHint) != "" {
				currentHint = strings.TrimSpace(newHint)
			}
			continue

		case ui.ActionCancel:
			ui.PrintWarn(msgBundle.CommitCancelled)
			return nil
		}
	}
}

func applyFlagOverrides(cfg *config.Config, opts *rootOptions) {
	if opts.provider != "" {
		cfg.Provider = opts.provider
		if opts.model == "" {
			cfg.Model = ""
		}
	}
	if opts.model != "" {
		cfg.Model = opts.model
	}
	if opts.locale != "" {
		cfg.Locale = opts.locale
	}
	if opts.maxLength > 0 {
		cfg.MaxLength = opts.maxLength
	}
}

func mergeExcludes(cfgExcludes, flagExcludes []string) []string {
	merged := append([]string{}, git.DefaultExcludes...)
	merged = append(merged, cfgExcludes...)
	merged = append(merged, flagExcludes...)
	return merged
}

func detectStaged(label string, repo *git.Repo, files, excludes []string) ([]string, string, string, error) {
	var stagedFiles []string
	var diff, stat string
	err := ui.RunWithSpinner(label, func() error {
		var e error
		stagedFiles, e = repo.StagedFiles(files, excludes)
		if e != nil {
			return e
		}
		if len(stagedFiles) == 0 {
			return nil
		}
		diff, e = repo.StagedDiff(files, excludes)
		if e != nil {
			return e
		}
		stat, _ = repo.StagedStat(files, excludes)
		return nil
	})
	return stagedFiles, diff, stat, err
}
