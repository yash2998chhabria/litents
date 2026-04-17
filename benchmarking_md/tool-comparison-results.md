# Litents vs Popular CLI Baselines

Generated on: 2026-04-17T05:14:08Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: c2cfea1+local
Zellij: zellij 0.44.1
Codex: codex-cli 0.120.0
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
Initialize control surface |       20 runs, mean=24.20ms (p50=23ms, p95=28ms, min=21ms, max=39ms) |       20 runs, mean=54.60ms (p50=51ms, p95=65ms, min=47ms, max=103ms) |       20 runs, mean=44.00ms (p50=44ms, p95=46ms, min=39ms, max=57ms) |       20 runs, mean=8.45ms (p50=8ms, p95=10ms, min=7ms, max=16ms)
Start one workload |       20 runs, mean=18.05ms (p50=17ms, p95=21ms, min=15ms, max=26ms) |       20 runs, mean=50.35ms (p50=47ms, p95=70ms, min=39ms, max=96ms) | N/A |       20 runs, mean=64.55ms (p50=61ms, p95=73ms, min=57ms, max=114ms)
Status/list/health poll |       20 runs, mean=11.70ms (p50=10ms, p95=11ms, min=9ms, max=50ms) |       20 runs, mean=19.40ms (p50=18ms, p95=24ms, min=16ms, max=25ms) |       20 runs, mean=7.60ms (p50=8ms, p95=9ms, min=6ms, max=10ms) |       20 runs, mean=26.50ms (p50=24ms, p95=30ms, min=21ms, max=68ms)
Stop control surface |       20 runs, mean=728.20ms (p50=726ms, p95=729ms, min=719ms, max=778ms) |       20 runs, mean=17.45ms (p50=15ms, p95=17ms, min=12ms, max=71ms) |       20 runs, mean=3.30ms (p50=3ms, p95=4ms, min=3ms, max=4ms) |       20 runs, mean=148.10ms (p50=147ms, p95=152ms, min=142ms, max=161ms)
Cleanup state files |       20 runs, mean=53.20ms (p50=43ms, p95=48ms, min=33ms, max=257ms) | N/A | N/A |       20 runs, mean=13.95ms (p50=14ms, p95=15ms, min=11ms, max=21ms)
