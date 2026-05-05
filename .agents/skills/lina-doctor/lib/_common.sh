#!/usr/bin/env bash
# Shared shell helpers for Lina Doctor scripts.

set -euo pipefail

IFS=$'\n\t'

if [ "${LINAPRO_DOCTOR_DEBUG:-}" = "1" ]; then
  set -x
fi

# _color returns an ANSI color code when stdout is connected to a terminal.
_color() {
  local code="$1"
  if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
    tput setaf "$code" 2>/dev/null || true
  fi
}

# _reset returns an ANSI reset code when stdout is connected to a terminal.
_reset() {
  if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
    tput sgr0 2>/dev/null || true
  fi
}

# log_info prints a normal progress message.
log_info() {
  local blue reset
  blue="$(_color 6)"
  reset="$(_reset)"
  printf '%s[lina-doctor] INFO:%s %s\n' "$blue" "$reset" "$*"
}

# log_warn prints a warning message that does not stop execution.
log_warn() {
  local yellow reset
  yellow="$(_color 3)"
  reset="$(_reset)"
  printf '%s[lina-doctor] WARN:%s %s\n' "$yellow" "$reset" "$*" >&2
}

# log_error prints an error message to stderr.
log_error() {
  local red reset
  red="$(_color 1)"
  reset="$(_reset)"
  printf '%s[lina-doctor] ERROR:%s %s\n' "$red" "$reset" "$*" >&2
}

# log_debug prints a debug message when LINAPRO_DOCTOR_DEBUG=1 is set.
log_debug() {
  if [ "${LINAPRO_DOCTOR_DEBUG:-}" = "1" ]; then
    printf '[lina-doctor] DEBUG: %s\n' "$*" >&2
  fi
}

# die prints an error message and exits the current process.
die() {
  log_error "$*"
  exit 1
}

# _semver_core removes a leading v and strips prerelease/build suffixes.
_semver_core() {
  local value="$1"
  value="${value#v}"
  value="${value#V}"
  value="${value%%-*}"
  value="${value%%+*}"
  printf '%s\n' "$value"
}

# version_ge returns success when semantic version a is greater than or equal to b.
version_ge() {
  local a b a_major a_minor a_patch b_major b_minor b_patch
  a="$(_semver_core "$1")"
  b="$(_semver_core "$2")"

  IFS=. read -r a_major a_minor a_patch <<EOF
$a
EOF
  IFS=. read -r b_major b_minor b_patch <<EOF
$b
EOF

  a_major="${a_major:-0}"
  a_minor="${a_minor:-0}"
  a_patch="${a_patch:-0}"
  b_major="${b_major:-0}"
  b_minor="${b_minor:-0}"
  b_patch="${b_patch:-0}"

  if [ "$a_major" -gt "$b_major" ]; then return 0; fi
  if [ "$a_major" -lt "$b_major" ]; then return 1; fi
  if [ "$a_minor" -gt "$b_minor" ]; then return 0; fi
  if [ "$a_minor" -lt "$b_minor" ]; then return 1; fi
  [ "$a_patch" -ge "$b_patch" ]
}

# retry runs a command with fixed-delay retries for transient failures.
retry() {
  local times="$1"
  local delay="$2"
  shift 2
  if [ "${1:-}" = "--" ]; then
    shift
  fi

  local attempt
  attempt=1
  while true; do
    if "$@"; then
      return 0
    fi
    if [ "$attempt" -ge "$times" ]; then
      return 1
    fi
    log_warn "Command failed; retrying in ${delay}s (${attempt}/${times}): $*"
    sleep "$delay"
    attempt=$((attempt + 1))
  done
}

# confirm asks for interactive y/N confirmation unless non-interactive mode is set.
confirm() {
  local message="$1"
  local answer
  if [ "${LINAPRO_DOCTOR_NON_INTERACTIVE:-}" = "1" ]; then
    return 0
  fi
  printf '%s [y/N] ' "$message"
  if ! read -r answer; then
    answer=""
  fi
  case "$answer" in
    y|Y|yes|YES) return 0 ;;
    *) return 1 ;;
  esac
}

# detect_os returns macos, linux, windows, or unsupported.
detect_os() {
  local kernel
  kernel="$(uname -s 2>/dev/null || printf 'unknown')"
  case "$kernel" in
    Darwin) printf 'macos\n' ;;
    Linux) printf 'linux\n' ;;
    MINGW*|MSYS*|CYGWIN*) printf 'windows\n' ;;
    *) printf 'unsupported\n' ;;
  esac
}

# detect_shell returns the active shell family.
detect_shell() {
  if [ -n "${BASH_VERSION:-}" ]; then
    printf 'bash\n'
    return 0
  fi
  if [ -n "${ZSH_VERSION:-}" ]; then
    printf 'zsh\n'
    return 0
  fi
  case "${SHELL:-}" in
    */zsh) printf 'zsh\n' ;;
    */bash) printf 'bash\n' ;;
    */fish) printf 'fish\n' ;;
    *) 
      if [ -n "${PSModulePath:-}" ]; then
        printf 'powershell\n'
      else
        printf 'unknown\n'
      fi
      ;;
  esac
}

# command_exists returns success when a command is available in PATH.
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# is_path_entry_present returns success when a directory is present in PATH.
is_path_entry_present() {
  local expected="$1"
  case ":${PATH:-}:" in
    *":$expected:"*) return 0 ;;
    *) return 1 ;;
  esac
}

# detect_package_manager returns the preferred package manager for the current platform.
detect_package_manager() {
  local os_name="${1:-}"
  if [ -z "$os_name" ]; then
    os_name="$(detect_os)"
  fi
  case "$os_name" in
    macos)
      if command_exists brew; then printf 'brew\n'; else printf 'none\n'; fi
      ;;
    linux)
      if command_exists apt-get; then printf 'apt-get\n'; return 0; fi
      if command_exists dnf; then printf 'dnf\n'; return 0; fi
      if command_exists yum; then printf 'yum\n'; return 0; fi
      if command_exists pacman; then printf 'pacman\n'; return 0; fi
      printf 'none\n'
      ;;
    windows)
      if command_exists winget; then printf 'winget\n'; return 0; fi
      if command_exists scoop; then printf 'scoop\n'; return 0; fi
      if command_exists choco; then printf 'choco\n'; return 0; fi
      printf 'none\n'
      ;;
    *)
      printf 'none\n'
      ;;
  esac
}

# detect_node_version_manager returns the first detected Node version manager.
detect_node_version_manager() {
  if [ -n "${NVM_DIR:-}" ] && { command_exists nvm || [ -s "$NVM_DIR/nvm.sh" ]; }; then
    printf 'nvm\n'
    return 0
  fi
  if [ -n "${FNM_DIR:-}" ] || command_exists fnm; then
    printf 'fnm\n'
    return 0
  fi
  if [ -n "${VOLTA_HOME:-}" ] || command_exists volta; then
    printf 'volta\n'
    return 0
  fi
  printf 'none\n'
}

# detect_npm_global_prefix returns npm's global prefix when npm is available.
detect_npm_global_prefix() {
  if command_exists npm; then
    npm config get prefix 2>/dev/null | awk 'NR == 1 {print $1}'
    return 0
  fi
  printf '\n'
}

# find_repo_root returns the current git repository root when available.
find_repo_root() {
  git rev-parse --show-toplevel 2>/dev/null || pwd -P
}

# repo_root_detected returns success when the path looks like a LinaPro repo root.
repo_root_detected() {
  local root="$1"
  [ -d "$root/apps/lina-core" ] && [ -d "$root/apps/lina-vben" ] && [ -d "$root/hack/tests" ]
}

# json_escape escapes a string for JSON output.
json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/\\n}"
  value="${value//$'\r'/\\r}"
  value="${value//$'\t'/\\t}"
  printf '%s' "$value"
}

# json_string prints a JSON string value or null for an empty value.
json_string() {
  local value="${1:-}"
  if [ -z "$value" ]; then
    printf 'null'
  else
    printf '"%s"' "$(json_escape "$value")"
  fi
}

# json_get_string extracts a simple top-level JSON string by key.
json_get_string() {
  local json="$1"
  local key="$2"
  printf '%s\n' "$json" | sed -E -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p" | head -n 1
}

# json_get_bool extracts a simple top-level JSON boolean by key.
json_get_bool() {
  local json="$1"
  local key="$2"
  printf '%s\n' "$json" | sed -E -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*(true|false).*/\1/p" | head -n 1
}

# json_tool_ok extracts the ok boolean for a tool from doctor-check JSON.
json_tool_ok() {
  local json="$1"
  local tool="$2"
  printf '%s\n' "$json" |
    sed -E -n "s/.*\"$tool\"[[:space:]]*:[[:space:]]*\\{[^}]*\"ok\"[[:space:]]*:[[:space:]]*(true|false).*/\1/p" |
    head -n 1
}

# print_path_fix_hint prints current-session and persistent PATH repair hints.
print_path_fix_hint() {
  local tool="$1"
  local bin_dir="$2"
  local shell_name="$3"
  printf 'PATH warning: %s was installed under %s but is not in PATH.\n' "$tool" "$bin_dir"
  printf "Current shell: export PATH=\"%s:\$PATH\"\n" "$bin_dir"
  case "$shell_name" in
    zsh)
      printf "Persistent: echo 'export PATH=\"%s:\$PATH\"' >> ~/.zshrc\n" "$bin_dir"
      ;;
    bash)
      printf "Persistent: echo 'export PATH=\"%s:\$PATH\"' >> ~/.bashrc\n" "$bin_dir"
      ;;
    fish)
      printf 'Persistent: fish_add_path "%s"\n' "$bin_dir"
      ;;
    powershell)
      printf "Persistent: add \"%s\" to \$PROFILE or the user PATH.\n" "$bin_dir"
      ;;
    *)
      printf 'Persistent: add %s to your shell profile PATH.\n' "$bin_dir"
      ;;
  esac
}
