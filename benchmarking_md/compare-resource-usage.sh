#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LITENTS_BIN="${LITENTS_BIN:-$ROOT_DIR/.tmp-litents-bench-bin}"
RUNS="${LITENTS_RESOURCE_RUNS:-10}"
WORKLOAD_SECONDS="${LITENTS_RESOURCE_SLEEP_SECONDS:-5}"
OUT_DIR="${LITENTS_BENCH_OUT:-$ROOT_DIR/benchmarking_md}"
RESULT_FILE="$OUT_DIR/resource-comparison-results.md"

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

HAS_AOE=0
if command -v aoe >/dev/null 2>&1; then
  HAS_AOE=1
fi

measure_resources() {
  local log_file="$1"
  shift

  local time_file rc rss_mib cpu_ms
  time_file="${log_file}.time"

  if [[ "$(uname -s)" == "Darwin" ]]; then
    set +e
    /usr/bin/time -l "$@" >"$log_file" 2>"$time_file"
    rc=$?
    set -e
    if (( rc != 0 )); then
      echo "Command failed (exit=$rc): $*" >&2
      echo "Log: $log_file" >&2
      tail -n 20 "$time_file" >&2 || true
      return "$rc"
    fi
    rss_mib=$(awk '/maximum resident set size/ {printf "%.2f", $1 / 1024 / 1024; found=1} END {if (!found) exit 1}' "$time_file")
    cpu_ms=$(awk '/ real / {printf "%.0f", ($3 + $5) * 1000; found=1} END {if (!found) exit 1}' "$time_file")
  else
    set +e
    /usr/bin/time -v "$@" >"$log_file" 2>"$time_file"
    rc=$?
    set -e
    if (( rc != 0 )); then
      echo "Command failed (exit=$rc): $*" >&2
      echo "Log: $log_file" >&2
      tail -n 20 "$time_file" >&2 || true
      return "$rc"
    fi
    rss_mib=$(awk -F: '/Maximum resident set size/ {gsub(/^[ \t]+/, "", $2); printf "%.2f", $2 / 1024; found=1} END {if (!found) exit 1}' "$time_file")
    cpu_ms=$(awk -F: '/User time/ {gsub(/^[ \t]+/, "", $2); user=$2} /System time/ {gsub(/^[ \t]+/, "", $2); sys=$2} END {if (user == "" || sys == "") exit 1; printf "%.0f", (user + sys) * 1000}' "$time_file")
  fi

  rm -f "$time_file"
  printf '%s %s\n' "$rss_mib" "$cpu_ms"
}

record_resources() {
  local rss_file="$1"
  local cpu_file="$2"
  local log_file="$3"
  shift 3

  local result rss cpu
  result=$(measure_resources "$log_file" "$@")
  rss="${result%% *}"
  cpu="${result##* }"
  echo "$rss" >>"$rss_file"
  echo "$cpu" >>"$cpu_file"
}

print_stats() {
  local file="$1"
  local unit="$2"
  if [[ ! -s "$file" ]]; then
    echo "n/a"
    return
  fi

  local count min max avg p50 p95 sorted
  count=$(wc -l <"$file")
  if (( count == 0 )); then
    echo "n/a"
    return
  fi

  sorted="$(mktemp)"
  sort -n "$file" >"$sorted"
  min=$(head -n 1 "$sorted")
  max=$(tail -n 1 "$sorted")
  avg=$(awk '{sum += $1} END {if (NR == 0) print 0; else printf "%.2f", sum / NR}' "$file")
  p50=$(awk -v n="$count" 'NR == int((n + 1) / 2){print; exit}' "$sorted")
  p95=$(awk -v n="$count" 'NR >= int((n - 1) * 0.95 + 1){print; exit}' "$sorted")
  rm -f "$sorted"

  echo "${count} runs, mean=${avg}${unit} (p50=${p50}${unit}, p95=${p95}${unit}, min=${min}${unit}, max=${max}${unit})"
}

make_tmp() {
  local file
  file="$(mktemp)"
  : >"$file"
  echo "$file"
}

litents_init_rss_tmp="$(make_tmp)"
litents_new_rss_tmp="$(make_tmp)"
litents_status_rss_tmp="$(make_tmp)"
litents_stop_rss_tmp="$(make_tmp)"
litents_clean_rss_tmp="$(make_tmp)"
zellij_init_rss_tmp="$(make_tmp)"
zellij_new_rss_tmp="$(make_tmp)"
zellij_list_rss_tmp="$(make_tmp)"
zellij_kill_rss_tmp="$(make_tmp)"
codex_app_start_rss_tmp="$(make_tmp)"
codex_app_health_rss_tmp="$(make_tmp)"
codex_app_stop_rss_tmp="$(make_tmp)"
aoe_init_rss_tmp="$(make_tmp)"
aoe_start_rss_tmp="$(make_tmp)"
aoe_status_rss_tmp="$(make_tmp)"
aoe_stop_rss_tmp="$(make_tmp)"
aoe_remove_rss_tmp="$(make_tmp)"

litents_init_cpu_tmp="$(make_tmp)"
litents_new_cpu_tmp="$(make_tmp)"
litents_status_cpu_tmp="$(make_tmp)"
litents_stop_cpu_tmp="$(make_tmp)"
litents_clean_cpu_tmp="$(make_tmp)"
zellij_init_cpu_tmp="$(make_tmp)"
zellij_new_cpu_tmp="$(make_tmp)"
zellij_list_cpu_tmp="$(make_tmp)"
zellij_kill_cpu_tmp="$(make_tmp)"
codex_app_start_cpu_tmp="$(make_tmp)"
codex_app_health_cpu_tmp="$(make_tmp)"
codex_app_stop_cpu_tmp="$(make_tmp)"
aoe_init_cpu_tmp="$(make_tmp)"
aoe_start_cpu_tmp="$(make_tmp)"
aoe_status_cpu_tmp="$(make_tmp)"
aoe_stop_cpu_tmp="$(make_tmp)"
aoe_remove_cpu_tmp="$(make_tmp)"

for i in $(seq 1 "$RUNS"); do
  work_dir="$(mktemp -d)"
  repo_dir="$work_dir/repo"
  state_root="$work_dir/state"
  config_root="$work_dir/config"
  run_log="$work_dir/run.log"
  session_suffix="resource-$i"
  project_name="repo"
  agent_name="agent-1"

  mkdir -p "$repo_dir" "$state_root" "$config_root/litents"
  git -C "$repo_dir" init -q

  cat >"$config_root/litents/config.json" <<EOF_CFG
{
  "tmux_session_prefix": "ltrsrc-${session_suffix}",
  "worktree_root": "$work_dir/worktrees",
  "default_base_branch": "main",
  "codex_command": "sh",
  "codex_args": ["-lc"],
  "notify_enabled": false,
  "notify_command": ""
}
EOF_CFG

  base_env=(XDG_CONFIG_HOME="$config_root" XDG_STATE_HOME="$state_root")
  litents_session="ltrsrc-${session_suffix}"
  agent_script="$work_dir/agent-workload.sh"
  cat >"$agent_script" <<EOF_WORKLOAD
#!/bin/sh
sleep $WORKLOAD_SECONDS
echo done
EOF_WORKLOAD
  chmod +x "$agent_script"

  if (( HAS_AOE == 1 )); then
    mkdir -p "$work_dir/bin" "$work_dir/home" "$work_dir/aoe-config" "$work_dir/aoe-data"
    cat >"$work_dir/bin/codex" <<EOF_CODEX
#!/bin/sh
sleep $WORKLOAD_SECONDS
echo done
EOF_CODEX
    chmod +x "$work_dir/bin/codex"
  fi

  record_resources "$litents_init_rss_tmp" "$litents_init_cpu_tmp" "$run_log" env "${base_env[@]}" "$LITENTS_BIN" init --no-watch --session "$litents_session" "$repo_dir"
  record_resources "$litents_new_rss_tmp" "$litents_new_cpu_tmp" "$run_log" env "${base_env[@]}" "$LITENTS_BIN" new --project "$project_name" --no-worktree --prompt "sleep $WORKLOAD_SECONDS; echo done" "$agent_name"
  record_resources "$litents_status_rss_tmp" "$litents_status_cpu_tmp" "$run_log" env "${base_env[@]}" "$LITENTS_BIN" status --project "$project_name"
  record_resources "$litents_stop_rss_tmp" "$litents_stop_cpu_tmp" "$run_log" env "${base_env[@]}" "$LITENTS_BIN" stop --force --project "$project_name" "$agent_name"
  record_resources "$litents_clean_rss_tmp" "$litents_clean_cpu_tmp" "$run_log" env "${base_env[@]}" "$LITENTS_BIN" clean --project "$project_name" --worktrees

  if (( HAS_ZELLIJ == 1 )); then
    zellij_session="ltzrsrc-${session_suffix}"
    zellij_config="$work_dir/zellij-config"
    zellij_data="$work_dir/zellij-data"
    mkdir -p "$zellij_config" "$zellij_data"

    record_resources "$zellij_init_rss_tmp" "$zellij_init_cpu_tmp" "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" zellij attach --create-background "$zellij_session"
    record_resources "$zellij_new_rss_tmp" "$zellij_new_cpu_tmp" "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" ZELLIJ_SESSION_NAME="$zellij_session" zellij action new-tab -n "$agent_name" -- "$agent_script"
    record_resources "$zellij_list_rss_tmp" "$zellij_list_cpu_tmp" "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" ZELLIJ_SESSION_NAME="$zellij_session" zellij action list-tabs
    record_resources "$zellij_kill_rss_tmp" "$zellij_kill_cpu_tmp" "$run_log" env ZELLIJ_CONFIG_DIR="$zellij_config" ZELLIJ_DATA_DIR="$zellij_data" zellij kill-session "$zellij_session"
  fi

  if (( HAS_CODEX_APP_SERVER == 1 )); then
    codex_port=$((46000 + (RANDOM % 9000)))
    codex_log="$work_dir/codex-app-server.log"
    codex_pid_file="$work_dir/codex-app-server.pid"
    codex_url="http://127.0.0.1:${codex_port}/healthz"
    codex_start_script="$work_dir/codex-start.sh"
    codex_stop_script="$work_dir/codex-stop.sh"

    cat >"$codex_start_script" <<EOF_CODEX_START
#!/bin/sh
set -e
codex app-server --listen "ws://127.0.0.1:${codex_port}" >"$codex_log" 2>&1 &
echo \$! >"$codex_pid_file"
for _ in \$(seq 1 100); do
  if curl -fsS "$codex_url" >/dev/null 2>&1; then
    exit 0
  fi
  if ! kill -0 "\$(cat "$codex_pid_file")" >/dev/null 2>&1; then
    exit 1
  fi
  sleep 0.02
done
exit 1
EOF_CODEX_START
    chmod +x "$codex_start_script"

    cat >"$codex_stop_script" <<EOF_CODEX_STOP
#!/bin/sh
set -e
if [ -s "$codex_pid_file" ]; then
  kill "\$(cat "$codex_pid_file")" >/dev/null 2>&1 || true
  wait "\$(cat "$codex_pid_file")" >/dev/null 2>&1 || true
fi
EOF_CODEX_STOP
    chmod +x "$codex_stop_script"

    record_resources "$codex_app_start_rss_tmp" "$codex_app_start_cpu_tmp" "$run_log" "$codex_start_script"
    record_resources "$codex_app_health_rss_tmp" "$codex_app_health_cpu_tmp" "$run_log" curl -fsS "$codex_url"
    record_resources "$codex_app_stop_rss_tmp" "$codex_app_stop_cpu_tmp" "$run_log" "$codex_stop_script"
  fi

  if (( HAS_AOE == 1 )); then
    aoe_repo_dir="$work_dir/aoe-repo"
    aoe_profile="litents-resource-${session_suffix}"
    aoe_session="litents-resource-aoe-${i}"
    aoe_env=(HOME="$work_dir/home" XDG_CONFIG_HOME="$work_dir/aoe-config" XDG_DATA_HOME="$work_dir/aoe-data" PATH="$work_dir/bin:$PATH")
    aoe_start_script="$work_dir/aoe-start.sh"

    mkdir -p "$aoe_repo_dir"
    git -C "$aoe_repo_dir" init -q

    cat >"$aoe_start_script" <<EOF_AOE_START
#!/bin/sh
set -e
aoe add -p "$aoe_profile" "$aoe_repo_dir" -t "$aoe_session" -c codex
aoe session start -p "$aoe_profile" "$aoe_session"
EOF_AOE_START
    chmod +x "$aoe_start_script"

    record_resources "$aoe_init_rss_tmp" "$aoe_init_cpu_tmp" "$run_log" env "${aoe_env[@]}" aoe init -p "$aoe_profile" "$aoe_repo_dir"
    record_resources "$aoe_start_rss_tmp" "$aoe_start_cpu_tmp" "$run_log" env "${aoe_env[@]}" "$aoe_start_script"
    record_resources "$aoe_status_rss_tmp" "$aoe_status_cpu_tmp" "$run_log" env "${aoe_env[@]}" aoe status -p "$aoe_profile" --json
    record_resources "$aoe_stop_rss_tmp" "$aoe_stop_cpu_tmp" "$run_log" env "${aoe_env[@]}" aoe session stop -p "$aoe_profile" "$aoe_session"
    record_resources "$aoe_remove_rss_tmp" "$aoe_remove_cpu_tmp" "$run_log" env "${aoe_env[@]}" aoe remove -p "$aoe_profile" "$aoe_session" --force
  fi

  rm -rf "$work_dir"
done

litents_init_rss_stats=$(print_stats "$litents_init_rss_tmp" "MiB")
litents_new_rss_stats=$(print_stats "$litents_new_rss_tmp" "MiB")
litents_status_rss_stats=$(print_stats "$litents_status_rss_tmp" "MiB")
litents_stop_rss_stats=$(print_stats "$litents_stop_rss_tmp" "MiB")
litents_clean_rss_stats=$(print_stats "$litents_clean_rss_tmp" "MiB")
zellij_init_rss_stats=$(print_stats "$zellij_init_rss_tmp" "MiB")
zellij_new_rss_stats=$(print_stats "$zellij_new_rss_tmp" "MiB")
zellij_list_rss_stats=$(print_stats "$zellij_list_rss_tmp" "MiB")
zellij_kill_rss_stats=$(print_stats "$zellij_kill_rss_tmp" "MiB")
codex_app_start_rss_stats=$(print_stats "$codex_app_start_rss_tmp" "MiB")
codex_app_health_rss_stats=$(print_stats "$codex_app_health_rss_tmp" "MiB")
codex_app_stop_rss_stats=$(print_stats "$codex_app_stop_rss_tmp" "MiB")
aoe_init_rss_stats=$(print_stats "$aoe_init_rss_tmp" "MiB")
aoe_start_rss_stats=$(print_stats "$aoe_start_rss_tmp" "MiB")
aoe_status_rss_stats=$(print_stats "$aoe_status_rss_tmp" "MiB")
aoe_stop_rss_stats=$(print_stats "$aoe_stop_rss_tmp" "MiB")
aoe_remove_rss_stats=$(print_stats "$aoe_remove_rss_tmp" "MiB")

litents_init_cpu_stats=$(print_stats "$litents_init_cpu_tmp" "ms")
litents_new_cpu_stats=$(print_stats "$litents_new_cpu_tmp" "ms")
litents_status_cpu_stats=$(print_stats "$litents_status_cpu_tmp" "ms")
litents_stop_cpu_stats=$(print_stats "$litents_stop_cpu_tmp" "ms")
litents_clean_cpu_stats=$(print_stats "$litents_clean_cpu_tmp" "ms")
zellij_init_cpu_stats=$(print_stats "$zellij_init_cpu_tmp" "ms")
zellij_new_cpu_stats=$(print_stats "$zellij_new_cpu_tmp" "ms")
zellij_list_cpu_stats=$(print_stats "$zellij_list_cpu_tmp" "ms")
zellij_kill_cpu_stats=$(print_stats "$zellij_kill_cpu_tmp" "ms")
codex_app_start_cpu_stats=$(print_stats "$codex_app_start_cpu_tmp" "ms")
codex_app_health_cpu_stats=$(print_stats "$codex_app_health_cpu_tmp" "ms")
codex_app_stop_cpu_stats=$(print_stats "$codex_app_stop_cpu_tmp" "ms")
aoe_init_cpu_stats=$(print_stats "$aoe_init_cpu_tmp" "ms")
aoe_start_cpu_stats=$(print_stats "$aoe_start_cpu_tmp" "ms")
aoe_status_cpu_stats=$(print_stats "$aoe_status_cpu_tmp" "ms")
aoe_stop_cpu_stats=$(print_stats "$aoe_stop_cpu_tmp" "ms")
aoe_remove_cpu_stats=$(print_stats "$aoe_remove_cpu_tmp" "ms")

zellij_version="not installed"
if (( HAS_ZELLIJ == 1 )); then
  zellij_version="$(zellij --version)"
fi

codex_version="not installed"
if command -v codex >/dev/null 2>&1; then
  codex_version="$(codex --version)"
fi

aoe_version="not installed"
if (( HAS_AOE == 1 )); then
  aoe_version="$(aoe --version)"
fi

litents_ref="$(git -C "$ROOT_DIR" rev-parse --short HEAD)"
if ! git -C "$ROOT_DIR" diff --quiet || ! git -C "$ROOT_DIR" diff --cached --quiet; then
  litents_ref="${litents_ref}+local"
fi

cat >"$RESULT_FILE" <<EOF_REPORT
# Litents Resource Usage Comparison

Generated on: $(date -u "+%Y-%m-%dT%H:%M:%SZ")

Host: $(uname -sm)
Go: $(go version | awk '{print $3}')
Litents binary: $LITENTS_BIN
Litents source: $litents_ref
Zellij: $zellij_version
Codex: $codex_version
Agent of Empires: $aoe_version

Method:
- Synthetic command workload: \`sleep ${WORKLOAD_SECONDS}; echo done\`
- Number of repeats: $RUNS
- Scope: peak RSS and CPU time for lifecycle commands in a headless shell harness.
- Memory unit: MiB, from \`/usr/bin/time -l\` on macOS or \`/usr/bin/time -v\` on Linux.
- CPU unit: user + system CPU milliseconds for the measured command.
- This is not a full desktop GUI memory profile and does not include terminal emulator memory.

### Peak RSS summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface | $litents_init_rss_stats | $zellij_init_rss_stats | $codex_app_start_rss_stats | $aoe_init_rss_stats
Start one workload | $litents_new_rss_stats | $zellij_new_rss_stats | N/A | $aoe_start_rss_stats
Status/list/health poll | $litents_status_rss_stats | $zellij_list_rss_stats | $codex_app_health_rss_stats | $aoe_status_rss_stats
Stop control surface | $litents_stop_rss_stats | $zellij_kill_rss_stats | $codex_app_stop_rss_stats | $aoe_stop_rss_stats
Cleanup state files | $litents_clean_rss_stats | N/A | N/A | $aoe_remove_rss_stats

### CPU time summary

Metric | Litents | Zellij | Codex app-server | Agent of Empires
---|---|---|---|---
Initialize control surface | $litents_init_cpu_stats | $zellij_init_cpu_stats | $codex_app_start_cpu_stats | $aoe_init_cpu_stats
Start one workload | $litents_new_cpu_stats | $zellij_new_cpu_stats | N/A | $aoe_start_cpu_stats
Status/list/health poll | $litents_status_cpu_stats | $zellij_list_cpu_stats | $codex_app_health_cpu_stats | $aoe_status_cpu_stats
Stop control surface | $litents_stop_cpu_stats | $zellij_kill_cpu_stats | $codex_app_stop_cpu_stats | $aoe_stop_cpu_stats
Cleanup state files | $litents_clean_cpu_stats | N/A | N/A | $aoe_remove_cpu_stats

### Interpretation

Litents has no resident daemon after lifecycle commands exit. These results are best read as command peak memory and CPU cost for orchestrating local agent sessions, not as a full memory census of terminal emulators or desktop apps.
EOF_REPORT

rm -f "$litents_init_rss_tmp" "$litents_new_rss_tmp" "$litents_status_rss_tmp" "$litents_stop_rss_tmp" "$litents_clean_rss_tmp" "$zellij_init_rss_tmp" "$zellij_new_rss_tmp" "$zellij_list_rss_tmp" "$zellij_kill_rss_tmp" "$codex_app_start_rss_tmp" "$codex_app_health_rss_tmp" "$codex_app_stop_rss_tmp" "$aoe_init_rss_tmp" "$aoe_start_rss_tmp" "$aoe_status_rss_tmp" "$aoe_stop_rss_tmp" "$aoe_remove_rss_tmp"
rm -f "$litents_init_cpu_tmp" "$litents_new_cpu_tmp" "$litents_status_cpu_tmp" "$litents_stop_cpu_tmp" "$litents_clean_cpu_tmp" "$zellij_init_cpu_tmp" "$zellij_new_cpu_tmp" "$zellij_list_cpu_tmp" "$zellij_kill_cpu_tmp" "$codex_app_start_cpu_tmp" "$codex_app_health_cpu_tmp" "$codex_app_stop_cpu_tmp" "$aoe_init_cpu_tmp" "$aoe_start_cpu_tmp" "$aoe_status_cpu_tmp" "$aoe_stop_cpu_tmp" "$aoe_remove_cpu_tmp"

printf 'Resource report written to: %s\n' "$RESULT_FILE"
