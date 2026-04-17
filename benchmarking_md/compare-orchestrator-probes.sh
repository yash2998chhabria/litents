#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LITENTS_BIN="${LITENTS_BIN:-$ROOT_DIR/.tmp-litents-bench-bin}"
RUNS="${LITENTS_PROBE_RUNS:-20}"
OUT_DIR="${LITENTS_BENCH_OUT:-$ROOT_DIR/benchmarking_md}"
RESULT_FILE="$OUT_DIR/orchestrator-probe-results.md"

mkdir -p "$OUT_DIR"

if [[ ! -x "$LITENTS_BIN" ]]; then
  go build -o "$LITENTS_BIN" "$ROOT_DIR/cmd/litents"
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
      tail -n 20 "$time_file" >&2 || true
      return "$rc"
    fi
    rss_mib=$(awk -F: '/Maximum resident set size/ {gsub(/^[ \t]+/, "", $2); printf "%.2f", $2 / 1024; found=1} END {if (!found) exit 1}' "$time_file")
    cpu_ms=$(awk -F: '/User time/ {gsub(/^[ \t]+/, "", $2); user=$2} /System time/ {gsub(/^[ \t]+/, "", $2); sys=$2} END {if (user == "" || sys == "") exit 1; printf "%.0f", (user + sys) * 1000}' "$time_file")
  fi

  rm -f "$time_file"
  printf '%s %s\n' "$rss_mib" "$cpu_ms"
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

app_version() {
  local app_path="$1"
  if [[ -d "$app_path" ]]; then
    defaults read "$app_path/Contents/Info" CFBundleShortVersionString 2>/dev/null || echo "installed"
  else
    echo "not installed"
  fi
}

measure_probe() {
  local name="$1"
  local status="$2"
  local category="$3"
  local command_display="$4"
  shift 4

  local rss_tmp cpu_tmp log_file result rss cpu rss_stats cpu_stats
  rss_tmp="$(mktemp)"
  cpu_tmp="$(mktemp)"
  : >"$rss_tmp"
  : >"$cpu_tmp"

  for i in $(seq 1 "$RUNS"); do
    log_file="$(mktemp)"
    result=$(measure_resources "$log_file" "$@")
    rss="${result%% *}"
    cpu="${result##* }"
    echo "$rss" >>"$rss_tmp"
    echo "$cpu" >>"$cpu_tmp"
    rm -f "$log_file"
  done

  rss_stats=$(print_stats "$rss_tmp" "MiB")
  cpu_stats=$(print_stats "$cpu_tmp" "ms")
  rm -f "$rss_tmp" "$cpu_tmp"

  printf '| %s | %s | %s | `%s` | %s | %s |\n' "$name" "$status" "$category" "$command_display" "$rss_stats" "$cpu_stats" >>"$RESULT_FILE"
}

cat >"$RESULT_FILE" <<EOF_REPORT
# Litents Orchestrator Probe Results

Generated on: $(date -u "+%Y-%m-%dT%H:%M:%SZ")

Host: $(uname -sm)
Go: $(go version | awk '{print $3}')
Runs per installed CLI probe: $RUNS

Method:
- This file covers every orchestrator named in the comparison list.
- CLI/TUI tools get a reproducible version/help command probe for peak RSS and CPU time.
- GUI-first products are reported with install/discovery status, not fake headless lifecycle numbers.
- Full lifecycle latency is in [tool-comparison-results.md](tool-comparison-results.md).
- Lifecycle command resource usage is in [resource-comparison-results.md](resource-comparison-results.md).

## Installed CLI/TUI probe measurements

| Product | Local status | Category | Probe command | Peak RSS | CPU time |
| --- | --- | --- | --- | ---: | ---: |
EOF_REPORT

measure_probe "Litents" "local source build" "project tool" "litents doctor" "$LITENTS_BIN" doctor

if command -v zellij >/dev/null 2>&1; then
  measure_probe "Zellij" "$(zellij --version)" "automated lifecycle peer" "zellij --version" zellij --version
fi

if command -v codex >/dev/null 2>&1; then
  measure_probe "Codex CLI" "$(codex --version)" "runtime/server substrate" "codex --version" codex --version
fi

if command -v aoe >/dev/null 2>&1; then
  measure_probe "Agent of Empires" "$(aoe --version)" "automated lifecycle peer" "aoe --version" aoe --version
fi

if command -v claude-squad >/dev/null 2>&1; then
  measure_probe "Claude Squad" "$(claude-squad version | head -n 1)" "installed CLI/TUI peer" "claude-squad version" claude-squad version
fi

if command -v ccmanager >/dev/null 2>&1; then
  measure_probe "CCManager" "$(ccmanager --version)" "installed CLI/TUI peer" "ccmanager --version" ccmanager --version
fi

sidecar_bin=""
if command -v sidecar >/dev/null 2>&1; then
  sidecar_bin="$(command -v sidecar)"
elif [[ -x "$HOME/go/bin/sidecar" ]]; then
  sidecar_bin="$HOME/go/bin/sidecar"
fi
if [[ -n "$sidecar_bin" ]]; then
  measure_probe "Sidecar Workspaces" "$("$sidecar_bin" --version)" "installed CLI/TUI peer" "sidecar --version" "$sidecar_bin" --version
fi

if command -v agent-hand >/dev/null 2>&1; then
  measure_probe "Agent Hand" "installed" "installed CLI/TUI peer" "agent-hand --version" agent-hand --version
fi

if command -v agent-deck >/dev/null 2>&1; then
  measure_probe "Agent Deck" "installed" "installed CLI/TUI peer" "agent-deck --version" agent-deck --version
fi

cat >>"$RESULT_FILE" <<EOF_REPORT

## Full comparison coverage matrix

| Product | Coverage in this repo | Local result |
| --- | --- | --- |
| Litents | Lifecycle latency, lifecycle resource usage, CLI probe | Source build benchmarked locally |
| Zellij | Lifecycle latency, lifecycle resource usage, CLI probe | $(if command -v zellij >/dev/null 2>&1; then zellij --version; else echo "not installed"; fi) |
| Codex app-server | Lifecycle latency, lifecycle resource usage, CLI probe | $(if command -v codex >/dev/null 2>&1; then codex --version; else echo "not installed"; fi) |
| Agent of Empires | Lifecycle latency, lifecycle resource usage, CLI probe | $(if command -v aoe >/dev/null 2>&1; then aoe --version; else echo "not installed"; fi) |
| Claude Squad | CLI/version/RSS/CPU probe | $(if command -v claude-squad >/dev/null 2>&1; then claude-squad version | head -n 1; else echo "not installed"; fi) |
| CCManager | CLI/version/RSS/CPU probe | $(if command -v ccmanager >/dev/null 2>&1; then ccmanager --version; else echo "not installed"; fi) |
| Sidecar Workspaces | CLI/version/RSS/CPU probe | $(if [[ -n "$sidecar_bin" ]]; then "$sidecar_bin" --version; else echo "not installed"; fi) |
| Crystal | GUI install/probe only | Crystal.app $(app_version "/Applications/Crystal.app") |
| Conductor | GUI product workflow only | $(app_version "/Applications/Conductor.app") |
| CodeAgentSwarm | GUI product workflow only | $(app_version "/Applications/CodeAgentSwarm.app") |
| Termyx | GUI product workflow only | $(app_version "/Applications/Termyx.app") |
| Agent Hand | Blocked/probe when installable | $(if command -v agent-hand >/dev/null 2>&1; then echo "installed"; else echo "not installed; installer/source/brew release asset blocked in this environment"; fi) |
| Agent Deck | Blocked/probe when installable | $(if command -v agent-deck >/dev/null 2>&1; then echo "installed"; else echo "not installed; release asset/brew fallback blocked in this environment"; fi) |

## Notes

- Lifecycle benchmarks are only reported where the tool exposes a reproducible non-interactive path.
- Version/help probes are not a substitute for lifecycle benchmarks, but they give a comparable lower-bound process startup, RSS, and CPU signal across CLI/TUI products.
- GUI-first products such as Crystal, Conductor, CodeAgentSwarm, and Termyx should be compared through workflow and UX case studies unless they expose a stable automation interface.
EOF_REPORT

printf 'Orchestrator probe report written to: %s\n' "$RESULT_FILE"
