# Litents vs Popular CLI Baselines

Generated on: 2026-04-17T04:49:38Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents commit: 2d1921f

Method:
- Synthetic command workload: `sleep 0.45; echo done`
- Number of repeats: 20
- Scope: one agent window in one project, no model/network calls.
- Litents config: `codex_command: sh`, `codex_args: ["-lc"]`, `--no-worktree`, `--no-watch`
- tmux baseline: one session + one additional window using the same workload script

### Raw timing summary

Metric | Litents | tmux
---|---|---
Initialize project/session |       20 runs, mean=38.20ms (p50=26ms, p95=47ms, min=24ms, max=244ms) |       20 runs, mean=7.25ms (p50=7ms, p95=9ms, min=6ms, max=9ms)
Start one agent workload |       20 runs, mean=20.15ms (p50=19ms, p95=25ms, min=18ms, max=27ms) |       20 runs, mean=6.75ms (p50=7ms, p95=8ms, min=6ms, max=8ms)
Status/list poll |       20 runs, mean=10.10ms (p50=10ms, p95=12ms, min=9ms, max=13ms) |       20 runs, mean=5.30ms (p50=5ms, p95=6ms, min=4ms, max=6ms)
Stop/cleanup command |       20 runs, mean=725.30ms (p50=726ms, p95=727ms, min=717ms, max=729ms) |       20 runs, mean=5.55ms (p50=5ms, p95=7ms, min=4ms, max=7ms)
Cleanup state files |       20 runs, mean=45.75ms (p50=43ms, p95=72ms, min=35ms, max=88ms) | N/A


