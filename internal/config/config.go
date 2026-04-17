package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config mirrors the JSON schema described in litents.md.
type Config struct {
	TmuxSessionPrefix         string   `json:"tmux_session_prefix"`
	WorktreeRoot              string   `json:"worktree_root"`
	DefaultBaseBranch         string   `json:"default_base_branch"`
	CodexCommand              string   `json:"codex_command"`
	CodexArgs                 []string `json:"codex_args"`
	NotifyEnabled             bool     `json:"notify_enabled"`
	NotifyCommand             string   `json:"notify_command"`
	WatchIntervalSeconds      int      `json:"watch_interval_seconds"`
	SilenceThresholdSeconds   int      `json:"silence_threshold_seconds"`
	ActivityNotifyCooldownSec int      `json:"activity_notify_cooldown_seconds"`
	NotifyOnQuiet             bool     `json:"notify_on_quiet"`
	WaitingRegexes            []string `json:"waiting_regexes"`
	DoneRegexes               []string `json:"done_regexes"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

var defaultConfig = Config{
	TmuxSessionPrefix:         "litents",
	WorktreeRoot:              "~/.local/share/litents/worktrees",
	DefaultBaseBranch:         "main",
	CodexCommand:              "codex",
	CodexArgs:                 nil,
	NotifyEnabled:             true,
	NotifyCommand:             "auto",
	WatchIntervalSeconds:      3,
	SilenceThresholdSeconds:   180,
	ActivityNotifyCooldownSec: 120,
	NotifyOnQuiet:             false,
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
	UpdatedAt: time.Now().UTC(),
}

func DefaultConfig() Config {
	cfg := defaultConfig
	cfg.UpdatedAt = time.Now().UTC()
	return cfg
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config json: %w", err)
	}

	if cfg.WorktreeRoot == "" {
		cfg.WorktreeRoot = defaultConfig.WorktreeRoot
	}
	if cfg.CodexCommand == "" {
		cfg.CodexCommand = defaultConfig.CodexCommand
	}
	if cfg.DefaultBaseBranch == "" {
		cfg.DefaultBaseBranch = defaultConfig.DefaultBaseBranch
	}
	if cfg.TmuxSessionPrefix == "" {
		cfg.TmuxSessionPrefix = defaultConfig.TmuxSessionPrefix
	}
	if cfg.WatchIntervalSeconds <= 0 {
		cfg.WatchIntervalSeconds = defaultConfig.WatchIntervalSeconds
	}
	if cfg.SilenceThresholdSeconds <= 0 {
		cfg.SilenceThresholdSeconds = defaultConfig.SilenceThresholdSeconds
	}
	if cfg.ActivityNotifyCooldownSec <= 0 {
		cfg.ActivityNotifyCooldownSec = defaultConfig.ActivityNotifyCooldownSec
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if cfg.CodexCommand == "" {
		cfg.CodexCommand = defaultConfig.CodexCommand
	}
	if cfg.WorktreeRoot == "" {
		cfg.WorktreeRoot = defaultConfig.WorktreeRoot
	}
	cfg.UpdatedAt = time.Now().UTC()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o600)
}

func EnsureDefaults(cfg Config) Config {
	if cfg.TmuxSessionPrefix == "" {
		cfg.TmuxSessionPrefix = defaultConfig.TmuxSessionPrefix
	}
	if strings.TrimSpace(cfg.WorktreeRoot) == "" {
		cfg.WorktreeRoot = defaultConfig.WorktreeRoot
	}
	if cfg.DefaultBaseBranch == "" {
		cfg.DefaultBaseBranch = defaultConfig.DefaultBaseBranch
	}
	if cfg.CodexCommand == "" {
		cfg.CodexCommand = defaultConfig.CodexCommand
	}
	if cfg.NotifyCommand == "" {
		cfg.NotifyCommand = defaultConfig.NotifyCommand
	}
	if cfg.WaitingRegexes == nil || len(cfg.WaitingRegexes) == 0 {
		cfg.WaitingRegexes = defaultConfig.WaitingRegexes
	}
	if cfg.DoneRegexes == nil || len(cfg.DoneRegexes) == 0 {
		cfg.DoneRegexes = defaultConfig.DoneRegexes
	}
	if cfg.WatchIntervalSeconds == 0 {
		cfg.WatchIntervalSeconds = defaultConfig.WatchIntervalSeconds
	}
	if cfg.SilenceThresholdSeconds == 0 {
		cfg.SilenceThresholdSeconds = defaultConfig.SilenceThresholdSeconds
	}
	if cfg.ActivityNotifyCooldownSec == 0 {
		cfg.ActivityNotifyCooldownSec = defaultConfig.ActivityNotifyCooldownSec
	}
	cfg.UpdatedAt = time.Now().UTC()
	return cfg
}
