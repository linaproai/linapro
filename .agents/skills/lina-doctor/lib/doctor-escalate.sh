#!/usr/bin/env bash
# Escalation helpers for Lina Doctor install failures.

set -euo pipefail

IFS=$'\n\t'

ESCALATE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$ESCALATE_DIR/_common.sh"

# infer_root_cause classifies a command log into a stable failure category.
infer_root_cause() {
  local log_file="$1"
  if [ ! -f "$log_file" ]; then
    printf 'unknown\n'
    return 0
  fi
  if grep -Eiq 'timeout|Could not resolve host|connection refused|proxy|i/o timeout' "$log_file"; then
    printf 'network\n'
    return 0
  fi
  if grep -Eiq 'EACCES|permission denied|operation not permitted|must be run as root' "$log_file"; then
    printf 'permission\n'
    return 0
  fi
  if grep -Eiq 'No such package|Unable to locate package|formula was not found|not in registry' "$log_file"; then
    printf 'package_not_found\n'
    return 0
  fi
  if grep -Eiq 'already installed|conflicts with|node command in PATH points to' "$log_file"; then
    printf 'shim_conflict\n'
    return 0
  fi
  printf 'unknown\n'
}

# recommended_action returns a concise manual action for the platform and root cause.
recommended_action() {
  local tool="$1"
  local root_cause="$2"
  case "$root_cause" in
    network)
      case "$tool" in
        gf) printf 'Set GOPROXY=https://goproxy.cn,direct and retry lina-doctor.\n' ;;
        playwright) printf 'Set PLAYWRIGHT_DOWNLOAD_HOST=https://npmmirror.com/mirrors/playwright/ and retry the optional step.\n' ;;
        *) printf 'Check network, proxy, DNS, and package registry connectivity, then retry.\n' ;;
      esac
      ;;
    permission)
      printf 'Use a user-writable install prefix or rerun the displayed command with the required elevated privileges.\n'
      ;;
    package_not_found)
      printf 'Confirm the package manager channel still provides the package, then use the manual install command from references/tool-matrix.md.\n'
      ;;
    shim_conflict)
      printf 'Inspect PATH and version-manager shims, then ensure the expected binary is resolved before retrying.\n'
      ;;
    *)
      printf 'Inspect the log file and rerun the failed command manually to collect the native package-manager error.\n'
      ;;
  esac
}

# emit_escalation prints a human-readable report and writes a JSON copy to /tmp.
emit_escalation() {
  local tool="$1"
  local command_text="$2"
  local package_manager="$3"
  local log_file="$4"
  local root_cause
  local action
  root_cause="$(infer_root_cause "$log_file")"
  action="$(recommended_action "$tool" "$root_cause")"

  printf 'ERROR: failed to install %s\n' "$tool"
  printf 'Command: %s\n' "$command_text"
  printf 'Package manager: %s\n' "$package_manager"
  printf 'Root cause: %s\n' "$root_cause"
  printf 'Last 50 lines of output:\n'
  if [ -f "$log_file" ]; then
    tail -n 50 "$log_file"
  else
    printf '(log file missing: %s)\n' "$log_file"
  fi
  printf 'Recommended action: %s\n' "$action"

  {
    printf '{\n'
    printf '  "tool": %s,\n' "$(json_string "$tool")"
    printf '  "command": %s,\n' "$(json_string "$command_text")"
    printf '  "package_manager": %s,\n' "$(json_string "$package_manager")"
    printf '  "root_cause": %s,\n' "$(json_string "$root_cause")"
    printf '  "log_file": %s,\n' "$(json_string "$log_file")"
    printf '  "recommended_action": %s\n' "$(json_string "$action")"
    printf '}\n'
  } >/tmp/lina-doctor-escalation.json
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  if [ "$#" -lt 4 ]; then
    die "Usage: doctor-escalate.sh <tool> <command> <package-manager> <log-file>"
  fi
  emit_escalation "$1" "$2" "$3" "$4"
fi
