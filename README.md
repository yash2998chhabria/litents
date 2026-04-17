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
- `dash`: open the terminal dashboard; `litents` with no args should land here
- `discover`: scan tmux and surface unmanaged Codex-like panes
- `adopt`: track an existing Codex pane without relaunching it
- `untrack`: remove Litents tracking without killing the underlying pane
- `status | ls`: show the agent table with status and attention-aware fields
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

# open the dashboard
litents

# list and watch all agents
litents status
litents watch --project myrepo

# discover or adopt existing panes
litents discover
litents adopt %12 --project myrepo --id existing-codex

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
  ├─ events.jsonl
  └─ output.log
```

The agent metadata carries attention-aware fields such as `source`, `attention_reason`, `attention_excerpt`, and `attention_since` so the dashboard and notifications can explain why a session needs help.

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

### Performance comparison

The main comparison is steady-state CPU/RAM while agents are already running. That is the practical question for day-to-day use: how much does the supervisor cost while you are living in it?

Latest run: 5 synthetic agents per tool, 5 samples, macOS `darwin/arm64`.

| Tool | Running agents | Mean RAM RSS | Mean CPU |
| --- | ---: | ---: | ---: |
| Litents | 5 | 35.72 MiB | 0.00% |
| Zellij | 5 | 149.26 MiB | 0.40% |
| Agent of Empires | 5 | 18.27 MiB | 0.00% |

Method: each tool launches five synthetic `sleep` agents; the harness samples the tool runtime plus managed pane process trees five times at one-second intervals. Litents has no resident daemon, so its running cost is the private tmux server plus managed panes.

### Control-plane latency

The lifecycle benchmark measures local orchestration commands with no model or network calls.

| Operation | Litents mean | Zellij mean | Agent of Empires mean |
| --- | ---: | ---: | ---: |
| Start one workload | 18.05 ms | 52.55 ms | 64.45 ms |
| Status/list poll | 10.20 ms | 21.30 ms | 23.95 ms |
| Dashboard render | 19.75 ms | N/A | N/A |
| Discover unmanaged panes | 11.55 ms | N/A | N/A |
| Adopt unmanaged pane | 35.10 ms | N/A | N/A |

Notes: raw `tmux` is not shown as a competitor because it is Litents' runtime substrate. Codex desktop is not shown because GUI apps are not reproducibly driven in this headless benchmark; the full harness includes a `codex app-server` substrate measurement.

Full benchmark data:

- [Running agents CPU/RAM](benchmarking_md/running-agents-resource-results.md)
- [Lifecycle latency](benchmarking_md/tool-comparison-results.md)
- [Command RSS/CPU](benchmarking_md/resource-comparison-results.md)
- [All-orchestrator probe coverage](benchmarking_md/orchestrator-probe-results.md)
- [Product comparison notes](benchmarking_md/product-comparison-results.md)

### Competitive landscape

| Product | Status in this repo | What it is good at | Litents comparison |
| --- | --- | --- | --- |
| [Zellij](https://zellij.dev/features/) | Automated lifecycle benchmark | Modern terminal sessions, layouts, resurrection | Terminal-native session peer with heavier workspace UX |
| [Codex CLI / app-server](https://openai.com/codex) | Automated local server benchmark | First-party Codex substrate across CLI, app, IDE, cloud | Agent runtime/server substrate, not a full orchestration peer |
| [Agent of Empires](https://www.agent-of-empires.com/) | Automated lifecycle benchmark | Parallel agents with tmux sessions, branches, worktrees, optional Docker | Direct terminal-native competitor with broader sandbox/worktree scope |
| [Claude Squad](https://github.com/smtg-ai/claude-squad) | Installed and version-probed | Multi-agent TUI for Claude Code, Codex, Gemini, Aider, Amp, and OpenCode | Very close CLI/TUI competitor; lifecycle benchmark needs interactive TUI automation |
| [CCManager](https://github.com/kbwo/ccmanager) | Installed and version-probed | Multi-agent sessions across worktrees, projects, and many agent CLIs | Direct manager competitor; lifecycle benchmark needs TUI/worktree automation |
| [Sidecar Workspaces](https://sidecar.haplab.com/docs/workspaces-plugin) | Installed and RSS/CPU-probed | Worktree branches, optional agents, tmux sessions, review/merge flow | Workflow-heavy peer; lifecycle benchmark needs plugin/workspace setup |
| [Agent Hand](https://weykon.github.io/agent-hand/) | Install attempted, blocked by upstream release/source paths | Fast tmux-backed terminal session manager for AI coding agents | Close terminal-native peer once install path is reproducible |
| [Agent Deck](https://asheshgoplani.github.io/agent-deck/) | Install attempted, blocked by upstream release asset and local CLT fallback | tmux command center for multiple coding-agent sessions | Session-visibility peer once install path is reproducible |
| [Crystal](https://github.com/stravu/crystal) | Installed GUI app, product/workflow comparison | Desktop app for Codex and Claude Code sessions in git worktrees | Direct desktop competitor with richer GUI UX |
| [Conductor](https://docs.conductor.build/) | Product/workflow comparison | Mac app for teams of coding agents in isolated workspaces | Productized desktop competitor |
| [CodeAgentSwarm](https://www.codeagentswarm.com/) | Product/workflow comparison | Multi-terminal macOS workspace with task board, history, and notifications | Integrated workspace competitor |
| [Termyx](https://termyx.dev/) | Product/workflow comparison | Native macOS git-worktree IDE for Claude Code | GUI worktree/session competitor |
| [Claude Code](https://www.claude.com/product/claude-code) | Runtime comparison | Strong single-agent coding loop and git workflow | Agent runtime Litents can orchestrate locally |
| [Gemini CLI](https://github.com/google-gemini/gemini-cli) | Runtime comparison | Google-backed terminal agent and GitHub workflow path | Agent runtime Litents can run inside panes |
| [Aider](https://github.com/Aider-AI/aider) | Runtime comparison | Git-centric terminal pair programming | Single-agent runtime Litents can manage |
| [OpenCode](https://opencode.ai/) | Runtime comparison | Open source terminal, desktop, and IDE coding agent | Agent runtime and potential managed command |

## Validation

Use this command set for end-to-end confidence:

```bash
go test ./...
go test ./... -bench . -run '^$' -benchmem
./benchmarking_md/compare-with-popular-tools.sh
./benchmarking_md/compare-running-agents-resource.sh
./benchmarking_md/compare-resource-usage.sh
./benchmarking_md/compare-orchestrator-probes.sh
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
