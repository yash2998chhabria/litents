package config

import (
	"path/filepath"
	"testing"
)

func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "config.json")
	payload := DefaultConfig()
	if err := Save(path, payload); err != nil {
		b.Fatalf("Save: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Load(path); err != nil {
			b.Fatalf("Load: %v", err)
		}
	}
}

func BenchmarkSaveConfig(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "config.json")
	cfg := DefaultConfig()
	cfg.CodexCommand = "codex"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.WorktreeRoot = "/tmp/worktrees/bench" + string('a'+rune(i%26))
		if err := Save(path, cfg); err != nil {
			b.Fatalf("Save: %v", err)
		}
	}
}
