# Litents vs Popular CLI Baselines

Generated on: 2026-04-17T05:39:52Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: d8d5fa8+local
Zellij: zellij 0.44.1
Codex: codex-cli 0.121.0
Agent of Empires: aoe 1.4.3

Method:
- Synthetic command workload: `sleep 0.45; echo done`
- Number of repeats: 20
- Scope: one agent window in one project, no model/network calls.
- Litents config: `codex_command: sh`, `codex_args: ["-lc"]`, `--no-worktree`, `--no-watch`
- Zellij baseline: one detached background session + one tab using the same workload script
- Codex app-server baseline: `codex app-server --listen ws://127.0.0.1:<port>` startup, health check, and process stop
- Agent of Empires baseline: temporary fake `codex` shim on `PATH`, `aoe init`, `aoe add`, `aoe session start`, JSON status, stop, and remove
- Codex desktop app itself is not measured here because launching and driving a macOS GUI app is not reproducible in this headless harness

### Raw timing summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface |       20 runs, mean=32.95ms (p50=24ms, p95=28ms, min=22ms, max=196ms) |       20 runs, mean=51.05ms (p50=49ms, p95=55ms, min=48ms, max=77ms) |       20 runs, mean=48.85ms (p50=44ms, p95=48ms, min=40ms, max=149ms) |       20 runs, mean=8.35ms (p50=7ms, p95=9ms, min=7ms, max=23ms)
Start one workload |       20 runs, mean=18.05ms (p50=18ms, p95=20ms, min=17ms, max=20ms) |       20 runs, mean=52.55ms (p50=46ms, p95=97ms, min=43ms, max=105ms) | N/A |       20 runs, mean=64.45ms (p50=60ms, p95=66ms, min=59ms, max=132ms)
Status/list/health poll |       20 runs, mean=10.20ms (p50=10ms, p95=11ms, min=9ms, max=11ms) |       20 runs, mean=21.30ms (p50=21ms, p95=27ms, min=16ms, max=31ms) |       20 runs, mean=7.30ms (p50=7ms, p95=8ms, min=7ms, max=9ms) |       20 runs, mean=23.95ms (p50=23ms, p95=27ms, min=22ms, max=29ms)
Dashboard render |       20 runs, mean=19.75ms (p50=19ms, p95=21ms, min=18ms, max=26ms) | N/A | N/A | N/A
Peek recent output |       20 runs, mean=9.20ms (p50=8ms, p95=10ms, min=7ms, max=28ms) | N/A | N/A | N/A
Discover unmanaged panes |       20 runs, mean=11.55ms (p50=11ms, p95=14ms, min=10ms, max=21ms) | N/A | N/A | N/A
Adopt unmanaged pane |       20 runs, mean=35.10ms (p50=34ms, p95=39ms, min=33ms, max=39ms) | N/A | N/A | N/A
Untrack adopted pane |       20 runs, mean=5.20ms (p50=5ms, p95=6ms, min=5ms, max=6ms) | N/A | N/A | N/A
Stop control surface |       20 runs, mean=725.15ms (p50=726ms, p95=728ms, min=721ms, max=729ms) |       20 runs, mean=14.35ms (p50=14ms, p95=16ms, min=13ms, max=16ms) |       20 runs, mean=3.80ms (p50=3ms, p95=7ms, min=3ms, max=9ms) |       20 runs, mean=147.70ms (p50=148ms, p95=150ms, min=144ms, max=151ms)
Cleanup state files |       20 runs, mean=40.30ms (p50=41ms, p95=43ms, min=36ms, max=44ms) | N/A | N/A |       20 runs, mean=13.65ms (p50=14ms, p95=15ms, min=11ms, max=15ms)
