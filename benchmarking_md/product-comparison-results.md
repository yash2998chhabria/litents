# Litents Product Comparison Notes

Generated on: 2026-04-17

This file separates reproducible timing data from broader product positioning. A tool is only placed in the automated timing table when it can run a deterministic, non-interactive local lifecycle from a shell harness.

## Automated headless lifecycle benchmarks

Source: [tool-comparison-results.md](tool-comparison-results.md)

Lifecycle resource usage: [resource-comparison-results.md](resource-comparison-results.md)

| Product | Version/status | Harness coverage | Result shape |
| --- | --- | --- | --- |
| Litents | local source build | `init`, `new`, `status`, `stop`, `clean` with a synthetic offline workload | Fastest measured workload launch and faster status than Zellij/AOE in this run |
| [Zellij](https://zellij.dev/features/) | 0.44.1 | detached session, new tab, list tabs, kill session | Useful terminal workspace peer with heavier lifecycle timings |
| [Codex app-server](https://openai.com/codex) | Codex CLI 0.120.0 | local app-server startup, health check, process stop | Health/server substrate only, not a full agent workspace lifecycle |
| [Agent of Empires](https://www.agent-of-empires.com/) | 1.4.3 | `aoe init`, `aoe add`, `aoe session start`, JSON status, stop, remove | Direct tmux/worktree manager peer; init/remove are fast, workload launch/status are heavier than Litents |

## Installed CLI/TUI probes for the wider competitor set

Source: [orchestrator-probe-results.md](orchestrator-probe-results.md)

These tools are real local competitors, but their useful workflow involves a TUI, plugin setup, or richer workspace state than a simple one-command timing harness can fairly drive today. The probe numbers are version/help command peak RSS and CPU, not full lifecycle timings.

| Product | Installed/probed status | Peak RSS probe | Why it is not in the lifecycle table yet |
| --- | --- | ---: | --- |
| [Claude Squad](https://github.com/smtg-ai/claude-squad) | `claude-squad 1.0.17` via Homebrew | `claude-squad version`: 20 runs, mean=5.91MiB | It is a close multi-agent TUI competitor; fair lifecycle timing needs scripted workspace/session creation rather than version startup |
| [CCManager](https://github.com/kbwo/ccmanager) | `ccmanager 4.1.7` via npm | `ccmanager --version`: 20 runs, mean=67.43MiB | It is a broad multi-agent manager; fair lifecycle timing needs non-interactive TUI/worktree automation |
| [Sidecar Workspaces](https://sidecar.haplab.com/docs/workspaces-plugin) | `sidecar v0.83.0` via `go install github.com/marcus/sidecar/cmd/sidecar@latest` | `sidecar --version`: 20 runs, mean=28.33MiB | The workspace plugin flow needs project/plugin setup before lifecycle timings are meaningful |

## Install attempted, blocked by upstream/local packaging

| Product | Official source | Attempted path | Result |
| --- | --- | --- | --- |
| [Agent Hand](https://weykon.github.io/agent-hand/) | tmux-backed AI coding-agent session manager | install script, `cargo install --git https://github.com/weykon/agent-hand agent-hand`, and `brew install weykon/tap/agent-hand` | installer path returned 404 earlier; Cargo GitHub fetch failed authentication; Homebrew formula release asset returned 404 |
| [Agent Deck](https://asheshgoplani.github.io/agent-deck/) | tmux command center for multiple agent sessions | official install script and Homebrew tap fallback | release asset returned 404; Homebrew fallback was blocked by outdated local Command Line Tools |

## Product/workflow comparisons only

These products are relevant competitors, but they are GUI/macOS-app-first or cloud/workspace-first. They should be compared with workflow coverage, UX, isolation, review/merge flow, and operator visibility, not a fake shell timing number.

| Product | Official source | Local status | Positioning | Litents comparison |
| --- | --- | --- | --- | --- |
| [Crystal](https://github.com/stravu/crystal) | GitHub | Installed as `Crystal.app 0.3.5` via `brew install --cask stravu-crystal` | Desktop app for Codex and Claude Code sessions in parallel git worktrees | Direct desktop competitor with richer GUI UX |
| [Conductor](https://docs.conductor.build/) | Docs | not installed; official flow is download app and drag to Applications | Mac app for orchestrating teams of coding agents in isolated workspaces | Productized desktop competitor |
| [CodeAgentSwarm](https://www.codeagentswarm.com/) | Site | not installed; official flow is macOS app download plus sign-in | macOS workspace for multiple Claude Code, Codex, and Gemini CLI terminals | Integrated workspace competitor |
| [Termyx](https://termyx.dev/) | Site | not installed; official flow is macOS app download | Native macOS git-worktree IDE for Claude Code | GUI worktree/session competitor |

## Current interpretation

Litents' strongest measured advantage is local lifecycle overhead for starting and monitoring synthetic agent workloads. Its weakest measured area is stop latency, because it deliberately gives the agent a graceful interrupt window before force cleanup.

The next fairest benchmark expansion would be a scripted TUI/worktree harness for Claude Squad, CCManager, and Sidecar Workspaces, followed by separate UX/workflow case studies for Crystal, Conductor, CodeAgentSwarm, and Termyx.
