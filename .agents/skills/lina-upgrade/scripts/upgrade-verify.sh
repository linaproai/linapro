#!/usr/bin/env bash
# Runs the LinaPro upgrade verification suite after regeneration and migration.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${LINAPRO_REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
LOG_DIR="$REPO_ROOT/temp/lina-upgrade"
LOG_FILE="$LOG_DIR/verify.log"

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

run_typecheck() {
  cd "$REPO_ROOT/apps/lina-vben"
  if node -e "const p=require('./package.json'); process.exit(p.scripts && p.scripts.typecheck ? 0 : 1)" >/dev/null 2>&1; then
    pnpm typecheck
  else
    pnpm run check:type
  fi
}

run_backend_build() {
  cd "$REPO_ROOT/apps/lina-core" && go build ./...
}

run_frontend_lint() {
  cd "$REPO_ROOT/apps/lina-vben" && pnpm lint
}

run_e2e_smoke() {
  cd "$REPO_ROOT/hack/tests" &&
    pnpm playwright test \
      e2e/auth/TC0001-login-success.ts \
      e2e/settings/dict/TC0012-dict-type-crud.ts \
      e2e/extension/plugin/TC0106-source-plugin-upgrade-governance.ts
}

run_logged "backend go build" run_backend_build
run_logged "frontend typecheck" run_typecheck
run_logged "frontend lint" run_frontend_lint
run_logged "e2e smoke" run_e2e_smoke
printf '[lina-upgrade] verification completed. Log: %s\n' "$LOG_FILE"
