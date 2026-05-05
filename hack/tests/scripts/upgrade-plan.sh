#!/usr/bin/env bash
# Unit-tests lina-upgrade plan generation against temporary git fixtures.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
PLANNER="$REPO_ROOT/.claude/skills/lina-upgrade/scripts/upgrade-plan.sh"

fail() {
  printf 'FAIL upgrade-plan: %s\n' "$*" >&2
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

write_metadata() {
  local repo="$1"
  local version="$2"
  mkdir -p "$repo/apps/lina-core/manifest/config"
  cat >"$repo/apps/lina-core/manifest/config/metadata.yaml" <<EOF
framework:
  version: "$version"
EOF
}

create_source_repo() {
  local source_repo="$1"
  git init "$source_repo" >/dev/null
  git -C "$source_repo" config user.email "fixture@example.test"
  git -C "$source_repo" config user.name "Fixture"

  write_metadata "$source_repo" v0.5.0
  mkdir -p "$source_repo/apps/lina-core/manifest/sql"
  printf 'CREATE TABLE IF NOT EXISTS upgrade_plan_base (id int);\n' >"$source_repo/apps/lina-core/manifest/sql/001-base.sql"
  git -C "$source_repo" add .
  git -C "$source_repo" commit -m "baseline v0.5.0" >/dev/null
  git -C "$source_repo" tag v0.5.0

  write_metadata "$source_repo" v0.6.0
  mkdir -p \
    "$source_repo/apps/lina-core/internal/dao" \
    "$source_repo/apps/lina-core/internal/service/upgrade" \
    "$source_repo/apps/lina-core/pkg/bizerr" \
    "$source_repo/apps/lina-plugins/plugin-alpha/backend" \
    "$source_repo/openspec/changes/archive/2026-05-05-upgrade-plan"
  printf 'CREATE TABLE IF NOT EXISTS upgrade_plan_next (id int);\n' >"$source_repo/apps/lina-core/manifest/sql/002-upgrade-plan.sql"
  printf 'package dao\n' >"$source_repo/apps/lina-core/internal/dao/sys_upgrade.go"
  printf 'package upgrade\n' >"$source_repo/apps/lina-core/internal/service/upgrade/service.go"
  printf 'package bizerr\n' >"$source_repo/apps/lina-core/pkg/bizerr/code.go"
  printf 'package backend\n' >"$source_repo/apps/lina-plugins/plugin-alpha/backend/plugin.go"
  printf '# Changelog\n\n## v0.6.0\n\n**BREAKING** Tier 1 contract changed in apps/lina-core/pkg/bizerr.\n' >"$source_repo/CHANGELOG.md"
  printf '## Why\n\nTier 1 migration note for plugin contracts.\n' >"$source_repo/openspec/changes/archive/2026-05-05-upgrade-plan/proposal.md"
  git -C "$source_repo" add .
  git -C "$source_repo" commit -m "target v0.6.0" >/dev/null
  git -C "$source_repo" tag v0.6.0

  write_metadata "$source_repo" v0.6.1
  printf 'package upgrade\n' >"$source_repo/apps/lina-core/internal/service/upgrade/no_sql.go"
  git -C "$source_repo" add .
  git -C "$source_repo" commit -m "target v0.6.1" >/dev/null
  git -C "$source_repo" tag v0.6.1
}

clone_fixture() {
  local temp_dir="$1"
  local baseline="$2"
  local source_repo="$temp_dir/source"
  local remote_repo="$temp_dir/remote.git"
  local work_repo="$temp_dir/work"

  create_source_repo "$source_repo"
  git clone --bare "$source_repo" "$remote_repo" >/dev/null
  git clone "$remote_repo" "$work_repo" >/dev/null
  git -C "$work_repo" config user.email "fixture@example.test"
  git -C "$work_repo" config user.name "Fixture"
  git -C "$work_repo" checkout -B user-work "$baseline" >/dev/null
  git -C "$work_repo" remote add upstream "$remote_repo"
  printf '%s\n' "$work_repo"
}

test_success() {
  local temp_dir work_repo output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"

  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$PLANNER" v0.6.0)"

  assert_contains "$output" '# LinaPro Upgrade Plan'
  assert_contains "$output" "Baseline version: \`v0.5.0\`"
  assert_contains "$output" "Target version: \`v0.6.0\`"
  assert_contains "$output" '**BREAKING** Tier 1 contract changed'
  assert_contains "$output" 'Tier 1 migration note for plugin contracts.'
  assert_contains "$output" 'tier1 apps/lina-core/pkg/bizerr/code.go'
  assert_contains "$output" 'tier1 apps/lina-plugins/plugin-alpha/backend/plugin.go'
  assert_contains "$output" 'tier2 apps/lina-core/internal/service/upgrade/service.go'
  assert_contains "$output" 'tier3 apps/lina-core/internal/dao/sys_upgrade.go'
  assert_contains "$output" 'apps/lina-core/manifest/sql/002-upgrade-plan.sql'
  printf 'PASS upgrade-plan success\n'
}

test_no_new_sql() {
  local temp_dir work_repo output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.6.0)"

  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$PLANNER" v0.6.1)"

  assert_contains "$output" "Target version: \`v0.6.1\`"
  assert_contains "$output" 'tier2 apps/lina-core/internal/service/upgrade/no_sql.go'
  assert_not_contains "$output" '003-'
  printf 'PASS upgrade-plan no-new-sql\n'
}

test_target_not_greater() {
  local temp_dir work_repo output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"

  set +e
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$PLANNER" v0.5.0 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "target-not-greater should fail"
  assert_contains "$output" 'ERR_TARGET_NOT_GREATER baseline=v0.5.0 target=v0.5.0'
  printf 'PASS upgrade-plan target-not-greater\n'
}

test_target_tag_missing() {
  local temp_dir work_repo output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"

  set +e
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$PLANNER" v0.7.0 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "target-tag-missing should fail"
  assert_contains "$output" 'ERR_TARGET_TAG_NOT_FOUND target=v0.7.0'
  printf 'PASS upgrade-plan target-tag-missing\n'
}

test_invalid_target() {
  local temp_dir work_repo output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"

  set +e
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$PLANNER" latest 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "invalid-target should fail"
  assert_contains "$output" 'ERR_TARGET_VERSION_INVALID target=latest'
  printf 'PASS upgrade-plan invalid-target\n'
}

test_success
test_no_new_sql
test_target_not_greater
test_target_tag_missing
test_invalid_target
