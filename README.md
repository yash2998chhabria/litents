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

### Reproducible comparison vs popular local managers

Litents is benchmarked against local orchestration peers using the same synthetic workload (`sleep 0.45; echo done`) with no model or network calls. The published harness no longer reports raw `tmux` as a competitor column: `tmux` is Litents' session substrate, not a product-level orchestration peer.

Command:

```bash
./benchmarking_md/compare-with-popular-tools.sh
```

Latest run (20 repeats, macOS, darwin/arm64, go1.26.2, Zellij 0.44.1, Codex CLI 0.120.0, Agent of Empires 1.4.3):

| Metric | Litents | Zellij | Codex app-server | Agent of Empires |
| --- | ---: | ---: | ---: | ---: |
| Initialize control surface | 20 runs, mean=24.20ms (p50=23ms, p95=28ms, min=21ms, max=39ms) | 20 runs, mean=54.60ms (p50=51ms, p95=65ms, min=47ms, max=103ms) | 20 runs, mean=44.00ms (p50=44ms, p95=46ms, min=39ms, max=57ms) | 20 runs, mean=8.45ms (p50=8ms, p95=10ms, min=7ms, max=16ms) |
| Start one workload | 20 runs, mean=18.05ms (p50=17ms, p95=21ms, min=15ms, max=26ms) | 20 runs, mean=50.35ms (p50=47ms, p95=70ms, min=39ms, max=96ms) | N/A | 20 runs, mean=64.55ms (p50=61ms, p95=73ms, min=57ms, max=114ms) |
| Status/list/health poll | 20 runs, mean=11.70ms (p50=10ms, p95=11ms, min=9ms, max=50ms) | 20 runs, mean=19.40ms (p50=18ms, p95=24ms, min=16ms, max=25ms) | 20 runs, mean=7.60ms (p50=8ms, p95=9ms, min=6ms, max=10ms) | 20 runs, mean=26.50ms (p50=24ms, p95=30ms, min=21ms, max=68ms) |
| Stop control surface | 20 runs, mean=728.20ms (p50=726ms, p95=729ms, min=719ms, max=778ms) | 20 runs, mean=17.45ms (p50=15ms, p95=17ms, min=12ms, max=71ms) | 20 runs, mean=3.30ms (p50=3ms, p95=4ms, min=3ms, max=4ms) | 20 runs, mean=148.10ms (p50=147ms, p95=152ms, min=142ms, max=161ms) |
| Cleanup state files | 20 runs, mean=53.20ms (p50=43ms, p95=48ms, min=33ms, max=257ms) | N/A | N/A | 20 runs, mean=13.95ms (p50=14ms, p95=15ms, min=11ms, max=21ms) |

### Readable summary

| Area | Takeaway |
| --- | --- |
| Startup | Litents initializes in 24.20ms mean, faster than Zellij and Codex app-server in this run while still creating project/session state. Agent of Empires has a faster config-only init path. |
| Workload launch | Litents is the fastest measured tool for starting the synthetic agent workload: 18.05ms mean vs 50.35ms for Zellij and 64.55ms for Agent of Empires. |
| Status polling | Litents status is 11.70ms mean while reading project/agent state. Codex app-server health is faster, but it is a health endpoint rather than an agent/workspace status table. |
| Stop behavior | Litents is intentionally slower because `stop` waits for graceful interrupt before force cleanup. |
| Coverage | The harness now includes Zellij, Codex app-server, and Agent of Empires. Claude Squad, CCManager, and Sidecar are installed/probed but need a richer TUI/worktree automation harness for lifecycle timings. |

The Codex desktop app itself is not measured because a macOS GUI app cannot be driven reproducibly in this headless shell benchmark. The harness measures `codex app-server`, which is the local headless server substrate exposed by Codex CLI.

Full timing output is in [benchmarking_md/tool-comparison-results.md](benchmarking_md/tool-comparison-results.md). Broader competitor notes are in [benchmarking_md/product-comparison-results.md](benchmarking_md/product-comparison-results.md).

### Resource usage

Resource benchmarks use the same headless lifecycle shape but measure peak RSS and CPU time for each lifecycle command. Litents has no resident daemon after commands exit, so these numbers represent orchestration command overhead rather than a terminal emulator or desktop GUI memory profile.

Command:

```bash
./benchmarking_md/compare-resource-usage.sh
```

Latest peak RSS summary (10 repeats, macOS, darwin/arm64):

| Metric | Litents | Zellij | Codex app-server | Agent of Empires |
| --- | ---: | ---: | ---: | ---: |
| Initialize control surface | 10 runs, mean=5.19MiB (p50=5.11MiB, p95=5.39MiB, min=4.92MiB, max=5.44MiB) | 10 runs, mean=14.74MiB (p50=14.72MiB, p95=14.80MiB, min=14.61MiB, max=14.83MiB) | 10 runs, mean=4.71MiB (p50=4.70MiB, p95=4.75MiB, min=4.70MiB, max=4.75MiB) | 10 runs, mean=8.18MiB (p50=8.17MiB, p95=8.20MiB, min=8.11MiB, max=8.23MiB) |
| Start one workload | 10 runs, mean=5.06MiB (p50=5.02MiB, p95=5.09MiB, min=4.98MiB, max=5.36MiB) | 10 runs, mean=13.78MiB (p50=13.78MiB, p95=13.83MiB, min=13.66MiB, max=13.86MiB) | N/A | 10 runs, mean=25.47MiB (p50=25.47MiB, p95=25.64MiB, min=25.23MiB, max=25.69MiB) |
| Status/list/health poll | 10 runs, mean=5.10MiB (p50=5.11MiB, p95=5.20MiB, min=4.92MiB, max=5.22MiB) | 10 runs, mean=13.59MiB (p50=13.59MiB, p95=13.67MiB, min=13.50MiB, max=13.69MiB) | 10 runs, mean=4.71MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.73MiB) | 10 runs, mean=8.76MiB (p50=8.75MiB, p95=8.80MiB, min=8.72MiB, max=8.83MiB) |
| Stop control surface | 10 runs, mean=5.32MiB (p50=5.28MiB, p95=5.48MiB, min=5.16MiB, max=5.48MiB) | 10 runs, mean=13.34MiB (p50=13.34MiB, p95=13.47MiB, min=13.22MiB, max=13.47MiB) | 10 runs, mean=1.86MiB (p50=1.86MiB, p95=1.86MiB, min=1.86MiB, max=1.86MiB) | 10 runs, mean=8.91MiB (p50=8.89MiB, p95=8.94MiB, min=8.86MiB, max=8.95MiB) |
| Cleanup state files | 10 runs, mean=5.32MiB (p50=5.27MiB, p95=5.52MiB, min=5.02MiB, max=5.53MiB) | N/A | N/A | 10 runs, mean=8.80MiB (p50=8.80MiB, p95=8.83MiB, min=8.72MiB, max=8.89MiB) |

All resource output is in [benchmarking_md/resource-comparison-results.md](benchmarking_md/resource-comparison-results.md).

### All-orchestrator probe coverage

Every named orchestrator is covered in the repo. CLI/TUI products get a repeatable probe for peak RSS and CPU; GUI-first products get install/discovery status rather than fabricated shell lifecycle numbers.

Command:

```bash
./benchmarking_md/compare-orchestrator-probes.sh
```

Latest installed CLI/TUI probe summary:

| Product | Local status | Probe | Peak RSS |
| --- | --- | --- | ---: |
| Litents | local source build | `litents doctor` | 20 runs, mean=4.34MiB (p50=4.31MiB, p95=4.42MiB, min=4.27MiB, max=4.44MiB) |
| Claude Squad | 1.0.17 | `claude-squad version` | 20 runs, mean=5.91MiB (p50=5.89MiB, p95=6.05MiB, min=5.80MiB, max=6.06MiB) |
| Agent of Empires | 1.4.3 | `aoe --version` | 20 runs, mean=8.15MiB (p50=8.16MiB, p95=8.16MiB, min=8.12MiB, max=8.16MiB) |
| Zellij | 0.44.1 | `zellij --version` | 20 runs, mean=10.30MiB (p50=10.30MiB, p95=10.33MiB, min=10.27MiB, max=10.34MiB) |
| Codex CLI | 0.120.0 | `codex --version` | 20 runs, mean=16.04MiB (p50=16.03MiB, p95=16.09MiB, min=15.98MiB, max=16.09MiB) |
| Sidecar Workspaces | 0.83.0 | `sidecar --version` | 20 runs, mean=28.33MiB (p50=28.27MiB, p95=28.84MiB, min=27.98MiB, max=29.19MiB) |
| CCManager | 4.1.7 | `ccmanager --version` | 20 runs, mean=67.43MiB (p50=67.44MiB, p95=67.55MiB, min=67.14MiB, max=67.61MiB) |

Full coverage matrix is in [benchmarking_md/orchestrator-probe-results.md](benchmarking_md/orchestrator-probe-results.md).

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
