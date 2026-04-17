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

### Reproducible comparison vs popular local managers

Litents is benchmarked against local orchestration peers using the same synthetic workload (`sleep 0.45; echo done`) with no model or network calls. The published harness no longer reports raw `tmux` as a competitor column: `tmux` is Litents' session substrate, not a product-level orchestration peer.

Command:

```bash
./benchmarking_md/compare-with-popular-tools.sh
```

Latest run (20 repeats, macOS, darwin/arm64, go1.26.2, Zellij 0.44.1, Codex CLI 0.121.0, Agent of Empires 1.4.3):

| Metric | Litents | Zellij | Codex app-server | Agent of Empires |
| --- | ---: | ---: | ---: | ---: |
| Initialize control surface | 20 runs, mean=32.95ms (p50=24ms, p95=28ms, min=22ms, max=196ms) | 20 runs, mean=51.05ms (p50=49ms, p95=55ms, min=48ms, max=77ms) | 20 runs, mean=48.85ms (p50=44ms, p95=48ms, min=40ms, max=149ms) | 20 runs, mean=8.35ms (p50=7ms, p95=9ms, min=7ms, max=23ms) |
| Start one workload | 20 runs, mean=18.05ms (p50=18ms, p95=20ms, min=17ms, max=20ms) | 20 runs, mean=52.55ms (p50=46ms, p95=97ms, min=43ms, max=105ms) | N/A | 20 runs, mean=64.45ms (p50=60ms, p95=66ms, min=59ms, max=132ms) |
| Status/list/health poll | 20 runs, mean=10.20ms (p50=10ms, p95=11ms, min=9ms, max=11ms) | 20 runs, mean=21.30ms (p50=21ms, p95=27ms, min=16ms, max=31ms) | 20 runs, mean=7.30ms (p50=7ms, p95=8ms, min=7ms, max=9ms) | 20 runs, mean=23.95ms (p50=23ms, p95=27ms, min=22ms, max=29ms) |
| Dashboard render | 20 runs, mean=19.75ms (p50=19ms, p95=21ms, min=18ms, max=26ms) | N/A | N/A | N/A |
| Peek recent output | 20 runs, mean=9.20ms (p50=8ms, p95=10ms, min=7ms, max=28ms) | N/A | N/A | N/A |
| Discover unmanaged panes | 20 runs, mean=11.55ms (p50=11ms, p95=14ms, min=10ms, max=21ms) | N/A | N/A | N/A |
| Adopt unmanaged pane | 20 runs, mean=35.10ms (p50=34ms, p95=39ms, min=33ms, max=39ms) | N/A | N/A | N/A |
| Untrack adopted pane | 20 runs, mean=5.20ms (p50=5ms, p95=6ms, min=5ms, max=6ms) | N/A | N/A | N/A |
| Stop control surface | 20 runs, mean=725.15ms (p50=726ms, p95=728ms, min=721ms, max=729ms) | 20 runs, mean=14.35ms (p50=14ms, p95=16ms, min=13ms, max=16ms) | 20 runs, mean=3.80ms (p50=3ms, p95=7ms, min=3ms, max=9ms) | 20 runs, mean=147.70ms (p50=148ms, p95=150ms, min=144ms, max=151ms) |
| Cleanup state files | 20 runs, mean=40.30ms (p50=41ms, p95=43ms, min=36ms, max=44ms) | N/A | N/A | 20 runs, mean=13.65ms (p50=14ms, p95=15ms, min=11ms, max=15ms) |

### Readable summary

| Area | Takeaway |
| --- | --- |
| Startup | Litents initializes in 32.95ms mean, faster than Zellij and Codex app-server in this run while still creating project/session state. Agent of Empires has a faster config-only init path. |
| Workload launch | Litents is the fastest measured tool for starting the synthetic agent workload: 18.05ms mean vs 52.55ms for Zellij and 64.45ms for Agent of Empires. |
| Status polling | Litents status is 10.20ms mean while reading project/agent state. Codex app-server health is faster, but it is a health endpoint rather than an agent/workspace status table. |
| Supervisor commands | Dashboard render is 19.75ms mean, peek is 9.20ms, discover is 11.55ms, adopt is 35.10ms, and untrack is 5.20ms in this run. |
| Stop behavior | Litents is intentionally slower because `stop` waits for graceful interrupt before force cleanup. |
| Coverage | The harness now includes Zellij, Codex app-server, and Agent of Empires. Claude Squad, CCManager, and Sidecar are installed/probed but need a richer TUI/worktree automation harness for lifecycle timings. |

The Codex desktop app itself is not measured because a macOS GUI app cannot be driven reproducibly in this headless shell benchmark. The harness measures `codex app-server`, which is the local headless server substrate exposed by Codex CLI.

Full timing output is in [benchmarking_md/tool-comparison-results.md](benchmarking_md/tool-comparison-results.md). Broader competitor notes are in [benchmarking_md/product-comparison-results.md](benchmarking_md/product-comparison-results.md).

### Resource usage

The cleanest performance comparison is steady-state CPU/RAM while agents are already running. This measures the tool runtime process tree plus managed synthetic agent panes, not command startup memory.

Command:

```bash
./benchmarking_md/compare-running-agents-resource.sh
```

Latest running-agent comparison (5 synthetic agents per tool, 5 samples, macOS, darwin/arm64):

| Tool | Running agents | RAM RSS | CPU |
| --- | ---: | ---: | ---: |
| Litents | 5 | 5 samples, mean=35.72MiB (p50=35.72MiB, p95=35.72MiB, min=35.72MiB, max=35.72MiB) | 5 samples, mean=0.00% (p50=0.00%, p95=0.00%, min=0.00%, max=0.00%) |
| Zellij | 5 | 5 samples, mean=149.26MiB (p50=149.25MiB, p95=149.25MiB, min=149.25MiB, max=149.28MiB) | 5 samples, mean=0.40% (p50=0.10%, p95=0.60%, min=0.00%, max=1.30%) |
| Agent of Empires | 5 | 5 samples, mean=18.27MiB (p50=18.27MiB, p95=18.27MiB, min=18.27MiB, max=18.27MiB) | 5 samples, mean=0.00% (p50=0.00%, p95=0.00%, min=0.00%, max=0.00%) |

Full running-agent output is in [benchmarking_md/running-agents-resource-results.md](benchmarking_md/running-agents-resource-results.md).

Command lifecycle resource benchmarks use the same headless lifecycle shape but measure peak RSS and CPU time for each lifecycle command. Litents has no resident daemon after commands exit, so these numbers represent orchestration command overhead rather than a terminal emulator or desktop GUI memory profile. The benchmark covers the operator surface too: dashboard, peek, discovery, adoption, and untracking.

Command:

```bash
./benchmarking_md/compare-resource-usage.sh
```

Latest peak RSS summary (10 repeats, macOS, darwin/arm64):

| Metric | Litents | Zellij | Codex app-server | Agent of Empires |
| --- | ---: | ---: | ---: | ---: |
| Initialize control surface | 10 runs, mean=5.19MiB (p50=5.11MiB, p95=5.50MiB, min=5.02MiB, max=5.59MiB) | 10 runs, mean=14.72MiB (p50=14.73MiB, p95=14.78MiB, min=14.62MiB, max=14.83MiB) | 10 runs, mean=4.72MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.75MiB) | 10 runs, mean=8.16MiB (p50=8.16MiB, p95=8.17MiB, min=8.14MiB, max=8.17MiB) |
| Start one workload | 10 runs, mean=5.25MiB (p50=5.20MiB, p95=5.38MiB, min=5.09MiB, max=5.42MiB) | 10 runs, mean=13.79MiB (p50=13.78MiB, p95=13.84MiB, min=13.73MiB, max=13.86MiB) | N/A | 10 runs, mean=25.56MiB (p50=25.53MiB, p95=25.77MiB, min=25.23MiB, max=26.02MiB) |
| Status/list/health poll | 10 runs, mean=5.28MiB (p50=5.27MiB, p95=5.39MiB, min=5.14MiB, max=5.41MiB) | 10 runs, mean=13.58MiB (p50=13.59MiB, p95=13.61MiB, min=13.52MiB, max=13.62MiB) | 10 runs, mean=4.72MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.77MiB) | 10 runs, mean=8.76MiB (p50=8.77MiB, p95=8.78MiB, min=8.70MiB, max=8.81MiB) |
| Dashboard render | 10 runs, mean=5.56MiB (p50=5.53MiB, p95=5.66MiB, min=5.47MiB, max=5.73MiB) | N/A | N/A | N/A |
| Peek recent output | 10 runs, mean=5.15MiB (p50=5.09MiB, p95=5.20MiB, min=5.00MiB, max=5.38MiB) | N/A | N/A | N/A |
| Discover unmanaged panes | 10 runs, mean=5.40MiB (p50=5.38MiB, p95=5.48MiB, min=5.27MiB, max=5.53MiB) | N/A | N/A | N/A |
| Adopt unmanaged pane | 10 runs, mean=5.86MiB (p50=5.86MiB, p95=6.20MiB, min=5.47MiB, max=6.23MiB) | N/A | N/A | N/A |
| Untrack adopted pane | 10 runs, mean=4.97MiB (p50=4.92MiB, p95=5.14MiB, min=4.83MiB, max=5.16MiB) | N/A | N/A | N/A |
| Stop control surface | 10 runs, mean=5.45MiB (p50=5.47MiB, p95=5.52MiB, min=5.22MiB, max=5.62MiB) | 10 runs, mean=13.33MiB (p50=13.34MiB, p95=13.39MiB, min=13.20MiB, max=13.39MiB) | 10 runs, mean=1.86MiB (p50=1.86MiB, p95=1.86MiB, min=1.86MiB, max=1.86MiB) | 10 runs, mean=8.92MiB (p50=8.91MiB, p95=8.95MiB, min=8.86MiB, max=9.02MiB) |
| Cleanup state files | 10 runs, mean=5.42MiB (p50=5.41MiB, p95=5.50MiB, min=5.28MiB, max=5.61MiB) | N/A | N/A | 10 runs, mean=8.78MiB (p50=8.78MiB, p95=8.81MiB, min=8.75MiB, max=8.83MiB) |

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
| Litents | local source build | `litents doctor` | 20 runs, mean=4.42MiB (p50=4.42MiB, p95=4.58MiB, min=4.31MiB, max=4.58MiB) |
| Claude Squad | 1.0.17 | `claude-squad version` | 20 runs, mean=5.92MiB (p50=5.91MiB, p95=6.09MiB, min=5.78MiB, max=6.31MiB) |
| Agent of Empires | 1.4.3 | `aoe --version` | 20 runs, mean=8.15MiB (p50=8.16MiB, p95=8.17MiB, min=8.12MiB, max=8.17MiB) |
| Zellij | 0.44.1 | `zellij --version` | 20 runs, mean=10.30MiB (p50=10.30MiB, p95=10.33MiB, min=10.27MiB, max=10.33MiB) |
| Codex CLI | 0.121.0 | `codex --version` | 20 runs, mean=16.19MiB (p50=16.19MiB, p95=16.23MiB, min=16.14MiB, max=16.23MiB) |
| Sidecar Workspaces | 0.83.0 | `sidecar --version` | 20 runs, mean=28.44MiB (p50=28.41MiB, p95=28.95MiB, min=28.00MiB, max=28.98MiB) |
| CCManager | 4.1.7 | `ccmanager --version` | 20 runs, mean=67.51MiB (p50=67.47MiB, p95=67.67MiB, min=67.33MiB, max=67.69MiB) |

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
