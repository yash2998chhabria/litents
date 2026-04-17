# Litents — Agent Handoff / Implementation Brief

You are continuing work on **Litents**.

The existing repository is a promising prototype, but the next iteration should turn it into a tool the user can actually live in day to day: a **lightweight Unix-native session supervisor for multiple Codex CLI sessions**.

This document is the implementation brief. Treat it as the product and architecture direction for the next phase.

---

## 1. Mission

Build Litents into the lightest serious way to:

- run multiple local Codex CLI sessions,
- keep tabs on all of them,
- get notified when one needs human input,
- resume/continue old sessions deterministically,
- track both Litents-created sessions and Codex sessions that Litents did **not** originally launch,
- manage everything from one lightweight terminal interface.

The user should feel like:

> “I’m just using Codex in multiple panes, and Litents quietly keeps track of everything for me.”

Not:

> “I’m using a whole orchestration framework.”

---

## 2. Product thesis

Litents should **not** try to replace Codex CLI.

Instead:

- **Codex CLI** is the real work surface.
- **tmux** is the persistence/runtime layer.
- **Litents** is the oversight layer.

That means:

- attach to Codex panes fast,
- show what is running,
- show what needs attention,
- store durable session metadata/history,
- make resume reliable,
- do not over-wrap or over-control Codex.

---

## 3. Non-negotiable constraints

1. **Keep it local-first.**
   - Linux/macOS first.
   - No browser UI.
   - No desktop app.
   - No heavy daemon required.

2. **Keep tmux as the runtime layer.**
   - Do not replace tmux with a custom scheduler.
   - tmux is the right persistence primitive.

3. **Keep Codex CLI as the primary interactive UX.**
   - The user must still be able to attach to a pane and use Codex normally.
   - Litents should not try to clone the full Codex UI.

4. **Do not rewrite this in Rust right now.**
   - Keep moving in Go.
   - The bottleneck is product shape, not language performance.

5. **Discovery/adoption is mandatory.**
   - Litents must track sessions it launched.
   - Litents must also discover/adopt Codex sessions it did not launch.

6. **Resume must be deterministic.**
   - Avoid “best guess” behavior when possible.
   - Prefer explicit session identity.

---

## 4. Current repo assessment

The repo already has the right foundation:

- tmux-backed sessions,
- git worktrees,
- JSON state,
- command surface for init/new/status/watch/attach/send/resume/history/stop/clean,
- notifications,
- local lightweight architecture.

But it is still too prototype-shaped.

### Main issues to fix

#### A. The core is still too monolithic
The current implementation still centralizes too much in one large app file. There are already multiple internal packages, but the actual behavior is not cleanly decomposed.

#### B. The product only fully understands sessions it launches itself
That is not enough for the intended UX.

#### C. Logging and live interactivity need to be cleaner
Litents should treat the tmux pane as the live source of truth.

#### D. Status/attention detection is too heuristic-heavy
Regex matching is useful, but it should not be the whole story.

#### E. There is no real interactive operator interface yet
The user needs a lightweight dashboard/TUI.

---

## 5. New product shape

Litents now has **two layers**:

### Layer A — Scriptable CLI
Keep the command-line interface for scripting and power use.

### Layer B — Interactive dashboard
Add a small terminal dashboard that becomes the default entry point.

Running `litents` with no arguments should open the dashboard.

This dashboard should let the user:

- see every tracked session,
- filter by attention/running/done,
- preview recent output,
- attach instantly,
- send input,
- resume,
- inspect history,
- adopt unmanaged sessions.

---

## 6. Target user experience

### Daily loop
The user starts Litents, sees a dashboard, and can immediately answer:

- What is running?
- Which session needs me?
- Which sessions are quiet/stuck?
- What did this agent do recently?
- Which old session should I resume?

### Interaction model
The user should usually do one of two things:

1. Stay in the dashboard and supervise.
2. Jump into a specific Codex pane and work normally.

### Litents should feel like

- a session inbox,
- a supervisor,
- a history/index,
- a jump tool.

Not like:

- a replacement shell,
- a replacement editor,
- a full orchestration framework.

---

## 7. Architecture direction

Refactor the repo into clearer responsibilities.

### Recommended package layout

```text
cmd/litents/
internal/app/
internal/cli/
internal/dashboard/
internal/discovery/
internal/supervisor/
internal/store/
internal/state/
internal/tmux/
internal/gitx/
internal/codex/
internal/notify/
internal/pathutil/
```

### Package roles

- `state` = canonical domain models
- `store` = persistence (JSON + append-only event log)
- `supervisor` = orchestration logic
- `discovery` = finding/adopting existing Codex/tmux sessions
- `dashboard` = terminal UI only
- `codex` = Codex session metadata helpers / resume helpers
- `tmux` = tmux adapter
- `notify` = notifications
- `gitx` = worktrees / repo helpers

---

## 8. Canonical state model

Create one canonical Agent model and use it everywhere.

### Minimum fields to add/standardize

```json
{
  "id": "impl-auth",
  "project": "myrepo",
  "source": "launched",
  "repo_path": "/repo",
  "worktree_path": "/worktrees/myrepo/impl-auth",
  "branch": "litents/impl-auth",
  "tmux_session": "litents-myrepo",
  "tmux_window": "impl-auth",
  "tmux_pane": "%12",
  "codex_session_id": "sess_123",
  "codex_thread_id": "thr_456",
  "model": "gpt-5.4",
  "approval_policy": "on-request",
  "sandbox_mode": "workspace-write",
  "prompt_file": ".../prompt.md",
  "log_file": ".../output.log",
  "events_file": ".../events.jsonl",
  "status": "waiting",
  "attention_reason": "approval",
  "attention_excerpt": "Allow command: pytest packages/auth -q",
  "last_error": "",
  "exit_code": null,
  "last_activity_at": "...",
  "created_at": "...",
  "updated_at": "...",
  "archived_at": null
}
```

### Add append-only event history
Each agent directory should contain:

```text
agent.json
prompt.md
events.jsonl
output.log
```

`agent.json` = latest snapshot  
`events.jsonl` = timeline of status/attention/resume transitions

---

## 9. Execution model changes

### Critical rule
**Do not make Litents the thing that replaces the live Codex pane.**

The pane is the primary surface.

### Logging strategy
Use one consistent approach:

- run Codex directly in the tmux pane,
- use tmux capture / pipe mechanisms for log collection,
- store structured events separately.

Do not build a confusing hybrid where the “real” session is hidden behind redirection and log scraping.

### Startup behavior
Session creation should be transactional:

- create worktree if needed,
- create metadata skeleton,
- create tmux window,
- confirm pane is alive,
- only then mark session active.

If something fails, leave a clean and understandable state.

---

## 10. Discovery and adoption

This is a required feature, not a nice-to-have.

### Add: `litents discover`
Scan tmux and surface candidate Codex sessions.

The command/dashboard should be able to show:

- tracked Litents sessions,
- unmanaged tmux panes that appear to be Codex,
- stale/dead sessions,
- archived sessions.

### Add: `litents adopt <pane-id>`
This should:

- create Litents metadata for an existing pane,
- infer project/repo/worktree if possible,
- attach pane/session/window identity,
- mark `source = adopted`.

### Add: `litents untrack <agent-id>`
Remove Litents tracking without killing the actual pane.

### Detection signals
Use a combination of:

- tmux pane metadata,
- pane current command,
- pane current path,
- pane title/window title if useful,
- Litents store matches,
- Codex session metadata if available.

---

## 11. Resume model

Current resume behavior is not strong enough for the long-term product.

### Desired resume order
When resuming a session:

1. If pane is alive → attach.
2. Else if explicit `codex_session_id` exists → resume that exact session.
3. Else if one strong match exists for that worktree/session metadata → use that.
4. Else open an interactive picker.

### Requirement
Litents should become good enough that the user is not guessing which session to continue.

---

## 12. Attention detection model

The user specifically wants to be pinged when a session needs input.

So build a better attention model.

### Move from only `status` to:

- `status`
- `attention_reason`
- `attention_excerpt`
- `attention_since`

### Suggested attention reasons

- `approval`
- `input_required`
- `error`
- `done`
- `stalled`
- `resume_needed`
- `untracked`

### Detection layers

#### Layer 1 — strongest signals
- pane alive/dead
- explicit exit code
- explicit session identity
- known attention text

#### Layer 2 — tmux signals
- activity
- bell
- silence
- pane/window status changes

#### Layer 3 — text heuristics
Keep regexes for:

- approvals,
- yes/no prompts,
- “press enter”,
- “waiting for input”,
- done markers,
- failure markers.

### Notification quality requirement
Do **not** send a notification that only says “status changed”.

Send notifications like:

- `impl-auth needs approval — Allow command: pytest packages/auth -q`
- `tests finished — 24 passed, 1 skipped`
- `review is quiet for 10m`

---

## 13. Dashboard requirements

Add `litents dash` and make `litents` default to it.

### Dashboard purpose
The dashboard is a supervisor, not a replacement terminal.

### Layout

#### Main session list
Each row should show at least:

- project
- agent/session name
- status
- attention indicator
- age
- last activity
- branch
- source (`launched` / `adopted`)
- worktree/cwd

#### Preview/details pane
Show:

- prompt summary
- last 20–60 lines of output
- attention excerpt
- tmux session/window/pane
- repo/worktree path
- codex session id if known

### Filters
Support quick filters for:

- all
- attention
- running
- waiting
- quiet
- done
- archived
- unmanaged/discovered

### Minimum keybindings

- `Enter` → attach
- `r` → resume
- `s` → send input
- `n` → new session
- `a` → adopt discovered pane
- `x` → stop session
- `d` → toggle attention-only filter
- `/` → search
- `h` → open history/timeline
- `p` → peek output
- `q` → quit

### Important UX rule
The dashboard should **preview and jump**.
It should not try to clone all of Codex’s UI behavior.

---

## 14. Command surface after this phase

### Keep
- `init`
- `new`
- `status`
- `attach`
- `send`
- `tail`
- `watch`
- `resume`
- `history`
- `stop`
- `clean`

### Add
- `dash`
- `discover`
- `adopt <pane-id>`
- `untrack <agent-id>`
- `peek <agent-id>`
- optional `inbox`

### Behavior change
- `litents` with no args should open the dashboard.

---

## 15. Immediate implementation plan

## Phase 1 — refactor core without changing product behavior too much

Deliver:

- one canonical Agent/Project state model,
- smaller packages,
- cleaner adapters for tmux/git/notify,
- cleaner logging strategy,
- existing CLI still works.

## Phase 2 — discovery/adoption

Deliver:

- `discover`,
- `adopt`,
- `untrack`,
- better session identity handling,
- source tracking.

## Phase 3 — dashboard

Deliver:

- terminal dashboard,
- attach/resume/send/new from UI,
- preview pane,
- attention filters,
- search.

## Phase 4 — better history/attention

Deliver:

- `events.jsonl`,
- better notifications,
- timeline/history view,
- archive model,
- improved resume reliability.

## Phase 5 — polish

Deliver:

- more robust tests,
- concurrency safety around store writes,
- better stale session cleanup,
- docs/screenshots,
- quality-of-life improvements.

---

## 16. Acceptance criteria

This phase is successful when:

1. I can run 5–10 Codex CLI sessions and see them from one lightweight interface.
2. I can instantly see which ones need attention.
3. Notifications tell me **why** a session needs me.
4. I can attach to a session with one action.
5. I can resume an old session without guessing.
6. I can track sessions Litents didn’t launch itself.
7. I do not need a full editor open just to know what my agents are doing.
8. The actual Codex pane still feels like normal Codex.

---

## 17. What not to do

Do **not**:

- rewrite to Rust now,
- build a browser UI,
- build a desktop app,
- build a heavyweight always-on daemon,
- replace tmux,
- replace Codex CLI,
- over-optimize before the dashboard + discovery + resume story works,
- make the product more complicated than it needs to be.

---

## 18. First concrete milestone for the next PR

If you want a very pragmatic next step, the next meaningful PR should do **exactly this**:

### PR 1 scope

1. Refactor core state into one canonical model.
2. Add `discover`.
3. Add `adopt`.
4. Add `source` + `attention_reason` + `attention_excerpt` to session state.
5. Fix logging/live-session handling so the Codex pane remains first-class.
6. Add a minimal dashboard that can:
   - list sessions,
   - filter attention-needed sessions,
   - preview output,
   - attach.

That alone would move Litents from “prototype launcher” to “something the user can genuinely use.”

---

## 19. One-line instruction

> Build Litents into a lightweight local Codex session supervisor: keep tmux and Codex as the runtime, add discovery/adoption, durable session identity/history, actionable attention notifications, and a small dashboard TUI so the user can track every session and jump into the right one instantly.
