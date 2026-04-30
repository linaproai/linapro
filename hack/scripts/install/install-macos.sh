#!/usr/bin/env bash
# Performs post-clone setup for LinaPro on macOS.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
# shellcheck source=hack/scripts/install/lib/_common.sh
source "$SCRIPT_DIR/lib/_common.sh"

log_info "macOS install hint for missing tools: brew install go node pnpm git make mysql"
run_standard_install "$ROOT_DIR" "macOS" 0
