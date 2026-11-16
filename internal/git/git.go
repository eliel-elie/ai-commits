package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultExcludes are noisy lock/generated files skipped from diffs unless explicitly requested via --file.
var DefaultExcludes = []string{
	"package-lock.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	"go.sum",
	"Cargo.lock",
	"composer.lock",
	"Gemfile.lock",
	"poetry.lock",
}

// Repo represents a validated Git repository scoped to a working directory.
type Repo struct {
	WorkDir string
	Root    string
}

// CommitOptions configures how a commit is executed.
type CommitOptions struct {
	NoVerify bool
	Files    []string
}

// Open validates that dir exists and is inside a Git repository.
func Open(dir string) (*Repo, error) {
	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve directory %q: %w", dir, err)
	}
	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("access directory %q: %w", absDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", absDir)
	}

	out, err := runGit(absDir, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("%q is not a git repository: %w", absDir, err)
	}

	return &Repo{
		WorkDir: absDir,
		Root:    strings.TrimSpace(out),
	}, nil
}

// StageAll stages all changes within WorkDir.
func (r *Repo) StageAll() error {
	_, err := runGit(r.WorkDir, nil, "add", "-A", "--", ".")
	return err
}

// UnstagedFiles returns modified, deleted, or untracked file paths within WorkDir.
func (r *Repo) UnstagedFiles() ([]string, error) {
	out, err := runGit(r.WorkDir, nil, "-c", "core.quotePath=false", "status", "--porcelain", "-uall", "--", ".")
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path != "" {
			result = append(result, path)
		}
	}
	return result, nil
}

// StageFiles stages the specified files relative to WorkDir.
func (r *Repo) StageFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, files...)
	_, err := runGit(r.WorkDir, nil, args...)
	return err
}

// StagedFiles returns the list of staged file paths.
func (r *Repo) StagedFiles(files []string, excludes []string) ([]string, error) {
	args := buildDiffArgs([]string{"diff", "--staged", "--name-only"}, files, excludes)
	out, err := runGit(r.WorkDir, nil, args...)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

// StagedDiff returns the staged unified diff.
func (r *Repo) StagedDiff(files []string, excludes []string) (string, error) {
	args := buildDiffArgs([]string{"diff", "--staged"}, files, excludes)
	return runGit(r.WorkDir, nil, args...)
}

// StagedStat returns `git diff --staged --stat` for summary context.
func (r *Repo) StagedStat(files []string, excludes []string) (string, error) {
	args := buildDiffArgs([]string{"diff", "--staged", "--stat"}, files, excludes)
	return runGit(r.WorkDir, nil, args...)
}

// Commit creates a commit reading the full message from stdin (-F -), avoiding shell escaping/length limits.
func (r *Repo) Commit(message string, opts CommitOptions) error {
	args := []string{"commit", "-F", "-"}
	if opts.NoVerify {
		args = append(args, "--no-verify")
	}
	if len(opts.Files) > 0 {
		args = append(args, "--")
		args = append(args, opts.Files...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.WorkDir
	cmd.Stdin = strings.NewReader(message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

func buildDiffArgs(base []string, files []string, excludes []string) []string {
	args := append([]string{}, base...)
	if len(files) > 0 {
		args = append(args, "--")
		args = append(args, files...)
		return args
	}
	if len(excludes) > 0 {
		args = append(args, "--", ".")
		for _, pattern := range excludes {
			pattern = strings.TrimSpace(pattern)
			if pattern != "" {
				args = append(args, ":(exclude)"+pattern)
			}
		}
	}
	return args
}

func runGit(dir string, stdin []byte, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], errMsg)
	}
	return stdout.String(), nil
}
