package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{0, 12 * time.Second, 78 * time.Second, 120 * time.Second, 3720 * time.Second, 4 * 24 * time.Hour}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = formatDuration(durations[i%len(durations)])
	}
}

func BenchmarkMatchesAny(b *testing.B) {
	app := &App{}
	text := strings.Repeat("approval needed\n", 128)
	patterns := []string{"approval", "done", "finished", "requires permission", "continue\\?", "press enter"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.matchesAny(text, patterns)
	}
}

func BenchmarkMatchDoneLog(b *testing.B) {
	text := strings.Repeat("task complete and verified\n", 64)
	patterns := []string{"(?i)task complete", "(?i)done", "(?i)finished"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matchDoneLog(text, patterns)
	}
}

func BenchmarkIsQuiet(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "output.log")
	if err := os.WriteFile(path, []byte("initial\n"), 0o644); err != nil {
		b.Fatalf("write output file: %v", err)
	}
	nowBase := time.Now()
	app := &App{
		Now:    func() time.Time { return nowBase.Add(200 * time.Second) },
		Config: Config{SilenceThresholdSeconds: 120},
	}
	lastActivity := nowBase.Add(-180 * time.Second)
	for i := 0; i < b.N; i++ {
		_ = app.isQuiet(lastActivity, path)
	}
}

func BenchmarkReadFileTailSafe(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "output.log")
	lines := make([]byte, 0, 4096)
	for i := 0; i < 1000; i++ {
		lines = append(lines, "line\n"...)
	}
	if err := os.WriteFile(path, lines, 0o644); err != nil {
		b.Fatalf("write output file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readFileTailSafe(path, 40)
	}
}
