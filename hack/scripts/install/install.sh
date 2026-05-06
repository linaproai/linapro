#!/usr/bin/env bash
# Self-contained installer. This file is the exact content deployed as https://linapro.ai/install.sh.

set -Eeuo pipefail

LINAPRO_REPO_URL="https://github.com/linaproai/linapro.git"

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

select_latest_version_tag() {
  awk '
    function version_gt(left, right, leftParts, rightParts) {
      split(substr(left, 2), leftParts, ".")
      split(substr(right, 2), rightParts, ".")
      if ((leftParts[1] + 0) != (rightParts[1] + 0)) {
        return (leftParts[1] + 0) > (rightParts[1] + 0)
      }
      if ((leftParts[2] + 0) != (rightParts[2] + 0)) {
        return (leftParts[2] + 0) > (rightParts[2] + 0)
      }
      return (leftParts[3] + 0) > (rightParts[3] + 0)
    }
    {
      tag = $2
      sub("^refs/tags/", "", tag)
      if (tag ~ /^v[0-9]+[.][0-9]+[.][0-9]+$/) {
        if (!found || version_gt(tag, latest)) {
          latest = tag
          found = 1
        }
      }
    }
    END {
      if (!found) {
        exit 1
      }
      print latest
    }
  '
}

resolve_version() {
  if [ -n "${LINAPRO_VERSION:-}" ]; then
    if ! is_stable_version_tag "$LINAPRO_VERSION"; then
      die "LINAPRO_VERSION must be a stable version tag like v0.x.y: $LINAPRO_VERSION"
    fi
    printf '%s\n' "$LINAPRO_VERSION"
    return 0
  fi

  local refs tag
  if ! refs="$(git ls-remote --tags --refs "$LINAPRO_REPO_URL" 'v*' 2>/dev/null)"; then
    die "Could not list LinaPro release tags with git. Retry with: LINAPRO_VERSION=v0.x.y curl -fsSL https://linapro.ai/install.sh | bash"
  fi
  tag="$(printf '%s\n' "$refs" | select_latest_version_tag || true)"
  if [ -z "$tag" ]; then
    die "Could not resolve the latest stable LinaPro release tag. Retry with: LINAPRO_VERSION=v0.x.y curl -fsSL https://linapro.ai/install.sh | bash"
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
  printf '\n'
  printf 'Upgrade later with Git tags:\n'
  printf '  git fetch --tags --force origin\n'
  printf '  git checkout --detach <new-version-tag>\n'
}

clone_release() {
  local version="$1"
  local target="$2"
  local clone_args

  clone_args=()
  if [ "${LINAPRO_SHALLOW:-}" = "1" ]; then
    clone_args+=(--depth 1 --no-single-branch)
    log_warn "Shallow clone enabled. Run git fetch --unshallow --tags --force origin before the first tag-based upgrade if Git reports a shallow-history limitation."
  fi

  log_info "Cloning $LINAPRO_REPO_URL."
  if ! git clone "${clone_args[@]}" "$LINAPRO_REPO_URL" "$target"; then
    die "git clone failed. Check network access and repository availability."
  fi

  if ! git -C "$target" config remote.origin.tagOpt --tags; then
    die "Failed to configure origin to fetch release tags."
  fi

  if ! git -C "$target" fetch --tags --force origin; then
    die "Failed to fetch release tags from origin."
  fi

  log_info "Checking out LinaPro $version."
  if ! git -C "$target" checkout --detach "$version"; then
    die "Failed to check out release tag '$version'. Check that the tag exists in origin."
  fi
}

main() {
  local os_name version target target_abs

  os_name="$(detect_os)"
  if ! command -v git >/dev/null 2>&1; then
    die "git is required before running this installer."
  fi

  version="$(resolve_version)"
  target="${LINAPRO_DIR:-./linapro}"
  print_banner "$version" "$target" "$os_name"

  prepare_target "$target"
  clone_release "$version" "$target"

  target_abs="$(cd "$target" && pwd -P)"
  print_next_steps "$target_abs"
}

main "$@"
