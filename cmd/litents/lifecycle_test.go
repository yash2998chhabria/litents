package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/litents/litents/internal/gitx"
	"github.com/litents/litents/internal/tmux"
)

type fakeCommand struct {
	name     string
	args     []string
	out      string
	err      error
	dir      string
}

type fakeCommandRunner struct {
	steps  []fakeCommand
	next   int
	t      *testing.T
}

func newFakeCommandRunner(t *testing.T, steps []fakeCommand) *fakeCommandRunner {
	t.Helper()
	return &fakeCommandRunner{
		steps:  steps,
		t:      t,
	}
}

func (r *fakeCommandRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	r.t.Helper()

	if r.next >= len(r.steps) {
		r.t.Fatalf("unexpected command call %q %q (no matching expected step)", name, args)
	}
	want := r.steps[r.next]
	r.next++
	if want.dir != dir {
		r.t.Fatalf("command %d dir: got %q want %q", r.next, dir, want.dir)
	}
	if want.name != name {
		r.t.Fatalf("command %d name: got %q want %q", r.next, name, want.name)
	}
	if !reflect.DeepEqual(want.args, args) {
		r.t.Fatalf("command %d args: got %#v want %#v", r.next, args, want.args)
	}
	return want.out, want.err
}

func (r *fakeCommandRunner) LookPath(name string) (string, error) {
	r.t.Helper()
	return "/usr/bin/" + name, nil
}

func (r *fakeCommandRunner) assertConsumed() {
	r.t.Helper()
	if r.next != len(r.steps) {
		r.t.Fatalf("command runner consumed %d of %d expected steps", r.next, len(r.steps))
	}
}

func TestCommandLifecycle_UsesExpectedCommandSequence(t *testing.T) {
	t.Parallel()

	steps := []fakeCommand{
		{name: "git", args: []string{"-C", "/tmp/repo", "rev-parse", "--show-toplevel"}, out: "/tmp/repo\n", err: nil},
		{name: "git", args: []string{"-C", "/tmp/repo", "worktree", "add", "-B", "litents/auth", "/tmp/worktrees/myrepo/auth", "main"}, out: "", err: nil},
		{name: "git", args: []string{"-C", "/tmp/repo", "status", "--porcelain"}, out: "", err: nil},
		{name: "tmux", args: []string{"has-session", "-t", "litents-myrepo"}, out: "", err: errors.New("missing")},
		{name: "tmux", args: []string{"new-session", "-d", "-s", "litents-myrepo", "-c", "/tmp/repo"}, out: "", err: nil},
		{name: "tmux", args: []string{"new-window", "-t", "litents-myrepo", "-n", "auth", "-c", "/tmp/worktrees/myrepo/auth", "codex --prompt \"implement authentication\""}, out: "", err: nil},
		{name: "tmux", args: []string{"pipe-pane", "-o", "-t", "litents-myrepo:auth", "cat >> \"/tmp/state/myrepo/agents/auth/output.log\""}, out: "", err: nil},
		{name: "tmux", args: []string{"list-panes", "-t", "litents-myrepo:auth", "-F", "#{pane_id}"}, out: "%7\n%8", err: nil},
	}
	runner := newFakeCommandRunner(t, steps)
	gitClient := gitx.New(runner)
	tmuxClient := tmux.New(runner)

	ctx := context.Background()

	repoRoot, err := gitClient.RepoRoot(ctx, "/tmp/repo")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	if repoRoot != "/tmp/repo" {
		t.Fatalf("repo root: got %q want %q", repoRoot, "/tmp/repo")
	}

	worktree := gitx.DefaultWorktreePath("/tmp/worktrees", "myrepo", "auth")
	if err := gitClient.WorktreeAdd(ctx, repoRoot, "litents/auth", worktree, "main"); err != nil {
		t.Fatalf("worktree add: %v", err)
	}

	clean, err := gitClient.IsClean(ctx, repoRoot)
	if err != nil {
		t.Fatalf("git clean: %v", err)
	}
	if !clean {
		t.Fatalf("expected repo clean")
	}

	if err := tmuxClient.EnsureSession(ctx, "litents-myrepo", repoRoot); err != nil {
		t.Fatalf("ensure session: %v", err)
	}
	if err := tmuxClient.CreateWindow(ctx, "litents-myrepo", "auth", worktree, fmt.Sprintf("codex --prompt %q", "implement authentication")); err != nil {
		t.Fatalf("create window: %v", err)
	}
	if err := tmuxClient.StartWindowLogging(ctx, "litents-myrepo", "auth", "/tmp/state/myrepo/agents/auth/output.log"); err != nil {
		t.Fatalf("start logging: %v", err)
	}
	panes, err := tmuxClient.ListPanes(ctx, "litents-myrepo", "auth")
	if err != nil {
		t.Fatalf("list panes: %v", err)
	}
	if len(panes) != 2 || panes[0] != "%7" || panes[1] != "%8" {
		t.Fatalf("panes: got %#v", panes)
	}

	runner.assertConsumed()
}

func TestCommandLifecycle_StopsWhenRepoRootMissing(t *testing.T) {
	t.Parallel()

	steps := []fakeCommand{
		{name: "git", args: []string{"-C", "/tmp/missing", "rev-parse", "--show-toplevel"}, out: "", err: errors.New("not a git repo")},
	}
	runner := newFakeCommandRunner(t, steps)
	gitClient := gitx.New(runner)

	if _, err := gitClient.RepoRoot(context.Background(), "/tmp/missing"); err == nil {
		t.Fatalf("expected RepoRoot to fail")
	}

	runner.assertConsumed()
}
