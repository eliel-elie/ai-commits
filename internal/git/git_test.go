package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	// Initial commit so HEAD exists.
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "chore: init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	return dir
}

func TestMultiFileAndLargeCommitMessage(t *testing.T) {
	dir := initTempRepo(t)

	// Create three files; we will stage and commit only a.txt and b.txt.
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content of "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", dir, err)
	}

	// Stage c.txt first, then commit only a.txt and b.txt via --file semantics.
	if err := repo.StageFiles([]string{"c.txt"}); err != nil {
		t.Fatalf("StageFiles(c.txt) failed: %v", err)
	}
	targetFiles := []string{"a.txt", "b.txt"}
	if err := repo.StageFiles(targetFiles); err != nil {
		t.Fatalf("StageFiles(a,b) failed: %v", err)
	}

	staged, err := repo.StagedFiles(targetFiles, nil)
	if err != nil {
		t.Fatalf("StagedFiles failed: %v", err)
	}
	if len(staged) != 2 {
		t.Fatalf("expected 2 staged files, got %v", staged)
	}

	// Build a large multi-line commit message (>10KB) with special shell characters.
	var sb strings.Builder
	sb.WriteString("feat(cli): support large messages, -C dir, and multiple --file flags\n\n")
	sb.WriteString("Details with special chars: `backticks`, \"double quotes\", 'single quotes', $HOME, %PATH%, & | ; !\n\n")
	for i := 0; i < 150; i++ {
		sb.WriteString("- Bullet item describing a detailed change across modules without truncation\n")
	}
	wantMsg := strings.TrimRight(sb.String(), "\n")

	if err := repo.Commit(wantMsg, CommitOptions{Files: targetFiles}); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify full commit message matches byte-for-byte (ignoring trailing newline added by git).
	cmd := exec.Command("git", "log", "-1", "--format=%B")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	gotMsg := strings.TrimRight(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	if gotMsg != wantMsg {
		t.Fatalf("commit message mismatch:\ngot len=%d\nwant len=%d", len(gotMsg), len(wantMsg))
	}

	// Verify c.txt is still staged and was NOT included in the commit.
	remaining, err := repo.StagedFiles(nil, nil)
	if err != nil {
		t.Fatalf("StagedFiles(nil) failed: %v", err)
	}
	if len(remaining) != 1 || remaining[0] != "c.txt" {
		t.Fatalf("expected c.txt to remain staged, got %v", remaining)
	}
}
