# Litents vs Popular CLI Baselines

Generated on: 2026-04-17T05:03:10Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: b1f347f+local
tmux: tmux 3.6a
Zellij: zellij 0.44.1
Codex: codex-cli 0.120.0

Method:
- Synthetic command workload: `sleep 0.45; echo done`
- Number of repeats: 20
- Scope: one agent window in one project, no model/network calls.
- Litents config: `codex_command: sh`, `codex_args: ["-lc"]`, `--no-worktree`, `--no-watch`
- tmux baseline: one session + one additional window using the same workload script
- Zellij baseline: one detached background session + one tab using the same workload script
- Codex app-server baseline: `codex app-server --listen ws://127.0.0.1:<port>` startup, health check, and process stop
- Codex desktop app itself is not measured here because launching and driving a macOS GUI app is not reproducible in this headless harness

### Raw timing summary

Metric | Litents | tmux | Zellij | Codex app-server
---|---|---|---|---
Initialize control surface |       20 runs, mean=38.80ms (p50=22ms, p95=28ms, min=20ms, max=358ms) |       20 runs, mean=7.05ms (p50=7ms, p95=10ms, min=5ms, max=13ms) |       20 runs, mean=49.10ms (p50=47ms, p95=59ms, min=43ms, max=70ms) |       20 runs, mean=41.70ms (p50=42ms, p95=44ms, min=38ms, max=49ms)
Start one workload |       20 runs, mean=16.95ms (p50=16ms, p95=21ms, min=14ms, max=24ms) |       20 runs, mean=8.65ms (p50=6ms, p95=21ms, min=5ms, max=28ms) |       20 runs, mean=49.10ms (p50=46ms, p95=65ms, min=41ms, max=89ms) | N/A
Status/list/health poll |       20 runs, mean=9.45ms (p50=9ms, p95=11ms, min=8ms, max=11ms) |       20 runs, mean=5.50ms (p50=5ms, p95=10ms, min=4ms, max=13ms) |       20 runs, mean=20.55ms (p50=20ms, p95=25ms, min=16ms, max=30ms) |       20 runs, mean=7.75ms (p50=7ms, p95=10ms, min=6ms, max=22ms)
Stop control surface |       20 runs, mean=724.40ms (p50=724ms, p95=728ms, min=721ms, max=728ms) |       20 runs, mean=6.05ms (p50=5ms, p95=9ms, min=5ms, max=14ms) |       20 runs, mean=13.80ms (p50=13ms, p95=14ms, min=12ms, max=31ms) |       20 runs, mean=6.05ms (p50=3ms, p95=6ms, min=3ms, max=56ms)
Cleanup state files |       20 runs, mean=37.25ms (p50=39ms, p95=45ms, min=28ms, max=49ms) | N/A | N/A | N/A

