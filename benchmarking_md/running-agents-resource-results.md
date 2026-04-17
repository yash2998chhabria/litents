# Running Agents CPU/RAM Comparison

Generated on: 2026-04-17T05:47:08Z

Host: Darwin arm64
Go: go1.26.2
Litents binary: /Users/yashchhabria/Projects/litents/.tmp-litents-bench-bin
Litents source: d517da0+local
Zellij: zellij 0.44.1
Agent of Empires: aoe 1.4.3

Method:
- Launch 5 synthetic running agents per tool.
- Each synthetic agent runs `sleep 25`.
- Sample process trees 5 times at 1s intervals while agents are alive.
- RSS is summed resident memory for the tool runtime process tree plus managed agent pane process trees.
- CPU is summed `ps %cpu` for the same sampled process tree.
- Litents has no resident daemon; its running cost is the private tmux server plus managed panes.
- Zellij is measured through an isolated background session.
- Agent of Empires is measured through a private tmux server and a temporary fake `codex` command.

## Steady-state running agents summary

Tool | Running agents | RAM RSS | CPU %
---|---:|---:|---:
Litents | 5 |        5 samples, mean=35.72MiB (p50=35.72MiB, p95=35.72MiB, min=35.72MiB, max=35.72MiB) |        5 samples, mean=0.00% (p50=0.00%, p95=0.00%, min=0.00%, max=0.00%)
Zellij | 5 |        5 samples, mean=149.26MiB (p50=149.25MiB, p95=149.25MiB, min=149.25MiB, max=149.28MiB) |        5 samples, mean=0.40% (p50=0.10%, p95=0.60%, min=0.00%, max=1.30%)
Agent of Empires | 5 |        5 samples, mean=18.27MiB (p50=18.27MiB, p95=18.27MiB, min=18.27MiB, max=18.27MiB) |        5 samples, mean=0.00% (p50=0.00%, p95=0.00%, min=0.00%, max=0.00%)

## Interpretation

This benchmark answers the steady-state operator question: how much CPU and RAM are consumed while agents are already running. It intentionally excludes GUI-only products and tools that do not expose a reproducible headless running-agent lifecycle.
