package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/litents/litents/internal/config"
	"github.com/litents/litents/internal/runner"
)

type TemplateData struct {
	Project  string
	Agent    string
	Status   string
	Message  string
	Worktree string
	LogFile  string
}

type Notifier struct {
	Runner runner.Runner
}

func New(r runner.Runner) *Notifier {
	return &Notifier{Runner: r}
}

func (n *Notifier) Send(ctx context.Context, cfg *config.Config, data TemplateData) error {
	if !cfg.NotifyEnabled {
		return nil
	}

	normalized := TemplateData{
		Project:  data.Project,
		Agent:    data.Agent,
		Status:   data.Status,
		Message:  data.Message,
		Worktree: data.Worktree,
		LogFile:  data.LogFile,
	}

	if normalized.Message == "" {
		normalized.Message = "Litents agent needs attention"
	}

	message := replaceTemplate(data.Message, normalized)
	body := fmt.Sprintf("%s in %s", normalized.Status, message)
	switch cfg.NotifyCommand {
	case "auto", "":
		if n.notifySendCommand(ctx, message) {
			return nil
		}
		if n.terminalNotifierCommand(ctx, normalized) {
			return nil
		}
		return n.osascriptCommand(ctx, normalized)
	default:
		cmd := renderTemplate(cfg.NotifyCommand, map[string]string{
			"{{project}}":   normalized.Project,
			"{{agent}}":     normalized.Agent,
			"{{status}}":    normalized.Status,
			"{{message}}":   message,
			"{{worktree}}":  normalized.Worktree,
			"{{log_file}}":  normalized.LogFile,
			"{{body}}":      body,
		})
		_, err := n.Runner.Run(ctx, "", "sh", "-c", cmd)
		return err
	}
}

func (n *Notifier) notifySendCommand(ctx context.Context, message string) bool {
	if _, err := n.Runner.LookPath("notify-send"); err != nil {
		return false
	}
	cmd := fmt.Sprintf("notify-send %q %q", "Litents", message)
	_, _ = n.Runner.Run(ctx, "", "sh", "-c", cmd)
	return true
}

func (n *Notifier) terminalNotifierCommand(ctx context.Context, data TemplateData) bool {
	if _, err := n.Runner.LookPath("terminal-notifier"); err != nil {
		return false
	}
	title := fmt.Sprintf("Litents: %s", data.Agent)
	cmd := fmt.Sprintf("terminal-notifier -title %q -subtitle %q -message %q", "Litents", title, data.Message)
	_, _ = n.Runner.Run(ctx, "", "sh", "-c", cmd)
	return true
}

func (n *Notifier) osascriptCommand(ctx context.Context, data TemplateData) error {
	cmd := fmt.Sprintf("osascript -e %q", fmt.Sprintf("display notification %q with title %q subtitle %q", data.Message, "Litents", data.Agent))
	_, err := n.Runner.Run(ctx, "", "sh", "-c", cmd)
	if err != nil {
		return err
	}
	return nil
}

func replaceTemplate(base string, data TemplateData) string {
	repl := strings.NewReplacer(
		"{{project}}", data.Project,
		"{{agent}}", data.Agent,
		"{{status}}", data.Status,
		"{{message}}", data.Message,
		"{{worktree}}", data.Worktree,
		"{{log_file}}", data.LogFile,
	)
	return repl.Replace(base)
}

func renderTemplate(input string, vars map[string]string) string {
	repl := strings.NewReplacer()
	pairs := make([]string, 0, len(vars)*2)
	for key, val := range vars {
		pairs = append(pairs, key, val)
	}
	repl = strings.NewReplacer(pairs...)
	return repl.Replace(input)
}

