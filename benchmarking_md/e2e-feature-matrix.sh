#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LITENTS_BIN="${LITENTS_BIN:-$ROOT_DIR/.tmp-litents-e2e-bin}"
RUN_ID="${RUN_ID:-$(date +%Y%m%d%H%M%S)-$$}"
TMP_ROOT="${LITENTS_E2E_TMP:-$(mktemp -d)}"
KEEP_TMP="${KEEP_TMP:-0}"

if [[ ! -x "$LITENTS_BIN" ]]; then
  go build -o "$LITENTS_BIN" "$ROOT_DIR/cmd/litents"
fi

cleanup() {
  if [[ -n "${TMUX_SESSION_NAME:-}" ]]; then
    tmux kill-session -t "$TMUX_SESSION_NAME" >/dev/null 2>&1 || true
  fi

  if [[ -n "${LITENTS_BIN}" && -f "$LITENTS_BIN" && "${KEEP_TMP}" == "0" && "$LITENTS_BIN" == *".tmp-litents-e2e-bin" ]]; then
    rm -f "$LITENTS_BIN"
  fi

  if [[ "$KEEP_TMP" == "0" && -d "$TMP_ROOT" ]]; then
    rm -rf "$TMP_ROOT"
  fi
}

trap cleanup EXIT

if [[ ! -d "$TMP_ROOT" ]]; then
  TMP_ROOT="$(mktemp -d)"
fi
RUN_ROOT="$TMP_ROOT"
REPO_ROOT="$RUN_ROOT/repo"
STATE_ROOT="$RUN_ROOT/state"
CONFIG_ROOT="$RUN_ROOT/config"
WORKTREE_ROOT="$RUN_ROOT/worktrees"
SHIMS_ROOT="$RUN_ROOT/shims"
SESSION_PREFIX="litents-e2e-$RUN_ID"
PROJECT_NAME="repo"
AGENT_NAME="agent-1"
CODEX_SHIM="$SHIMS_ROOT/codex"

mkdir -p "$REPO_ROOT" "$STATE_ROOT" "$CONFIG_ROOT/litents" "$WORKTREE_ROOT" "$SHIMS_ROOT"
cat > "$REPO_ROOT/.gitignore" <<'EOF_GITIGNORE'
*
!.gitignore
EOF_GITIGNORE
cat > "$REPO_ROOT/README.md" <<'EOF_README'
# Litents e2e test repo
EOF_README
git -C "$REPO_ROOT" init -q

cat > "$CODEX_SHIM" <<'EOF_SHIM'
#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then
  exit 0
fi

if [ "$1" = "resume" ]; then
  echo "litents resume command: $*"
  exit 0
fi

sh -c "$1"
EOF_SHIM
chmod +x "$CODEX_SHIM"

cat > "$CONFIG_ROOT/litents/config.json" <<EOF_CFG
{
  "tmux_session_prefix": "$SESSION_PREFIX",
  "worktree_root": "$WORKTREE_ROOT",
  "default_base_branch": "main",
  "codex_command": "$CODEX_SHIM",
  "codex_args": [],
  "notify_enabled": false,
  "notify_command": "",
  "watch_interval_seconds": 1,
  "silence_threshold_seconds": 120,
  "activity_notify_cooldown_seconds": 30
}
EOF_CFG

export XDG_CONFIG_HOME="$CONFIG_ROOT"
export XDG_STATE_HOME="$STATE_ROOT"

TMUX_SESSION_NAME="$SESSION_PREFIX-$PROJECT_NAME"

status_check() {
  local cmd=("$@")
  local out
  if ! out="$("${cmd[@]}" 2>&1)"; then
    echo "$out"
    echo "[e2e] command failed: ${cmd[*]}"
    exit 1
  fi
  echo "$out"
}

echo "[e2e] running doctor check"
status_check "$LITENTS_BIN" doctor

echo "[e2e] initializing test project"
status_check "$LITENTS_BIN" init --no-watch --session "$TMUX_SESSION_NAME" "$REPO_ROOT"

echo "[e2e] creating long-running agent"
status_check "$LITENTS_BIN" new --project "$PROJECT_NAME" --no-worktree --prompt 'while true; do echo "agent tick $(date +%s)"; sleep 0.2; done' "$AGENT_NAME"

sleep 1

echo "[e2e] checking status/ls output"
status_check "$LITENTS_BIN" status --project "$PROJECT_NAME" | grep -q "$AGENT_NAME"
status_check "$LITENTS_BIN" ls --project "$PROJECT_NAME" | grep -q "$AGENT_NAME"
status_check "$LITENTS_BIN" dash --project "$PROJECT_NAME" --preview "$AGENT_NAME" --n 5 | grep -q "Litents dashboard"

echo "[e2e] checking tail output capture"
if ! status_check "$LITENTS_BIN" tail --project "$PROJECT_NAME" --n 5 "$AGENT_NAME" | grep -q "agent tick"; then
  echo "[e2e] expected initial log output missing"
  exit 1
fi
if ! status_check "$LITENTS_BIN" peek --project "$PROJECT_NAME" --n 5 "$AGENT_NAME" | grep -q "agent tick"; then
  echo "[e2e] expected peek output missing"
  exit 1
fi

echo "[e2e] discovering and adopting unmanaged Codex pane"
tmux new-window -t "$TMUX_SESSION_NAME" -n unmanaged-codex -c "$REPO_ROOT" "$CODEX_SHIM" 'while true; do echo "unmanaged codex tick"; sleep 0.3; done'
sleep 1
UNMANAGED_PANE="$(tmux list-panes -t "$TMUX_SESSION_NAME:unmanaged-codex" -F '#{pane_id}' | head -n 1)"
status_check "$LITENTS_BIN" discover | grep -q "$UNMANAGED_PANE"
status_check "$LITENTS_BIN" adopt "$UNMANAGED_PANE" --project "$PROJECT_NAME" --id adopted-codex
status_check "$LITENTS_BIN" status --project "$PROJECT_NAME" | grep -q "adopted"
status_check "$LITENTS_BIN" untrack --project "$PROJECT_NAME" adopted-codex
if status_check "$LITENTS_BIN" status --project "$PROJECT_NAME" | grep -q "adopted-codex"; then
  echo "[e2e] expected adopted pane untracked"
  exit 1
fi

echo "[e2e] sending input to agent pane"
status_check "$LITENTS_BIN" send --project "$PROJECT_NAME" "$AGENT_NAME" "hello from e2e"

echo "[e2e] validating notification test path"
status_check "$LITENTS_BIN" notify test

echo "[e2e] validating history query"
status_check "$LITENTS_BIN" history --project "$PROJECT_NAME" | grep -q "$AGENT_NAME"

echo "[e2e] stopping, resuming, and stopping again"
status_check "$LITENTS_BIN" stop --force --project "$PROJECT_NAME" "$AGENT_NAME"
status_check "$LITENTS_BIN" resume --project "$PROJECT_NAME" "$AGENT_NAME"
status_check "$LITENTS_BIN" stop --force --project "$PROJECT_NAME" "$AGENT_NAME"

echo "[e2e] validating watch command output"
(
  "$LITENTS_BIN" watch --project "$PROJECT_NAME" > "$RUN_ROOT/watch.log" 2>&1
) &
WATCH_PID=$!
sleep 2
kill "$WATCH_PID" >/dev/null 2>&1 || true
wait "$WATCH_PID" >/dev/null 2>&1 || true
if [[ ! -s "$RUN_ROOT/watch.log" ]]; then
  echo "[e2e] watch output missing"
  exit 1
fi

echo "[e2e] cleaning finished agents"
status_check "$LITENTS_BIN" clean --project "$PROJECT_NAME"

if status_check "$LITENTS_BIN" ls --project "$PROJECT_NAME" | grep -q "$AGENT_NAME"; then
  echo "[e2e] expected agent removed after cleanup"
  exit 1
fi

echo "[e2e] feature matrix complete"
