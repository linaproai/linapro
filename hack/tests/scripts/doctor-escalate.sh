#!/usr/bin/env bash
# Unit tests for lina-doctor escalation root-cause inference.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
DOCTOR_ESCALATE="$REPO_ROOT/.claude/skills/lina-doctor/lib/doctor-escalate.sh"

fail() {
  printf 'FAIL doctor-escalate: %s\n' "$*" >&2
  exit 1
}

run_case() {
  local name="$1"
  local text="$2"
  local expected="$3"
  local temp_dir log_file output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  log_file="$temp_dir/$name.log"
  printf '%s\n' "$text" >"$log_file"
  output="$(bash "$DOCTOR_ESCALATE" "$name" "fake command" "test" "$log_file")"
  printf '%s\n' "$output" | grep -F "Root cause: $expected" >/dev/null ||
    fail "$name expected $expected: $output"
  printf 'PASS doctor-escalate %s\n' "$name"
}

run_case permission 'EACCES: permission denied' permission
run_case network 'timeout while connecting to proxy.golang.org' network
run_case package 'Unable to locate package golang-go' package_not_found
