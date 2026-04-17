package state

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func writeJSON(path string, v any) error {
	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o600)
}

func TestValidateAgentID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "lowercase letters", id: "planner", wantErr: false},
		{name: "with dash", id: "planner-2", wantErr: false},
		{name: "with underscore", id: "plan_2", wantErr: false},
		{name: "numeric", id: "123", wantErr: false},
		{name: "starts with dash", id: "-bad", wantErr: true},
		{name: "upper case", id: "Bad", wantErr: true},
		{name: "punctuation", id: "bad/id", wantErr: true},
		{name: "space", id: "bad id", wantErr: true},
		{name: "empty", id: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAgentID(tc.id)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAgentID(%q) error = %v, wantErr %v", tc.id, err, tc.wantErr)
			}
		})
	}
}

func TestProjectNameFromRepoAndNormalizePath(t *testing.T) {
	t.Parallel()

	got := ProjectNameFromRepo("/Users/me/projects/myrepo/")
	if got != "myrepo" {
		t.Fatalf("project name: got %q want %q", got, "myrepo")
	}

	abs, err := NormalizePath(".")
	if err != nil {
		t.Fatalf("NormalizePath(\".\") unexpected error: %v", err)
	}
	if !filepath.IsAbs(abs) {
		t.Fatalf("NormalizePath(\".\") expected absolute path, got %q", abs)
	}

	if _, err := NormalizePath(""); err == nil {
		t.Fatalf("expected NormalizePath(\"\") to fail")
	}
}

func TestProjectAndAgentPathHelpers(t *testing.T) {
	t.Parallel()

	stateRoot := filepath.Join(t.TempDir(), "state")
	projectDir := ProjectDir(stateRoot, "myrepo")
	if got := AgentDir(stateRoot, "myrepo", "agent-1"); got != filepath.Join(projectDir, "agents", "agent-1") {
		t.Fatalf("agent dir: got %q want %q", got, filepath.Join(projectDir, "agents", "agent-1"))
	}
	if got := AgentPath(stateRoot, "myrepo", "agent-1"); got != filepath.Join(projectDir, "agents", "agent-1", "agent.json") {
		t.Fatalf("agent path: got %q want %q", got, filepath.Join(projectDir, "agents", "agent-1", "agent.json"))
	}
	if got := ProjectPath(stateRoot, "myrepo"); got != filepath.Join(projectDir, "project.json") {
		t.Fatalf("project path: got %q want %q", got, filepath.Join(projectDir, "project.json"))
	}
}

func TestSaveLoadProjectAndLoadProjectsSorted(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	projects := []Project{
		{Name: "zeta", RepoPath: filepath.Join(root, "zeta"), TmuxSession: "litents-zeta"},
		{Name: "alpha", RepoPath: filepath.Join(root, "alpha"), TmuxSession: "litents-alpha"},
	}

	for _, p := range projects {
		if err := SaveProject(root, p); err != nil {
			t.Fatalf("SaveProject(%q) unexpected error: %v", p.Name, err)
		}
	}

	got, err := LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects() unexpected error: %v", err)
	}
	want := []string{"alpha", "zeta"}
	if len(got) != len(want) {
		t.Fatalf("project count: got %d want %d", len(got), len(want))
	}
	for i, exp := range want {
		if got[i].Name != exp {
			t.Fatalf("sorted project[%d] = %q want %q", i, got[i].Name, exp)
		}
	}
}

func TestProjectForProjectOrRepo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if _, err := ProjectForProjectOrRepo(root, "", ""); err == nil {
		t.Fatalf("expected error when no projects")
	}

	if err := SaveProject(root, Project{Name: "alpha", RepoPath: filepath.Clean("/repo/alpha")}); err != nil {
		t.Fatalf("SaveProject() unexpected error: %v", err)
	}
	p, err := ProjectForProjectOrRepo(root, "/repo/alpha", "")
	if err != nil {
		t.Fatalf("lookup by repo error: %v", err)
	}
	if p.Name != "alpha" {
		t.Fatalf("repo project: got %q want %q", p.Name, "alpha")
	}

	p, err = ProjectForProjectOrRepo(root, "", "alpha")
	if err != nil {
		t.Fatalf("lookup by project error: %v", err)
	}
	if p.Name != "alpha" {
		t.Fatalf("project name: got %q want %q", p.Name, "alpha")
	}

	p, err = ProjectForProjectOrRepo(root, "", "")
	if err != nil {
		t.Fatalf("fallback lookup expected with single project: %v", err)
	}
	if p.Name != "alpha" {
		t.Fatalf("single project fallback: got %q want %q", p.Name, "alpha")
	}

	if err := SaveProject(root, Project{Name: "bravo", RepoPath: filepath.Clean("/repo/bravo")}); err != nil {
		t.Fatalf("SaveProject() unexpected error: %v", err)
	}
	if _, err := ProjectForProjectOrRepo(root, "", ""); err == nil {
		t.Fatalf("expected error when multiple projects exist")
	}
}

func TestSaveLoadAgentAndResolveAcrossProjects(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	alpha := Project{Name: "alpha", RepoPath: filepath.Clean("/repo/alpha")}
	bravo := Project{Name: "bravo", RepoPath: filepath.Clean("/repo/bravo")}
	if err := SaveProject(root, alpha); err != nil {
		t.Fatalf("save project alpha: %v", err)
	}
	if err := SaveProject(root, bravo); err != nil {
		t.Fatalf("save project bravo: %v", err)
	}

	agentAlpha := Agent{
		ID:          "planner",
		Project:     alpha.Name,
		Role:        "impl",
		WorktreePath: filepath.Join(root, "alpha", "planner"),
		CreatedAt:   time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 4, 16, 12, 0, 1, 0, time.UTC),
	}
	agentBravo := Agent{
		ID:          "writer",
		Project:     bravo.Name,
		Role:        "notes",
		WorktreePath: filepath.Join(root, "bravo", "writer"),
	}

	if err := SaveAgent(root, agentAlpha); err != nil {
		t.Fatalf("save agent alpha: %v", err)
	}
	if err := SaveAgent(root, agentBravo); err != nil {
		t.Fatalf("save agent bravo: %v", err)
	}

	loaded, err := LoadAgent(root, alpha.Name, agentAlpha.ID)
	if err != nil {
		t.Fatalf("LoadAgent() unexpected error: %v", err)
	}
	if loaded.ID != agentAlpha.ID || loaded.Project != alpha.Name {
		t.Fatalf("LoadAgent mismatch: got %#v", loaded)
	}

	agents, err := LoadAgents(root, alpha.Name)
	if err != nil {
		t.Fatalf("LoadAgents() unexpected error: %v", err)
	}
	if len(agents) != 1 || agents[0].ID != "planner" {
		t.Fatalf("LoadAgents list mismatch: %#v", agents)
	}

	_, _, err = ResolveAgentAcrossProjects(root, "", "planner")
	if err != nil {
		t.Fatalf("ResolveAgentAcrossProjects expected to find unique planner: %v", err)
	}

	if err := SaveAgent(root, Agent{ID: "planner", Project: bravo.Name}); err != nil {
		t.Fatalf("save duplicate planner in bravo: %v", err)
	}
	_, _, err = ResolveAgentAcrossProjects(root, "", "planner")
	if err == nil {
		t.Fatalf("expected error for duplicate agent IDs across projects")
	}

	if _, err := LoadProject(root, "missing"); err == nil {
		t.Fatalf("expected missing project load to fail")
	}

	all, err := LoadAllAgents(root)
	if err != nil {
		t.Fatalf("LoadAllAgents() unexpected error: %v", err)
	}
	if len(all) < 2 {
		t.Fatalf("LoadAllAgents expected >=2 entries, got %d", len(all))
	}

	if err := DeleteAgent(root, bravo.Name, "writer"); err != nil {
		t.Fatalf("DeleteAgent() unexpected error: %v", err)
	}
	if _, err := LoadAgent(root, bravo.Name, "writer"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected deleted agent file missing, got err=%v", err)
	}
}

func TestFindProjectByRepoAndResolveAgentMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := Project{Name: "myrepo", RepoPath: filepath.Clean("/repo/myrepo")}
	if err := SaveProject(root, p); err != nil {
		t.Fatalf("save project: %v", err)
	}

	if _, err := FindProjectByRepo(root, "/repo/nope"); err == nil {
		t.Fatalf("expected no repo match")
	}

	if _, _, err := ResolveAgentAcrossProjects(root, "myrepo", "missing"); err == nil {
		t.Fatalf("expected missing agent error")
	}
}

func TestLoadProjectsWithInvalidEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	base := filepath.Join(root, "projects", "alpha")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(base, "project.json"), []byte("{bad"), 0o600); err != nil {
		t.Fatalf("write invalid project: %v", err)
	}

	got, err := LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects() unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected invalid project to be skipped, got %d entries", len(got))
	}

	agents, err := LoadAgents(root, "missing")
	if err != nil {
		t.Fatalf("LoadAgents(missing) unexpected error: %v", err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected zero agents for missing project, got %d", len(agents))
	}
}

func TestLoadAllAgentsSortsByLoadOrder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := SaveProject(root, Project{Name: "myrepo"}); err != nil {
		t.Fatalf("save project: %v", err)
	}
	if err := SaveAgent(root, Agent{ID: "z", Project: "myrepo", WorktreePath: "/tmp"}); err != nil {
		t.Fatalf("save agent z: %v", err)
	}
	if err := SaveAgent(root, Agent{ID: "a", Project: "myrepo", WorktreePath: "/tmp"}); err != nil {
		t.Fatalf("save agent a: %v", err)
	}

	agents, err := LoadAgents(root, "myrepo")
	if err != nil {
		t.Fatalf("LoadAgents() unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("agent count: got %d want %d", len(agents), 2)
	}
	if got := []string{agents[0].ID, agents[1].ID}; !reflect.DeepEqual(got, []string{"a", "z"}) {
		t.Fatalf("agent sort order: got %v want [a z]", got)
	}
}
