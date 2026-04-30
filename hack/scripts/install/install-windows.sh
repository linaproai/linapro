#!/usr/bin/env bash
# Performs post-clone setup for LinaPro on Windows. This script must run in Git Bash or WSL.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
# shellcheck source=hack/scripts/install/lib/_common.sh
source "$SCRIPT_DIR/lib/_common.sh"

kernel="$(uname -s 2>/dev/null || printf 'unknown')"
case "$kernel" in
  MINGW*|MSYS*|CYGWIN*)
    log_info "Detected Windows Git Bash runtime: $kernel"
    if command -v cygpath >/dev/null 2>&1; then
      log_debug "Windows repository path: $(cygpath -w "$ROOT_DIR")"
    fi
    ;;
  Linux)
    if grep -qi microsoft /proc/version 2>/dev/null; then
      log_info "Detected WSL runtime."
    else
      die "install-windows.sh must be run from Git Bash or WSL."
    fi
    ;;
  *)
    die "install-windows.sh must be run from Git Bash or WSL."
    ;;
esac

log_info "Windows install hint for missing tools: winget install GoLang.Go OpenJS.NodeJS, npm i -g pnpm, scoop install make, or choco install make mysql-cli."
run_standard_install "$ROOT_DIR" "Windows Git Bash or WSL" 1
