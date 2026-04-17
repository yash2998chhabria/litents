package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	StatusStarting = "starting"
	StatusRunning  = "running"
	StatusWaiting  = "waiting"
	StatusQuiet    = "quiet"
	StatusDone     = "done"
	StatusFailed   = "failed"
	StatusUnknown  = "unknown"
)

type Project struct {
	Name        string    `json:"name"`
	RepoPath    string    `json:"repo_path"`
	TmuxSession string    `json:"tmux_session"`
	CreatedAt   time.Time `json:"created_at"`
}

type Agent struct {
	ID               string     `json:"id"`
	Project          string     `json:"project"`
	Role             string     `json:"role"`
	Source           string     `json:"source"`
	RepoPath         string     `json:"repo_path"`
	WorktreePath     string     `json:"worktree_path"`
	Branch           string     `json:"branch"`
	TmuxSession      string     `json:"tmux_session"`
	TmuxWindow       string     `json:"tmux_window"`
	TmuxPane         string     `json:"tmux_pane"`
	CodexSessionID   string     `json:"codex_session_id"`
	CodexThreadID    string     `json:"codex_thread_id"`
	Model            string     `json:"model"`
	ApprovalPolicy   string     `json:"approval_policy"`
	SandboxMode      string     `json:"sandbox_mode"`
	PromptFile       string     `json:"prompt_file"`
	LogFile          string     `json:"log_file"`
	EventsFile       string     `json:"events_file"`
	Status           string     `json:"status"`
	LastStatus       string     `json:"last_status"`
	AttentionReason  string     `json:"attention_reason"`
	AttentionExcerpt string     `json:"attention_excerpt"`
	AttentionSince   *time.Time `json:"attention_since"`
	LastError        string     `json:"last_error"`
	ExitCode         *int       `json:"exit_code"`
	LastActivityAt   time.Time  `json:"last_activity_at"`
	LastNotifiedAt   *time.Time `json:"last_notified_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ArchivedAt       *time.Time `json:"archived_at"`
	RunnerScriptFile string     `json:"runner_script"`
	SummaryFile      string     `json:"summary_file"`
}

var agentIDRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

func ValidateAgentID(id string) error {
	if !agentIDRegex.MatchString(id) {
		return fmt.Errorf("invalid agent id %q; use lowercase, numbers, dash or underscore", id)
	}
	return nil
}

func ProjectNameFromRepo(repoPath string) string {
	repoPath = filepath.Clean(repoPath)
	parts := strings.Split(repoPath, string(filepath.Separator))
	return parts[len(parts)-1]
}

func NormalizePath(in string) (string, error) {
	if in == "" {
		return "", errors.New("empty path")
	}
	abs, err := filepath.Abs(in)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func ProjectDir(stateRoot, projectName string) string {
	return filepath.Join(stateRoot, "projects", projectName)
}

func AgentDir(stateRoot, projectName, agentID string) string {
	return filepath.Join(ProjectDir(stateRoot, projectName), "agents", agentID)
}

func AgentPath(stateRoot, projectName, agentID string) string {
	return filepath.Join(AgentDir(stateRoot, projectName, agentID), "agent.json")
}

func ProjectPath(stateRoot, projectName string) string {
	return filepath.Join(ProjectDir(stateRoot, projectName), "project.json")
}

func loadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func saveJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func SaveProject(stateRoot string, project Project) error {
	return saveJSON(ProjectPath(stateRoot, project.Name), project)
}

func LoadProject(stateRoot, projectName string) (Project, error) {
	var p Project
	if err := loadJSON(ProjectPath(stateRoot, projectName), &p); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Project{}, fs.ErrNotExist
		}
		return Project{}, err
	}
	return p, nil
}

func LoadProjects(stateRoot string) ([]Project, error) {
	root := filepath.Join(stateRoot, "projects")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Project{}, nil
		}
		return nil, err
	}

	projects := make([]Project, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		p, err := LoadProject(stateRoot, entry.Name())
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Name < projects[j].Name })
	return projects, nil
}

func FindProjectByRepo(stateRoot, repoPath string) (Project, error) {
	projects, err := LoadProjects(stateRoot)
	if err != nil {
		return Project{}, err
	}
	repoPath = filepath.Clean(repoPath)
	for _, project := range projects {
		if filepath.Clean(project.RepoPath) == repoPath {
			return project, nil
		}
	}
	return Project{}, fs.ErrNotExist
}

func ProjectForProjectOrRepo(stateRoot, repoPath, projectName string) (Project, error) {
	if projectName != "" {
		return LoadProject(stateRoot, projectName)
	}
	if repoPath != "" {
		abs, err := NormalizePath(repoPath)
		if err != nil {
			return Project{}, err
		}
		return FindProjectByRepo(stateRoot, abs)
	}
	projects, err := LoadProjects(stateRoot)
	if err != nil {
		return Project{}, err
	}
	if len(projects) == 0 {
		return Project{}, fmt.Errorf("no projects initialized")
	}
	if len(projects) > 1 {
		return Project{}, fmt.Errorf("multiple projects exist; pass --project or --repo")
	}
	return projects[0], nil
}

func SaveAgent(stateRoot string, agent Agent) error {
	return saveJSON(AgentPath(stateRoot, agent.Project, agent.ID), agent)
}

func LoadAgent(stateRoot, projectName, agentID string) (Agent, error) {
	var a Agent
	if err := loadJSON(AgentPath(stateRoot, projectName, agentID), &a); err != nil {
		return Agent{}, err
	}
	return a, nil
}

func LoadAgents(stateRoot, projectName string) ([]Agent, error) {
	base := filepath.Join(ProjectDir(stateRoot, projectName), "agents")
	entries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Agent{}, nil
		}
		return nil, err
	}

	agents := make([]Agent, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		a, err := LoadAgent(stateRoot, projectName, entry.Name())
		if err != nil {
			continue
		}
		agents = append(agents, a)
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})
	return agents, nil
}

func LoadAllAgents(stateRoot string) ([]Agent, error) {
	projects, err := LoadProjects(stateRoot)
	if err != nil {
		return nil, err
	}
	all := make([]Agent, 0)
	for _, p := range projects {
		agents, err := LoadAgents(stateRoot, p.Name)
		if err != nil {
			return nil, err
		}
		all = append(all, agents...)
	}
	return all, nil
}

func DeleteAgent(stateRoot, projectName, agentID string) error {
	return os.RemoveAll(AgentDir(stateRoot, projectName, agentID))
}

func ResolveAgentAcrossProjects(stateRoot, project, agentID string) (Project, Agent, error) {
	if project != "" {
		p, err := LoadProject(stateRoot, project)
		if err != nil {
			return Project{}, Agent{}, err
		}
		a, err := LoadAgent(stateRoot, p.Name, agentID)
		if err != nil {
			return Project{}, Agent{}, err
		}
		return p, a, nil
	}

	projects, err := LoadProjects(stateRoot)
	if err != nil {
		return Project{}, Agent{}, err
	}

	var found []struct {
		project Project
		agent   Agent
	}
	for _, p := range projects {
		a, err := LoadAgent(stateRoot, p.Name, agentID)
		if err != nil {
			continue
		}
		found = append(found, struct {
			project Project
			agent   Agent
		}{p, a})
	}

	if len(found) == 0 {
		return Project{}, Agent{}, fmt.Errorf("agent %q not found", agentID)
	}
	if len(found) > 1 {
		return Project{}, Agent{}, fmt.Errorf("agent %q exists in multiple projects; pass --project", agentID)
	}
	return found[0].project, found[0].agent, nil
}
