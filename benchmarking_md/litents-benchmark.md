# Litents Benchmark Plan

> Goal: prove that **Litents** is the **lowest-overhead, fastest local orchestrator for raw Codex CLI agents on Unix** across a clearly defined benchmark suite.

## 1. What we are actually trying to prove

Litents should optimize for this exact job:

- managing many **local** Codex CLI agents,
- keeping tabs on what is running,
- surfacing when an agent needs input,
- preserving history and resuming old work,
- doing all of that with the **least possible RAM, CPU, wakeups, dependencies, and UI overhead**.

### Honest claim language

Do **not** claim:

- “best tool in the world”
- “fastest at all AI coding tasks”
- “universally better than every alternative”

Do claim, if the data supports it:

- “Litents had the **lowest orchestration overhead** among the tested local multi-agent tools.”
- “Litents had the **fastest operator workflows** for monitoring, resuming, and responding to waiting agents among the tested local baselines.”
- “Litents stayed **closest to raw tmux + Codex CLI** while adding supervision, notifications, and history.”

This benchmark can prove Litents is best **for a defined class of workloads and baselines**, not for every possible workflow.

---

## 2. The key insight: separate orchestrator speed from model speed

If you benchmark only with real model calls, you will mostly benchmark:

- model latency,
- network latency,
- repo complexity,
- prompt variability,
- reasoning effort.

That does **not** tell you whether Litents is a better orchestrator.

So the benchmark suite must have **two tracks**:

### Track A — Orchestrator overhead benchmark

Use a **deterministic mock agent** (`codex-shim`) that behaves like a terminal agent but has no real model latency.

This track measures the thing we care about most:

- startup overhead,
- refresh overhead,
- notification latency,
- memory use,
- CPU wakeups,
- history indexing,
- resume performance,
- crash recovery.

### Track B — Real Codex end-to-end benchmark

Use actual Codex CLI with the **same**:

- model,
- prompts,
- repo,
- reasoning level,
- approval mode,
- sandbox mode,
- network conditions.

This track shows whether Litents stays low-overhead in the real world.

**Important:** Track A is where Litents should “win.” Track B is where Litents should show that it adds very little overhead on top of raw Codex CLI.

---

## 3. Baselines to compare against

### Required baselines

1. **Litents**
   - release build
   - local-only mode
   - event-driven where possible
   - no daemon unless explicitly required

2. **Raw tmux + Codex CLI + shell scripts**
   - this is the true lightweight baseline
   - Litents should aim to stay as close to this as possible while adding orchestration features
   - status: baseline run implemented
   - script: [compare-with-popular-tools.sh](compare-with-popular-tools.sh)
   - latest data: [tool-comparison-results.md](tool-comparison-results.md)

3. **Zellij + Codex CLI**
   - modern terminal-native session manager baseline
   - good comparison for “lightweight but more productized than tmux”
   - status: detached session + tab baseline implemented in [compare-with-popular-tools.sh](compare-with-popular-tools.sh)
   - latest data: [tool-comparison-results.md](tool-comparison-results.md)

4. **Codex app**
   - official OpenAI desktop multi-agent baseline
   - compare only on platforms where it exists
   - status: headless `codex app-server` baseline implemented in [compare-with-popular-tools.sh](compare-with-popular-tools.sh)
   - note: desktop GUI launch is not included because it is not reproducible in a headless shell harness
   - latest data: [tool-comparison-results.md](tool-comparison-results.md)

### Optional baselines

5. **Mux local mode**
   - compare only if it can be run reproducibly in a local-only workflow

6. **Litents-Go prototype**
   - only if you build an early Go version and want a language/architecture shootout

### What not to benchmark as a primary baseline

- plain terminal emulators alone (Ghostty, WezTerm, Alacritty)
- IDEs
- cloud agent products
- anything that solves a meaningfully different problem

A terminal emulator is not the orchestration layer.

---

## 4. Platform matrix

Run benchmarks separately by OS.

### macOS matrix

Include:

- Litents
- raw tmux + Codex CLI
- Zellij + Codex CLI
- Codex app
- optional Mux

### Linux matrix

Include:

- Litents
- raw tmux + Codex CLI
- Zellij + Codex CLI
- optional Mux

Do **not** include Codex app on Linux. As of the current official docs, the Codex app is available on **macOS and Windows**, and OpenAI only offers “Get notified for Linux.”

---

## 5. Keep the benchmark fair

### Pin versions

Record and pin:

- Litents commit SHA
- Codex CLI version
- Codex app version
- tmux version
- Zellij version
- OS version
- shell version
- repo commit SHA

### Pin the real-agent configuration

For Track B, all systems must use the same:

- model
- reasoning level
- approval mode
- sandbox mode
- prompt text
- repo commit
- environment variables
- MCP configuration
- network policy

### Use the same model across systems

Use:

- **`gpt-5.4`** as the main real-agent benchmark model
- optional second pass with **`gpt-5.3-codex-spark`** for interactive responsiveness

Do **not** compare one tool on `gpt-5.4` and another on `gpt-5.3-codex-spark`.

### Separate cold and warm runs

For every benchmark, capture:

- cold run: no relevant sessions exist yet
- warm run: project/session metadata already exists
- resume run: prior sessions/history already exist

### Publish raw data

Every result must ship with:

- raw logs
- JSON metrics
- machine spec
- exact commands
- benchmark harness revision

If a result cannot be reproduced, it does not count.

---

## 6. Benchmark workloads

## A. Orchestrator overhead benchmarks (use `codex-shim`)

Build a deterministic mock terminal agent called `codex-shim` that:

- prints startup banners,
- simulates planning/output bursts,
- optionally goes silent,
- optionally emits a waiting marker,
- optionally exits with success/failure,
- can write enough output to stress pane scrollback and history capture.

### A1. Cold start: launch many agents

Measure:

- time to launch 1, 4, 8, 16, 32 agents
- time to first agent visible
- time to all agents visible
- orchestrator RSS during launch
- orchestrator CPU during launch

### A2. Idle supervision cost

Launch N idle agents and measure for 15 minutes:

- idle RSS
- idle CPU%
- wakeups/sec if available
- power draw if available
- background processes spawned by the orchestrator

Run at:

- 4 agents
- 8 agents
- 16 agents
- 32 agents

### A3. Notification latency

Have `codex-shim` emit a deterministic line like:

```text
LITENTS_WAITING_FOR_INPUT
```

Measure:

- timestamp of line emitted
- timestamp of orchestrator state transition
- timestamp of user notification fired

Report:

- p50 latency
- p95 latency
- missed notifications
- duplicate notifications

### A4. Status refresh latency under load

With 16 and 32 running panes, measure:

- time to refresh status view
- time to filter/search an agent
- time to switch to a target agent
- time to render agent detail/history view

### A5. History and resume performance

Create archived projects with:

- 10 agents
- 50 agents
- 100 agents

Measure:

- time to list all prior agents
- time to search for one by name/status/branch
- time to resume the last session
- time to resume an arbitrary session from history

### A6. Crash and recovery behavior

Simulate:

- Litents crash while agents keep running
- terminal disconnect
- tmux server restart if possible in the harness

Measure:

- time to detect and recover state
- % of agents recovered correctly
- notification state preserved or not
- history continuity preserved or not

### A7. Stop/cleanup performance

Measure:

- time to stop 8/16/32 agents
- time to archive state
- time to clean worktrees
- time to remove dead sessions from the UI/state store

---

## B. Real Codex end-to-end benchmarks

Use actual Codex CLI and pin everything.

### B1. Multi-agent repo scan

Task:

- start 4 agents in separate worktrees
- each agent scans a different part of the repo
- one summarizer agent collects findings

Measure:

- time to get all agents running
- orchestrator idle overhead while agents work
- time to identify which agent is waiting or done
- operator time to jump to a specific agent

### B2. Implementation fan-out

Task:

- start 3 implementation agents on separate branches/worktrees
- each agent modifies a different subsystem
- one test agent runs the test suite continuously

Measure:

- spawn latency
- worktree creation latency
- review/switch latency
- total system overhead above raw Codex CLI

### B3. Needs-input scenario

Task:

- force one agent to request approval or clarification
- operator must detect that, switch to it, answer, and return

Measure:

- time from agent request to notification
- time from notification to operator landing in the right context
- number of steps/keystrokes
- operator error rate

### B4. Resume-yesterday workflow

Task:

- archive 10+ historical agents
- return the next day
- find the last stuck agent and resume it

Measure:

- time to locate the right agent
- time to inspect its history
- time to resume it with the original context intact

### B5. Large-session oversight

Task:

- run 8–16 real Codex agents
- operator periodically checks status, answers 1–2 waiting agents, and reviews diffs

Measure:

- overhead over 30 minutes
- memory growth over time
- state accuracy
- whether any agent becomes “lost” from the operator’s view

---

## 7. Success metrics

## Primary metrics

1. **Idle RSS of orchestrator only**
2. **Incremental RSS above raw Codex CLI baseline**
3. **Idle CPU% of orchestrator only**
4. **p95 notification latency**
5. **p95 switch-to-agent latency**
6. **p95 resume-from-history latency**
7. **time to N ready agents**
8. **recovery success rate after orchestrator crash/disconnect**

## Secondary metrics

9. binary size
10. dependency count
11. installation time
12. number of helper/background processes
13. lines of config needed to get useful behavior
14. keystrokes/clicks for common operator flows
15. power draw on battery-capable systems

## Suggested headline thresholds

These are not laws, but good targets:

- Litents should be **within 0–10% of raw tmux** on cold start and idle overhead.
- Litents should beat **Codex app** on:
  - orchestrator RSS
  - idle CPU
  - notification latency
  - agent-switch latency
- Litents should beat **Zellij** on at least one of:
  - idle RSS
  - refresh latency
  - notification latency
- Litents should add **near-zero overhead per additional managed agent** relative to the raw tmux baseline.

If Litents is much heavier than raw tmux, it is drifting away from the core goal.

---

## 8. Instrumentation plan

### General timing

Capture monotonic timestamps for:

- launch requested
- pane created
- process started
- first output observed
- waiting marker observed
- notification fired
- operator attached
- session resumed
- process exit

Write these to newline-delimited JSON.

### Process/resource sampling

Sample at 1s intervals:

- RSS per relevant PID
- CPU% per relevant PID
- total process count per tool

Use standard OS tools, for example:

- `ps`
- `/usr/bin/time`
- `top` or equivalent
- platform-specific power tools when available

### Power/battery sampling

Do this only on hardware where it is meaningful.

- On macOS laptops, collect a repeatable power sample during a fixed idle and active window.
- On Linux laptops, use a repeatable power source when supported by the hardware.
- On desktops or servers, treat power as optional and focus on memory/CPU.

### UI/interaction timing

Automate the operator flows where possible:

- find waiting agent
- switch to waiting agent
- open history for a named agent
- resume last historical session

For GUI baselines, use scripted UI automation only if it is stable enough to be fair.
If not, run a human-timed protocol with multiple repeats and publish video.

---

## 9. Build a dedicated benchmark harness repo

Create a separate repo, for example:

```text
litents-bench/
  harness/
  codex-shim/
  workloads/
  repos/
  scripts/
  results/
  docs/
```

## `codex-shim`

The shim should support flags like:

```bash
codex-shim \
  --startup-ms 200 \
  --burst-lines 100 \
  --wait-after-ms 5000 \
  --emit-waiting \
  --exit-code 0
```

It should emulate:

- a fast planning agent
- a quiet long-running agent
- an agent that needs user input
- an agent that fails
- a very chatty agent

## Generated test repos

Use at least three repo sizes:

1. **small** — toy app, 1k–5k LOC
2. **medium** — realistic service/app, 20k–100k LOC
3. **large** — repo slice or generated monorepo-like tree, 250k+ LOC

### Why generated repos matter

They make file counts, path depths, and worktree creation reproducible.

---

## 10. Reporting format

Every benchmark result should produce:

- `summary.md`
- `results.json`
- `raw/` logs
- `machine.json`
- `versions.json`

## Example headline table

| Benchmark | Litents | raw tmux | Zellij | Codex app | Winner |
|---|---:|---:|---:|---:|---|
| idle RSS @ 16 agents | | | | | |
| p95 notification latency | | | | | |
| switch to waiting agent | | | | | |
| resume last historical agent | | | | | |
| launch 16 agents cold | | | | | |
| recovery after orchestrator crash | | | | | |

## Example conclusion language

Good:

> On macOS, Litents had the lowest orchestration RSS and fastest notification path among the tested local multi-agent tools, while staying within 6% of raw tmux on cold-start overhead.

Bad:

> Litents is objectively the best AI agent tool.

---

## 11. Product decisions this benchmark should drive

Use this benchmark to answer:

### A. Is Rust actually worth it?

Do **not** assume the language alone wins.

Test:

- Rust prototype
- Go prototype (optional)
- identical mock harness

Measure:

- idle RSS
- event-loop latency
- child-process supervision overhead
- snapshot serialization overhead
- build and install ergonomics

### B. Should Litents use polling or event streams?

Benchmark:

- periodic polling of `tmux list-*`
- `tmux` control mode / long-lived command channel
- tail-based log scanning

Choose the architecture with the lowest idle CPU and best p95 alert latency.

### C. Should notifications be driven by terminal output or deeper hooks?

Prefer the simplest stable method first.

Use terminal-output/state heuristics by default.
Only use deeper Codex hooks if they materially improve results and remain stable enough to trust.

### D. How much persistent state is enough?

Compare:

- JSON files only
- JSON + small index
- SQLite

The MVP should choose the lightest option that keeps resume/search fast enough.

---

## 12. Recommended implementation strategy

### Recommendation

Build Litents in **Rust** first.

Why:

- single native binary
- good fit for low-overhead Unix tooling
- good process/control-plane ergonomics
- aligns with the general low-overhead philosophy of Codex CLI itself

### But benchmark architecture before celebrating the language

The architecture decisions will matter more than the language:

- event-driven vs polling
- tmux control mode vs repeated shell-outs
- JSON files vs DB
- no daemon vs daemon
- how often the UI refreshes

A sloppy Rust architecture can still lose to a well-designed Go or shell tool.

---

## 13. How to compare against Codex app fairly

The Codex app is not the same kind of tool as Litents. It is a desktop “command center” with worktrees, automations, Git functions, review UI, and more.

So do **not** compare Litents to Codex app on things Litents intentionally does not do.

Compare them on the exact overlapping job:

- watch many agents
- notice when one needs attention
- switch to the right agent
- keep history organized
- resume old work
- do all of that with as little overhead as possible

### Fair Codex app test rules

- keep the same repo size
- use local workflows only where possible
- disable optional features not used in the scenario
- do not count feature depth as “speed”
- count operator-visible latency and machine overhead

---

## 14. Acceptance criteria for Litents v1

Ship v1 only if it can hit most of these:

- launches and supervises **16 local agents** reliably on Unix
- resumes historical agents quickly enough to feel instant or near-instant
- notifies accurately when an agent needs input
- survives Litents restarts without losing the agents
- uses clearly less memory than Codex app in equivalent local scenarios
- stays close to raw tmux in idle resource cost
- does not require a database, web server, or Electron shell

---

## 15. What to hand to the agent building this

Tell the coding agent to build in this order:

1. `codex-shim`
2. benchmark harness
3. raw tmux baseline scripts
4. Litents minimal Rust prototype
5. Litents event-driven prototype
6. benchmark report generator
7. only then polish the product UX

Reason: if you do not build the benchmark harness first, you will accidentally optimize vibes instead of overhead.

---

## 16. Sources / grounding notes

These are the product facts that matter for the benchmark framing.

- OpenAI says **Codex CLI** runs locally from your terminal, is open source, and is **built in Rust for speed and efficiency**:  
  https://developers.openai.com/codex/cli
- OpenAI says the **Codex app** is a desktop experience for working on Codex threads in parallel with built-in worktree support, automations, and Git features:  
  https://developers.openai.com/codex/app
- OpenAI’s current docs say the **Codex app is available on macOS and Windows**, with Linux not yet generally available:  
  https://developers.openai.com/codex/app
- OpenAI’s Codex model docs say to start with **`gpt-5.4`** for most tasks and describe **`gpt-5.3-codex-spark`** as a research-preview, near-instant coding model:  
  https://developers.openai.com/codex/models  
  https://openai.com/index/introducing-gpt-5-3-codex-spark/
- OpenAI’s Codex CLI docs document **`codex resume`**, `--last`, and `--all`, which matters for the history/resume benchmarks:  
  https://developers.openai.com/codex/cli/features
- OpenAI’s Codex hooks are currently **experimental**, so the benchmark should not depend on them for the MVP notification path:  
  https://developers.openai.com/codex/hooks
- tmux’s official docs describe it as a terminal multiplexer with detach/reattach and multiple panes/sessions, and tmux also has a **control mode** for machine-readable integration:  
  https://man.openbsd.org/tmux  
  https://github.com/tmux/tmux/wiki/Control-Mode
- Git’s official docs describe **`git worktree`** as a way for one repository to support multiple working trees at once, which is the right primitive for parallel local agents:  
  https://git-scm.com/docs/git-worktree
- Zellij’s official docs highlight session management, background session handling, and session resurrection, which is why it is a reasonable lightweight benchmark baseline:  
  https://zellij.dev/features/  
  https://zellij.dev/documentation/session-resurrection.html
- Mux describes itself as a local-or-remote parallel agentic development tool, so it is a reasonable optional comparison baseline if you want a more purpose-built local manager in the mix:  
  https://github.com/coder/mux

---

## 17. One-sentence north star

**Litents should feel like raw tmux + Codex CLI at rest, but with better supervision, notifications, history, and resume than any other local tool in its class.**
