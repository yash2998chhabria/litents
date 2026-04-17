# Litents — Lightweight Local Orchestrator for Raw Codex CLI Agents

> Build a tiny Unix-first control plane for running, supervising, resuming, and getting notified by multiple local Codex CLI agents. Litents should feel like `tmux` grew a purpose-built agent manager, not like a heavyweight desktop app.

## 1. Product Summary

**Name:** `litents`  
**Tagline:** Lightweight local intent/agent supervisor for Codex CLI.  
**Primary user:** A power user on Linux or macOS who runs many Codex CLI agents in parallel and wants maximum use of a Unix machine with minimal RAM, battery, and UI overhead.

Litents is **not** an AI agent framework. It does not replace Codex CLI. It orchestrates raw Codex CLI processes using durable Unix primitives:

- `tmux` for sessions, windows, panes, detach/reattach, and visibility.
- `git worktree` for isolated parallel checkouts.
- Codex CLI for actual agent behavior.
- Lightweight polling and tmux hooks for notifications.
- Local JSON metadata and pane logs for history.

The end result should let me run something like:

```bash
litents init ~/code/myrepo
litents new planner --prompt "Inspect the repo and propose a plan"
litents new impl-auth --from-branch main --prompt-file prompts/auth.md
litents new tests --prompt "Run tests, find failures, and propose fixes"
litents status
litents attach impl-auth
litents resume impl-auth
```

…and get pinged when an agent needs input, approval, or attention.

---

## 2. Design Principles

1. **Lightweight first.** No Electron, no web server by default, no database requirement, no long-running daemon unless explicitly requested.
2. **Unix-native.** Use `tmux`, process inspection, shell commands, files, and git worktrees.
3. **Codex-native.** Do not wrap Codex so heavily that normal Codex workflows break. A user should still be able to enter the pane and use Codex directly.
4. **Recoverable.** If Litents crashes, tmux panes and Codex sessions should still exist.
5. **Inspectable.** Every agent has a worktree, original prompt, output log, metadata file, and tmux pane/window name.
6. **No magic hidden state.** Store state in simple JSON files under XDG-compatible paths.
7. **Good defaults, configurable escape hatches.** Detection regexes, notification commands, tmux session name, and worktree paths should be configurable.

---

## 3. Non-goals

Do **not** build these in the MVP:

- A desktop app.
- A browser UI.
- A replacement terminal emulator.
- A new autonomous agent runtime.
- A full task queue with distributed workers.
- A cloud service.
- A custom model provider layer.
- A heavy TUI dependency unless absolutely necessary.
- A background daemon that must run for normal usage.

Litents should be useful even as a small CLI plus tmux session.

---

## 4. Target Environment

Support:

- macOS, Apple Silicon and Intel.
- Linux, x86_64 and ARM64.
- `bash` or `zsh` shells.
- `tmux` installed.
- `git` installed.
- `codex` CLI installed and authenticated.

Optional integrations:

- Linux desktop notifications via `notify-send`.
- macOS notifications via `terminal-notifier` if present, otherwise `osascript`.
- Headless notifications via a user-provided command such as `ntfy`, `curl`, `wall`, or a shell script.

---

## 5. Recommended Implementation Choice

Implement as a **single small Go CLI** unless the target repo already has a strong preference for another systems language.

Why Go:

- Easy cross-platform static-ish binary.
- Good `os/exec` support for calling `tmux`, `git`, and `codex`.
- Good stdlib JSON support.
- No runtime dependency like Python.
- No heavy TUI dependency required for the MVP.

MVP should use only Go standard library where practical. Avoid SQLite for the first version; use JSON metadata files and raw logs.

Suggested binary:

```bash
litents
```

Suggested package layout:

```text
litents/
  cmd/litents/main.go
  internal/agent/
  internal/config/
  internal/gitx/
  internal/notify/
  internal/state/
  internal/tmux/
  internal/watch/
  README.md
  go.mod
```

If implementing in shell instead, keep the same CLI surface and storage layout.

---

## 6. Core Concepts

### 6.1 Project

A project is a source repo managed by Litents.

Fields:

```json
{
  "name": "myrepo",
  "repo_path": "/Users/me/code/myrepo",
  "tmux_session": "litents-myrepo",
  "created_at": "2026-04-17T12:00:00Z"
}
```

### 6.2 Agent

An agent is one Codex CLI process running inside a tmux pane, usually in its own git worktree.

Fields:

```json
{
  "id": "impl-auth",
  "project": "myrepo",
  "role": "implementation",
  "repo_path": "/Users/me/code/myrepo",
  "worktree_path": "/Users/me/code/.litents-worktrees/myrepo/impl-auth",
  "branch": "litents/impl-auth",
  "tmux_session": "litents-myrepo",
  "tmux_window": "impl-auth",
  "tmux_pane": "%12",
  "prompt_file": "/Users/me/.local/state/litents/projects/myrepo/agents/impl-auth/prompt.md",
  "log_file": "/Users/me/.local/state/litents/projects/myrepo/agents/impl-auth/output.log",
  "status": "running",
  "last_status": "running",
  "last_activity_at": "2026-04-17T12:00:00Z",
  "last_notified_at": null,
  "created_at": "2026-04-17T12:00:00Z",
  "updated_at": "2026-04-17T12:00:00Z"
}
```

### 6.3 Status values

Use these statuses:

- `starting` — tmux pane/worktree/process is being created.
- `running` — process is alive and output is changing.
- `waiting` — likely waiting for user input or approval.
- `quiet` — process alive but no output for a configured silence threshold.
- `done` — process exited successfully or returned to shell.
- `failed` — process exited with error or setup failed.
- `unknown` — state cannot be determined.

---

## 7. Storage Layout

Use XDG paths on Linux and reasonable macOS equivalents.

Default state root:

```bash
${XDG_STATE_HOME:-$HOME/.local/state}/litents
```

Default config root:

```bash
${XDG_CONFIG_HOME:-$HOME/.config}/litents
```

Example:

```text
~/.local/state/litents/
  projects/
    myrepo/
      project.json
      agents/
        impl-auth/
          agent.json
          prompt.md
          output.log
          summary.md
        tests/
          agent.json
          prompt.md
          output.log
          summary.md

~/.config/litents/
  config.json
```

Do not require a central database for the MVP.

---

## 8. Configuration

Config file:

```json
{
  "tmux_session_prefix": "litents",
  "worktree_root": "~/.local/share/litents/worktrees",
  "default_base_branch": "main",
  "codex_command": "codex",
  "codex_args": [],
  "notify_enabled": true,
  "notify_command": "auto",
  "watch_interval_seconds": 3,
  "silence_threshold_seconds": 180,
  "activity_notify_cooldown_seconds": 120,
  "waiting_regexes": [
    "(?i)approval",
    "(?i)allow.*command",
    "(?i)requires.*permission",
    "(?i)permission required",
    "(?i)continue\\?",
    "(?i)press enter",
    "(?i)waiting for input",
    "(?i)do you want",
    "(?i)yes/no",
    "(?i)y/n",
    "❯",
    ">\\s*$"
  ],
  "done_regexes": [
    "(?i)task complete",
    "(?i)done",
    "(?i)finished"
  ]
}
```

The regexes must be user-editable. Codex output can change over time, so do not hardcode assumptions deeply.

---

## 9. CLI Commands

### 9.1 `litents doctor`

Check dependencies and print actionable diagnostics.

```bash
litents doctor
```

Checks:

- `tmux` exists.
- `git` exists.
- `codex` exists.
- Current shell can run commands.
- Notification backend available or fallback configured.
- State/config directories writable.

Example output:

```text
Litents doctor
✓ tmux: /opt/homebrew/bin/tmux
✓ git: /usr/bin/git
✓ codex: /opt/homebrew/bin/codex
✓ state dir: /Users/me/.local/state/litents
✓ notify: osascript fallback
```

### 9.2 `litents init [repo]`

Initialize Litents for a repo and create the tmux session if missing.

```bash
litents init ~/code/myrepo
```

Behavior:

- Resolve repo root using `git rev-parse --show-toplevel`.
- Create project state directory.
- Create a tmux session named `litents-<repo-name>` if missing.
- Create a `home` window with a shell in repo root.
- Optionally create a `watch` window running `litents watch --project <name>`.

Flags:

```bash
--session NAME
--no-watch
--worktree-root PATH
```

### 9.3 `litents new <agent-id>`

Create a new agent in a new tmux window and, by default, a new git worktree.

```bash
litents new impl-auth --prompt "Fix the auth bug described in issue #123"
litents new tests --prompt-file prompts/test-failures.md
```

Behavior:

1. Validate agent ID: lowercase letters, numbers, dash, underscore.
2. Resolve project/repo.
3. Create branch `litents/<agent-id>` unless `--branch` provided.
4. Create worktree under configured worktree root.
5. Create tmux window named `<agent-id>`.
6. Start Codex CLI in that worktree.
7. Pipe pane output to the agent log file.
8. Write `agent.json` and `prompt.md`.

Suggested tmux command shape:

```bash
tmux new-window -t litents-myrepo -n impl-auth -c /path/to/worktree \
  'codex "$(cat /path/to/prompt.md)"; echo "[litents] codex exited with status $?"; read -r -p "Press enter to close..."'
```

Prefer safer implementation that avoids shell injection:

- Write a tiny generated runner script per agent.
- Make it executable.
- Run the script in tmux.

Flags:

```bash
--repo PATH
--project NAME
--prompt TEXT
--prompt-file PATH
--base-branch main
--branch litents/my-task
--no-worktree
--window NAME
--codex-arg ARG       # repeatable
--profile NAME       # passes codex profile if supported by user config
```

### 9.4 `litents ls` / `litents status`

Show all agents and their statuses.

```bash
litents status
litents status --watch
```

Example output:

```text
PROJECT  AGENT       STATUS   AGE    LAST ACTIVITY  WORKTREE
myrepo   planner     waiting  42m    12s ago        .../planner
myrepo   impl-auth   running  31m    3s ago         .../impl-auth
myrepo   tests       quiet    18m    6m ago         .../tests
```

`--watch` should redraw every few seconds without a heavy TUI dependency.

### 9.5 `litents attach <agent-id>`

Attach to a running agent’s tmux window/pane.

```bash
litents attach impl-auth
```

Behavior:

- If not already in tmux, attach to the project session and select the agent window.
- If already inside tmux, switch client to the window.

### 9.6 `litents send <agent-id> <text>`

Send text to an agent pane.

```bash
litents send impl-auth "Use the smaller patch and rerun the auth tests."
```

Behavior:

- Use `tmux send-keys` to type the message and press Enter.
- Refuse empty messages unless `--enter` is passed.

Flags:

```bash
--enter-only
--no-enter
```

### 9.7 `litents tail <agent-id>`

Print recent output from the agent log or tmux capture.

```bash
litents tail impl-auth
litents tail impl-auth -n 80
litents tail impl-auth --follow
```

### 9.8 `litents notify test`

Test notification backend.

```bash
litents notify test
```

### 9.9 `litents watch`

Run a lightweight watcher that updates statuses and sends notifications.

```bash
litents watch
litents watch --project myrepo
```

Behavior:

- Poll known agents every `watch_interval_seconds`.
- Capture last N lines from each tmux pane.
- Run waiting/done regexes.
- Check last activity time.
- Update `agent.json`.
- Notify when status transitions into `waiting`, `failed`, or optionally `done`.
- Do not spam. Respect cooldowns and notify only on state transitions or new matching output.

This can run as a tmux window called `watch`. Do not require launchd/systemd in the MVP.

### 9.10 `litents resume <agent-id>`

Resume or reattach to prior work.

```bash
litents resume impl-auth
```

Behavior:

- If the tmux pane is still alive, attach to it.
- If pane/window is gone but worktree exists, open a new tmux window in the worktree and run:

```bash
codex resume --last
```

- If that fails or user wants a picker, run:

```bash
codex resume
```

Flags:

```bash
--all       # use codex resume --all when appropriate
--picker    # force picker instead of --last
```

### 9.11 `litents history`

List previous agents and sessions.

```bash
litents history
litents history --project myrepo
```

Show:

- Agent ID.
- Status.
- Created time.
- Worktree path.
- Log path.
- Prompt summary.

### 9.12 `litents stop <agent-id>`

Stop an agent safely.

```bash
litents stop impl-auth
```

Behavior:

- Send Ctrl-C to the pane.
- Wait briefly.
- If still running and `--force`, kill pane/process.

### 9.13 `litents clean`

Clean dead agents and optionally remove worktrees.

```bash
litents clean
litents clean --worktrees --merged-only
```

Be conservative. Never delete uncommitted work without an explicit confirmation flag.

---

## 10. tmux Integration

Litents should treat tmux as the main UI and persistence layer.

### 10.1 Session naming

Default session:

```text
litents-<repo-name>
```

Example:

```text
litents-myrepo
```

### 10.2 Window naming

One window per agent is easier to manage than many panes in one window.

Default:

```text
<agent-id>
```

Examples:

```text
planner
impl-auth
tests
review
```

### 10.3 Pane logging

On agent creation, start logging pane output.

Possible command:

```bash
tmux pipe-pane -o -t "$PANE" "cat >> '$LOG_FILE'"
```

Raw logs are enough for MVP. Timestamping can be added later.

### 10.4 Activity monitoring

Set tmux options where useful:

```bash
tmux set-window-option -t "$WINDOW" monitor-activity on
tmux set-window-option -t "$WINDOW" monitor-bell on
```

Also consider `monitor-silence` for quiet detection, but do not rely on it exclusively because the watcher can compute silence from captured output.

### 10.5 Status line hints

MVP can update window names with a short prefix:

```text
!impl-auth   # waiting/needs attention
*impl-auth   # running/recent activity
.impl-auth   # quiet
✓impl-auth   # done
```

Keep this optional/configurable, because some users dislike mutated window names.

---

## 11. Notification Detection

This is the most important feature.

### 11.1 Waiting detection

An agent is likely waiting if recent pane output matches configurable regexes such as:

```text
approval
allow command
requires permission
permission required
continue?
press enter
waiting for input
do you want
y/n
yes/no
```

Also support user-supplied regexes.

### 11.2 Silence detection

If a pane has no output for `silence_threshold_seconds`, mark it as `quiet`. Do not always notify on quiet, but optionally allow:

```json
"notify_on_quiet": false
```

### 11.3 State transitions

Notify only when status changes:

```text
running -> waiting
running -> failed
quiet -> waiting
running -> done   # optional
```

Do not notify repeatedly for the same waiting state unless output changes or cooldown expires.

### 11.4 Notification backends

Implement `auto` backend:

Linux:

```bash
notify-send "Litents: impl-auth needs input" "Approval requested in myrepo"
```

macOS preferred if installed:

```bash
terminal-notifier -title "Litents" -subtitle "impl-auth" -message "Needs input in myrepo"
```

macOS fallback:

```bash
osascript -e 'display notification "Needs input in myrepo" with title "Litents" subtitle "impl-auth"'
```

Headless/custom:

```json
{
  "notify_command": "/Users/me/bin/litents-notify {{project}} {{agent}} {{status}} {{message}}"
}
```

Template variables:

```text
{{project}}
{{agent}}
{{status}}
{{message}}
{{worktree}}
{{log_file}}
```

---

## 12. Git Worktree Behavior

By default, each agent should get its own worktree:

```bash
git worktree add -b litents/impl-auth /path/to/worktrees/myrepo/impl-auth main
```

Rules:

- Never create multiple agents in the same worktree unless `--no-worktree` is explicitly passed.
- Refuse to clean/delete a dirty worktree unless `--force` is passed.
- Branch naming default: `litents/<agent-id>`.
- Worktree path default:

```text
~/.local/share/litents/worktrees/<repo-name>/<agent-id>
```

Commands:

```bash
litents new impl-auth --base-branch main
litents clean --worktrees --merged-only
```

---

## 13. Codex CLI Behavior

Litents should invoke Codex CLI as transparently as possible.

Default launch:

```bash
codex "$(cat prompt.md)"
```

Safer generated runner script:

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$WORKTREE"
echo "[litents] starting codex for $AGENT_ID"
codex "$(cat "$PROMPT_FILE")"
status=$?
echo "[litents] codex exited with status $status"
exit "$status"
```

Do not assume all users want the same Codex approval or sandbox mode. Allow passthrough args and profiles.

Examples:

```bash
litents new review --codex-arg "--profile" --codex-arg "readonly_quiet" --prompt "Review this repo"
litents new impl --profile full_auto --prompt-file prompts/impl.md
```

Resume behavior:

```bash
codex resume --last
codex resume
codex resume --all
```

If the exact Codex resume behavior changes, do not hard fail. Fall back to plain `codex resume` and let the user pick.

---

## 14. History and Continuation

Litents must maintain its own lightweight history independent of Codex internals.

For each agent, store:

- Original prompt.
- Worktree path.
- Branch name.
- tmux session/window/pane identifiers.
- Output log.
- Status transitions.
- Timestamps.

`litents history` should work even when tmux is not running.

`litents resume <agent>` should:

1. Reattach if the pane exists.
2. Recreate a tmux window in the same worktree if pane is gone.
3. Run `codex resume --last` in that worktree by default.
4. Fall back to `codex resume` picker.

Optional later enhancement:

- Parse Codex history metadata to associate a Litents agent with an exact Codex session ID. Do not require this for MVP.

---

## 15. Minimal Dashboard UX

Avoid heavy TUI for MVP. A plain table is enough.

```bash
litents status --watch
```

Example:

```text
Litents: myrepo                                      q quit | a attach | r refresh

AGENT       STATUS    LAST ACTIVITY  AGE    BRANCH              NOTES
planner     waiting   11s ago        42m    litents/planner     approval prompt
impl-auth   running   2s ago         31m    litents/impl-auth   active output
tests       quiet     7m ago         18m    litents/tests       no output
review      done      14m ago        1h     litents/review      exited
```

Nice-to-have interactive keys later:

- `Enter`: attach selected agent.
- `s`: send message.
- `t`: tail log.
- `k`: stop agent.
- `r`: resume selected agent.
- `/`: filter.

Do not block MVP on interactive dashboard keys.

---

## 16. Safety Requirements

1. Never run arbitrary shell text without escaping or a generated script.
2. Never delete a git worktree with uncommitted changes unless `--force` is explicit.
3. Never hide Codex approval prompts from the user.
4. Do not store secrets in Litents metadata.
5. Avoid capturing user keystrokes beyond pane output logs.
6. Log file permissions should be user-only where possible.
7. When sending notifications, do not include sensitive output by default. Use generic text such as “Agent needs input.”

---

## 17. MVP Acceptance Criteria

The MVP is done when all of these work on Linux and macOS:

### Setup

```bash
litents doctor
litents init ~/code/myrepo
```

Expected:

- Project state exists.
- tmux session exists.
- Doctor reports dependencies clearly.

### Start agents

```bash
litents new planner --prompt "Inspect this repo and summarize architecture."
litents new impl-a --prompt "Make a small safe improvement."
```

Expected:

- Two git worktrees exist.
- Two tmux windows exist.
- Two Codex CLI processes start.
- Logs are written.
- Metadata is written.

### Observe agents

```bash
litents status
litents status --watch
```

Expected:

- Table shows both agents.
- Status updates over time.
- Last activity changes as output appears.

### Attach/send/tail

```bash
litents attach planner
litents send planner "Continue with option 2."
litents tail planner -n 50
```

Expected:

- User can jump to the agent.
- User can send input.
- User can inspect recent output without opening an editor.

### Notify

```bash
litents notify test
litents watch
```

Expected:

- Test notification appears.
- When a pane output matches waiting regex, notification appears once.
- Repeated notifications are suppressed by cooldown.

### Resume

```bash
litents resume planner
```

Expected:

- If pane exists, attach to it.
- If pane is gone, start a new window in the agent worktree and run Codex resume.

---

## 18. Suggested Build Phases

### Phase 1 — Skeleton

- Go CLI with subcommands.
- Config load/save.
- State directories.
- Dependency doctor.

### Phase 2 — tmux and agent launch

- Create tmux session.
- Create windows.
- Launch generated runner scripts.
- Start pane logging.

### Phase 3 — git worktrees

- Resolve repo root.
- Create worktree per agent.
- Store branch/worktree metadata.
- Conservative cleanup behavior.

### Phase 4 — status and history

- Inspect tmux panes.
- Capture last output.
- Show status table.
- List historical agents from state files.

### Phase 5 — watcher and notifications

- Poll agents.
- Regex-based waiting detection.
- Status transitions.
- Notification backends.
- Cooldown/suppression logic.

### Phase 6 — resume and polish

- `attach`, `send`, `tail`, `resume`, `stop`, `clean`.
- Better errors.
- README with examples.

---

## 19. Testing Strategy

Use integration tests where possible, but keep them practical.

Mockable pieces:

- Command runner interface for `tmux`, `git`, `codex`.
- Notification backend.
- State store.
- Regex detector.

Unit tests:

- Config loading.
- Agent ID validation.
- Waiting regex detection.
- Status transition notification suppression.
- Worktree path generation.
- JSON state read/write.

Manual integration test:

```bash
mkdir /tmp/litents-demo
cd /tmp/litents-demo
git init
echo '# demo' > README.md
git add README.md
git commit -m init
litents init .
litents new planner --prompt "Read this repo and ask me for approval before changing anything."
litents status --watch
```

For CI, provide a fake `codex` script that prints known output and sleeps, so waiting detection can be tested without real Codex calls.

Example fake Codex:

```bash
#!/usr/bin/env bash
echo "Starting fake Codex"
sleep 1
echo "Approval required: allow command? y/n"
sleep 30
```

---

## 20. README Requirements

The README should include:

- What Litents is.
- Why it exists.
- Install instructions.
- Quickstart.
- Command reference.
- Example multi-agent workflow.
- Notification setup for macOS and Linux.
- How to resume agents.
- How to safely clean worktrees.
- Known limitations.

README example workflow:

```bash
litents doctor
litents init ~/code/myrepo
litents new planner --prompt "Map the repo and propose three safe tasks."
litents new tests --prompt "Run tests and identify the smallest failing area."
litents status --watch
litents attach planner
litents send planner "Start with the first task."
litents resume planner
```

---

## 21. Known Limitations to Document

Be honest in the README:

- “Needs input” detection is best-effort because Codex CLI output is not a stable machine API.
- tmux must be installed and working.
- Notifications depend on local OS tools.
- Litents does not guarantee that agents avoid conflicts unless worktrees are used.
- Codex session IDs may not be exposed in a stable way; Litents resumes by worktree and Codex’s own resume command.

---

## 22. Stretch Goals

Only after MVP:

1. Exact Codex session ID association.
2. `litents tui` with keyboard navigation.
3. `litents web` optional local-only web UI.
4. `litents summarize <agent>` that asks Codex to summarize a log.
5. `litents compact-history` to generate durable summaries.
6. Queueing with `max_parallel_agents`.
7. Cost/token usage aggregation if Codex exposes it reliably.
8. GitHub issue/PR integration.
9. Remote SSH host mode.
10. Agent templates: planner, implementer, reviewer, tester.

---

## 23. Default Agent Templates

Support optional templates later. MVP can just store examples.

### planner

```text
You are the planning agent. Inspect the repo. Do not edit files. Produce a concise plan with risks, test strategy, and suggested worktree tasks.
```

### implementer

```text
You are the implementation agent. Make minimal, high-confidence changes for the assigned task. Run relevant tests. Avoid broad refactors.
```

### reviewer

```text
You are the review agent. Inspect the current branch diff. Look for correctness, regressions, security issues, and missing tests. Do not edit files unless asked.
```

### tester

```text
You are the testing agent. Run the relevant test suite, identify failures, and report the smallest reproducible issue. Prefer diagnosis before edits.
```

---

## 24. Example Generated Runner Script

For each agent, generate something like:

```bash
#!/usr/bin/env bash
set -euo pipefail

AGENT_ID="impl-auth"
WORKTREE="/Users/me/.local/share/litents/worktrees/myrepo/impl-auth"
PROMPT_FILE="/Users/me/.local/state/litents/projects/myrepo/agents/impl-auth/prompt.md"
LOG_PREFIX="[litents:$AGENT_ID]"

cd "$WORKTREE"
echo "$LOG_PREFIX cwd: $PWD"
echo "$LOG_PREFIX started: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

if ! command -v codex >/dev/null 2>&1; then
  echo "$LOG_PREFIX error: codex not found"
  exit 127
fi

codex "$(cat "$PROMPT_FILE")"
status=$?

echo "$LOG_PREFIX exited: $(date -u +%Y-%m-%dT%H:%M:%SZ) status=$status"
exit "$status"
```

Keep the runner file in the agent state directory.

---

## 25. Final Build Instruction for the Coding Agent

Build `litents` as a tiny, local, Unix-native orchestration CLI for multiple Codex CLI agents.

Prioritize this MVP order:

1. `doctor`
2. `init`
3. `new`
4. `status`
5. `attach`
6. `tail`
7. `send`
8. `watch` with notifications
9. `resume`
10. `stop` and `clean`

Keep the implementation boring and reliable. Use tmux as the interface, git worktrees as the isolation layer, and JSON files as the state store. Avoid heavy UI or daemon architecture until the core workflow is solid.
