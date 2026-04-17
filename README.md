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

### Competitive comparison vs popular local managers

Litents is benchmarked against local orchestration baselines using the same synthetic workload (`sleep 0.45; echo done`) with no model or network calls:

Command:

```bash
./benchmarking_md/compare-with-popular-tools.sh
```

Latest run (20 repeats, macOS, darwin/arm64, go1.26.2, tmux 3.6a, Zellij 0.44.1, Codex CLI 0.120.0):

| Metric | Litents | tmux | Zellij | Codex app-server |
| --- | ---: | ---: | ---: | ---: |
| Initialize control surface | 20 runs, mean=38.80ms (p50=22ms, p95=28ms, min=20ms, max=358ms) | 20 runs, mean=7.05ms (p50=7ms, p95=10ms, min=5ms, max=13ms) | 20 runs, mean=49.10ms (p50=47ms, p95=59ms, min=43ms, max=70ms) | 20 runs, mean=41.70ms (p50=42ms, p95=44ms, min=38ms, max=49ms) |
| Start one workload | 20 runs, mean=16.95ms (p50=16ms, p95=21ms, min=14ms, max=24ms) | 20 runs, mean=8.65ms (p50=6ms, p95=21ms, min=5ms, max=28ms) | 20 runs, mean=49.10ms (p50=46ms, p95=65ms, min=41ms, max=89ms) | N/A |
| Status/list/health poll | 20 runs, mean=9.45ms (p50=9ms, p95=11ms, min=8ms, max=11ms) | 20 runs, mean=5.50ms (p50=5ms, p95=10ms, min=4ms, max=13ms) | 20 runs, mean=20.55ms (p50=20ms, p95=25ms, min=16ms, max=30ms) | 20 runs, mean=7.75ms (p50=7ms, p95=10ms, min=6ms, max=22ms) |
| Stop control surface | 20 runs, mean=724.40ms (p50=724ms, p95=728ms, min=721ms, max=728ms) | 20 runs, mean=6.05ms (p50=5ms, p95=9ms, min=5ms, max=14ms) | 20 runs, mean=13.80ms (p50=13ms, p95=14ms, min=12ms, max=31ms) | 20 runs, mean=6.05ms (p50=3ms, p95=6ms, min=3ms, max=56ms) |
| Cleanup state files | 20 runs, mean=37.25ms (p50=39ms, p95=45ms, min=28ms, max=49ms) | N/A | N/A | N/A |

### Readable summary

| Area | Takeaway |
| --- | --- |
| Startup | Litents starts faster than Zellij in this run and close to Codex app-server, while doing project state initialization. |
| Workload launch | Litents is about 2x raw tmux, but roughly 3x faster than Zellij for this synthetic one-tab workload. |
| Status polling | Litents is single-digit milliseconds while reading agent state and refreshing status. |
| Stop behavior | Litents is intentionally slower because `stop` waits for graceful interrupt before force cleanup. |

The `tmux` control-plane benchmark is intentionally a lower bound. Litents is not trying to beat `tmux` at being `tmux`; it is trying to stay terminal-native while adding project metadata, agent logs, worktree isolation, status tracking, resume, and cleanup.

The Codex desktop app itself is not measured because a macOS GUI app cannot be driven reproducibly in this headless shell benchmark. The harness measures `codex app-server`, which is the local headless server substrate exposed by Codex CLI.
Full comparison output is in [benchmarking_md/tool-comparison-results.md](benchmarking_md/tool-comparison-results.md).

### Competitive landscape

| Product | Model | What it is good at | How Litents should compare |
| --- | --- | --- | --- |
| [tmux](https://github.com/tmux/tmux/wiki) | Terminal multiplexer | Fastest local session/window primitive | Performance floor, not feature peer |
| [Zellij](https://zellij.dev/features/) | Terminal workspace | Modern terminal sessions, layouts, resurrection | Terminal-native session peer with heavier workspace UX |
| [Codex CLI / app](https://openai.com/codex) | Agent CLI + app | First-party Codex agent across CLI, app, IDE, cloud | Agent runtime that Litents can orchestrate locally |
| [Claude Code](https://www.claude.com/product/claude-code) | Agent CLI | Strong single-agent coding loop and git workflow | Agent runtime that Litents can orchestrate locally |
| [Gemini CLI](https://github.com/google-gemini/gemini-cli) | Agent CLI | Google-backed terminal agent and GitHub workflow path | Agent runtime baseline, not orchestration peer |
| [Cursor Agent CLI](https://docs.cursor.com/en/cli) | Agent CLI | Headless Cursor agent in terminal/automation | Agent runtime baseline, not orchestration peer |
| [Aider](https://github.com/Aider-AI/aider) | Agent CLI | Git-centric terminal pair programming | Single-agent baseline Litents can run inside panes |
| [OpenCode](https://opencode.ai/) | Agent CLI/app/IDE | Open source terminal, desktop, and IDE coding agent | Agent runtime baseline and potential managed command |
| [Claude Squad](https://github.com/smtg-ai/claude-squad) | Multi-agent TUI | Manages multiple agent sessions in separate workspaces | Direct local multi-agent manager competitor |
| [Crystal](https://github.com/stravu/crystal) | Desktop worktree manager | Multiple Codex/Claude sessions in isolated worktrees | Direct competitor with richer desktop UX |
| [CCManager](https://github.com/kbwo/ccmanager) | Multi-agent manager | Multi-agent sessions across worktrees and projects | Direct competitor with broader agent support |
| [Conductor](https://docs.conductor.build/) | Mac app | Parallel Codex/Claude agents in isolated workspaces | Productized desktop competitor |
| [CodeAgentSwarm](https://www.codeagentswarm.com/) | Agent workspace | Multi-terminal visibility, history, task integrations | More integrated workspace competitor |
| [Termyx](https://termyx.dev/) | Worktree IDE | Claude Code status per worktree in a GUI | GUI worktree/session competitor |
| [Sidecar Workspaces](https://sidecar.haplab.com/docs/workspaces-plugin) | Workspace plugin | Worktree branches, optional agents, review/merge flow | Workflow-heavy competitor |
| [Agent Hand](https://weykon.github.io/agent-hand/) | tmux-backed manager | Fast terminal session manager for coding agents | Close terminal-native peer |
| [Agent Deck](https://sushantvema.github.io/notes/agent_deck) | tmux command center | Visibility across many tmux-backed agent sessions | Session-visibility peer |
| [Agent of Empires](https://www.agent-of-empires.com/docs/index.html) | tmux/worktree manager | Parallel agents with branches, worktrees, optional Docker | Heavier sandbox-capable peer |
| [Goose](https://goose-docs.ai/) | Local agent runtime | General-purpose desktop/CLI/API agent with MCP | Heavier runtime/platform comparison |
| [OpenHands](https://docs.openhands.dev/sdk/index) | Agent platform/SDK | Local/remote software-agent execution platform | Heavyweight platform comparison |
| [Factory](https://factory.ai/) | Cloud/enterprise agents | Delegated coding tasks, PRs, traceability, analytics | Enterprise workflow comparison |

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
