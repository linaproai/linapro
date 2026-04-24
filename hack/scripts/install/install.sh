#!/usr/bin/env bash
set -euo pipefail

DEFAULT_REPO="gqcn/linapro"
DEFAULT_FALLBACK_REF="main"
SCRIPT_NAME="$(basename "$0")"

REPO="$DEFAULT_REPO"
REF=""
INSTALL_DIR=""
INSTALL_NAME=""
USE_CURRENT_DIR=0
FORCE_OVERLAY=0
WORK_DIR=""
DOWNLOADED_FROM=""
REF_EXPLICIT=0
REF_SOURCE=""

cleanup() {
  if [ -n "$WORK_DIR" ] && [ -d "$WORK_DIR" ]; then
    rm -rf "$WORK_DIR"
  fi
}
trap cleanup EXIT

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<USAGE
LinaPro installer for macOS and Linux.

Usage:
  $SCRIPT_NAME [options]

Options:
  --repo <owner/name>   GitHub repository to download. Default: $DEFAULT_REPO
  --ref <value>         Branch, tag, or commit reference. Default: latest stable tag
                        Fallback: $DEFAULT_FALLBACK_REF when no stable tag is available.
  --dir <path>          Install into the specified directory.
  --name <directory>    Install into a new child directory under the current path.
  --current-dir         Install directly into the current working directory.
  --force               Allow overlay install into a non-empty target directory.
  --help                Show this help message.

Examples:
  $SCRIPT_NAME
  $SCRIPT_NAME --ref v0.1.0 --name linapro-v0.1.0
  $SCRIPT_NAME --dir /opt/workspaces/linapro
  $SCRIPT_NAME --current-dir --force

Advanced environment variables:
  LINAPRO_INSTALL_ARCHIVE_PATH  Use a local .tar.gz archive instead of downloading.
  LINAPRO_INSTALL_STABLE_REF    Override the auto-detected stable tag.
USAGE
}

normalize_repo() {
  local input="$1"
  local repo="$input"

  if [[ "$repo" =~ ^https?://github\.com/([^/]+)/([^/]+)\.git/?$ ]]; then
    repo="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
  elif [[ "$repo" =~ ^https?://github\.com/([^/]+)/([^/]+)/?$ ]]; then
    repo="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
  elif [[ "$repo" =~ ^git@github\.com:([^/]+)/([^/]+)(\.git)?$ ]]; then
    repo="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
  fi

  repo="${repo%.git}"

  if [[ ! "$repo" =~ ^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$ ]]; then
    fail "unsupported repository value '$input'; use owner/name or a GitHub URL"
  fi

  printf '%s\n' "$repo"
}

require_downloader() {
  if command -v curl >/dev/null 2>&1; then
    printf 'curl\n'
    return 0
  fi

  if command -v wget >/dev/null 2>&1; then
    printf 'wget\n'
    return 0
  fi

  fail "curl or wget is required to download the source archive"
}

download_with_tool() {
  local tool="$1"
  local url="$2"
  local out_file="$3"

  if [ "$tool" = "curl" ]; then
    curl -fsSL "$url" -o "$out_file" >/dev/null 2>&1
    return $?
  fi

  wget -q -O "$out_file" "$url" >/dev/null 2>&1
}

fetch_text_with_tool() {
  local tool="$1"
  local url="$2"

  if [ "$tool" = "curl" ]; then
    curl -fsSL "$url"
    return $?
  fi

  wget -q -O - "$url"
}

build_archive_candidates() {
  local repo="$1"
  local ref="$2"

  printf 'https://codeload.github.com/%s/tar.gz/refs/heads/%s\n' "$repo" "$ref"
  printf 'https://codeload.github.com/%s/tar.gz/refs/tags/%s\n' "$repo" "$ref"
  printf 'https://codeload.github.com/%s/tar.gz/%s\n' "$repo" "$ref"
}

extract_tag_names() {
  grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' | \
    sed 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/'
}

select_latest_stable_tag() {
  local names="$1"
  local stable_tag=""

  stable_tag="$(
    printf '%s\n' "$names" | \
      awk '
        /^[vV]?[0-9]+\.[0-9]+\.[0-9]+$/ {
          tag=$0
          normalized=$0
          sub(/^[vV]/, "", normalized)
          split(normalized, parts, ".")
          printf "%010d.%010d.%010d %s\n", parts[1], parts[2], parts[3], tag
        }
      ' | \
      sort | \
      tail -n 1 | \
      awk '{print $2}'
  )"

  printf '%s\n' "$stable_tag"
}

resolve_effective_ref() {
  if [ "$REF_EXPLICIT" -eq 1 ]; then
    REF_SOURCE="user provided"
    return 0
  fi

  if [ -n "${LINAPRO_INSTALL_STABLE_REF:-}" ]; then
    REF="$LINAPRO_INSTALL_STABLE_REF"
    REF_SOURCE="stable override"
    return 0
  fi

  local downloader
  downloader="$(require_downloader)"

  local tags_json=""
  local tag_names=""
  local stable_tag=""

  tags_json="$(fetch_text_with_tool "$downloader" "https://api.github.com/repos/$REPO/tags?per_page=100" 2>/dev/null || true)"
  if [ -n "$tags_json" ]; then
    tag_names="$(printf '%s' "$tags_json" | extract_tag_names || true)"
    stable_tag="$(select_latest_stable_tag "$tag_names")"
    if [ -n "$stable_tag" ]; then
      REF="$stable_tag"
      REF_SOURCE="latest stable tag"
      return 0
    fi
  fi

  REF="$DEFAULT_FALLBACK_REF"
  REF_SOURCE="fallback branch"
}

download_archive() {
  local repo="$1"
  local ref="$2"
  local archive_file="$3"

  if [ -n "${LINAPRO_INSTALL_ARCHIVE_PATH:-}" ]; then
    if [ ! -f "$LINAPRO_INSTALL_ARCHIVE_PATH" ]; then
      fail "LINAPRO_INSTALL_ARCHIVE_PATH points to a missing file: $LINAPRO_INSTALL_ARCHIVE_PATH"
    fi

    cp "$LINAPRO_INSTALL_ARCHIVE_PATH" "$archive_file"
    DOWNLOADED_FROM="local archive: $LINAPRO_INSTALL_ARCHIVE_PATH"
    return 0
  fi

  local downloader
  downloader="$(require_downloader)"

  local candidate
  while IFS= read -r candidate; do
    if download_with_tool "$downloader" "$candidate" "$archive_file"; then
      DOWNLOADED_FROM="$candidate"
      return 0
    fi
    rm -f "$archive_file"
  done < <(build_archive_candidates "$repo" "$ref")

  fail "failed to download archive for repository '$repo' and ref '$ref'"
}

directory_has_entries() {
  local dir="$1"
  local entries=()

  if [ ! -d "$dir" ]; then
    return 1
  fi

  shopt -s nullglob dotglob
  entries=("$dir"/*)
  shopt -u nullglob dotglob

  [ "${#entries[@]}" -gt 0 ]
}

resolve_target_dir() {
  local repo="$1"
  local target_name=""

  if [ "$USE_CURRENT_DIR" -eq 1 ]; then
    pwd -P
    return 0
  fi

  if [ -n "$INSTALL_DIR" ]; then
    local parent_dir
    parent_dir="$(dirname "$INSTALL_DIR")"
    mkdir -p "$parent_dir"
    parent_dir="$(cd "$parent_dir" && pwd -P)"
    printf '%s/%s\n' "$parent_dir" "$(basename "$INSTALL_DIR")"
    return 0
  fi

  if [ -n "$INSTALL_NAME" ]; then
    target_name="$INSTALL_NAME"
  else
    target_name="$(basename "$repo")"
  fi

  printf '%s/%s\n' "$(pwd -P)" "$target_name"
}

extract_source_dir() {
  local archive_file="$1"
  local extract_root="$2"

  mkdir -p "$extract_root"
  tar -xzf "$archive_file" -C "$extract_root"

  local children=()
  shopt -s nullglob dotglob
  children=("$extract_root"/*)
  shopt -u nullglob dotglob

  if [ "${#children[@]}" -ne 1 ] || [ ! -d "${children[0]}" ]; then
    fail "expected the archive to extract into a single top-level directory"
  fi

  printf '%s\n' "${children[0]}"
}

copy_source_contents() {
  local source_dir="$1"
  local target_dir="$2"

  mkdir -p "$target_dir"
  cp -R "$source_dir"/. "$target_dir"/
}

print_check_line() {
  local label="$1"
  local command_name="$2"
  shift 2

  if command -v "$command_name" >/dev/null 2>&1; then
    local output
    output="$("$command_name" "$@" 2>/dev/null | sed -n '1p' || true)"
    if [ -n "$output" ]; then
      printf '  [OK] %s: %s\n' "$label" "$output"
    else
      printf '  [OK] %s: detected\n' "$label"
    fi
    return 0
  fi

  printf '  [MISSING] %s\n' "$label"
  return 1
}

run_environment_check() {
  local missing_count=0

  log
  log "Environment check:"
  print_check_line "Go" go version || missing_count=$((missing_count + 1))
  print_check_line "Node.js" node --version || missing_count=$((missing_count + 1))
  print_check_line "pnpm" pnpm --version || missing_count=$((missing_count + 1))
  print_check_line "MySQL" mysql --version || missing_count=$((missing_count + 1))
  print_check_line "make" make --version || missing_count=$((missing_count + 1))

  if [ "$missing_count" -gt 0 ]; then
    log
    log "Some dependencies are missing. Install them before running the LinaPro bootstrap commands."
  fi
}

print_next_steps() {
  local target_dir="$1"

  log
  log "Project directory: $target_dir"
  log "Archive source: $DOWNLOADED_FROM"
  log
  log "Next steps:"
  printf '  1. cd "%s"\n' "$target_dir"
  log "  2. make init confirm=init"
  log "  3. make mock confirm=mock"
  log "  4. make dev"
  log
  log "The installer only bootstraps the source tree and environment check."
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --repo)
        [ "$#" -ge 2 ] || fail "--repo requires a value"
        REPO="$2"
        shift 2
        ;;
      --ref)
        [ "$#" -ge 2 ] || fail "--ref requires a value"
        REF="$2"
        REF_EXPLICIT=1
        shift 2
        ;;
      --dir)
        [ "$#" -ge 2 ] || fail "--dir requires a value"
        INSTALL_DIR="$2"
        shift 2
        ;;
      --name)
        [ "$#" -ge 2 ] || fail "--name requires a value"
        INSTALL_NAME="$2"
        shift 2
        ;;
      --current-dir)
        USE_CURRENT_DIR=1
        shift
        ;;
      --force)
        FORCE_OVERLAY=1
        shift
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        fail "unknown option: $1"
        ;;
    esac
  done

  if [ "$USE_CURRENT_DIR" -eq 1 ] && [ -n "$INSTALL_DIR" ]; then
    fail "--current-dir and --dir cannot be used together"
  fi

  if [ "$USE_CURRENT_DIR" -eq 1 ] && [ -n "$INSTALL_NAME" ]; then
    fail "--current-dir and --name cannot be used together"
  fi

  if [ -n "$INSTALL_DIR" ] && [ -n "$INSTALL_NAME" ]; then
    fail "--dir and --name cannot be used together"
  fi
}

main() {
  parse_args "$@"
  REPO="$(normalize_repo "$REPO")"
  resolve_effective_ref

  WORK_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t linapro-install)"
  local archive_file="$WORK_DIR/source.tar.gz"
  local extract_root="$WORK_DIR/extract"
  local target_dir=""
  local source_dir=""

  target_dir="$(resolve_target_dir "$REPO")"

  log "Repository: $REPO"
  log "Resolved ref: $REF [$REF_SOURCE]"
  log "Target directory: $target_dir"

  if [ -d "$target_dir" ] && directory_has_entries "$target_dir" && [ "$FORCE_OVERLAY" -ne 1 ]; then
    fail "target directory '$target_dir' is not empty; rerun with --force to overlay the source tree"
  fi

  download_archive "$REPO" "$REF" "$archive_file"
  source_dir="$(extract_source_dir "$archive_file" "$extract_root")"
  copy_source_contents "$source_dir" "$target_dir"

  log "LinaPro source bootstrapped successfully."
  run_environment_check
  print_next_steps "$target_dir"
}

main "$@"
