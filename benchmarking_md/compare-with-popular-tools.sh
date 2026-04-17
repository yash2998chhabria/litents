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

HAS_ZELLIJ=0
if command -v zellij >/dev/null 2>&1; then
  HAS_ZELLIJ=1
fi

HAS_CODEX_APP_SERVER=0
if command -v codex >/dev/null 2>&1 && command -v curl >/dev/null 2>&1; then
  HAS_CODEX_APP_SERVER=1
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
zellij_init_tmp="$(mktemp)"
zellij_new_tmp="$(mktemp)"
zellij_list_tmp="$(mktemp)"
zellij_kill_tmp="$(mktemp)"
codex_app_start_tmp="$(mktemp)"
codex_app_health_tmp="$(mktemp)"
codex_app_stop_tmp="$(mktemp)"

: >"$litents_init_tmp"
: >"$litents_new_tmp"
: >"$litents_status_tmp"
: >"$litents_stop_tmp"
: >"$litents_clean_tmp"
: >"$tmux_init_tmp"
: >"$tmux_new_tmp"
: >"$tmux_list_tmp"
: >"$tmux_kill_tmp"
: >"$zellij_init_tmp"
: >"$zellij_new_tmp"
: >"$zellij_list_tmp"
: >"$zellij_kill_tmp"
: >"$codex_app_start_tmp"
: >"$codex_app_health_tmp"
: >"$codex_app_stop_tmp"

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

  if (( HAS_ZELLIJ == 1 )); then
    zellij_session="ltzel-${session_suffix}"
    zellij_config="$work_dir/zellij-config"
    zellij_data="$work_dir/zellij-data"
    mkdir -p "$zellij_config" "$zellij_data"

    zellij_init_ms=$(measure_ms "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" zellij attach --create-background "$zellij_session")
    echo "$zellij_init_ms" >> "$zellij_init_tmp"

    zellij_new_ms=$(measure_ms "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" ZELLIJ_SESSION_NAME="$zellij_session" zellij action new-tab -n "$agent_name" -- "$agent_script")
    echo "$zellij_new_ms" >> "$zellij_new_tmp"

    zellij_list_ms=$(measure_ms "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" ZELLIJ_SESSION_NAME="$zellij_session" zellij action list-tabs)
    echo "$zellij_list_ms" >> "$zellij_list_tmp"

    zellij_kill_ms=$(measure_ms "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" zellij kill-session "$zellij_session")
    echo "$zellij_kill_ms" >> "$zellij_kill_tmp"
  fi

  if (( HAS_CODEX_APP_SERVER == 1 )); then
    codex_port=$((45000 + (RANDOM % 10000)))
    codex_log="$work_dir/codex-app-server.log"
    codex_url="http://127.0.0.1:${codex_port}/healthz"

    start=$(now_ns)
    codex app-server --listen "ws://127.0.0.1:${codex_port}" >"$codex_log" 2>&1 &
    codex_pid=$!
    codex_ready=0
    for _ in $(seq 1 100); do
      if curl -fsS "$codex_url" >/dev/null 2>&1; then
        codex_ready=1
        break
      fi
      if ! kill -0 "$codex_pid" >/dev/null 2>&1; then
        break
      fi
      sleep 0.02
    done
    end=$(now_ns)
    if (( codex_ready != 1 )); then
      echo "codex app-server failed to become ready" >&2
      echo "Log: $codex_log" >&2
      tail -n 20 "$codex_log" >&2 || true
      kill "$codex_pid" >/dev/null 2>&1 || true
      wait "$codex_pid" >/dev/null 2>&1 || true
      exit 1
    fi
    echo $(( (end - start) / 1000000 )) >> "$codex_app_start_tmp"

    codex_health_ms=$(measure_ms "$run_log" curl -fsS "$codex_url")
    echo "$codex_health_ms" >> "$codex_app_health_tmp"

    start=$(now_ns)
    kill "$codex_pid" >/dev/null 2>&1 || true
    wait "$codex_pid" >/dev/null 2>&1 || true
    end=$(now_ns)
    echo $(( (end - start) / 1000000 )) >> "$codex_app_stop_tmp"
  fi

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
zellij_init_stats=$(print_stats "$zellij_init_tmp")
zellij_new_stats=$(print_stats "$zellij_new_tmp")
zellij_list_stats=$(print_stats "$zellij_list_tmp")
zellij_kill_stats=$(print_stats "$zellij_kill_tmp")
codex_app_start_stats=$(print_stats "$codex_app_start_tmp")
codex_app_health_stats=$(print_stats "$codex_app_health_tmp")
codex_app_stop_stats=$(print_stats "$codex_app_stop_tmp")

zellij_version="not installed"
if (( HAS_ZELLIJ == 1 )); then
  zellij_version="$(zellij --version)"
fi

codex_version="not installed"
if command -v codex >/dev/null 2>&1; then
  codex_version="$(codex --version)"
fi

litents_ref="$(git -C "$ROOT_DIR" rev-parse --short HEAD)"
if ! git -C "$ROOT_DIR" diff --quiet || ! git -C "$ROOT_DIR" diff --cached --quiet; then
  litents_ref="${litents_ref}+local"
fi

cat >"$RESULT_FILE" <<EOF_REPORT
# Litents vs Popular CLI Baselines

Generated on: $(date -u "+%Y-%m-%dT%H:%M:%SZ")

Host: $(uname -sm)
Go: $(go version | awk '{print $3}')
Litents binary: $LITENTS_BIN
Litents source: $litents_ref
tmux: $(tmux -V)
Zellij: $zellij_version
Codex: $codex_version

Method:
- Synthetic command workload: \`sleep ${WORKLOAD_SECONDS}; echo done\`
- Number of repeats: $RUNS
- Scope: one agent window in one project, no model/network calls.
- Litents config: \`codex_command: sh\`, \`codex_args: ["-lc"]\`, \`--no-worktree\`, \`--no-watch\`
- tmux baseline: one session + one additional window using the same workload script
- Zellij baseline: one detached background session + one tab using the same workload script
- Codex app-server baseline: \`codex app-server --listen ws://127.0.0.1:<port>\` startup, health check, and process stop
- Codex desktop app itself is not measured here because launching and driving a macOS GUI app is not reproducible in this headless harness

### Raw timing summary

Metric | Litents | tmux | Zellij | Codex app-server
---|---|---|---|---
Initialize control surface | $litents_init_stats | $tmux_init_stats | $zellij_init_stats | $codex_app_start_stats
Start one workload | $litents_new_stats | $tmux_new_stats | $zellij_new_stats | N/A
Status/list/health poll | $litents_status_stats | $tmux_list_stats | $zellij_list_stats | $codex_app_health_stats
Stop control surface | $litents_stop_stats | $tmux_kill_stats | $zellij_kill_stats | $codex_app_stop_stats
Cleanup state files | $litents_clean_stats | N/A | N/A | N/A


EOF_REPORT

rm -f "$litents_init_tmp" "$litents_new_tmp" "$litents_status_tmp" "$litents_stop_tmp" "$litents_clean_tmp" "$tmux_init_tmp" "$tmux_new_tmp" "$tmux_list_tmp" "$tmux_kill_tmp" "$zellij_init_tmp" "$zellij_new_tmp" "$zellij_list_tmp" "$zellij_kill_tmp" "$codex_app_start_tmp" "$codex_app_health_tmp" "$codex_app_stop_tmp"

printf 'Comparison report written to: %s\n' "$RESULT_FILE"
