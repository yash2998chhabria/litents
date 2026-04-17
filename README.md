# Litents

A lightweight Unix-native CLI for running multiple Codex agents in parallel with tmux + git worktrees.

[![Go Test](https://github.com/yash2998chhabria/litents/actions/workflows/ci.yml/badge.svg)](https://github.com/yash2998chhabria/litents/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/yash2998chhabria/litents)](https://github.com/yash2998chhabria/litents/releases)

## Why Litents

Litents keeps Codex workflows fast, inspectable, and shell-native:

- one light binary
- simple JSON state on disk
- one tmux session per repository
- one window per agent
- per-agent logs, prompts, and metadata persisted in state

## Features

- `init`: initialize a repo and create a tmux session
- `new`: launch a new agent in its own worktree
- `status | ls`: show agent table with status
- `watch`: poll agent states and notify on state transitions
- `attach`: jump to a running agent window
- `send`: send text / input to an agent
- `tail`: read latest agent output
- `notify test`: validate notification backend
- `resume`: recover a dead/closed pane
- `history`: list previous agents
- `stop`: stop an agent pane
- `clean`: prune finished agents and optionally remove worktrees

## Repository layout

```text
litents/
├─ cmd/litents/main.go
├─ internal/
│  ├─ core/
│  ├─ config/
│  ├─ gitx/
│  ├─ notify/
│  ├─ pathutil/
│  ├─ runner/
│  ├─ state/
│  └─ tmux/
├─ .github/workflows/
├─ litents.md
└─ README.md
```

## Requirements

- Go 1.22+
- `tmux`
- `git`
- `codex` CLI auth'd on machine

## Quick start

```bash
# initialize repository for orchestration
litents init ~/code/myrepo

# create a new agent
litents new planner --prompt "Inspect repository and draft a fix strategy."

# list and watch all agents
litents status
litents watch --project myrepo

# view output
litents tail planner
```

## Data paths

Defaults:

- State: `${XDG_STATE_HOME:-$HOME/.local/state}/litents`
- Config: `${XDG_CONFIG_HOME:-$HOME/.config}/litents/config.json`

Example per-agent files:

```text
~/.local/state/litents/projects/<project>/agents/<agent-id>/
  ├─ agent.json
  ├─ prompt.md
  └─ output.log
```

## Configuration

Defaults live in:

- `~/.config/litents/config.json`

Key options:

- `tmux_session_prefix`
- `worktree_root`
- `default_base_branch`
- `codex_command`
- `codex_args`
- `notify_enabled`
- `notify_command`
- `watch_interval_seconds`
- `silence_threshold_seconds`
- `activity_notify_cooldown_seconds`
- `waiting_regexes`
- `done_regexes`

## Notifications

Notification backend is controlled by `notify_command`:

- `auto` picks platform defaults
- custom command supports placeholders:
  - `{{project}}`, `{{agent}}`, `{{status}}`, `{{message}}`

## Development

Run tests:

```bash
go test ./...
```

Build locally:

```bash
go build -o litents ./cmd/litents
```



## Benchmarks

### Why these were benchmarked

Benchmarking here is focused on core Litents overhead so we can separate orchestrator overhead from external model latency. The benchmark plan is in [benchmarking_md/litents-benchmark.md](benchmarking_md/litents-benchmark.md).

### Benchmark scope

We run package-level microbenchmarks for config, core parsing/refresh helpers, and state persistence to quantify the control-plane cost before and independent of real `codex` runtime.

### Command

```bash
go test ./... -bench . -run '^$' -benchmem
```

### Latest results (macOS, go1.26.2, apple m4)

- `internal/config`
  - `BenchmarkDefaultConfig`: `33.33 ns/op`, `0 B/op`, `0 allocs/op`
  - `BenchmarkLoadConfig`: `15.84 µs/op`, `2112 B/op`, `33 allocs/op`
  - `BenchmarkSaveConfig`: `36.97 µs/op`, `2916 B/op`, `10 allocs/op`
- `internal/core`
  - `BenchmarkFormatDuration`: `28.39 ns/op`, `2 B/op`, `0 allocs/op`
  - `BenchmarkMatchesAny`: `656.7 ns/op`, `2105 B/op`, `18 allocs/op`
  - `BenchmarkMatchDoneLog`: `876.7 ns/op`, `2153 B/op`, `17 allocs/op`
  - `BenchmarkIsQuiet`: `11.29 ns/op`, `0 B/op`, `0 allocs/op`
  - `BenchmarkReadFileTailSafe`: `31.80 µs/op`, `9224 B/op`, `1006 allocs/op`
- `internal/state`
  - `BenchmarkSaveLoadProjects` (100 projects): `1.31 ms/op`, `176.7 KB/op`, `1719 allocs/op`
  - `BenchmarkLoadAgents` (100 agents): `1.66 ms/op`, `271.2 KB/op`, `2820 allocs/op`
  - `BenchmarkSaveAgentOverwrite100x`: `38.4 µs/op`, `3180 B/op`, `14 allocs/op`

### End-to-end justification

These results show Litents core orchestration operations are tiny and predictable compared to external work:
- configuration and state helpers are microsecond-scale;
- status/rule matching and log tail sampling stay in the low-microsecond range;
- list/load paths for 100 agents are low-millisecond, validating fast local recovery and history scanning.

That gives us confidence that Litents adds minimal local orchestration overhead and stays lightweight for the operator-facing workflows defined in the project spec.

Full raw output: [benchmarking_md/last_bench_results.md](benchmarking_md/last_bench_results.md)

### Competitive comparison vs popular local managers

Litents is currently benchmarked against a lean `tmux` baseline using the same synthetic workload (`sleep 0.45; echo done`) with no Codex model calls:

Command:

```bash
./benchmarking_md/compare-with-popular-tools.sh
```

Latest run (20 repeats, macOS, darwin/arm64, go1.26.2):

| Metric | Litents | tmux |
| --- | ---: | ---: |
| Initialize project/session | 20 runs, mean=38.20ms (p50=26ms, p95=47ms, min=24ms, max=244ms) | 20 runs, mean=7.25ms (p50=7ms, p95=9ms, min=6ms, max=9ms) |
| Start one agent workload | 20 runs, mean=20.15ms (p50=19ms, p95=25ms, min=18ms, max=27ms) | 20 runs, mean=6.75ms (p50=7ms, p95=8ms, min=6ms, max=8ms) |
| Status/list poll | 20 runs, mean=10.10ms (p50=10ms, p95=12ms, min=9ms, max=13ms) | 20 runs, mean=5.30ms (p50=5ms, p95=6ms, min=4ms, max=6ms) |
| Stop/cleanup command | 20 runs, mean=725.30ms (p50=726ms, p95=727ms, min=717ms, max=729ms) | 20 runs, mean=5.55ms (p50=5ms, p95=7ms, min=4ms, max=7ms) |
| Cleanup state files | 20 runs, mean=45.75ms (p50=43ms, p95=72ms, min=35ms, max=88ms) | N/A |

### Readable summary

| Area | Litents | tmux baseline | Takeaway |
| --- | --- | --- | --- |
| Control-plane startup | +427% slower than tmux baseline | +0% | expected for richer state/session orchestration |
| Agent launch | +198% slower | +0% | still sub-25ms average for startup path |
| Status polling | +91% slower | +0% | additional refresh + state persistence |
| Stop/cleanup | +13,100%+ | +0% | includes state/status reconciliation + cleanup safety |

The `tmux` control-plane benchmark is intentionally lower bound; Litents currently shows more overhead mostly in state persistence + command lifecycle work (`stop` + `clean`) while preserving richer process state than tmux-only control.

Zellij and Codex app comparisons are not part of the current automated harness yet (both require additional interactive/session handling for reproducible headless execution).
Full comparison output is in [benchmarking_md/tool-comparison-results.md](benchmarking_md/tool-comparison-results.md).

## Validation

Use this command set for end-to-end confidence:

```bash
go test ./...
go test ./... -bench . -run '^$' -benchmem
./benchmarking_md/compare-with-popular-tools.sh
./benchmarking_md/e2e-feature-matrix.sh
```

## Release

CI and release workflows:

- `/.github/workflows/ci.yml`
- `/.github/workflows/release.yml`

Release builds target:

- `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
- artifacts are `.tar.gz` with SHA256 checksums

## Design source

Primary architecture notes are in
[litents.md](litents.md).
