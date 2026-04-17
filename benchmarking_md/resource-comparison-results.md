# Litents Resource Usage Comparison

Generated on: 2026-04-17T05:40:20Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: d8d5fa8+local
Zellij: zellij 0.44.1
Codex: codex-cli 0.121.0
Agent of Empires: aoe 1.4.3

Method:
- Synthetic command workload: `sleep 5; echo done`
- Number of repeats: 10
- Scope: peak RSS and CPU time for lifecycle commands in a headless shell harness.
- Memory unit: MiB, from `/usr/bin/time -l` on macOS or `/usr/bin/time -v` on Linux.
- CPU unit: user + system CPU milliseconds for the measured command.
- This is not a full desktop GUI memory profile and does not include terminal emulator memory.

### Peak RSS summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface |       10 runs, mean=5.19MiB (p50=5.11MiB, p95=5.50MiB, min=5.02MiB, max=5.59MiB) |       10 runs, mean=14.72MiB (p50=14.73MiB, p95=14.78MiB, min=14.62MiB, max=14.83MiB) |       10 runs, mean=4.72MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.75MiB) |       10 runs, mean=8.16MiB (p50=8.16MiB, p95=8.17MiB, min=8.14MiB, max=8.17MiB)
Start one workload |       10 runs, mean=5.25MiB (p50=5.20MiB, p95=5.38MiB, min=5.09MiB, max=5.42MiB) |       10 runs, mean=13.79MiB (p50=13.78MiB, p95=13.84MiB, min=13.73MiB, max=13.86MiB) | N/A |       10 runs, mean=25.56MiB (p50=25.53MiB, p95=25.77MiB, min=25.23MiB, max=26.02MiB)
Status/list/health poll |       10 runs, mean=5.28MiB (p50=5.27MiB, p95=5.39MiB, min=5.14MiB, max=5.41MiB) |       10 runs, mean=13.58MiB (p50=13.59MiB, p95=13.61MiB, min=13.52MiB, max=13.62MiB) |       10 runs, mean=4.72MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.77MiB) |       10 runs, mean=8.76MiB (p50=8.77MiB, p95=8.78MiB, min=8.70MiB, max=8.81MiB)
Dashboard render |       10 runs, mean=5.56MiB (p50=5.53MiB, p95=5.66MiB, min=5.47MiB, max=5.73MiB) | N/A | N/A | N/A
Peek recent output |       10 runs, mean=5.15MiB (p50=5.09MiB, p95=5.20MiB, min=5.00MiB, max=5.38MiB) | N/A | N/A | N/A
Discover unmanaged panes |       10 runs, mean=5.40MiB (p50=5.38MiB, p95=5.48MiB, min=5.27MiB, max=5.53MiB) | N/A | N/A | N/A
Adopt unmanaged pane |       10 runs, mean=5.86MiB (p50=5.86MiB, p95=6.20MiB, min=5.47MiB, max=6.23MiB) | N/A | N/A | N/A
Untrack adopted pane |       10 runs, mean=4.97MiB (p50=4.92MiB, p95=5.14MiB, min=4.83MiB, max=5.16MiB) | N/A | N/A | N/A
Stop control surface |       10 runs, mean=5.45MiB (p50=5.47MiB, p95=5.52MiB, min=5.22MiB, max=5.62MiB) |       10 runs, mean=13.33MiB (p50=13.34MiB, p95=13.39MiB, min=13.20MiB, max=13.39MiB) |       10 runs, mean=1.86MiB (p50=1.86MiB, p95=1.86MiB, min=1.86MiB, max=1.86MiB) |       10 runs, mean=8.92MiB (p50=8.91MiB, p95=8.95MiB, min=8.86MiB, max=9.02MiB)
Cleanup state files |       10 runs, mean=5.42MiB (p50=5.41MiB, p95=5.50MiB, min=5.28MiB, max=5.61MiB) | N/A | N/A |       10 runs, mean=8.78MiB (p50=8.78MiB, p95=8.81MiB, min=8.75MiB, max=8.83MiB)

### CPU time summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=33.00ms (p50=30ms, p95=30ms, min=30ms, max=60ms) |       10 runs, mean=14.00ms (p50=10ms, p95=20ms, min=0ms, max=20ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms)
Start one workload |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) | N/A |       10 runs, mean=63.00ms (p50=60ms, p95=70ms, min=50ms, max=70ms)
Status/list/health poll |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=16.00ms (p50=20ms, p95=20ms, min=10ms, max=20ms)
Dashboard render |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) | N/A | N/A | N/A
Peek recent output |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) | N/A | N/A | N/A
Discover unmanaged panes |       10 runs, mean=1.00ms (p50=0ms, p95=0ms, min=0ms, max=10ms) | N/A | N/A | N/A
Adopt unmanaged pane |       10 runs, mean=20.00ms (p50=20ms, p95=20ms, min=20ms, max=20ms) | N/A | N/A | N/A
Untrack adopted pane |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) | N/A | N/A | N/A
Stop control surface |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=30.00ms (p50=30ms, p95=30ms, min=30ms, max=30ms)
Cleanup state files |       10 runs, mean=20.00ms (p50=20ms, p95=20ms, min=20ms, max=20ms) | N/A | N/A |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms)

### Interpretation

Litents has no resident daemon after lifecycle commands exit. These results are best read as command peak memory and CPU cost for orchestrating local agent sessions, not as a full memory census of terminal emulators or desktop apps.
