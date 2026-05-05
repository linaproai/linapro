#!/usr/bin/env bash
# Unit-tests lina-upgrade verification orchestration with fake toolchain commands.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
VERIFIER="$REPO_ROOT/.claude/skills/lina-upgrade/scripts/upgrade-verify.sh"

fail() {
  printf 'FAIL upgrade-verify: %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local text="$1"
  local pattern="$2"
  printf '%s\n' "$text" | grep -F "$pattern" >/dev/null ||
    fail "expected output to contain '$pattern'"
}

assert_not_contains() {
  local text="$1"
  local pattern="$2"
  if printf '%s\n' "$text" | grep -F "$pattern" >/dev/null; then
    fail "expected output not to contain '$pattern'"
  fi
}

create_fixture_repo() {
  local fixture_repo="$1"
  mkdir -p "$fixture_repo/apps/lina-core" "$fixture_repo/apps/lina-vben" "$fixture_repo/hack/tests"
  printf '{"scripts":{"typecheck":"typecheck"}}\n' >"$fixture_repo/apps/lina-vben/package.json"
}

write_fake_tools() {
  local fake_bin="$1"
  mkdir -p "$fake_bin"
  cat >"$fake_bin/go" <<'SCRIPT'
#!/usr/bin/env bash
if [ "${1:-}" = "build" ]; then
  printf 'backend build\n' >> "${TRACE:?}"
  exit 0
fi
printf 'unexpected go args: %s\n' "$*" >&2
exit 2
SCRIPT
  cat >"$fake_bin/node" <<'SCRIPT'
#!/usr/bin/env bash
if [ "${1:-}" = "-e" ]; then
  exit 0
fi
printf 'unexpected node args: %s\n' "$*" >&2
exit 2
SCRIPT
  cat >"$fake_bin/pnpm" <<'SCRIPT'
#!/usr/bin/env bash
case "${1:-}" in
  typecheck)
    printf 'frontend typecheck\n' >> "${TRACE:?}"
    ;;
  lint)
    printf 'frontend lint\n' >> "${TRACE:?}"
    if [ "${FAIL_LINT:-0}" = "1" ]; then
      printf 'lint fixture failed\n' >&2
      exit 9
    fi
    ;;
  playwright)
    if [ "${2:-}" = "test" ]; then
      printf 'e2e smoke\n' >> "${TRACE:?}"
    else
      printf 'unexpected playwright args: %s\n' "$*" >&2
      exit 2
    fi
    ;;
  *)
    printf 'unexpected pnpm args: %s\n' "$*" >&2
    exit 2
    ;;
esac
SCRIPT
  chmod +x "$fake_bin"/*
}

test_success() {
  local temp_dir fixture_repo fake_bin trace_file output log_file trace
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fixture_repo="$temp_dir/repo"
  fake_bin="$temp_dir/bin"
  trace_file="$temp_dir/trace.log"
  create_fixture_repo "$fixture_repo"
  write_fake_tools "$fake_bin"

  output="$(TRACE="$trace_file" PATH="$fake_bin:/usr/bin:/bin" LINAPRO_REPO_ROOT="$fixture_repo" bash "$VERIFIER")"
  log_file="$fixture_repo/temp/lina-upgrade/verify.log"
  trace="$(cat "$trace_file")"

  assert_contains "$output" '[lina-upgrade] verification completed.'
  assert_contains "$trace" 'backend build'
  assert_contains "$trace" 'frontend typecheck'
  assert_contains "$trace" 'frontend lint'
  assert_contains "$trace" 'e2e smoke'
  assert_contains "$(cat "$log_file")" '[lina-upgrade] backend go build'
  assert_contains "$(cat "$log_file")" '[lina-upgrade] e2e smoke'
  printf 'PASS upgrade-verify success\n'
}

test_lint_failure() {
  local temp_dir fixture_repo fake_bin trace_file output rc log_file trace
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fixture_repo="$temp_dir/repo"
  fake_bin="$temp_dir/bin"
  trace_file="$temp_dir/trace.log"
  create_fixture_repo "$fixture_repo"
  write_fake_tools "$fake_bin"

  set +e
  output="$(TRACE="$trace_file" FAIL_LINT=1 PATH="$fake_bin:/usr/bin:/bin" LINAPRO_REPO_ROOT="$fixture_repo" bash "$VERIFIER" 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "lint failure should return non-zero"
  log_file="$fixture_repo/temp/lina-upgrade/verify.log"
  trace="$(cat "$trace_file")"
  assert_contains "$output" '[lina-upgrade] frontend lint failed. See'
  assert_contains "$(cat "$log_file")" 'lint fixture failed'
  assert_contains "$trace" 'backend build'
  assert_contains "$trace" 'frontend typecheck'
  assert_contains "$trace" 'frontend lint'
  assert_not_contains "$trace" 'e2e smoke'
  printf 'PASS upgrade-verify lint-failure\n'
}

test_success
test_lint_failure
