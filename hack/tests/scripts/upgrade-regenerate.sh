#!/usr/bin/env bash
# Unit-tests lina-upgrade regeneration command orchestration and failure output.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
REGENERATOR="$REPO_ROOT/.claude/skills/lina-upgrade/scripts/upgrade-regenerate.sh"

fail() {
  printf 'FAIL upgrade-regenerate: %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local text="$1"
  local pattern="$2"
  printf '%s\n' "$text" | grep -F "$pattern" >/dev/null ||
    fail "expected output to contain '$pattern'"
}

create_fixture_repo() {
  local fixture_repo="$1"
  mkdir -p "$fixture_repo/apps/lina-core"
  cat >"$fixture_repo/apps/lina-core/Makefile" <<'MAKEFILE'
.PHONY: dao ctrl

dao:
	@printf 'dao\n' >> ../../trace.log

ctrl:
	@if [ "$$FAIL_CTRL" = "1" ]; then \
		printf 'ctrl fixture failed\n' >&2; \
		exit 42; \
	fi
	@printf 'ctrl\n' >> ../../trace.log
MAKEFILE
}

test_success() {
  local temp_dir fixture_repo output log_file trace
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fixture_repo="$temp_dir/repo"
  create_fixture_repo "$fixture_repo"

  output="$(FAIL_CTRL=0 LINAPRO_REPO_ROOT="$fixture_repo" bash "$REGENERATOR")"
  log_file="$fixture_repo/temp/lina-upgrade/regenerate.log"
  trace="$(cat "$fixture_repo/trace.log")"

  assert_contains "$output" '[lina-upgrade] regeneration completed.'
  assert_contains "$trace" 'dao'
  assert_contains "$trace" 'ctrl'
  assert_contains "$(cat "$log_file")" '[lina-upgrade] make dao'
  assert_contains "$(cat "$log_file")" '[lina-upgrade] make ctrl'
  printf 'PASS upgrade-regenerate success\n'
}

test_ctrl_failure() {
  local temp_dir fixture_repo output rc log_file trace
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fixture_repo="$temp_dir/repo"
  create_fixture_repo "$fixture_repo"

  set +e
  output="$(FAIL_CTRL=1 LINAPRO_REPO_ROOT="$fixture_repo" bash "$REGENERATOR" 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "ctrl failure should return non-zero"
  log_file="$fixture_repo/temp/lina-upgrade/regenerate.log"
  trace="$(cat "$fixture_repo/trace.log")"
  assert_contains "$trace" 'dao'
  assert_contains "$output" '[lina-upgrade] make ctrl failed. See'
  assert_contains "$(cat "$log_file")" 'ctrl fixture failed'
  printf 'PASS upgrade-regenerate ctrl-failure\n'
}

test_success
test_ctrl_failure
