package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// Paths captures XDG-style roots used by litents.
type Paths struct {
	ConfigRoot string
	StateRoot  string
}

// Default returns the filesystem roots for config and state.
func Default() Paths {
	cfgHome := envOrDefault("XDG_CONFIG_HOME", filepath.Join(userHome(), ".config"))
	stateHome := envOrDefault("XDG_STATE_HOME", filepath.Join(userHome(), ".local", "state"))

	return Paths{
		ConfigRoot: filepath.Join(cfgHome, "litents"),
		StateRoot:  filepath.Join(stateHome, "litents"),
	}
}

func (p Paths) ConfigPath() string {
	return filepath.Join(p.ConfigRoot, "config.json")
}

func (p Paths) ProjectRoot(name string) string {
	return filepath.Join(p.StateRoot, "projects", name)
}

func (p Paths) WorktreeRoot(fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return ExpandPath(fallback)
	}
	return filepath.Join(dataHome(), "worktrees")
}

func dataHome() string {
	dataHome := envOrDefault("XDG_DATA_HOME", filepath.Join(userHome(), ".local", "share"))
	return filepath.Join(dataHome, "litents")
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

// ExpandPath expands ~ and environment variables in path strings.
func ExpandPath(in string) string {
	if in == "" {
		return in
	}

	expanded := os.ExpandEnv(in)
	if strings.HasPrefix(expanded, "~") {
		home := userHome()
		expanded = filepath.Join(home, strings.TrimPrefix(expanded, "~"))
	}

	expanded = filepath.Clean(expanded)
	return expanded
}
