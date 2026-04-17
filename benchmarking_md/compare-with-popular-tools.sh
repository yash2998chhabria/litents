#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LITENTS_BIN="${LITENTS_BIN:-$ROOT_DIR/.tmp-litents-bench-bin}"
RUNS="${LITENTS_BENCH_RUNS:-20}"
WORKLOAD_SECONDS="${LITENTS_BENCH_SLEEP_SECONDS:-0.45}"
OUT_DIR="${LITENTS_BENCH_OUT:-$ROOT_DIR/benchmarking_md}"
RESULT_FILE="$OUT_DIR/tool-comparison-results.md"

mkdir -p "$OUT_DIR"

if [[ ! -x "$LITENTS_BIN" ]]; then
  go build -o "$LITENTS_BIN" "$ROOT_DIR/cmd/litents"
fi

now_ns() {
  date +%s%N
}

measure_ms() {
  local log_file="$1"
  shift

  local start end
  start=$(now_ns)
  "$@" >"$log_file" 2>&1
  local rc=$?
  end=$(now_ns)

  if (( rc != 0 )); then
    echo "Command failed (exit=$rc): $*" >&2
    echo "Log: $log_file" >&2
    tail -n 20 "$log_file" >&2 || true
    return "$rc"
  fi

  echo $(( (end - start) / 1000000 ))
}

print_stats() {
  local file="$1"
  if [[ ! -s "$file" ]]; then
    echo "n/a"
    return
  fi

  local count min max avg p50 p95 sorted
  count=$(wc -l < "$file")

  if (( count == 0 )); then
    echo "n/a"
    return
  fi

  sorted="$(mktemp)"
  sort -n "$file" > "$sorted"

  min=$(head -n 1 "$sorted")
  max=$(tail -n 1 "$sorted")
  avg=$(awk '{sum += $1} END {if (NR == 0) print 0; else printf "%.2f", sum / NR}' "$file")
  p50=$(awk -v n="$count" 'NR == int((n + 1) / 2){print; exit}' "$sorted")
  p95=$(awk -v n="$count" 'NR >= int((n - 1) * 0.95 + 1){print; exit}' "$sorted")
  rm -f "$sorted"

  echo "${count} runs, mean=${avg}ms (p50=${p50}ms, p95=${p95}ms, min=${min}ms, max=${max}ms)"
}

litents_init_tmp="$(mktemp)"
litents_new_tmp="$(mktemp)"
litents_status_tmp="$(mktemp)"
litents_stop_tmp="$(mktemp)"
litents_clean_tmp="$(mktemp)"
tmux_init_tmp="$(mktemp)"
tmux_new_tmp="$(mktemp)"
tmux_list_tmp="$(mktemp)"
tmux_kill_tmp="$(mktemp)"

: >"$litents_init_tmp"
: >"$litents_new_tmp"
: >"$litents_status_tmp"
: >"$litents_stop_tmp"
: >"$litents_clean_tmp"
: >"$tmux_init_tmp"
: >"$tmux_new_tmp"
: >"$tmux_list_tmp"
: >"$tmux_kill_tmp"

for i in $(seq 1 "$RUNS"); do
  work_dir="$(mktemp -d)"
  repo_dir="$work_dir/repo"
  state_root="$work_dir/state"
  config_root="$work_dir/config"
  run_log="$work_dir/run.log"
  session_suffix="bench-$i"
  project_name="repo"
  agent_name="agent-1"

  mkdir -p "$repo_dir" "$state_root" "$config_root/litents"
  git -C "$repo_dir" init -q

  cat >"$config_root/litents/config.json" <<EOF_CFG
{
  "tmux_session_prefix": "ltcmp-${session_suffix}",
  "worktree_root": "$work_dir/worktrees",
  "default_base_branch": "main",
  "codex_command": "sh",
  "codex_args": ["-lc"],
  "notify_enabled": false,
  "notify_command": ""
}
EOF_CFG

  base_env=(XDG_CONFIG_HOME="$config_root" XDG_STATE_HOME="$state_root")
  litents_session="ltcmp-${session_suffix}"
  tmux_session="ltmux-${session_suffix}"
  agent_script="$work_dir/agent-workload.sh"
  cat >"$agent_script" <<EOF_WORKLOAD
#!/bin/sh
sleep $WORKLOAD_SECONDS
echo done
EOF_WORKLOAD
  chmod +x "$agent_script"

  litents_init_ms=$(measure_ms "$run_log" env "${base_env[@]}" "$LITENTS_BIN" init --no-watch --session "$litents_session" "$repo_dir")
  echo "$litents_init_ms" >> "$litents_init_tmp"

  litents_new_ms=$(measure_ms "$run_log" env "${base_env[@]}" "$LITENTS_BIN" new --project "$project_name" --no-worktree --prompt "sleep $WORKLOAD_SECONDS; echo done" "$agent_name")
  echo "$litents_new_ms" >> "$litents_new_tmp"

  sleep 0.05
  litents_status_ms=$(measure_ms "$run_log" env "${base_env[@]}" "$LITENTS_BIN" status --project "$project_name")
  echo "$litents_status_ms" >> "$litents_status_tmp"

  litents_stop_ms=$(measure_ms "$run_log" env "${base_env[@]}" "$LITENTS_BIN" stop --force --project "$project_name" "$agent_name")
  echo "$litents_stop_ms" >> "$litents_stop_tmp"

  litents_clean_ms=$(measure_ms "$run_log" env "${base_env[@]}" "$LITENTS_BIN" clean --project "$project_name" --worktrees)
  echo "$litents_clean_ms" >> "$litents_clean_tmp"

  tmux_init_ms=$(measure_ms "$run_log" tmux new-session -d -s "$tmux_session" -n "home" -c "$repo_dir" "$agent_script")
  echo "$tmux_init_ms" >> "$tmux_init_tmp"

  tmux_new_ms=$(measure_ms "$run_log" tmux new-window -t "$tmux_session" -n "$agent_name" -c "$repo_dir" "$agent_script")
  echo "$tmux_new_ms" >> "$tmux_new_tmp"

  tmux_list_ms=$(measure_ms "$run_log" tmux list-windows -t "$tmux_session")
  echo "$tmux_list_ms" >> "$tmux_list_tmp"

  tmux_kill_ms=$(measure_ms "$run_log" tmux kill-session -t "$tmux_session")
  echo "$tmux_kill_ms" >> "$tmux_kill_tmp"

  rm -rf "$work_dir"

done

litents_init_stats=$(print_stats "$litents_init_tmp")
litents_new_stats=$(print_stats "$litents_new_tmp")
litents_status_stats=$(print_stats "$litents_status_tmp")
litents_stop_stats=$(print_stats "$litents_stop_tmp")
litents_clean_stats=$(print_stats "$litents_clean_tmp")
tmux_init_stats=$(print_stats "$tmux_init_tmp")
tmux_new_stats=$(print_stats "$tmux_new_tmp")
tmux_list_stats=$(print_stats "$tmux_list_tmp")
tmux_kill_stats=$(print_stats "$tmux_kill_tmp")

cat >"$RESULT_FILE" <<EOF_REPORT
# Litents vs Popular CLI Baselines

Generated on: $(date -u "+%Y-%m-%dT%H:%M:%SZ")

Host: $(uname -sm)
Go: $(go version | awk '{print $3}')
Litents binary: $LITENTS_BIN
Litents commit: $(git -C "$ROOT_DIR" rev-parse --short HEAD)

Method:
- Synthetic command workload: \`sleep ${WORKLOAD_SECONDS}; echo done\`
- Number of repeats: $RUNS
- Scope: one agent window in one project, no model/network calls.
- Litents config: \`codex_command: sh\`, \`codex_args: ["-lc"]\`, \`--no-worktree\`, \`--no-watch\`
- tmux baseline: one session + one additional window using the same workload script

### Raw timing summary

Metric | Litents | tmux
---|---|---
Initialize project/session | $litents_init_stats | $tmux_init_stats
Start one agent workload | $litents_new_stats | $tmux_new_stats
Status/list poll | $litents_status_stats | $tmux_list_stats
Stop/cleanup command | $litents_stop_stats | $tmux_kill_stats
Cleanup state files | $litents_clean_stats | N/A


EOF_REPORT

rm -f "$litents_init_tmp" "$litents_new_tmp" "$litents_status_tmp" "$litents_stop_tmp" "$litents_clean_tmp" "$tmux_init_tmp" "$tmux_new_tmp" "$tmux_list_tmp" "$tmux_kill_tmp"

printf 'Comparison report written to: %s\n' "$RESULT_FILE"
