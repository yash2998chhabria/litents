# Litents Product Comparison Notes

Generated on: 2026-04-17

This file separates reproducible timing data from broader product positioning. A tool is only placed in the automated timing table when it can run a deterministic, non-interactive local lifecycle from a shell harness.

## Automated headless lifecycle benchmarks

Source: [tool-comparison-results.md](tool-comparison-results.md)

| Product | Version/status | Harness coverage | Result shape |
| --- | --- | --- | --- |
| Litents | local source build | `init`, `new`, `status`, `stop`, `clean` with a synthetic offline workload | Fastest measured workload launch and faster status than Zellij/AOE in this run |
| [Zellij](https://zellij.dev/features/) | 0.44.1 | detached session, new tab, list tabs, kill session | Useful terminal workspace peer with heavier lifecycle timings |
| [Codex app-server](https://openai.com/codex) | Codex CLI 0.120.0 | local app-server startup, health check, process stop | Health/server substrate only, not a full agent workspace lifecycle |
| [Agent of Empires](https://www.agent-of-empires.com/) | 1.4.3 | `aoe init`, `aoe add`, `aoe session start`, JSON status, stop, remove | Direct tmux/worktree manager peer; init/remove are fast, workload launch/status are heavier than Litents |

## Installed and probed, not yet in lifecycle harness

These tools are real local competitors, but their useful workflow involves a TUI, plugin setup, or richer workspace state than a simple one-command timing harness can fairly drive today.

| Product | Installed/probed status | Probe result | Why it is not in the lifecycle table yet |
| --- | --- | ---: | --- |
| [Claude Squad](https://github.com/smtg-ai/claude-squad) | `claude-squad 1.0.17` via Homebrew | version command: 20 runs, mean=7.85ms (p50=8ms, p95=9ms, min=7ms, max=11ms) | It is a close multi-agent TUI competitor; fair lifecycle timing needs scripted workspace/session creation rather than version startup |
| [CCManager](https://github.com/kbwo/ccmanager) | `ccmanager 4.1.7` via npm | version command: 20 runs, mean=106.70ms (p50=105ms, p95=117ms, min=103ms, max=117ms) | It is a broad multi-agent manager; fair lifecycle timing needs non-interactive TUI/worktree automation |
| [Sidecar Workspaces](https://sidecar.haplab.com/docs/workspaces-plugin) | `sidecar v0.83.0` via `go install github.com/marcus/sidecar/cmd/sidecar@latest` | version command: 20 runs, mean=18.00ms (p50=17ms, p95=21ms, min=16ms, max=24ms) | The workspace plugin flow needs project/plugin setup before lifecycle timings are meaningful |

## Install attempted, blocked by upstream/local packaging

| Product | Official source | Attempted path | Result |
| --- | --- | --- | --- |
| [Agent Hand](https://weykon.github.io/agent-hand/) | tmux-backed AI coding-agent session manager | install script plus `cargo install --git https://github.com/weykon/agent-hand agent-hand` | installer path returned 404 earlier; Cargo GitHub fetch failed authentication in this environment |
| [Agent Deck](https://asheshgoplani.github.io/agent-deck/) | tmux command center for multiple agent sessions | official install script and Homebrew tap fallback | release asset returned 404; Homebrew fallback was blocked by outdated local Command Line Tools |

## Product/workflow comparisons only

These products are relevant competitors, but they are GUI/macOS-app-first or cloud/workspace-first. They should be compared with workflow coverage, UX, isolation, review/merge flow, and operator visibility, not a fake shell timing number.

| Product | Official source | Positioning | Litents comparison |
| --- | --- | --- | --- |
| [Crystal](https://github.com/stravu/crystal) | GitHub | Desktop app for Codex and Claude Code sessions in parallel git worktrees | Direct desktop competitor with richer GUI UX |
| [Conductor](https://docs.conductor.build/) | Docs | Mac app for orchestrating teams of coding agents in isolated workspaces | Productized desktop competitor |
| [CodeAgentSwarm](https://www.codeagentswarm.com/) | Site | macOS workspace for multiple Claude Code, Codex, and Gemini CLI terminals | Integrated workspace competitor |
| [Termyx](https://termyx.dev/) | Site | Native macOS git-worktree IDE for Claude Code | GUI worktree/session competitor |

## Current interpretation

Litents' strongest measured advantage is local lifecycle overhead for starting and monitoring synthetic agent workloads. Its weakest measured area is stop latency, because it deliberately gives the agent a graceful interrupt window before force cleanup.

The next fairest benchmark expansion would be a scripted TUI/worktree harness for Claude Squad, CCManager, and Sidecar Workspaces, followed by separate UX/workflow case studies for Crystal, Conductor, CodeAgentSwarm, and Termyx.
