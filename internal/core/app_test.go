package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "zero", d: 0, want: "0s"},
		{name: "seconds", d: 12 * time.Second, want: "12s"},
		{name: "minutes", d: 3*time.Minute + 4*time.Second, want: "3m"},
		{name: "hours", d: 2*time.Hour + 1*time.Minute, want: "2h"},
		{name: "days", d: 2*24*time.Hour + 3*time.Hour, want: "2d"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDuration(tc.d); got != tc.want {
				t.Fatalf("formatDuration(%v): got %q want %q", tc.d, got, tc.want)
			}
		})
	}
}

func TestMatchesAny(t *testing.T) {
	t.Parallel()

	a := &App{}
	if !a.matchesAny("approval requested", []string{"approval", "done"}) {
		t.Fatalf("expected approval regex to match")
	}
	if a.matchesAny("no match", []string{"approval", "done"}) {
		t.Fatalf("expected no match")
	}
}

func TestMatchDoneLog(t *testing.T) {
	t.Parallel()

	if !matchDoneLog("task complete now", []string{"(?i)task complete"}) {
		t.Fatalf("expected done regex match")
	}
	if matchDoneLog("still running", []string{"(?i)task complete"}) {
		t.Fatalf("expected no done regex match")
	}
}

func TestMatchExitStatus(t *testing.T) {
	t.Parallel()

	if got := matchExitStatus("[litents] codex exited with status 0", statusFailed, statusDone); got != statusDone {
		t.Fatalf("status mismatch: got %q", got)
	}
	if got := matchExitStatus("[litents] codex exited with status 42", statusFailed, statusDone); got != statusFailed {
		t.Fatalf("status mismatch: got %q", got)
	}
	if got := matchExitStatus("", statusFailed, statusDone); got != statusDone {
		t.Fatalf("default status mismatch: got %q", got)
	}
}

func TestIsQuiet(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	app := &App{
		Now: func() time.Time { return base.Add(190 * time.Second) },
		Config: Config{SilenceThresholdSeconds: 120},
	}
	if !app.isQuiet(base, filepath.Join(t.TempDir(), "missing.log")) {
		t.Fatalf("expected quiet when last activity is older than threshold")
	}
}

func TestReadFileTailSafe(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.log")
	if err := os.WriteFile(path, []byte("a\nb\nc\nd\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	got := readFileTailSafe(path, 2)
	if got != "c\nd" {
		t.Fatalf("readFileTailSafe: got %q want %q", got, "c\nd")
	}
}
