package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeProject(i int) Project {
	return Project{
		Name:        fmt.Sprintf("project-%03d", i),
		RepoPath:    fmt.Sprintf("/tmp/repo-%03d", i),
		TmuxSession: fmt.Sprintf("litents-project-%03d", i),
		CreatedAt:   time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC),
	}
}

func makeAgent(i int, project string) Agent {
	return Agent{
		ID:            fmt.Sprintf("agent-%04d", i),
		Project:       project,
		WorktreePath:  fmt.Sprintf("/tmp/worktrees/%s/agent-%04d", project, i),
		Branch:        fmt.Sprintf("litents/agent-%04d", i),
		Role:          "impl",
		RepoPath:      "/tmp/repo",
		TmuxSession:   "litents-main",
		TmuxWindow:    fmt.Sprintf("agent-%04d", i),
		TmuxPane:      fmt.Sprintf("%%%d", i+1),
		PromptFile:    filepath.Join("/tmp", project, fmt.Sprintf("agent-%04d", i), "prompt.md"),
		LogFile:       filepath.Join("/tmp", project, fmt.Sprintf("agent-%04d", i), "output.log"),
		Status:        "running",
		LastStatus:    "running",
		LastActivityAt: time.Now(),
		CreatedAt:     time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 4, 16, 12, 0, 1, 0, time.UTC),
	}
}

func BenchmarkSaveLoadProjects(b *testing.B) {
	root := b.TempDir()
	projects := make([]Project, 100)
	for i := 0; i < 100; i++ {
		projects[i] = makeProject(i)
		if err := SaveProject(root, projects[i]); err != nil {
			b.Fatalf("SaveProject: %v", err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := LoadProjects(root); err != nil {
			b.Fatalf("LoadProjects: %v", err)
		}
	}
}

func BenchmarkLoadAgents100(b *testing.B) {
	root := b.TempDir()
	project := "myrepo"
	if err := SaveProject(root, Project{Name: project, RepoPath: "/tmp/myrepo", TmuxSession: "litents-myrepo"}); err != nil {
		b.Fatalf("SaveProject: %v", err)
	}
	for i := 0; i < 100; i++ {
		agent := makeAgent(i, project)
		if err := SaveAgent(root, agent); err != nil {
			b.Fatalf("SaveAgent: %v", err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agents, err := LoadAgents(root, project)
		if err != nil {
			b.Fatalf("LoadAgents: %v", err)
		}
		if len(agents) != 100 {
			b.Fatalf("expected 100 agents, got %d", len(agents))
		}
	}
}

func BenchmarkSaveAgentOverwrite100x(b *testing.B) {
	root := b.TempDir()
	project := "myrepo"
	if err := SaveProject(root, Project{Name: project, RepoPath: "/tmp/myrepo", TmuxSession: "litents-myrepo"}); err != nil {
		b.Fatalf("SaveProject: %v", err)
	}
	base := makeAgent(1, project)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent := base
		agent.UpdatedAt = time.Date(2026, 4, 16, 12, 0, i%60, 0, time.UTC)
		if err := SaveAgent(root, agent); err != nil {
			b.Fatalf("SaveAgent: %v", err)
		}
	}
	_ = os.RemoveAll(root)
}
