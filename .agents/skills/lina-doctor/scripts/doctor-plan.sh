#!/usr/bin/env bash
# Builds a topologically ordered Lina Doctor install plan from doctor-check JSON.

set -euo pipefail

IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$SKILL_DIR/lib/_common.sh"

repo_root="$(find_repo_root)"
cd "$repo_root"

input_file=""
if [ "${1:-}" = "--input" ]; then
  input_file="${2:-}"
  [ -n "$input_file" ] || die "--input requires a file path"
fi
if [ -n "$input_file" ]; then
  check_json="$(cat "$input_file")"
else
  check_json="$(cat)"
fi

os_name="$(json_get_string "$check_json" os)"
package_manager="$(json_get_string "$check_json" package_manager)"
repo_detected="$(json_get_bool "$check_json" repo_root_detected)"
node_version_manager="$(bash "$SCRIPT_DIR/doctor-detect.sh" | sed -n 's/.*"node_version_manager"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"

extract_mirror_hints() {
  printf '%s\n' "$check_json" | awk '
    /"mirror_hints"[[:space:]]*:/ { flag=1; next }
    flag && /^[[:space:]]*]/ { flag=0; next }
    flag { print }
  '
}

tool_ok() {
  local tool="$1"
  local value
  value="$(json_tool_ok "$check_json" "$tool")"
  [ "$value" = "true" ]
}

pm_command() {
  local tool="$1"
  case "$os_name:$package_manager:$tool" in
    macos:brew:go) printf 'brew install go' ;;
    macos:brew:node) printf 'brew install node' ;;
    macos:brew:git) printf 'brew install git' ;;
    macos:brew:make) printf 'brew install make' ;;
    macos:brew:openspec) printf 'brew install openspec' ;;
    linux:apt-get:go) printf 'sudo apt-get update && sudo apt-get install -y golang-go' ;;
    linux:apt-get:node) printf 'sudo apt-get update && sudo apt-get install -y nodejs npm' ;;
    linux:apt-get:git) printf 'sudo apt-get update && sudo apt-get install -y git' ;;
    linux:apt-get:make) printf 'sudo apt-get update && sudo apt-get install -y make' ;;
    linux:dnf:go) printf 'sudo dnf install -y golang' ;;
    linux:dnf:node) printf 'sudo dnf install -y nodejs npm' ;;
    linux:dnf:git) printf 'sudo dnf install -y git' ;;
    linux:dnf:make) printf 'sudo dnf install -y make' ;;
    linux:yum:go) printf 'sudo yum install -y golang' ;;
    linux:yum:node) printf 'sudo yum install -y nodejs npm' ;;
    linux:yum:git) printf 'sudo yum install -y git' ;;
    linux:yum:make) printf 'sudo yum install -y make' ;;
    linux:pacman:go) printf 'sudo pacman -Sy --needed go' ;;
    linux:pacman:node) printf 'sudo pacman -Sy --needed nodejs npm' ;;
    linux:pacman:git) printf 'sudo pacman -Sy --needed git' ;;
    linux:pacman:make) printf 'sudo pacman -Sy --needed make' ;;
    windows:winget:go) printf 'powershell.exe -NoProfile -Command "winget install GoLang.Go"' ;;
    windows:winget:node) printf 'powershell.exe -NoProfile -Command "winget install OpenJS.NodeJS"' ;;
    windows:winget:git) printf 'powershell.exe -NoProfile -Command "winget install Git.Git"' ;;
    windows:winget:make) printf 'powershell.exe -NoProfile -Command "winget install ezwinports.make"' ;;
    windows:scoop:go) printf 'powershell.exe -NoProfile -Command "scoop install go"' ;;
    windows:scoop:node) printf 'powershell.exe -NoProfile -Command "scoop install nodejs-lts"' ;;
    windows:scoop:git) printf 'powershell.exe -NoProfile -Command "scoop install git"' ;;
    windows:scoop:make) printf 'powershell.exe -NoProfile -Command "scoop install make"' ;;
    windows:choco:go) printf 'powershell.exe -NoProfile -Command "choco install golang -y"' ;;
    windows:choco:node) printf 'powershell.exe -NoProfile -Command "choco install nodejs-lts -y"' ;;
    windows:choco:git) printf 'powershell.exe -NoProfile -Command "choco install git -y"' ;;
    windows:choco:make) printf 'powershell.exe -NoProfile -Command "choco install make -y"' ;;
    *) printf '' ;;
  esac
}

# shellcheck disable=SC2016
node_command() {
  case "$node_version_manager" in
    nvm) printf '[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"; nvm install 20 && nvm use 20' ;;
    fnm) printf 'fnm install 20 && fnm use 20' ;;
    volta) printf 'volta install node@20' ;;
    *) pm_command node ;;
  esac
}

requires_sudo() {
  local command_text="$1"
  case "$command_text" in
    sudo\ *) printf 'true' ;;
    *) printf 'false' ;;
  esac
}

plan_items=""
add_item() {
  local tool="$1"
  local command_text="$2"
  local depends_on="$3"
  local optional="$4"
  [ -n "$command_text" ] || return 0
  if [ -n "$plan_items" ]; then
    plan_items="$plan_items,"
  fi
  plan_items="$plan_items
    { \"tool\": \"$(json_escape "$tool")\", \"command\": \"$(json_escape "$command_text")\", \"package_manager\": \"$(json_escape "$package_manager")\", \"requires_sudo\": $(requires_sudo "$command_text"), \"depends_on\": [$(json_escape "$depends_on" | awk -v q='"' 'NF {split($0,a,","); for (i=1;i<=length(a);i++) {gsub(/^ +| +$/, "", a[i]); printf "%s%s%s%s", (i>1?", ":""), q, a[i], q}}')], \"optional\": $optional }"
}

if ! tool_ok git; then add_item git "$(pm_command git)" "" false; fi
if ! tool_ok make; then add_item make "$(pm_command make)" "" false; fi
if ! tool_ok go; then add_item go "$(pm_command go)" "" false; fi
if ! tool_ok node; then add_item node "$(node_command)" "" false; fi
if ! tool_ok gf; then add_item gf 'go install github.com/gogf/gf/v2/cmd/gf@latest' "go" false; fi
if ! tool_ok pnpm; then add_item pnpm 'npm i -g pnpm' "node" false; fi
if ! tool_ok openspec; then
  if [ "$os_name" = "macos" ] && [ "$package_manager" = "brew" ]; then
    add_item openspec 'brew install openspec' "" false
  else
    add_item openspec 'npm i -g @fission-ai/openspec@latest' "node" false
  fi
fi
if [ "${LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL:-}" != "1" ] && ! tool_ok goframe-v2; then
  add_item goframe-v2 'npx skills add github.com/gogf/skills -g' "node" true
fi
if [ "${LINAPRO_DOCTOR_SKIP_PLAYWRIGHT:-}" != "1" ] && [ "$repo_detected" = "true" ] && ! tool_ok playwright; then
  add_item playwright 'cd hack/tests && pnpm exec playwright install' "pnpm" true
fi

printf '{\n'
mirror_hints="$(extract_mirror_hints)"
printf '  "mirror_hints": ['
if [ -n "$mirror_hints" ]; then
  printf '\n%s\n' "$mirror_hints"
  printf '  ],\n'
else
  printf '],\n'
fi
printf '  "items": ['
if [ -n "$plan_items" ]; then
  printf '%s\n' "$plan_items"
  printf '  ]\n'
else
  printf ']\n'
fi
printf '}\n'
