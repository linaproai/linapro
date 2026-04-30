#!/usr/bin/env bash
# Shared installer helpers used by macOS, Linux, and Windows Git Bash entrypoints.

set -euo pipefail

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
  printf '%s[linapro] INFO:%s %s\n' "$blue" "$reset" "$*"
}

# log_warn prints a warning message that does not stop installation.
log_warn() {
  local yellow reset
  yellow="$(_color 3)"
  reset="$(_reset)"
  printf '%s[linapro] WARN:%s %s\n' "$yellow" "$reset" "$*" >&2
}

# log_error prints an error message to stderr.
log_error() {
  local red reset
  red="$(_color 1)"
  reset="$(_reset)"
  printf '%s[linapro] ERROR:%s %s\n' "$red" "$reset" "$*" >&2
}

# log_debug prints a debug message when LINAPRO_DEBUG=1 is set.
log_debug() {
  if [ "${LINAPRO_DEBUG:-}" = "1" ]; then
    printf '[linapro] DEBUG: %s\n' "$*" >&2
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

# require_command verifies that a command exists and prints the install hint on failure.
require_command() {
  local command_name="$1"
  local install_hint="$2"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    die "Missing required command '$command_name'. Install hint: $install_hint"
  fi
}

# retry runs a command with fixed-delay retries for transient network failures.
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

# confirm asks for interactive y/N confirmation unless LINAPRO_NON_INTERACTIVE=1.
confirm() {
  local message="$1"
  local answer
  if [ "${LINAPRO_NON_INTERACTIVE:-}" = "1" ]; then
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

# is_port_in_use returns success when a TCP port appears to have a listener.
is_port_in_use() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -iTCP:"$port" -sTCP:LISTEN -n -P >/dev/null 2>&1
    return $?
  fi
  if command -v netstat >/dev/null 2>&1; then
    netstat -an 2>/dev/null | grep -E "[.:]${port}[[:space:]].*LISTEN" >/dev/null 2>&1
    return $?
  fi
  return 1
}

# run_prereq_check runs the shared prerequisite probe and interprets its exit code.
run_prereq_check() {
  local prereq_script="$1"
  local rc
  set +e
  bash "$prereq_script"
  rc=$?
  set -e

  case "$rc" in
    0) return 0 ;;
    2)
      log_warn "Prerequisite check completed with warnings; installation will continue."
      return 0
      ;;
    *)
      die "Prerequisite check failed. Install missing critical tools, then rerun this script."
      ;;
  esac
}

# copy_default_config creates config.yaml from the template only when missing.
copy_default_config() {
  local repo_root="$1"
  local config_file template_file
  config_file="$repo_root/apps/lina-core/manifest/config/config.yaml"
  template_file="$repo_root/apps/lina-core/manifest/config/config.template.yaml"

  if [ -f "$config_file" ]; then
    log_info "Config already exists: $config_file"
    return 0
  fi
  if [ ! -f "$template_file" ]; then
    die "Missing config template: $template_file"
  fi
  cp "$template_file" "$config_file"
  log_info "Created config from template: $config_file"
}

# run_core_fallback executes host init/mock commands when make is unavailable on Git Bash.
run_core_fallback() {
  local repo_root="$1"
  local target="$2"
  case "$target" in
    init)
      (cd "$repo_root/apps/lina-core" && go run main.go init --confirm=init --sql-source=local)
      ;;
    mock)
      (cd "$repo_root/apps/lina-core" && go run main.go mock --confirm=mock --sql-source=local)
      ;;
    *)
      die "Unsupported fallback target: $target"
      ;;
  esac
}

# download_go_modules installs backend module dependencies.
download_go_modules() {
  local repo_root="$1"
  cd "$repo_root/apps/lina-core" && go mod download
}

# install_frontend_deps installs frontend workspace dependencies.
install_frontend_deps() {
  local repo_root="$1"
  cd "$repo_root/apps/lina-vben" && pnpm install
}

# run_make_or_fallback executes a root make target or the Windows Git Bash fallback.
run_make_or_fallback() {
  local repo_root="$1"
  local target="$2"
  local allow_fallback="$3"

  if command -v make >/dev/null 2>&1; then
    (cd "$repo_root" && make "$target" confirm="$target")
    return 0
  fi
  if [ "$allow_fallback" = "1" ]; then
    log_warn "make is unavailable; using GoFrame command fallback for '$target'."
    run_core_fallback "$repo_root" "$target"
    return 0
  fi
  die "make is required for '$target'."
}

# run_standard_install performs the shared post-clone setup flow.
run_standard_install() {
  local repo_root="$1"
  local platform_name="$2"
  local allow_make_fallback="${3:-0}"

  log_info "Running LinaPro installer for $platform_name."
  run_prereq_check "$repo_root/hack/scripts/install/checks/prereq.sh"

  log_info "Downloading backend Go modules."
  retry 2 3 -- download_go_modules "$repo_root"

  log_info "Installing frontend dependencies."
  retry 2 3 -- install_frontend_deps "$repo_root"

  copy_default_config "$repo_root"

  log_info "Initializing database schema and seed data."
  run_make_or_fallback "$repo_root" init "$allow_make_fallback"

  if [ -n "${LINAPRO_SKIP_MOCK:-}" ]; then
    log_info "Skipping mock data because LINAPRO_SKIP_MOCK is set."
  else
    log_info "Loading mock data."
    run_make_or_fallback "$repo_root" mock "$allow_make_fallback"
  fi

  for port in 5666 8080; do
    if is_port_in_use "$port"; then
      log_warn "Port $port is already in use. Stop the conflicting process before running make dev if needed."
    fi
  done

  printf '\n'
  log_info "LinaPro installation completed."
  printf 'Project directory: %s\n' "$repo_root"
  printf 'Default admin: admin / admin123\n'
  printf 'Next command: make dev\n'
}
