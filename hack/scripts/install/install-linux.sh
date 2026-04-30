#!/usr/bin/env bash
# Performs post-clone setup for LinaPro on Linux and WSL.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
# shellcheck source=hack/scripts/install/lib/_common.sh
source "$SCRIPT_DIR/lib/_common.sh"

log_info "Linux install hint for missing tools: apt-get install -y golang nodejs git make default-mysql-client, or use your distribution equivalents."
run_standard_install "$ROOT_DIR" "Linux" 0
