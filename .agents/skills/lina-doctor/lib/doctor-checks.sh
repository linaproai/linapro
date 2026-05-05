#!/usr/bin/env bash
# Shared verification helpers backed by doctor-check.sh JSON output.

set -euo pipefail

IFS=$'\n\t'

CHECKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$CHECKS_DIR/_common.sh"

# doctor_check_json runs doctor-check.sh and returns its JSON even for warning exits.
doctor_check_json() {
  local skill_dir="$1"
  local check_json
  set +e
  check_json="$(bash "$skill_dir/scripts/doctor-check.sh" 2>/dev/null)"
  set -e
  printf '%s\n' "$check_json"
}

# doctor_tool_ok returns success when the named tool is ok in a check JSON document.
doctor_tool_ok() {
  local check_json="$1"
  local tool="$2"
  [ "$(json_tool_ok "$check_json" "$tool")" = "true" ]
}
