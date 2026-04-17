#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LITENTS_BIN="${LITENTS_BIN:-$ROOT_DIR/.tmp-litents-bench-bin}"
AGENTS="${LITENTS_RUNNING_AGENTS:-5}"
SAMPLES="${LITENTS_RUNNING_SAMPLES:-5}"
SAMPLE_INTERVAL="${LITENTS_RUNNING_SAMPLE_INTERVAL:-1}"
OUT_DIR="${LITENTS_BENCH_OUT:-$ROOT_DIR/benchmarking_md}"
RESULT_FILE="$OUT_DIR/running-agents-resource-results.md"
WORKLOAD_SECONDS=$((SAMPLES * SAMPLE_INTERVAL + 20))

mkdir -p "$OUT_DIR"

if [[ ! -x "$LITENTS_BIN" ]]; then
  go build -o "$LITENTS_BIN" "$ROOT_DIR/cmd/litents"
fi

HAS_ZELLIJ=0
if command -v zellij >/dev/null 2>&1; then
  HAS_ZELLIJ=1
fi

HAS_AOE=0
if command -v aoe >/dev/null 2>&1; then
  HAS_AOE=1
fi

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

  echo "${count} samples, mean=${avg}${unit} (p50=${p50}${unit}, p95=${p95}${unit}, min=${min}${unit}, max=${max}${unit})"
}

unique_numbers() {
  tr ' ' '\n' | awk 'NF && $1 ~ /^[0-9]+$/ && !seen[$1]++ {print $1}' | tr '\n' ' '
}

descendants() {
  local root="$1"
  local queue children child
  queue="$root"
  while [[ -n "$queue" ]]; do
    child="${queue%% *}"
    queue="${queue#"$child"}"
    queue="${queue# }"
    children="$(pgrep -P "$child" 2>/dev/null | tr '\n' ' ' || true)"
    if [[ -n "$children" ]]; then
      printf '%s ' $children
      queue="$queue $children"
    fi
  done
}

process_tree() {
  local roots="$*"
  local all="$roots"
  local pid kids
  for pid in $roots; do
    kids="$(descendants "$pid")"
    all="$all $kids"
  done
  printf '%s\n' "$all" | unique_numbers
}

sample_roots() {
  local roots="$1"
  local rss_file="$2"
  local cpu_file="$3"
  local i pids rows

  for i in $(seq 1 "$SAMPLES"); do
    pids="$(process_tree "$roots")"
    if [[ -z "${pids// /}" ]]; then
      echo "0" >>"$rss_file"
      echo "0" >>"$cpu_file"
    else
      rows="$(ps -o rss= -o pcpu= -p "$(echo "$pids" | tr ' ' ',')" 2>/dev/null || true)"
      if [[ -z "$rows" ]]; then
        echo "0" >>"$rss_file"
        echo "0" >>"$cpu_file"
      else
        awk '
          {rss += $1; cpu += $2}
          END {
            printf "%.2f\n", rss / 1024
            printf "%.2f\n", cpu
          }
        ' <<<"$rows" | {
          read -r rss
          read -r cpu
          echo "$rss" >>"$rss_file"
          echo "$cpu" >>"$cpu_file"
        }
      fi
    fi
    sleep "$SAMPLE_INTERVAL"
  done
}

tmux_server_pid() {
  local session="$1"
  TMUX="" tmux display-message -p -t "$session" '#{pid}' 2>/dev/null || true
}

tmux_session_pane_pids() {
  local session="$1"
  TMUX="" tmux list-panes -a -F '#{session_name} #{pane_pid}' 2>/dev/null | awk -v s="$session" '$1 == s {print $2}' | tr '\n' ' '
}

new_tmp_file() {
  local path
  path="$(mktemp)"
  : >"$path"
  echo "$path"
}

litents_rss_tmp="$(new_tmp_file)"
litents_cpu_tmp="$(new_tmp_file)"
zellij_rss_tmp="$(new_tmp_file)"
zellij_cpu_tmp="$(new_tmp_file)"
aoe_rss_tmp="$(new_tmp_file)"
aoe_cpu_tmp="$(new_tmp_file)"

cleanup_paths=()
cleanup() {
  for item in "${cleanup_paths[@]}"; do
    case "$item" in
      litents:*) TMUX_TMPDIR="${item#litents:}" TMUX="" tmux kill-server >/dev/null 2>&1 || true ;;
      aoe:*) TMUX_TMPDIR="${item#aoe:}" TMUX="" tmux kill-server >/dev/null 2>&1 || true ;;
      zellij:*)
        zellij_env="${item#zellij:}"
        ZELLIJ_CONFIG_DIR="${zellij_env%%|*}" ZELLIJ_DATA_DIR="${zellij_env#*|}" zellij kill-all-sessions >/dev/null 2>&1 || true
        ;;
      dir:*) rm -rf "${item#dir:}" ;;
    esac
  done
  rm -f "$litents_rss_tmp" "$litents_cpu_tmp" "$zellij_rss_tmp" "$zellij_cpu_tmp" "$aoe_rss_tmp" "$aoe_cpu_tmp"
}
trap cleanup EXIT

run_litents() {
  local work_dir repo_dir state_root config_root tmux_tmp session project roots i
  work_dir="$(mktemp -d)"
  cleanup_paths+=("dir:$work_dir")
  repo_dir="$work_dir/repo"
  state_root="$work_dir/state"
  config_root="$work_dir/config"
  tmux_tmp="$work_dir/tmux"
  session="ltrun-litents-$$"
  project="repo"
  mkdir -p "$repo_dir" "$state_root" "$config_root/litents" "$tmux_tmp"
  git -C "$repo_dir" init -q
  cat >"$config_root/litents/config.json" <<EOF_CFG
{
  "tmux_session_prefix": "ltrun",
  "worktree_root": "$work_dir/worktrees",
  "default_base_branch": "main",
  "codex_command": "sh",
  "codex_args": ["-lc"],
  "notify_enabled": false,
  "notify_command": ""
}
EOF_CFG
  cleanup_paths+=("litents:$tmux_tmp")
  env XDG_CONFIG_HOME="$config_root" XDG_STATE_HOME="$state_root" TMUX_TMPDIR="$tmux_tmp" "$LITENTS_BIN" init --no-watch --session "$session" "$repo_dir" >/dev/null
  for i in $(seq 1 "$AGENTS"); do
    env XDG_CONFIG_HOME="$config_root" XDG_STATE_HOME="$state_root" TMUX_TMPDIR="$tmux_tmp" "$LITENTS_BIN" new --project "$project" --no-worktree --prompt "sleep $WORKLOAD_SECONDS" "agent-$i" >/dev/null
  done
  sleep 1
  roots="$(TMUX_TMPDIR="$tmux_tmp" tmux_server_pid "$session") $(TMUX_TMPDIR="$tmux_tmp" tmux_session_pane_pids "$session")"
  sample_roots "$roots" "$litents_rss_tmp" "$litents_cpu_tmp"
}

run_zellij() {
  if (( HAS_ZELLIJ != 1 )); then
    return
  fi
  local work_dir config_dir data_dir session script before after roots i
  work_dir="$(mktemp -d)"
  cleanup_paths+=("dir:$work_dir")
  config_dir="$work_dir/zellij-config"
  data_dir="$work_dir/zellij-data"
  session="ltrun-zellij-$$"
  mkdir -p "$config_dir" "$data_dir"
  cat >"$work_dir/workload.sh" <<EOF_WORKLOAD
#!/bin/sh
sleep $WORKLOAD_SECONDS
EOF_WORKLOAD
  script="$work_dir/workload.sh"
  chmod +x "$script"
  before="$(pgrep -x zellij 2>/dev/null | tr '\n' ' ' || true)"
  cleanup_paths+=("zellij:$config_dir|$data_dir")
  env ZELLIJ_CONFIG_DIR="$config_dir" ZELLIJ_DATA_DIR="$data_dir" zellij attach --create-background "$session" >/dev/null 2>&1
  for i in $(seq 1 "$AGENTS"); do
    env ZELLIJ_CONFIG_DIR="$config_dir" ZELLIJ_DATA_DIR="$data_dir" ZELLIJ_SESSION_NAME="$session" zellij action new-tab -n "agent-$i" -- "$script" >/dev/null 2>&1
  done
  sleep 1
  after="$(pgrep -x zellij 2>/dev/null | tr '\n' ' ' || true)"
  roots="$(tr ' ' '\n' <<<"$after" | awk -v before="$before" '
    BEGIN {
      split(before, b, " ")
      for (i in b) seen[b[i]] = 1
    }
    NF && !seen[$1] {print $1}
  ' | tr '\n' ' ')"
  if [[ -z "${roots// /}" ]]; then
    roots="$after"
  fi
  sample_roots "$roots" "$zellij_rss_tmp" "$zellij_cpu_tmp"
  env ZELLIJ_CONFIG_DIR="$config_dir" ZELLIJ_DATA_DIR="$data_dir" zellij kill-session "$session" >/dev/null 2>&1 || true
}

run_aoe() {
  if (( HAS_AOE != 1 )); then
    return
  fi
  local work_dir repo_dir tmux_tmp profile roots i
  work_dir="$(mktemp -d)"
  cleanup_paths+=("dir:$work_dir")
  repo_dir="$work_dir/repo"
  tmux_tmp="$work_dir/tmux"
  profile="ltrun-aoe-$$"
  mkdir -p "$repo_dir" "$work_dir/bin" "$work_dir/home" "$work_dir/config" "$work_dir/data" "$tmux_tmp"
  git -C "$repo_dir" init -q
  cat >"$work_dir/bin/codex" <<EOF_CODEX
#!/bin/sh
sleep $WORKLOAD_SECONDS
EOF_CODEX
  chmod +x "$work_dir/bin/codex"
  cleanup_paths+=("aoe:$tmux_tmp")
  env HOME="$work_dir/home" XDG_CONFIG_HOME="$work_dir/config" XDG_DATA_HOME="$work_dir/data" TMUX_TMPDIR="$tmux_tmp" PATH="$work_dir/bin:$PATH" aoe init -p "$profile" "$repo_dir" >/dev/null
  for i in $(seq 1 "$AGENTS"); do
    env HOME="$work_dir/home" XDG_CONFIG_HOME="$work_dir/config" XDG_DATA_HOME="$work_dir/data" TMUX_TMPDIR="$tmux_tmp" PATH="$work_dir/bin:$PATH" aoe add -p "$profile" "$repo_dir" -t "agent-$i" -c codex >/dev/null
    env HOME="$work_dir/home" XDG_CONFIG_HOME="$work_dir/config" XDG_DATA_HOME="$work_dir/data" TMUX_TMPDIR="$tmux_tmp" PATH="$work_dir/bin:$PATH" aoe session start -p "$profile" "agent-$i" >/dev/null
  done
  sleep 1
  roots="$(TMUX_TMPDIR="$tmux_tmp" TMUX="" tmux display-message -p '#{pid}' 2>/dev/null || true) $(TMUX_TMPDIR="$tmux_tmp" TMUX="" tmux list-panes -a -F '#{pane_pid}' 2>/dev/null | tr '\n' ' ')"
  sample_roots "$roots" "$aoe_rss_tmp" "$aoe_cpu_tmp"
}

run_litents
run_zellij
run_aoe

litents_rss_stats="$(print_stats "$litents_rss_tmp" "MiB")"
litents_cpu_stats="$(print_stats "$litents_cpu_tmp" "%")"
zellij_rss_stats="$(print_stats "$zellij_rss_tmp" "MiB")"
zellij_cpu_stats="$(print_stats "$zellij_cpu_tmp" "%")"
aoe_rss_stats="$(print_stats "$aoe_rss_tmp" "MiB")"
aoe_cpu_stats="$(print_stats "$aoe_cpu_tmp" "%")"

zellij_version="not installed"
if (( HAS_ZELLIJ == 1 )); then
  zellij_version="$(zellij --version)"
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
# Running Agents CPU/RAM Comparison

Generated on: $(date -u "+%Y-%m-%dT%H:%M:%SZ")

Host: $(uname -sm)
Go: $(go version | awk '{print $3}')
Litents binary: $LITENTS_BIN
Litents source: $litents_ref
Zellij: $zellij_version
Agent of Empires: $aoe_version

Method:
- Launch $AGENTS synthetic running agents per tool.
- Each synthetic agent runs \`sleep $WORKLOAD_SECONDS\`.
- Sample process trees $SAMPLES times at ${SAMPLE_INTERVAL}s intervals while agents are alive.
- RSS is summed resident memory for the tool runtime process tree plus managed agent pane process trees.
- CPU is summed \`ps %cpu\` for the same sampled process tree.
- Litents has no resident daemon; its running cost is the private tmux server plus managed panes.
- Zellij is measured through an isolated background session.
- Agent of Empires is measured through a private tmux server and a temporary fake \`codex\` command.

## Steady-state running agents summary

Tool | Running agents | RAM RSS | CPU %
---|---:|---:|---:
Litents | $AGENTS | $litents_rss_stats | $litents_cpu_stats
Zellij | $AGENTS | $zellij_rss_stats | $zellij_cpu_stats
Agent of Empires | $AGENTS | $aoe_rss_stats | $aoe_cpu_stats

## Interpretation

This benchmark answers the steady-state operator question: how much CPU and RAM are consumed while agents are already running. It intentionally excludes GUI-only products and tools that do not expose a reproducible headless running-agent lifecycle.
EOF_REPORT

printf 'Running agents resource report written to: %s\n' "$RESULT_FILE"
