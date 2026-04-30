#!/usr/bin/env bash
# Regenerates GoFrame artifacts after LinaPro upgrade conflict handling.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${LINAPRO_REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
LOG_DIR="$REPO_ROOT/temp/lina-upgrade"
LOG_FILE="$LOG_DIR/regenerate.log"

mkdir -p "$LOG_DIR"
: >"$LOG_FILE"

run_logged() {
  local label="$1"
  shift
  printf '[lina-upgrade] %s\n' "$label" | tee -a "$LOG_FILE"
  if ! "$@" >>"$LOG_FILE" 2>&1; then
    printf '[lina-upgrade] %s failed. See %s\n' "$label" "$LOG_FILE" >&2
    exit 1
  fi
}

run_make_dao() {
  cd "$REPO_ROOT/apps/lina-core" && make dao
}

run_make_ctrl() {
  cd "$REPO_ROOT/apps/lina-core" && make ctrl
}

run_logged "make dao" run_make_dao
run_logged "make ctrl" run_make_ctrl
printf '[lina-upgrade] regeneration completed. Log: %s\n' "$LOG_FILE"
