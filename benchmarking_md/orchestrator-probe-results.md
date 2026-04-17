# Litents Orchestrator Probe Results

Generated on: 2026-04-17T05:25:56Z

Host: Darwin arm64
Go: go1.26.2
Runs per installed CLI probe: 20

Method:
- This file covers every orchestrator named in the comparison list.
- CLI/TUI tools get a reproducible version/help command probe for peak RSS and CPU time.
- GUI-first products are reported with install/discovery status, not fake headless lifecycle numbers.
- Full lifecycle latency is in [tool-comparison-results.md](tool-comparison-results.md).
- Lifecycle command resource usage is in [resource-comparison-results.md](resource-comparison-results.md).

## Installed CLI/TUI probe measurements

| Product | Local status | Category | Probe command | Peak RSS | CPU time |
| --- | --- | --- | --- | ---: | ---: |
| Litents | local source build | project tool | `litents doctor` |       20 runs, mean=4.34MiB (p50=4.31MiB, p95=4.42MiB, min=4.27MiB, max=4.44MiB) |       20 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |
| Zellij | zellij 0.44.1 | automated lifecycle peer | `zellij --version` |       20 runs, mean=10.30MiB (p50=10.30MiB, p95=10.33MiB, min=10.27MiB, max=10.34MiB) |       20 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |
| Codex CLI | codex-cli 0.120.0 | runtime/server substrate | `codex --version` |       20 runs, mean=16.04MiB (p50=16.03MiB, p95=16.09MiB, min=15.98MiB, max=16.09MiB) |       20 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |
| Agent of Empires | aoe 1.4.3 | automated lifecycle peer | `aoe --version` |       20 runs, mean=8.15MiB (p50=8.16MiB, p95=8.16MiB, min=8.12MiB, max=8.16MiB) |       20 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |
| Claude Squad | claude-squad version 1.0.17 | installed CLI/TUI peer | `claude-squad version` |       20 runs, mean=5.91MiB (p50=5.89MiB, p95=6.05MiB, min=5.80MiB, max=6.06MiB) |       20 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |
| CCManager | 4.1.7 | installed CLI/TUI peer | `ccmanager --version` |       20 runs, mean=67.43MiB (p50=67.44MiB, p95=67.55MiB, min=67.14MiB, max=67.61MiB) |       20 runs, mean=107.00ms (p50=110ms, p95=110ms, min=100ms, max=120ms) |
| Sidecar Workspaces | sidecar version v0.83.0 | installed CLI/TUI peer | `sidecar --version` |       20 runs, mean=28.33MiB (p50=28.27MiB, p95=28.84MiB, min=27.98MiB, max=29.19MiB) |       20 runs, mean=10.00ms (p50=10ms, p95=10ms, min=10ms, max=10ms) |

## Full comparison coverage matrix

| Product | Coverage in this repo | Local result |
| --- | --- | --- |
| Litents | Lifecycle latency, lifecycle resource usage, CLI probe | Source build benchmarked locally |
| Zellij | Lifecycle latency, lifecycle resource usage, CLI probe | zellij 0.44.1 |
| Codex app-server | Lifecycle latency, lifecycle resource usage, CLI probe | codex-cli 0.120.0 |
| Agent of Empires | Lifecycle latency, lifecycle resource usage, CLI probe | aoe 1.4.3 |
| Claude Squad | CLI/version/RSS/CPU probe | claude-squad version 1.0.17 |
| CCManager | CLI/version/RSS/CPU probe | 4.1.7 |
| Sidecar Workspaces | CLI/version/RSS/CPU probe | sidecar version v0.83.0 |
| Crystal | GUI install/probe only | Crystal.app 0.3.5 |
| Conductor | GUI product workflow only | not installed |
| CodeAgentSwarm | GUI product workflow only | not installed |
| Termyx | GUI product workflow only | not installed |
| Agent Hand | Blocked/probe when installable | not installed; installer/source/brew release asset blocked in this environment |
| Agent Deck | Blocked/probe when installable | not installed; release asset/brew fallback blocked in this environment |

## Notes

- Lifecycle benchmarks are only reported where the tool exposes a reproducible non-interactive path.
- Version/help probes are not a substitute for lifecycle benchmarks, but they give a comparable lower-bound process startup, RSS, and CPU signal across CLI/TUI products.
- GUI-first products such as Crystal, Conductor, CodeAgentSwarm, and Termyx should be compared through workflow and UX case studies unless they expose a stable automation interface.
