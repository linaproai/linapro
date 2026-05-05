#!/usr/bin/env bash
# Executes a Lina Doctor install plan with per-step confirmation and verification.

set -euo pipefail

IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$SKILL_DIR/lib/_common.sh"
# shellcheck source=.claude/skills/lina-doctor/lib/doctor-checks.sh
source "$SKILL_DIR/lib/doctor-checks.sh"
# shellcheck source=.claude/skills/lina-doctor/lib/doctor-escalate.sh
source "$SKILL_DIR/lib/doctor-escalate.sh"

repo_root="$(find_repo_root)"
cd "$repo_root"

plan_file=""
if [ "${1:-}" = "--plan" ]; then
  plan_file="${2:-}"
  [ -n "$plan_file" ] || die "--plan requires a file path"
fi
if [ -n "$plan_file" ]; then
  plan_json="$(cat "$plan_file")"
else
  plan_json="$(cat)"
fi

run_command() {
  local tool="$1"
  local command_text="$2"
  local log_file="$3"
  local timeout_duration="${LINAPRO_DOCTOR_TIMEOUT:-300s}"
  local timeout_seconds
  local rc

  timeout_seconds="${timeout_duration%s}"
  case "$timeout_seconds" in
    ''|*[!0-9]*)
      die "LINAPRO_DOCTOR_TIMEOUT must be expressed in seconds, for example 300s"
      ;;
  esac

  rc=0
  if command_exists timeout; then
    timeout "$timeout_duration" bash -lc "$command_text" >"$log_file" 2>&1 || rc=$?
  elif command_exists perl; then
    perl -e 'alarm shift; exec @ARGV' "$timeout_seconds" bash -lc "$command_text" >"$log_file" 2>&1 || rc=$?
  else
    bash -lc "$command_text" >"$log_file" 2>&1 || rc=$?
  fi
  if [ "$rc" -eq 124 ] || [ "$rc" -eq 142 ]; then
    printf '\ntimeout after %ss\n' "$timeout_seconds" >>"$log_file"
  fi
  return "$rc"
}

verify_tool() {
  local tool="$1"
  local check_json
  check_json="$(doctor_check_json "$SKILL_DIR")"
  doctor_tool_ok "$check_json" "$tool"
}

handle_path_after_install() {
  local tool="$1"
  local shell_name
  shell_name="$(detect_shell)"
  case "$tool" in
    gf)
      if [ -x "$HOME/go/bin/gf" ] && ! command_exists gf; then
        export PATH="$HOME/go/bin:$PATH"
        print_path_fix_hint gf "$HOME/go/bin" "$shell_name"
      fi
      ;;
    pnpm|openspec)
      npm_prefix="$(detect_npm_global_prefix)"
      if [ -n "$npm_prefix" ] && [ -d "$npm_prefix/bin" ] && ! is_path_entry_present "$npm_prefix/bin"; then
        export PATH="$npm_prefix/bin:$PATH"
        print_path_fix_hint "$tool" "$npm_prefix/bin" "$shell_name"
      fi
      ;;
  esac
}

extract_field() {
  local line="$1"
  local field="$2"
  printf '%s\n' "$line" | sed -E -n "s/.*\"$field\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p"
}

extract_bool_field() {
  local line="$1"
  local field="$2"
  printf '%s\n' "$line" | sed -E -n "s/.*\"$field\"[[:space:]]*:[[:space:]]*(true|false).*/\1/p"
}

if ! printf '%s\n' "$plan_json" | grep -F '"tool"' >/dev/null 2>&1; then
  printf 'All planned tools are already satisfied.\n'
  bash "$SCRIPT_DIR/doctor-verify.sh"
  exit 0
fi

printf '%s\n' "$plan_json" | grep -F '"tool"' | while IFS= read -r item_line; do
  tool="$(extract_field "$item_line" tool)"
  command_text="$(extract_field "$item_line" command)"
  package_manager="$(extract_field "$item_line" package_manager)"
  optional="$(extract_bool_field "$item_line" optional)"
  log_file="/tmp/lina-doctor-${tool}.log"

  printf 'Tool: %s\n' "$tool"
  printf 'Package manager: %s\n' "$package_manager"
  printf 'Command: %s\n' "$command_text"
  if ! confirm "Run install step for $tool?"; then
    printf 'Skipped %s by user choice.\n' "$tool"
    continue
  fi

  set +e
  run_command "$tool" "$command_text" "$log_file"
  rc=$?
  set -e

  if [ "$rc" -ne 0 ] && [ "$tool" = "openspec" ] && [ "$package_manager" = "brew" ]; then
    fallback='npm i -g @fission-ai/openspec@latest'
    printf 'brew install openspec failed; retrying fallback: %s\n' "$fallback"
    set +e
    run_command "$tool" "$fallback" "$log_file"
    rc=$?
    command_text="$fallback"
    set -e
  fi

  handle_path_after_install "$tool"

  if [ "$rc" -ne 0 ] || ! verify_tool "$tool"; then
    emit_escalation "$tool" "$command_text" "$package_manager" "$log_file"
    if [ "$optional" = "true" ]; then
      printf 'Optional tool %s failed; continuing.\n' "$tool"
      continue
    fi
    exit 1
  fi
  printf 'Installed and verified %s.\n' "$tool"
done

bash "$SCRIPT_DIR/doctor-verify.sh"
