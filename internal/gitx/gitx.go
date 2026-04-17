package gitx

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/litents/litents/internal/runner"
)

type Client struct {
	Runner runner.Runner
}

func New(r runner.Runner) *Client {
	return &Client{Runner: r}
}

func (c *Client) RepoRoot(ctx context.Context, path string) (string, error) {
	out, err := c.Runner.Run(ctx, "", "git", "-C", path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (c *Client) WorktreeAdd(ctx context.Context, repoRoot, branchName, worktreePath, baseBranch string) error {
	if _, err := c.Runner.Run(ctx, "", "git", "-C", repoRoot, "worktree", "add", "-B", branchName, worktreePath, baseBranch); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}
	return nil
}

func (c *Client) WorktreeRemove(ctx context.Context, repoRoot, worktreePath string, force bool) error {
	args := []string{"-C", repoRoot, "worktree", "remove", worktreePath}
	if force {
		args = append(args, "--force")
	}
	if _, err := c.Runner.Run(ctx, "", "git", args...); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	return nil
}

func (c *Client) IsClean(ctx context.Context, path string) (bool, error) {
	out, err := c.Runner.Run(ctx, "", "git", "-C", path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("check git clean: %w", err)
	}
	return strings.TrimSpace(out) == "", nil
}

func (c *Client) IsBranchMerged(ctx context.Context, repoRoot, branch, base string) (bool, error) {
	_, err := c.Runner.Run(ctx, "", "git", "-C", repoRoot, "merge-base", "--is-ancestor", branch, base)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) IsRepository(ctx context.Context, path string) (bool, error) {
	_, err := c.Runner.Run(ctx, "", "git", "-C", path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func DefaultWorktreePath(root, projectName, agentID string) string {
	return filepath.Join(root, projectName, agentID)
}
