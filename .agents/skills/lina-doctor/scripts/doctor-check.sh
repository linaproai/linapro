#!/usr/bin/env bash
# Checks LinaPro development tool availability and emits a single JSON document.

set -euo pipefail

IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$SKILL_DIR/lib/_common.sh"

repo_root="$(find_repo_root)"
cd "$repo_root"

if [ "${1:-}" = "--check-only" ]; then
  shift
fi
if [ "$#" -gt 0 ]; then
  die "Unsupported doctor-check argument: $1"
fi

detect_json="$(bash "$SCRIPT_DIR/doctor-detect.sh")"
os_name="$(json_get_string "$detect_json" os)"
package_manager="$(json_get_string "$detect_json" package_manager)"
shell_name="$(json_get_string "$detect_json" shell)"
repo_detected="$(json_get_bool "$detect_json" repo_root_detected)"
npm_prefix="$(json_get_string "$detect_json" npm_global_prefix)"

critical_failures=0
optional_warnings=0
path_warning_count=0
mirror_warning_count=0

bool() {
  if [ "$1" = "1" ]; then printf 'true'; else printf 'false'; fi
}

version_from_command() {
  local command_name="$1"
  case "$command_name" in
    go) go version 2>/dev/null | awk '{print $3}' | sed 's/^go//' ;;
    node) node --version 2>/dev/null | sed 's/^v//' ;;
    pnpm) pnpm --version 2>/dev/null | awk 'NR == 1 {print $1}' ;;
    git) git --version 2>/dev/null | awk '{print $3}' ;;
    make) make --version 2>/dev/null | awk 'NR == 1 {print $NF}' ;;
    openspec) openspec --version 2>/dev/null | awk 'NR == 1 {print $NF}' ;;
    gf) gf version 2>/dev/null | awk 'NR == 1 {print $NF}' ;;
    playwright) (cd "$repo_root/hack/tests" 2>/dev/null && pnpm exec playwright --version 2>/dev/null | awk 'NR == 1 {print $NF}') ;;
    *) printf '' ;;
  esac
}

check_versioned_tool() {
  local command_name="$1"
  local minimum="$2"
  local present=0
  local ok=0
  local version=""
  if command_exists "$command_name"; then
    present=1
    version="$(version_from_command "$command_name")"
    if [ -n "$version" ] && version_ge "$version" "$minimum"; then
      ok=1
    fi
  fi
  if [ "$ok" = "0" ]; then
    critical_failures=$((critical_failures + 1))
  fi
  printf '    "%s": { "present": %s, "version": %s, "min_version": %s, "ok": %s }' \
    "$command_name" "$(bool "$present")" "$(json_string "$version")" "$(json_string "$minimum")" "$(bool "$ok")"
}

check_presence_tool() {
  local command_name="$1"
  local critical="$2"
  local present=0
  local version=""
  if command_exists "$command_name"; then
    present=1
    version="$(version_from_command "$command_name")"
  fi
  if [ "$present" = "0" ] && [ "$critical" = "1" ]; then
    critical_failures=$((critical_failures + 1))
  fi
  printf '    "%s": { "present": %s, "version": %s, "min_version": null, "ok": %s }' \
    "$command_name" "$(bool "$present")" "$(json_string "$version")" "$(bool "$present")"
}

check_playwright() {
  local present=0
  local version=""
  local browser_present=0
  local cache_dir
  if [ "$repo_detected" = "true" ] && command_exists pnpm && [ -d "$repo_root/hack/tests" ]; then
    version="$(version_from_command playwright)"
    for cache_dir in "$HOME/Library/Caches/ms-playwright" "$HOME/.cache/ms-playwright" "$repo_root/hack/tests/node_modules/.cache/ms-playwright"; do
      if [ -d "$cache_dir" ] && find "$cache_dir" -maxdepth 2 -type d -name '*chromium*' -print -quit 2>/dev/null | grep . >/dev/null 2>&1; then
        browser_present=1
        break
      fi
    done
    if [ -n "$version" ] && [ "$browser_present" = "1" ]; then
      present=1
    fi
  fi
  if [ "$present" = "0" ] && [ "${LINAPRO_DOCTOR_SKIP_PLAYWRIGHT:-}" != "1" ]; then
    optional_warnings=$((optional_warnings + 1))
  fi
  printf '    "playwright": { "present": %s, "version": %s, "min_version": null, "ok": %s }' \
    "$(bool "$present")" "$(json_string "$version")" "$(bool "$present")"
}

check_goframe_skill() {
  local present=0
  if [ -f "$HOME/.claude/skills/goframe-v2/SKILL.md" ]; then
    present=1
  fi
  if [ "$present" = "0" ] && [ "${LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL:-}" != "1" ]; then
    optional_warnings=$((optional_warnings + 1))
  fi
  printf '    "goframe-v2": { "present": %s, "version": null, "min_version": null, "ok": %s }' \
    "$(bool "$present")" "$(bool "$present")"
}

path_issue_json=""
add_path_issue() {
  local tool="$1"
  local expected="$2"
  if [ -z "$expected" ]; then
    return 0
  fi
  if ! is_path_entry_present "$expected"; then
    if [ -n "$path_issue_json" ]; then
      path_issue_json="$path_issue_json,"
    fi
    path_issue_json="$path_issue_json
    { \"tool\": \"$(json_escape "$tool")\", \"expected_in\": \"$(json_escape "$expected")\", \"in_path\": false }"
    path_warning_count=$((path_warning_count + 1))
  fi
}

if [ -x "$HOME/go/bin/gf" ]; then
  add_path_issue gf "$HOME/go/bin"
fi
if [ -n "$npm_prefix" ]; then
  add_path_issue pnpm "$npm_prefix/bin"
  add_path_issue openspec "$npm_prefix/bin"
fi

mirror_hint_json=""
add_mirror_hint() {
  local var_name="$1"
  local current="$2"
  local suggested="$3"
  if [ -n "$mirror_hint_json" ]; then
    mirror_hint_json="$mirror_hint_json,"
  fi
  if [ -z "$current" ]; then
    current="(unset)"
  fi
  mirror_hint_json="$mirror_hint_json
    { \"var\": \"$(json_escape "$var_name")\", \"current\": \"$(json_escape "$current")\", \"suggested\": \"$(json_escape "$suggested")\" }"
  mirror_warning_count=$((mirror_warning_count + 1))
}

goproxy="${GOPROXY:-}"
if [ -z "$goproxy" ] || [ "$goproxy" = "off" ]; then
  add_mirror_hint GOPROXY "$goproxy" "https://goproxy.cn,direct"
fi
npm_registry=""
if command_exists npm; then
  npm_registry="$(npm config get registry 2>/dev/null | awk 'NR == 1 {print $1}')"
  if [ -z "$npm_registry" ] || [ "$npm_registry" = "https://registry.npmjs.org/" ]; then
    add_mirror_hint npm_registry "$npm_registry" "https://registry.npmmirror.com"
  fi
fi
if [ -z "${PLAYWRIGHT_DOWNLOAD_HOST:-}" ]; then
  add_mirror_hint PLAYWRIGHT_DOWNLOAD_HOST "" "https://npmmirror.com/mirrors/playwright/"
fi

printf '{\n'
printf '  "os": %s,\n' "$(json_string "$os_name")"
printf '  "package_manager": %s,\n' "$(json_string "$package_manager")"
printf '  "shell": %s,\n' "$(json_string "$shell_name")"
printf '  "repo_root_detected": %s,\n' "$repo_detected"
printf '  "tools": {\n'
check_versioned_tool go 1.22.0; printf ',\n'
check_versioned_tool node 20.19.0; printf ',\n'
check_versioned_tool pnpm 8.0.0; printf ',\n'
check_presence_tool git 1; printf ',\n'
check_presence_tool make 1; printf ',\n'
check_presence_tool openspec 1; printf ',\n'
check_presence_tool gf 1; printf ',\n'
check_playwright; printf ',\n'
check_goframe_skill; printf '\n'
printf '  },\n'
printf '  "path_issues": ['
if [ -n "$path_issue_json" ]; then
  printf '%s\n' "$path_issue_json"
  printf '  ],\n'
else
  printf '],\n'
fi
printf '  "mirror_hints": ['
if [ -n "$mirror_hint_json" ]; then
  printf '%s\n' "$mirror_hint_json"
  printf '  ]\n'
else
  printf ']\n'
fi
printf '}\n'

if [ "$critical_failures" -gt 0 ]; then
  exit 1
fi
if [ "$path_warning_count" -gt 0 ] || [ "$mirror_warning_count" -gt 0 ] || [ "$optional_warnings" -gt 0 ]; then
  exit 2
fi
exit 0
