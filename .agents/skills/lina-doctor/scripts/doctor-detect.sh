#!/usr/bin/env bash
# Detects host metadata for Lina Doctor without modifying the environment.

set -euo pipefail

IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$SKILL_DIR/lib/_common.sh"

repo_root="$(find_repo_root)"
cd "$repo_root"

os_name="$(detect_os)"
package_manager="$(detect_package_manager "$os_name")"
shell_name="$(detect_shell)"
node_version_manager="$(detect_node_version_manager)"
npm_prefix="$(detect_npm_global_prefix)"
goproxy="${GOPROXY:-}"
playwright_host="${PLAYWRIGHT_DOWNLOAD_HOST:-}"
npm_registry=""
if command_exists npm; then
  npm_registry="$(npm config get registry 2>/dev/null | awk 'NR == 1 {print $1}')"
fi
if repo_root_detected "$repo_root"; then
  repo_detected=true
else
  repo_detected=false
fi

printf '{\n'
printf '  "os": %s,\n' "$(json_string "$os_name")"
printf '  "package_manager": %s,\n' "$(json_string "$package_manager")"
printf '  "shell": %s,\n' "$(json_string "$shell_name")"
printf '  "node_version_manager": %s,\n' "$(json_string "$node_version_manager")"
printf '  "npm_global_prefix": %s,\n' "$(json_string "$npm_prefix")"
printf '  "repo_root": %s,\n' "$(json_string "$repo_root")"
printf '  "repo_root_detected": %s,\n' "$repo_detected"
printf '  "mirrors": {\n'
printf '    "goproxy": %s,\n' "$(json_string "$goproxy")"
printf '    "npm_registry": %s,\n' "$(json_string "$npm_registry")"
printf '    "playwright_download_host": %s\n' "$(json_string "$playwright_host")"
printf '  }\n'
printf '}\n'
