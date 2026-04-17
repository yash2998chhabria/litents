package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadMissingFileReturnsDefaultConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing-config.json")
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := DefaultConfig()
	if got.TmuxSessionPrefix != want.TmuxSessionPrefix {
		t.Fatalf("tmux session prefix: got %q want %q", got.TmuxSessionPrefix, want.TmuxSessionPrefix)
	}
	if got.WorktreeRoot != want.WorktreeRoot {
		t.Fatalf("worktree root: got %q want %q", got.WorktreeRoot, want.WorktreeRoot)
	}
	if got.CodexCommand != want.CodexCommand {
		t.Fatalf("codex command: got %q want %q", got.CodexCommand, want.CodexCommand)
	}
	if got.DefaultBaseBranch != want.DefaultBaseBranch {
		t.Fatalf("default base branch: got %q want %q", got.DefaultBaseBranch, want.DefaultBaseBranch)
	}
	if got.NotifyCommand != want.NotifyCommand {
		t.Fatalf("notify command: got %q want %q", got.NotifyCommand, want.NotifyCommand)
	}
	if got.WatchIntervalSeconds != want.WatchIntervalSeconds {
		t.Fatalf("watch interval: got %d want %d", got.WatchIntervalSeconds, want.WatchIntervalSeconds)
	}
	if got.SilenceThresholdSeconds != want.SilenceThresholdSeconds {
		t.Fatalf("silence threshold: got %d want %d", got.SilenceThresholdSeconds, want.SilenceThresholdSeconds)
	}
	if got.ActivityNotifyCooldownSec != want.ActivityNotifyCooldownSec {
		t.Fatalf("activity cooldown: got %d want %d", got.ActivityNotifyCooldownSec, want.ActivityNotifyCooldownSec)
	}
	if !reflect.DeepEqual(got.WaitingRegexes, want.WaitingRegexes) {
		t.Fatalf("waiting regexes mismatch")
	}
	if !reflect.DeepEqual(got.DoneRegexes, want.DoneRegexes) {
		t.Fatalf("done regexes mismatch")
	}
	if got.UpdatedAt.IsZero() {
		t.Fatalf("updated at should be set")
	}
}

func TestLoadInvalidJSONReturnsError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad-config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
}

	if _, err := Load(path); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestLoadAppliesDefaultsForMissingFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "partial-config.json")
	payload := `{
  "tmux_session_prefix": "cli",
  "worktree_root": "",
  "codex_command": "",
  "watch_interval_seconds": 0,
  "silence_threshold_seconds": 0,
  "activity_notify_cooldown_seconds": 0,
  "waiting_regexes": ["custom"],
  "done_regexes": ["custom done"]
}`
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := DefaultConfig()
	if got.TmuxSessionPrefix != "cli" {
		t.Fatalf("tmux session prefix: got %q want %q", got.TmuxSessionPrefix, "cli")
	}
	if got.WorktreeRoot != want.WorktreeRoot {
		t.Fatalf("worktree root: got %q want %q", got.WorktreeRoot, want.WorktreeRoot)
	}
	if got.CodexCommand != want.CodexCommand {
		t.Fatalf("codex command: got %q want %q", got.CodexCommand, want.CodexCommand)
	}
	if got.WatchIntervalSeconds != want.WatchIntervalSeconds {
		t.Fatalf("watch interval: got %d want %d", got.WatchIntervalSeconds, want.WatchIntervalSeconds)
	}
	if got.SilenceThresholdSeconds != want.SilenceThresholdSeconds {
		t.Fatalf("silence threshold: got %d want %d", got.SilenceThresholdSeconds, want.SilenceThresholdSeconds)
	}
	if got.ActivityNotifyCooldownSec != want.ActivityNotifyCooldownSec {
		t.Fatalf("activity cooldown: got %d want %d", got.ActivityNotifyCooldownSec, want.ActivityNotifyCooldownSec)
	}
	if !reflect.DeepEqual(got.WaitingRegexes, []string{"custom"}) {
		t.Fatalf("waiting regexes: got %#v want %#v", got.WaitingRegexes, []string{"custom"})
	}
	if !reflect.DeepEqual(got.DoneRegexes, []string{"custom done"}) {
		t.Fatalf("done regexes custom value not preserved: got %#v", got.DoneRegexes)
	}
}

func TestSaveAppliesDefaultsAndUpdatesTimestamp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config", "config.json")
	cfg := Config{
		TmuxSessionPrefix:         "custom",
		DefaultBaseBranch:         "main",
		CodexCommand:             "",
		WorktreeRoot:              "",
		WatchIntervalSeconds:      3,
		SilenceThresholdSeconds:   180,
		ActivityNotifyCooldownSec: 120,
		NotifyEnabled:             true,
		NotifyCommand:             "auto",
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	var got Config
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode config: %v", err)
	}

	if got.TmuxSessionPrefix != "custom" {
		t.Fatalf("tmux session prefix: got %q want %q", got.TmuxSessionPrefix, "custom")
	}
	want := DefaultConfig()
	if got.WorktreeRoot != want.WorktreeRoot {
		t.Fatalf("worktree root: got %q want %q", got.WorktreeRoot, want.WorktreeRoot)
	}
	if got.CodexCommand != want.CodexCommand {
		t.Fatalf("codex command: got %q want %q", got.CodexCommand, want.CodexCommand)
	}
	if got.UpdatedAt.IsZero() {
		t.Fatalf("updated at should be set")
	}
	if got.UpdatedAt.After(time.Now().Add(time.Second)) {
		t.Fatalf("updated at should not be far in the future")
	}
}

func TestEnsureDefaultsPopulatesMissingValues(t *testing.T) {
	t.Parallel()

	input := Config{
		NotifyEnabled: true,
		UpdatedAt:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	got := EnsureDefaults(input)

	want := DefaultConfig()
	if got.TmuxSessionPrefix != want.TmuxSessionPrefix {
		t.Fatalf("tmux session prefix: got %q want %q", got.TmuxSessionPrefix, want.TmuxSessionPrefix)
	}
	if strings.TrimSpace(got.WorktreeRoot) == "" {
		t.Fatalf("worktree root should not be blank")
	}
	if got.WorktreeRoot != want.WorktreeRoot {
		t.Fatalf("worktree root: got %q want %q", got.WorktreeRoot, want.WorktreeRoot)
	}
	if got.DefaultBaseBranch != want.DefaultBaseBranch {
		t.Fatalf("default base branch: got %q want %q", got.DefaultBaseBranch, want.DefaultBaseBranch)
	}
	if got.CodexCommand != want.CodexCommand {
		t.Fatalf("codex command: got %q want %q", got.CodexCommand, want.CodexCommand)
	}
	if got.NotifyCommand != want.NotifyCommand {
		t.Fatalf("notify command: got %q want %q", got.NotifyCommand, want.NotifyCommand)
	}
	if got.WatchIntervalSeconds != want.WatchIntervalSeconds {
		t.Fatalf("watch interval: got %d want %d", got.WatchIntervalSeconds, want.WatchIntervalSeconds)
	}
	if got.SilenceThresholdSeconds != want.SilenceThresholdSeconds {
		t.Fatalf("silence threshold: got %d want %d", got.SilenceThresholdSeconds, want.SilenceThresholdSeconds)
	}
	if got.ActivityNotifyCooldownSec != want.ActivityNotifyCooldownSec {
		t.Fatalf("activity cooldown: got %d want %d", got.ActivityNotifyCooldownSec, want.ActivityNotifyCooldownSec)
	}
	if len(got.WaitingRegexes) == 0 || !reflect.DeepEqual(got.WaitingRegexes, want.WaitingRegexes) {
		t.Fatalf("waiting regex defaults not applied")
	}
	if len(got.DoneRegexes) == 0 || !reflect.DeepEqual(got.DoneRegexes, want.DoneRegexes) {
		t.Fatalf("done regex defaults not applied")
	}
	if !got.UpdatedAt.After(input.UpdatedAt) {
		t.Fatalf("updated at not refreshed")
	}
}
