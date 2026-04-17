# Litents Resource Usage Comparison

Generated on: 2026-04-17T05:25:46Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: 43c656e
Zellij: zellij 0.44.1
Codex: codex-cli 0.120.0
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
Initialize control surface |       10 runs, mean=5.19MiB (p50=5.11MiB, p95=5.39MiB, min=4.92MiB, max=5.44MiB) |       10 runs, mean=14.74MiB (p50=14.72MiB, p95=14.80MiB, min=14.61MiB, max=14.83MiB) |       10 runs, mean=4.71MiB (p50=4.70MiB, p95=4.75MiB, min=4.70MiB, max=4.75MiB) |       10 runs, mean=8.18MiB (p50=8.17MiB, p95=8.20MiB, min=8.11MiB, max=8.23MiB)
Start one workload |       10 runs, mean=5.06MiB (p50=5.02MiB, p95=5.09MiB, min=4.98MiB, max=5.36MiB) |       10 runs, mean=13.78MiB (p50=13.78MiB, p95=13.83MiB, min=13.66MiB, max=13.86MiB) | N/A |       10 runs, mean=25.47MiB (p50=25.47MiB, p95=25.64MiB, min=25.23MiB, max=25.69MiB)
Status/list/health poll |       10 runs, mean=5.10MiB (p50=5.11MiB, p95=5.20MiB, min=4.92MiB, max=5.22MiB) |       10 runs, mean=13.59MiB (p50=13.59MiB, p95=13.67MiB, min=13.50MiB, max=13.69MiB) |       10 runs, mean=4.71MiB (p50=4.70MiB, p95=4.73MiB, min=4.70MiB, max=4.73MiB) |       10 runs, mean=8.76MiB (p50=8.75MiB, p95=8.80MiB, min=8.72MiB, max=8.83MiB)
Stop control surface |       10 runs, mean=5.32MiB (p50=5.28MiB, p95=5.48MiB, min=5.16MiB, max=5.48MiB) |       10 runs, mean=13.34MiB (p50=13.34MiB, p95=13.47MiB, min=13.22MiB, max=13.47MiB) |       10 runs, mean=1.86MiB (p50=1.86MiB, p95=1.86MiB, min=1.86MiB, max=1.86MiB) |       10 runs, mean=8.91MiB (p50=8.89MiB, p95=8.94MiB, min=8.86MiB, max=8.95MiB)
Cleanup state files |       10 runs, mean=5.32MiB (p50=5.27MiB, p95=5.52MiB, min=5.02MiB, max=5.53MiB) | N/A | N/A |       10 runs, mean=8.80MiB (p50=8.80MiB, p95=8.83MiB, min=8.72MiB, max=8.89MiB)

### CPU time summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface |       10 runs, mean=3.00ms (p50=0ms, p95=10ms, min=0ms, max=20ms) |       10 runs, mean=33.00ms (p50=30ms, p95=40ms, min=30ms, max=40ms) |       10 runs, mean=21.00ms (p50=20ms, p95=30ms, min=10ms, max=40ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms)
Start one workload |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) | N/A |       10 runs, mean=71.00ms (p50=70ms, p95=70ms, min=70ms, max=80ms)
Status/list/health poll |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=18.00ms (p50=20ms, p95=20ms, min=10ms, max=20ms)
Stop control surface |       10 runs, mean=1.00ms (p50=0ms, p95=0ms, min=0ms, max=10ms) |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms) |       10 runs, mean=3.00ms (p50=0ms, p95=10ms, min=0ms, max=10ms) |       10 runs, mean=30.00ms (p50=30ms, p95=30ms, min=30ms, max=30ms)
Cleanup state files |       10 runs, mean=20.00ms (p50=20ms, p95=20ms, min=20ms, max=20ms) | N/A | N/A |       10 runs, mean=0.00ms (p50=0ms, p95=0ms, min=0ms, max=0ms)

### Interpretation

Litents has no resident daemon after lifecycle commands exit. These results are best read as command peak memory and CPU cost for orchestrating local agent sessions, not as a full memory census of terminal emulators or desktop apps.
