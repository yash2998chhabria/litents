package tmux

import (
	"context"
	"errors"
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

func (c *Client) HasSession(ctx context.Context, session string) bool {
	_, err := c.Runner.Run(ctx, "", "tmux", "has-session", "-t", session)
	return err == nil
}

func (c *Client) EnsureSession(ctx context.Context, session, workdir string) error {
	if c.HasSession(ctx, session) {
		return nil
	}
	_, err := c.Runner.Run(ctx, "", "tmux", "new-session", "-d", "-s", session, "-c", workdir)
	return err
}

func (c *Client) CreateWindow(ctx context.Context, session, name, workdir, command string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "new-window", "-t", session, "-n", name, "-c", workdir, command)
	return err
}

func (c *Client) WindowExists(ctx context.Context, session, window string) bool {
	target := fmt.Sprintf("%s:%s", session, window)
	_, err := c.Runner.Run(ctx, "", "tmux", "list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return false
	}
	_, err = c.Runner.Run(ctx, "", "tmux", "list-windows", "-t", target, "-F", "#{window_name}")
	return err == nil
}

func (c *Client) ListWindows(ctx context.Context, session string) ([]string, error) {
	out, err := c.Runner.Run(ctx, "", "tmux", "list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return nil, err
	}
	lines := filterEmpty(strings.Split(strings.TrimSpace(out), "\n"))
	return lines, nil
}

func (c *Client) ListPanes(ctx context.Context, session, window string) ([]string, error) {
	target := fmt.Sprintf("%s:%s", session, window)
	out, err := c.Runner.Run(ctx, "", "tmux", "list-panes", "-t", target, "-F", "#{pane_id}")
	if err != nil {
		return nil, err
	}
	return filterEmpty(strings.Split(strings.TrimSpace(out), "\n")), nil
}

func (c *Client) PaneID(ctx context.Context, session, window string) (string, error) {
	panes, err := c.ListPanes(ctx, session, window)
	if err != nil {
		return "", err
	}
	if len(panes) == 0 {
		return "", errors.New("no panes")
	}
	return panes[0], nil
}

func (c *Client) StartWindowLogging(ctx context.Context, session, window, logfile string) error {
	target := fmt.Sprintf("%s:%s", session, window)
	cmd := fmt.Sprintf("cat >> %q", filepath.Clean(logfile))
	_, err := c.Runner.Run(ctx, "", "tmux", "pipe-pane", "-o", "-t", target, cmd)
	return err
}

func (c *Client) CapturePane(ctx context.Context, session, window string, lastLines int) (string, error) {
	target := fmt.Sprintf("%s:%s", session, window)
	out, err := c.Runner.Run(ctx, "", "tmux", "capture-pane", "-p", "-t", target)
	if err != nil {
		return "", err
	}
	all := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if lastLines <= 0 || lastLines >= len(all) {
		return strings.Join(all, "\n"), nil
	}
	return strings.Join(all[len(all)-lastLines:], "\n"), nil
}

func (c *Client) Attach(ctx context.Context, session string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "attach", "-t", session)
	return err
}

func (c *Client) SwitchClient(ctx context.Context, session string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "switch-client", "-t", session)
	return err
}

func (c *Client) SelectWindow(ctx context.Context, session, window string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "select-window", "-t", fmt.Sprintf("%s:%s", session, window))
	return err
}

func (c *Client) SendKeys(ctx context.Context, session, window, text string, enter bool) error {
	target := fmt.Sprintf("%s:%s", session, window)
	args := []string{"send-keys", "-t", target, text}
	if enter {
		args = append(args, "Enter")
	}
	_, err := c.Runner.Run(ctx, "", "tmux", args...)
	return err
}

func (c *Client) KillWindow(ctx context.Context, session, window string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "kill-window", "-t", fmt.Sprintf("%s:%s", session, window))
	return err
}

func (c *Client) KillPane(ctx context.Context, session, window string) error {
	_, err := c.Runner.Run(ctx, "", "tmux", "kill-pane", "-t", fmt.Sprintf("%s:%s", session, window))
	return err
}

func filterEmpty(in []string) []string {
	result := make([]string, 0, len(in))
	for _, item := range in {
		if strings.TrimSpace(item) == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
