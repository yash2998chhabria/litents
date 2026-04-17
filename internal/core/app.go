package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type App struct {
	Out    io.Writer
	ErrOut io.Writer
	Now    func() time.Time
	Config Config
}

type Config struct {
	TmuxSessionPrefix      string   `json:"tmux_session_prefix"`
	WorktreeRoot           string   `json:"worktree_root"`
	DefaultBaseBranch      string   `json:"default_base_branch"`
	CodexCommand           string   `json:"codex_command"`
	CodexArgs              []string `json:"codex_args"`
	NotifyEnabled          bool     `json:"notify_enabled"`
	NotifyCommand          string   `json:"notify_command"`
	WatchIntervalSeconds   int      `json:"watch_interval_seconds"`
	SilenceThresholdSeconds int     `json:"silence_threshold_seconds"`
	ActivityNotifyCooldown int     `json:"activity_notify_cooldown_seconds"`
	WaitingRegexes         []string `json:"waiting_regexes"`
	DoneRegexes            []string `json:"done_regexes"`
	NotifyOnQuiet          bool     `json:"notify_on_quiet"`
}

type Project struct {
	Name        string    `json:"name"`
	RepoPath    string    `json:"repo_path"`
	TmuxSession string    `json:"tmux_session"`
	CreatedAt   time.Time `json:"created_at"`
}

type Agent struct {
	ID             string     `json:"id"`
	Project        string     `json:"project"`
	Role           string     `json:"role"`
	RepoPath       string     `json:"repo_path"`
	WorktreePath   string     `json:"worktree_path"`
	Branch         string     `json:"branch"`
	TmuxSession    string     `json:"tmux_session"`
	TmuxWindow     string     `json:"tmux_window"`
	TmuxPane       string     `json:"tmux_pane"`
	PromptFile     string     `json:"prompt_file"`
	LogFile        string     `json:"log_file"`
	Status         string     `json:"status"`
	LastStatus     string     `json:"last_status"`
	LastActivityAt time.Time  `json:"last_activity_at"`
	LastNotifiedAt *time.Time `json:"last_notified_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

const (
	statusStarting = "starting"
	statusRunning  = "running"
	statusWaiting  = "waiting"
	statusQuiet    = "quiet"
	statusDone     = "done"
	statusFailed   = "failed"
	statusUnknown  = "unknown"
)

var agentIDRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func NewApp(out, errOut io.Writer) *App {
	a := &App{
		Out:    out,
		ErrOut: errOut,
		Now:    time.Now,
	}
	a.Config = loadConfigWithDefaults()
	a.loadConfig()
	return a
}

func loadConfigWithDefaults() Config {
	home, _ := os.UserHomeDir()
	return Config{
		TmuxSessionPrefix:      "litents",
		WorktreeRoot:           filepath.Join(home, ".local", "share", "litents", "worktrees"),
		DefaultBaseBranch:      "main",
		CodexCommand:           "codex",
		CodexArgs:              []string{},
		NotifyEnabled:          true,
		NotifyCommand:          "auto",
		WatchIntervalSeconds:   3,
		SilenceThresholdSeconds: 180,
		ActivityNotifyCooldown: 120,
		WaitingRegexes: []string{
			"(?i)approval",
			"(?i)allow.*command",
			"(?i)requires.*permission",
			"(?i)permission required",
			"(?i)continue\\?",
			"(?i)press enter",
			"(?i)waiting for input",
			"(?i)do you want",
			"(?i)yes/no",
			"(?i)y/n",
			"❯",
			">\\s*$",
		},
		DoneRegexes: []string{
			"(?i)task complete",
			"(?i)done",
			"(?i)finished",
		},
		NotifyOnQuiet:          false,
	}
}

func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return a.printUsage()
	}

	switch args[0] {
	case "doctor":
		return a.handleDoctor(args[1:])
	case "init":
		return a.handleInit(args[1:])
	case "new":
		return a.handleNew(args[1:])
	case "ls", "status":
		return a.handleStatus(args[1:])
	case "attach":
		return a.handleAttach(args[1:])
	case "send":
		return a.handleSend(args[1:])
	case "tail":
		return a.handleTail(args[1:])
	case "notify":
		return a.handleNotify(args[1:])
	case "watch":
		return a.handleWatch(args[1:])
	case "resume":
		return a.handleResume(args[1:])
	case "history":
		return a.handleHistory(args[1:])
	case "stop":
		return a.handleStop(args[1:])
	case "clean":
		return a.handleClean(args[1:])
	case "help":
		return a.printUsage()
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (a *App) printUsage() error {
	usage := `litents [command]

Core commands:
  doctor                    Check dependencies and directories
  init [repo]               Initialize a project for a repo
  new <agent-id>            Create a new agent
  status | ls               Show agent table
  attach <agent-id>         Attach to an agent window
  send <agent-id> <text>    Send text to agent
  tail <agent-id>           Print agent log
  notify test               Test notification command
  watch                     Poll and print agent status
  resume <agent-id>         Resume an agent pane from worktree
  history                   Show past agents
  stop <agent-id>           Stop an agent
  clean                     Remove dead agent state and optional worktrees
`
	_, err := io.WriteString(a.Out, usage)
	return err
}

func (a *App) handleDoctor(args []string) error {
	_ = args
	paths := []string{
		"tmux",
		"git",
		"codex",
	}
	var b strings.Builder
	b.WriteString("Litents doctor\n")
	for _, p := range paths {
		lookup, err := exec.LookPath(p)
		if err != nil {
			b.WriteString(fmt.Sprintf("✗ %s: not found\n", p))
			continue
		}
		b.WriteString(fmt.Sprintf("✓ %s: %s\n", p, lookup))
	}
	state := a.stateRoot()
	config := a.configDir()
	if err := ensureDirWritable(state); err != nil {
		b.WriteString(fmt.Sprintf("✗ state dir: %s (%v)\n", state, err))
	} else {
		b.WriteString(fmt.Sprintf("✓ state dir: %s\n", state))
	}
	if err := ensureDirWritable(config); err != nil {
		b.WriteString(fmt.Sprintf("✗ config dir: %s (%v)\n", config, err))
	} else {
		b.WriteString(fmt.Sprintf("✓ config dir: %s\n", config))
	}
	_, err := io.WriteString(a.Out, b.String())
	return err
}

func (a *App) handleInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	session := fs.String("session", "", "tmux session name")
	noWatch := fs.Bool("no-watch", false, "skip watch window")
	worktreeRoot := fs.String("worktree-root", "", "override configured worktree root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	repoArg := "."
	if fs.NArg() > 1 {
		return errors.New("usage: litents init [repo-path]")
	}
	if fs.NArg() == 1 {
		repoArg = fs.Arg(0)
	}
	repoPath, err := resolveRepoRoot(repoArg)
	if err != nil {
		return err
	}

	cfg := a.Config
	if *worktreeRoot != "" {
		cfg.WorktreeRoot = expandPath(*worktreeRoot)
	}
	name := filepath.Base(repoPath)
	sessionName := cfg.TmuxSessionPrefix + "-" + name
	if *session != "" {
		sessionName = *session
	}
	project := &Project{
		Name:        name,
		RepoPath:    repoPath,
		TmuxSession: sessionName,
		CreatedAt:   a.Now().UTC(),
	}
	if err := a.writeProject(project); err != nil {
		return err
	}

	if !tmuxHasSession(sessionName) {
		if err := runCommand(a.ErrOut, "tmux", "new-session", "-d", "-s", sessionName, "-n", "home", "-c", repoPath, "bash"); err != nil {
			return fmt.Errorf("tmux new-session: %w", err)
		}
	} else {
		_ = runCommand(a.ErrOut, "tmux", "new-window", "-t", sessionName, "-n", "home", "-c", repoPath, "bash")
	}

	if !*noWatch {
		windows, _ := tmuxListWindows(sessionName)
		if !containsString(windows, "watch") {
			cmd := []string{"tmux", "new-window", "-t", sessionName, "-n", "watch", "-c", repoPath, "sh", "-c", quoteForShell(os.Args[0]) + " watch --project " + quoteForShell(name)}
			_ = runCommand(a.ErrOut, cmd[0], cmd[1:]...)
		}
	}
	_, err = io.WriteString(a.Out, "✓ initialized "+name+"\n")
	return err
}

func (a *App) handleNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	repoPath := fs.String("repo", "", "repo root path")
	promptText := fs.String("prompt", "", "prompt text")
	promptFile := fs.String("prompt-file", "", "prompt file path")
	baseBranch := fs.String("base-branch", a.Config.DefaultBaseBranch, "base branch")
	branch := fs.String("branch", "", "branch name")
	noWorktree := fs.Bool("no-worktree", false, "do not create a git worktree")
	windowName := fs.String("window", "", "tmux window name")
	profile := fs.String("profile", "", "codex profile")
	var codexArgs stringSliceFlag
	fs.Var(&codexArgs, "codex-arg", "repeatable codex args")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: litents new <agent-id> [flags]")
	}
	agentID := fs.Arg(0)
	if !agentIDRegex.MatchString(agentID) {
		return fmt.Errorf("agent id must match %s", agentIDRegex.String())
	}

	project, err := a.resolveProject(*projectName, *repoPath)
	if err != nil {
		return err
	}

	prompt, err := a.agentPrompt(*promptText, *promptFile)
	if err != nil {
		return err
	}

	b := strings.TrimSpace(*baseBranch)
	if b == "" {
		b = a.Config.DefaultBaseBranch
	}
	agentBranch := strings.TrimSpace(*branch)
	if agentBranch == "" {
		agentBranch = "litents/" + agentID
	}

	worktreePath := project.RepoPath
	if !*noWorktree {
		worktreePath = filepath.Join(a.Config.WorktreeRoot, project.Name, agentID)
		if err := ensureDir(filepath.Dir(worktreePath)); err != nil {
			return err
		}
	if _, statErr := os.Stat(worktreePath); os.IsNotExist(statErr) {
		if err := runCommand(a.ErrOut, "git", "-C", project.RepoPath, "worktree", "add", "-B", agentBranch, worktreePath, b); err != nil {
			return fmt.Errorf("git worktree add failed: %w", err)
		}
	}
	}

	window := agentID
	if strings.TrimSpace(*windowName) != "" {
		window = strings.TrimSpace(*windowName)
	}
	if exists, err := tmuxHasWindow(project.TmuxSession, window); err == nil && exists {
		return fmt.Errorf("tmux window already exists: %s", window)
	}
	now := a.Now().UTC()
	agent := &Agent{
		ID:           agentID,
		Project:      project.Name,
		Role:         "",
		RepoPath:     project.RepoPath,
		WorktreePath: worktreePath,
		Branch:       agentBranch,
		TmuxSession:  project.TmuxSession,
		TmuxWindow:   window,
		PromptFile:   a.agentPromptPath(project.Name, agentID),
		LogFile:      a.agentLogPath(project.Name, agentID),
		Status:       statusStarting,
		LastStatus:   statusStarting,
		LastActivityAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := a.writeAgent(agent); err != nil {
		return err
	}
	if err := ensureDir(filepath.Dir(agent.PromptFile)); err != nil {
		return err
	}
	if err := os.WriteFile(agent.PromptFile, []byte(prompt), 0o644); err != nil {
		return err
	}

	argsForCodex := append([]string{}, a.Config.CodexArgs...)
	argsForCodex = append(argsForCodex, codexArgs...)
	if strings.TrimSpace(*profile) != "" {
		argsForCodex = append(argsForCodex, "--profile", strings.TrimSpace(*profile))
	}
	runnerPath := a.agentRunnerPath(project.Name, agentID)
	if err := writeRunnerFromPrompt(a, runnerPath, agent, a.Config.CodexCommand, argsForCodex); err != nil {
		return err
	}

	out, err := runCommandOutput(a.ErrOut, "tmux", "new-window", "-t", project.TmuxSession, "-n", window, "-c", worktreePath, "-P", "-F", "#{pane_id}", "sh", runnerPath)
	if err != nil {
		return err
	}
	if err := runCommand(a.ErrOut, "tmux", "pipe-pane", "-o", "-t", fmt.Sprintf("%s:%s", project.TmuxSession, window), fmt.Sprintf("cat >> %q", agent.LogFile)); err != nil {
		return err
	}
	agent.TmuxPane = strings.TrimSpace(out)
	agent.UpdatedAt = a.Now().UTC()
	agent.LastActivityAt = a.Now().UTC()
	if err := a.writeAgent(agent); err != nil {
		return err
	}
	_, err = io.WriteString(a.Out, "✓ created agent "+agentID+"\n")
	return err
}

func (a *App) handleStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	watch := fs.Bool("watch", false, "watch mode")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tick := time.Duration(a.Config.WatchIntervalSeconds) * time.Second
	if tick <= 0 {
		tick = 3 * time.Second
	}

	for {
		agents, err := a.loadAgentsByProject(*projectName)
		if err != nil {
			return err
		}
		updated := make([]*Agent, 0, len(agents))
		for _, agent := range agents {
			updated = append(updated, a.refreshAgentStatus(agent))
		}
		a.printStatusRows(updated)
		if !*watch {
			return nil
		}
		time.Sleep(tick)
	}
}

func (a *App) handleAttach(args []string) error {
	fs := flag.NewFlagSet("attach", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: litents attach <agent-id> [--project name]")
	}
	agent, err := a.findAgent(*projectName, fs.Arg(0))
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%s:%s", agent.TmuxSession, agent.TmuxWindow)
	if _, exists := os.LookupEnv("TMUX"); exists {
		if err := runCommand(a.ErrOut, "tmux", "select-window", "-t", target); err != nil {
			return err
		}
		_, err = io.WriteString(a.Out, "✓ attached\n")
		return err
	}
	if err := runCommand(a.ErrOut, "tmux", "attach-session", "-t", agent.TmuxSession); err != nil {
		return err
	}
	if err := runCommand(a.ErrOut, "tmux", "select-window", "-t", target); err != nil {
		return err
	}
	_, err = io.WriteString(a.Out, "✓ attached\n")
	return err
}

func (a *App) handleSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	enterOnly := fs.Bool("enter-only", false, "send only Enter")
	noEnter := fs.Bool("no-enter", false, "skip Enter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 || fs.NArg() > 2 {
		return errors.New("usage: litents send <agent-id> [text] [--project name]")
	}
	agentID := fs.Arg(0)
	message := ""
	if fs.NArg() == 2 {
		message = fs.Arg(1)
	}
	if strings.TrimSpace(message) == "" && !*enterOnly {
		return errors.New("empty message requires --enter-only")
	}
	agent, err := a.findAgent(*projectName, agentID)
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%s:%s", agent.TmuxSession, agent.TmuxWindow)
	if strings.TrimSpace(message) != "" {
		if err := runCommand(a.ErrOut, "tmux", "send-keys", "-l", "-t", target, message); err != nil {
			return err
		}
	}
	if !*noEnter {
		if err := runCommand(a.ErrOut, "tmux", "send-keys", "-t", target, "Enter"); err != nil {
			return err
		}
	}
	agent.LastActivityAt = a.Now().UTC()
	agent.UpdatedAt = a.Now().UTC()
	if err := a.writeAgent(agent); err != nil {
		return err
	}
	_, err = io.WriteString(a.Out, "✓ sent\n")
	return err
}

func (a *App) handleTail(args []string) error {
	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	lines := fs.Int("n", 80, "number of lines")
	follow := fs.Bool("follow", false, "follow output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: litents tail <agent-id> [--n N] [--follow]")
	}
	agent, err := a.findAgent(*projectName, fs.Arg(0))
	if err != nil {
		return err
	}
	if _, err := os.Stat(agent.LogFile); err != nil {
		return fmt.Errorf("log not found for %s", agent.ID)
	}
	if *follow {
		return tailFollow(agent.LogFile, *lines, a.Out, a.ErrOut, a.Now)
	}
	text, err := tailLines(agent.LogFile, *lines)
	if err != nil {
		return err
	}
	for _, l := range text {
		_, _ = io.WriteString(a.Out, l+"\n")
	}
	return nil
}

func (a *App) handleNotify(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: litents notify test")
	}
	if args[0] != "test" {
		return fmt.Errorf("unknown notify subcommand: %s", args[0])
	}
	return a.sendNotification("system", "notify-test", statusDone, "litents notification test")
}

func (a *App) handleWatch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	interval := time.Duration(a.Config.WatchIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 3 * time.Second
	}
	lastNotified := map[string]time.Time{}
	for {
		agents, err := a.loadAgentsByProject(*projectName)
		if err != nil {
			return err
		}
		changed := make([]*Agent, 0, len(agents))
		for _, agent := range agents {
			before := agent.Status
			updated := a.refreshAgentStatus(agent)
			if err := a.writeAgent(updated); err != nil {
				return err
			}
			if before != updated.Status {
				changed = append(changed, updated)
				shouldNotify := updated.Status == statusWaiting || updated.Status == statusFailed || updated.Status == statusDone
				if updated.Status == statusQuiet && a.Config.NotifyOnQuiet {
					shouldNotify = true
				}
				if shouldNotify {
					now := a.Now().UTC()
					last, ok := lastNotified[agentStatusKey(agent.Project, agent.ID)]
					if !ok || now.Sub(last).Seconds() >= float64(a.Config.ActivityNotifyCooldown) {
						_ = a.sendNotification(updated.Project, updated.ID, updated.Status, "status changed")
						lastNotified[agentStatusKey(agent.Project, agent.ID)] = now
					}
				}
			}
		}
		a.printStatusRows(changed)
		time.Sleep(interval)
	}
}

func agentStatusKey(project, id string) string {
	return project + "/" + id
}

func (a *App) handleResume(args []string) error {
	fs := flag.NewFlagSet("resume", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	picker := fs.Bool("picker", false, "force picker")
	all := fs.Bool("all", false, "resume all sessions")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: litents resume <agent-id> [--project name]")
	}
	agent, err := a.findAgent(*projectName, fs.Arg(0))
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%s:%s", agent.TmuxSession, agent.TmuxWindow)
	exists, err := tmuxHasPane(agent.TmuxPane, target)
	if err == nil && exists {
		if err := a.handleAttach([]string{agent.ID, "--project", agent.Project}); err != nil {
			return err
		}
		return nil
	}
	argsResume := []string{"resume"}
	if *all {
		argsResume = append(argsResume, "--all")
	} else if !*picker {
		argsResume = append(argsResume, "--last")
	}
	runnerPath := filepath.Join(a.agentDir(agent.Project, agent.ID), "resume-runner.sh")
	if err := writeRunnerFromCommand(a, runnerPath, agent.WorktreePath, agent.LogFile, a.Config.CodexCommand, argsResume); err != nil {
		return err
	}
	out, err := runCommandOutput(a.ErrOut, "tmux", "new-window", "-t", agent.TmuxSession, "-n", agent.TmuxWindow, "-c", agent.WorktreePath, "-P", "-F", "#{pane_id}", "sh", runnerPath)
	if err != nil {
		return err
	}
	agent.TmuxPane = strings.TrimSpace(out)
	agent.Status = statusStarting
	agent.UpdatedAt = a.Now().UTC()
	if err := a.writeAgent(agent); err != nil {
		return err
	}
	_, err = io.WriteString(a.Out, "✓ resumed "+agent.ID+"\n")
	return err
}

func (a *App) handleHistory(args []string) error {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	agents, err := a.loadAgentsByProject(*projectName)
	if err != nil {
		return err
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].CreatedAt.After(agents[j].CreatedAt)
	})
	header := "PROJECT\tAGENT\tSTATUS\tCREATED\tWORKTREE\tPROMPT\n"
	if _, err := io.WriteString(a.Out, header); err != nil {
		return err
	}
	for _, agent := range agents {
		promptSummary := promptSummary(agent.PromptFile)
		if promptSummary == "" {
			promptSummary = "-"
		}
		created := agent.CreatedAt.Format(time.RFC3339)
		_, _ = io.WriteString(a.Out, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", agent.Project, agent.ID, agent.Status, created, agent.WorktreePath, promptSummary))
	}
	return nil
}

func (a *App) handleStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	force := fs.Bool("force", false, "force kill pane")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: litents stop <agent-id> [--project name]")
	}
	agent, err := a.findAgent(*projectName, fs.Arg(0))
	if err != nil {
		return err
	}
	target := agent.TmuxPane
	if strings.TrimSpace(target) != "" {
		_ = runCommand(a.ErrOut, "tmux", "send-keys", "-t", target, "C-c")
		time.Sleep(700 * time.Millisecond)
		alive, _ := tmuxHasPane(target, fmt.Sprintf("%s:%s", agent.TmuxSession, agent.TmuxWindow))
		if alive {
			if *force {
				_ = runCommand(a.ErrOut, "tmux", "kill-pane", "-t", target)
			} else {
				return fmt.Errorf("pane still alive; rerun with --force")
			}
		}
	}
	agent.Status = statusDone
	agent.UpdatedAt = a.Now().UTC()
	if err := a.writeAgent(agent); err != nil {
		return err
	}
	_, err = io.WriteString(a.Out, "✓ stopped "+agent.ID+"\n")
	return err
}

func (a *App) handleClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	projectName := fs.String("project", "", "project name")
	removeWorktrees := fs.Bool("worktrees", false, "remove worktrees")
	mergedOnly := fs.Bool("merged-only", false, "only remove merged worktrees")
	force := fs.Bool("force", false, "force worktree deletion")
	if err := fs.Parse(args); err != nil {
		return err
	}
	agents, err := a.loadAgentsByProject(*projectName)
	if err != nil {
		return err
	}
	removed := 0
	for _, agent := range agents {
		if agent.Status != statusDone && agent.Status != statusFailed {
			continue
		}
		if *removeWorktrees {
			if err := a.removeWorktreeForAgent(agent, *mergedOnly, *force); err == nil {
				removed++
			}
		}
		_ = os.RemoveAll(a.agentDir(agent.Project, agent.ID))
	}
	_, _ = a.writeString(a.Out, fmt.Sprintf("✓ cleaned %d agents\n", len(agents)))
	_, _ = a.writeString(a.ErrOut, fmt.Sprintf("removed worktrees: %d\n", removed))
	return nil
}

func (a *App) writeString(w io.Writer, s string) (int, error) {
	return io.WriteString(w, s)
}

func (a *App) resolveProject(name, repoArg string) (*Project, error) {
	if strings.TrimSpace(name) != "" {
		project, err := a.loadProject(strings.TrimSpace(name))
		if err != nil {
			return nil, err
		}
		return project, nil
	}
	if strings.TrimSpace(repoArg) == "" {
		repoArg = "."
	}
	repo, err := resolveRepoRoot(repoArg)
	if err != nil {
		return nil, err
	}
	if project, err := a.projectFromRepo(repo); err == nil {
		return project, nil
	}
	project := &Project{
		Name:        filepath.Base(repo),
		RepoPath:    repo,
		TmuxSession: a.Config.TmuxSessionPrefix + "-" + filepath.Base(repo),
		CreatedAt:   a.Now().UTC(),
	}
	if err := a.writeProject(project); err != nil {
		return nil, err
	}
	return project, nil
}

func (a *App) agentPrompt(promptText, promptFile string) (string, error) {
	if strings.TrimSpace(promptText) != "" {
		return promptText, nil
	}
	if strings.TrimSpace(promptFile) == "" {
		return "", errors.New("provide --prompt or --prompt-file")
	}
	data, err := os.ReadFile(expandPath(promptFile))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) printStatusRows(agents []*Agent) {
	sort.Slice(agents, func(i, j int) bool {
		if agents[i].Project == agents[j].Project {
			return agents[i].ID < agents[j].ID
		}
		return agents[i].Project < agents[j].Project
	})
	_, _ = io.WriteString(a.Out, "PROJECT\tAGENT\tSTATUS\tAGE\tLAST ACTIVITY\tWORKTREE\n")
	for _, agent := range agents {
		age := formatDuration(a.Now().UTC().Sub(agent.CreatedAt))
		lastActivity := "n/a"
		if !agent.LastActivityAt.IsZero() {
			lastActivity = formatDuration(a.Now().UTC().Sub(agent.LastActivityAt)) + " ago"
		}
		_, _ = io.WriteString(a.Out, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", agent.Project, agent.ID, agent.Status, age, lastActivity, agent.WorktreePath))
	}
}

func (a *App) loadAgentsByProject(project string) ([]*Agent, error) {
	if strings.TrimSpace(project) != "" {
		return a.loadAgentsForProject(strings.TrimSpace(project))
	}
	projects, err := a.listProjects()
	if err != nil {
		return nil, err
	}
	agents := []*Agent{}
	for _, p := range projects {
		ps, err := a.loadAgentsForProject(p.Name)
		if err != nil {
			continue
		}
		agents = append(agents, ps...)
	}
	return agents, nil
}

func (a *App) loadAgentsForProject(project string) ([]*Agent, error) {
	project = strings.TrimSpace(project)
	base := a.projectAgentsBase(project)
	dirs, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return []*Agent{}, nil
	}
	if err != nil {
		return nil, err
	}
	agents := make([]*Agent, 0, len(dirs))
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		path := filepath.Join(base, d.Name(), "agent.json")
		agent, err := readAgent(path)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func (a *App) findAgent(project, id string) (*Agent, error) {
	candidates, err := a.loadAgentsByProject(project)
	if err != nil {
		return nil, err
	}
	matches := []*Agent{}
	for _, a := range candidates {
		if a.ID == id {
			matches = append(matches, a)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("agent not found: %s", id)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple agents named %s; pass --project", id)
	}
	return matches[0], nil
}

func (a *App) refreshAgentStatus(agent *Agent) *Agent {
	target := fmt.Sprintf("%s:%s", agent.TmuxSession, agent.TmuxWindow)
	logTail := readFileTailSafe(agent.LogFile, 40)
	paneAlive, err := tmuxHasPane(agent.TmuxPane, target)
	if err == nil && paneAlive {
		if a.matchesAny(logTail, a.Config.WaitingRegexes) {
			agent.Status = statusWaiting
		} else if a.Config.SilenceThresholdSeconds > 0 && a.isQuiet(agent.LastActivityAt, agent.LogFile) {
			agent.Status = statusQuiet
		} else {
			agent.Status = statusRunning
		}
	} else {
		if matchExitStatus(logTail, statusFailed, statusDone) == statusFailed {
			agent.Status = statusFailed
		} else if matchDoneLog(logTail, a.Config.DoneRegexes) {
			agent.Status = statusDone
		} else if agent.Status == statusStarting || agent.Status == statusRunning || agent.Status == statusWaiting || agent.Status == statusQuiet {
			agent.Status = statusDone
		}
	}
	logInfo, err := os.Stat(agent.LogFile)
	if err == nil {
		agent.LastActivityAt = logInfo.ModTime().UTC()
	}
	agent.UpdatedAt = a.Now().UTC()
	if agent.Status != agent.LastStatus {
		agent.LastStatus = agent.Status
	}
	return agent
}

func (a *App) writeAgent(agent *Agent) error {
	return writeJSON(a.agentPath(agent.Project, agent.ID), agent)
}

func (a *App) agentPath(project, id string) string {
	return filepath.Join(a.projectAgentsBase(project), id, "agent.json")
}

func (a *App) agentPromptPath(project, id string) string {
	return filepath.Join(a.projectAgentsBase(project), id, "prompt.md")
}

func (a *App) agentLogPath(project, id string) string {
	return filepath.Join(a.projectAgentsBase(project), id, "output.log")
}

func (a *App) agentRunnerPath(project, id string) string {
	return filepath.Join(a.projectAgentsBase(project), id, "runner.sh")
}

func (a *App) agentDir(project, id string) string {
	return filepath.Join(a.projectAgentsBase(project), id)
}

func (a *App) projectAgentsBase(project string) string {
	return filepath.Join(a.projectsRoot(), project, "agents")
}

func (a *App) projectsRoot() string {
	return filepath.Join(a.stateRoot(), "projects")
}

func (a *App) loadProject(name string) (*Project, error) {
	return readProject(filepath.Join(a.projectsRoot(), name, "project.json"))
}

func (a *App) writeProject(project *Project) error {
	path := filepath.Join(a.projectsRoot(), project.Name, "project.json")
	return writeJSON(path, project)
}

func (a *App) listProjects() ([]Project, error) {
	entries, err := os.ReadDir(a.projectsRoot())
	if os.IsNotExist(err) {
		return []Project{}, nil
	}
	if err != nil {
		return nil, err
	}
	projects := []Project{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(a.projectsRoot(), e.Name(), "project.json")
		p, err := readProject(path)
		if err != nil {
			continue
		}
		projects = append(projects, *p)
	}
	return projects, nil
}

func (a *App) projectFromRepo(repo string) (*Project, error) {
	projects, err := a.listProjects()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if samePath(p.RepoPath, repo) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project not initialized for repo %s", repo)
}

func (a *App) stateRoot() string {
	return xdgStateRoot()
}

func (a *App) configDir() string {
	return xdgConfigRoot()
}

func (a *App) configFile() string {
	return filepath.Join(a.configDir(), "config.json")
}

func (a *App) loadConfig() {
	data, err := os.ReadFile(a.configFile())
	if err != nil {
		return
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	if cfg.TmuxSessionPrefix != "" {
		a.Config.TmuxSessionPrefix = cfg.TmuxSessionPrefix
	}
	if cfg.WorktreeRoot != "" {
		a.Config.WorktreeRoot = expandPath(cfg.WorktreeRoot)
	}
	if cfg.DefaultBaseBranch != "" {
		a.Config.DefaultBaseBranch = cfg.DefaultBaseBranch
	}
	if cfg.CodexCommand != "" {
		a.Config.CodexCommand = cfg.CodexCommand
	}
	if len(cfg.CodexArgs) > 0 {
		a.Config.CodexArgs = cfg.CodexArgs
	}
	if cfg.NotifyCommand != "" {
		a.Config.NotifyCommand = cfg.NotifyCommand
	}
	if cfg.WatchIntervalSeconds > 0 {
		a.Config.WatchIntervalSeconds = cfg.WatchIntervalSeconds
	}
	if cfg.SilenceThresholdSeconds > 0 {
		a.Config.SilenceThresholdSeconds = cfg.SilenceThresholdSeconds
	}
	if cfg.ActivityNotifyCooldown > 0 {
		a.Config.ActivityNotifyCooldown = cfg.ActivityNotifyCooldown
	}
	if len(cfg.WaitingRegexes) > 0 {
		a.Config.WaitingRegexes = cfg.WaitingRegexes
	}
	if len(cfg.DoneRegexes) > 0 {
		a.Config.DoneRegexes = cfg.DoneRegexes
	}
	a.Config.NotifyOnQuiet = cfg.NotifyOnQuiet
	a.Config.NotifyEnabled = cfg.NotifyEnabled
}

func resolveRepoRoot(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	path = expandPath(path)
	out, err := runCommandOutput(nil, "git", "-C", path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed for %s: %w", path, err)
	}
	return strings.TrimSpace(out), nil
}

func runCommand(errOut io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = errOut
	cmd.Stderr = errOut
	return cmd.Run()
}

func runCommandOutput(errOut io.Writer, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && errOut != nil {
			_, _ = errOut.Write(ee.Stderr)
		}
		return "", err
	}
	return string(out), nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func ensureDirWritable(path string) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	tmp := filepath.Join(path, ".litents-writecheck")
	if err := os.WriteFile(tmp, []byte("ok"), 0o644); err != nil {
		return err
	}
	_ = os.Remove(tmp)
	return nil
}

func writeJSON(path string, value any) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readProject(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func readAgent(path string) (*Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a Agent
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func tmuxHasSession(name string) bool {
	if err := runCommand(io.Discard, "tmux", "has-session", "-t", name); err != nil {
		return false
	}
	return true
}

func tmuxListWindows(session string) ([]string, error) {
	out, err := runCommandOutput(io.Discard, "tmux", "list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return nil, err
	}
	lines := []string{}
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

func tmuxHasWindow(session, window string) (bool, error) {
	windows, err := tmuxListWindows(session)
	if err != nil {
		return false, err
	}
	for _, w := range windows {
		if w == window {
			return true, nil
		}
	}
	return false, nil
}

func tmuxHasPane(paneID, target string) (bool, error) {
	if strings.TrimSpace(paneID) == "" {
		return false, nil
	}
	out, err := runCommandOutput(io.Discard, "tmux", "list-panes", "-t", target, "-F", "#{pane_id}")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == paneID {
			return true, nil
		}
	}
	return false, nil
}

func promptSummary(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(b))
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return ""
	}
	sum := strings.TrimSpace(lines[0])
	if len(sum) > 48 {
		return sum[:48] + "..."
	}
	return sum
}

func (a *App) isQuiet(lastActivity time.Time, logfile string) bool {
	if lastActivity.IsZero() {
		info, err := os.Stat(logfile)
		if err != nil {
			return false
		}
		lastActivity = info.ModTime().UTC()
	}
	if lastActivity.IsZero() || a.Config.SilenceThresholdSeconds <= 0 {
		return false
	}
	return a.Now().UTC().Sub(lastActivity).Seconds() > float64(a.Config.SilenceThresholdSeconds)
}

func (a *App) matchesAny(text string, patterns []string) bool {
	if text == "" {
		return false
	}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func matchDoneLog(text string, patterns []string) bool {
	if text == "" {
		return false
	}
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func matchExitStatus(text string, failedState string, doneState string) string {
	re := regexp.MustCompile(`(?i)codex exited with status\s+([0-9]+)`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return doneState
	}
	if match[1] == "0" {
		return doneState
	}
	return failedState
}

func readFileTailSafe(path string, n int) string {
	lines, err := tailLines(path, n)
	if err != nil {
		return ""
	}
	return strings.Join(lines, "\n")
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}
	seconds := int64(d.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", seconds)
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func writeRunnerFromPrompt(a *App, scriptPath string, agent *Agent, codex string, args []string) error {
	var sb strings.Builder
	sb.WriteString("#!/bin/sh\n")
	sb.WriteString("set -eu\n")
	sb.WriteString("WORKTREE=" + quoteForShell(agent.WorktreePath) + "\n")
	sb.WriteString("PROMPT_FILE=" + quoteForShell(agent.PromptFile) + "\n")
	sb.WriteString("LOG_FILE=" + quoteForShell(agent.LogFile) + "\n")
	sb.WriteString("cd \"$WORKTREE\"\n")
	sb.WriteString("mkdir -p \"$(dirname \"$LOG_FILE\")\"\n")
	sb.WriteString("prompt=\"$(cat \"$PROMPT_FILE\")\"\n")
	sb.WriteString("status=0\n")
	c := []string{quoteForShell(codex)}
	for _, arg := range args {
		c = append(c, quoteForShell(arg))
	}
	sb.WriteString(strings.Join(c, " "))
	sb.WriteString(" \"$prompt\" >> \"$LOG_FILE\" 2>&1 || status=$?\n")
	sb.WriteString("echo \"[litents] codex exited with status $status\" >> \"$LOG_FILE\"\n")
	sb.WriteString("exit ${status}\n")
	if err := ensureDir(filepath.Dir(scriptPath)); err != nil {
		return err
	}
	if err := os.WriteFile(scriptPath, []byte(sb.String()), 0o755); err != nil {
		return err
	}
	return nil
}

func writeRunnerFromCommand(a *App, scriptPath, worktree, logFile, command string, args []string) error {
	_ = a
	var sb strings.Builder
	sb.WriteString("#!/bin/sh\n")
	sb.WriteString("set -eu\n")
	sb.WriteString("WORKTREE=" + quoteForShell(worktree) + "\n")
	sb.WriteString("LOG_FILE=" + quoteForShell(logFile) + "\n")
	sb.WriteString("cd \"$WORKTREE\"\n")
	sb.WriteString("mkdir -p \"$(dirname \"$LOG_FILE\")\"\n")
	c := []string{quoteForShell(command)}
	for _, arg := range args {
		c = append(c, quoteForShell(arg))
	}
	sb.WriteString(strings.Join(c, " "))
	sb.WriteString(" >> \"$LOG_FILE\" 2>&1 || true\n")
	if err := ensureDir(filepath.Dir(scriptPath)); err != nil {
		return err
	}
	return os.WriteFile(scriptPath, []byte(sb.String()), 0o755)
}

func quoteForShell(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "'\"'\"'") + "'"
}

func tailLines(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	lines := make([]string, 0, n)
	for sc.Scan() {
		if len(lines) < n {
			lines = append(lines, sc.Text())
			continue
		}
		copy(lines, lines[1:])
		lines[n-1] = sc.Text()
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func tailFollow(path string, n int, out io.Writer, errOut io.Writer, now func() time.Time) error {
	_ = now
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	pos, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if n > 0 {
		_, _ = f.Seek(0, io.SeekStart)
		lines := make([]string, 0, n)
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if len(lines) < n {
				lines = append(lines, sc.Text())
			} else {
				copy(lines, lines[1:])
				lines[n-1] = sc.Text()
			}
		}
		for _, l := range lines {
			_, _ = io.WriteString(out, l+"\n")
		}
		if sc.Err() != nil {
			return sc.Err()
		}
	}
	if err := writeStringNoop(errOut); err != nil {
		_ = err
	}
	for {
		time.Sleep(700 * time.Millisecond)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Size() < pos {
			pos = 0
			_, err := f.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
		}
		if info.Size() == pos {
			continue
		}
		buf := make([]byte, info.Size()-pos)
		nr, err := f.ReadAt(buf, pos)
		if err != nil && err != io.EOF {
			return err
		}
		if nr > 0 {
			if _, err := out.Write(buf[:nr]); err != nil {
				return err
			}
			pos += int64(nr)
		}
	}
}

func writeStringNoop(w io.Writer) error {
	if w == nil {
		return nil
	}
	_, err := io.WriteString(w, "")
	return err
}

func xdgStateRoot() string {
	if v := os.Getenv("XDG_STATE_HOME"); v != "" {
		return expandPath(v + "/litents")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "litents")
}

func xdgConfigRoot() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return expandPath(v + "/litents")
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "litents")
	}
	return filepath.Join(home, ".config", "litents")
}

func expandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return os.ExpandEnv(p)
}

func (a *App) sendNotification(project, agent, status, message string) error {
	if !a.Config.NotifyEnabled {
		return nil
	}
	cmd := strings.TrimSpace(a.Config.NotifyCommand)
	if cmd == "" || cmd == "auto" {
		cmd = detectNotifyCommand()
	}
	if cmd == "" {
		return nil
	}
	command := strings.NewReplacer(
		"{{project}}", project,
		"{{agent}}", agent,
		"{{status}}", status,
		"{{message}}", message,
	).Replace(cmd)
	if err := runCommand(a.ErrOut, "sh", "-c", command); err != nil {
		return err
	}
	return nil
}

func detectNotifyCommand() string {
	if runtime.GOOS == "darwin" && cmdExists("terminal-notifier") {
		return `terminal-notifier -title litents -message "{{message}}"`
	}
	if runtime.GOOS == "darwin" && cmdExists("osascript") {
		return `osascript -e 'display notification "{{message}}" with title "litents"'`
	}
	if cmdExists("notify-send") {
		return `notify-send "litents" "{{message}}"`
	}
	return ""
}

func cmdExists(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func (a *App) removeWorktreeForAgent(agent *Agent, mergedOnly bool, force bool) error {
	if !mergedOnly {
		return a.removeWorktreeDirect(agent, force)
	}
	if agent.Branch == "" {
		return nil
	}
	out, err := runCommandOutput(a.ErrOut, "git", "-C", agent.RepoPath, "merge-base", "--is-ancestor", agent.Branch, "HEAD")
	if err != nil {
		_ = out
		return nil
	}
	return a.removeWorktreeDirect(agent, force)
}

func (a *App) removeWorktreeDirect(agent *Agent, force bool) error {
	if agent.WorktreePath == "" {
		return nil
	}
	if !force {
		changed, err := hasDirtyWorktree(agent.WorktreePath)
		if err == nil && changed {
			return fmt.Errorf("skip %s: dirty worktree (use --force)", agent.ID)
		}
	}
	_ = runCommand(a.ErrOut, "git", "-C", agent.RepoPath, "worktree", "remove", agent.WorktreePath)
	return os.RemoveAll(agent.WorktreePath)
}

func hasDirtyWorktree(worktree string) (bool, error) {
	out, err := runCommandOutput(io.Discard, "git", "-C", worktree, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}
