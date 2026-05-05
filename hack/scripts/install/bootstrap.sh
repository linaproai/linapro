#!/usr/bin/env bash
# Self-contained installer bootstrap. This file is the exact content deployed as https://linapro.ai/install.sh.

set -Eeuo pipefail

LINAPRO_REPO_URL="https://github.com/linaproai/linapro.git"
LINAPRO_LATEST_URL="https://github.com/linaproai/linapro/releases/latest"

log_info() {
  printf '[linapro] INFO: %s\n' "$*"
}

log_warn() {
  printf '[linapro] WARN: %s\n' "$*" >&2
}

log_error() {
  printf '[linapro] ERROR: %s\n' "$*" >&2
}

die() {
  log_error "$*"
  exit 1
}

on_error() {
  local rc="$?"
  local line="${1:-unknown}"
  log_error "Unexpected installer error at line $line (exit $rc)."
  exit "$rc"
}

on_exit() {
  local rc="$?"
  if [ "$rc" -ne 0 ]; then
    log_error "Installation failed. Rerun with LINAPRO_VERSION=v0.x.y if version discovery failed."
  fi
}

trap 'on_error "$LINENO"' ERR
trap 'on_exit' EXIT

detect_os() {
  local kernel
  kernel="$(uname -s 2>/dev/null || printf 'unknown')"
  case "$kernel" in
    Darwin) printf 'macos\n' ;;
    Linux) printf 'linux\n' ;;
    MINGW*|MSYS*|CYGWIN*) printf 'windows\n' ;;
    *) die "Unsupported OS. LinaPro supports macOS, Linux, and Windows via Git Bash or WSL." ;;
  esac
}

is_stable_version_tag() {
  local tag="$1"
  printf '%s\n' "$tag" | grep -E '^v[0-9]+[.][0-9]+[.][0-9]+$' >/dev/null 2>&1
}

resolve_version() {
  if [ -n "${LINAPRO_VERSION:-}" ]; then
    printf '%s\n' "$LINAPRO_VERSION"
    return 0
  fi

  if ! command -v curl >/dev/null 2>&1; then
    die "curl is required to resolve the latest release. Example: LINAPRO_VERSION=v0.x.y curl -fsSL https://linapro.ai/install.sh | bash"
  fi

  local headers location tag
  headers="$(curl -sIL "$LINAPRO_LATEST_URL" || true)"
  location="$(printf '%s\n' "$headers" | awk 'BEGIN{IGNORECASE=1} /^location:/ {print $2}' | tr -d '\r' | tail -n 1)"
  tag="${location##*/}"
  if [ -z "$tag" ] || [ "$tag" = "$location" ]; then
    die "Could not resolve the latest LinaPro release. Retry with: LINAPRO_VERSION=v0.x.y curl -fsSL https://linapro.ai/install.sh | bash"
  fi
  if ! is_stable_version_tag "$tag"; then
    die "Resolved release target is not a stable version tag: $tag. Retry with: LINAPRO_VERSION=v0.x.y curl -fsSL https://linapro.ai/install.sh | bash"
  fi
  printf '%s\n' "$tag"
}

target_is_non_empty() {
  local target="$1"
  [ -d "$target" ] && [ -n "$(find "$target" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]
}

target_path_depth() {
  local target_abs="$1"
  local trimmed
  trimmed="${target_abs#/}"
  if [ -z "$trimmed" ]; then
    printf '0\n'
    return 0
  fi
  awk -F/ '{print NF}' <<EOF
$trimmed
EOF
}

assert_safe_force_target() {
  local target="$1"
  local target_abs="$2"
  local current_abs="$3"
  local home_abs path_depth

  if [ "$target_abs" = "/" ]; then
    die "Refusing to overwrite unsafe target directory: $target"
  fi

  if [ -n "${HOME:-}" ] && [ -d "$HOME" ]; then
    home_abs="$(cd "$HOME" && pwd -P)"
    if [ "$target_abs" = "$home_abs" ]; then
      die "Refusing to overwrite unsafe target directory: $target"
    fi
  fi

  path_depth="$(target_path_depth "$target_abs")"
  if [ "$path_depth" -lt 3 ]; then
    die "Refusing to overwrite unsafe target directory: $target"
  fi

  case "$current_abs" in
    "$target_abs"|"$target_abs"/*)
      die "Refusing to overwrite unsafe target directory: $target"
      ;;
  esac
}

prepare_target() {
  local target="$1"
  local parent target_abs current_abs
  parent="$(dirname "$target")"
  mkdir -p "$parent"

  if [ -e "$target" ] && [ ! -d "$target" ]; then
    die "Target path exists but is not a directory: $target"
  fi

  if target_is_non_empty "$target"; then
    if [ "${LINAPRO_FORCE:-}" != "1" ]; then
      die "Target directory is not empty: $target. Choose another LINAPRO_DIR or rerun with LINAPRO_FORCE=1."
    fi
    target_abs="$(cd "$target" && pwd -P)"
    current_abs="$(pwd -P)"
    assert_safe_force_target "$target" "$target_abs" "$current_abs"
    rm -rf "$target"
  fi
}

print_banner() {
  local version="$1"
  local target="$2"
  local os_name="$3"
  printf '\n'
  printf 'LinaPro installer\n'
  printf 'Version: %s\n' "$version"
  printf 'Target:  %s\n' "$target"
  printf 'OS:      %s\n' "$os_name"
  printf '\n'
}

print_next_steps() {
  local project_dir="$1"
  printf '\n'
  printf 'LinaPro source downloaded successfully.\n'
  printf 'Project directory: %s\n' "$project_dir"
  printf 'Default admin: admin / admin123\n'
  printf '\n'
  printf 'Next steps:\n'
  printf '  cd "%s"\n' "$project_dir"
  printf '  ask Claude Code "run lina-doctor to set up my LinaPro environment"\n'
  printf '  make init && make dev\n'
}

main() {
  local os_name version target target_abs
  local clone_args

  os_name="$(detect_os)"
  version="$(resolve_version)"
  target="${LINAPRO_DIR:-./linapro}"
  print_banner "$version" "$target" "$os_name"

  if ! command -v git >/dev/null 2>&1; then
    die "git is required before running this installer."
  fi

  prepare_target "$target"

  clone_args=(--branch "$version")
  if [ "${LINAPRO_SHALLOW:-}" = "1" ]; then
    clone_args+=(--depth 1)
    log_warn "Shallow clone enabled. The lina-upgrade skill will require git fetch --unshallow before the first upgrade."
  fi

  log_info "Cloning $LINAPRO_REPO_URL at $version."
  if ! git clone "${clone_args[@]}" "$LINAPRO_REPO_URL" "$target"; then
    die "git clone failed. Check network access and verify that tag '$version' exists."
  fi

  target_abs="$(cd "$target" && pwd -P)"
  print_next_steps "$target_abs"
}

main "$@"
