#!/usr/bin/env bash
# Unit-tests lina-upgrade baseline validation against temporary git fixtures.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CHECKER="$REPO_ROOT/.claude/skills/lina-upgrade/scripts/upgrade-baseline-check.sh"

fail() {
  printf 'FAIL upgrade-baseline-check: %s\n' "$*" >&2
  exit 1
}

create_source_repo() {
  local source_repo="$1"
  local version="$2"
  git init "$source_repo" >/dev/null
  git -C "$source_repo" config user.email "fixture@example.test"
  git -C "$source_repo" config user.name "Fixture"
  mkdir -p "$source_repo/apps/lina-core/manifest/config" "$source_repo/apps/lina-core/manifest/sql"
  cat >"$source_repo/apps/lina-core/manifest/config/metadata.yaml" <<EOF
framework:
  version: "$version"
EOF
  printf 'CREATE TABLE IF NOT EXISTS fixture (id int);\n' >"$source_repo/apps/lina-core/manifest/sql/001-fixture.sql"
  git -C "$source_repo" add .
  git -C "$source_repo" commit -m "baseline $version" >/dev/null
  git -C "$source_repo" tag "$version"
}

clone_fixture() {
  local temp_dir="$1"
  local version="$2"
  local source_repo="$temp_dir/source"
  local remote_repo="$temp_dir/remote.git"
  local work_repo="$temp_dir/work"
  create_source_repo "$source_repo" "$version"
  git clone --bare "$source_repo" "$remote_repo" >/dev/null
  git clone "$remote_repo" "$work_repo" >/dev/null
  git -C "$work_repo" config user.email "fixture@example.test"
  git -C "$work_repo" config user.name "Fixture"
  git -C "$work_repo" remote add upstream "$remote_repo"
  printf '%s\n' "$work_repo"
}

test_ok() {
  local temp_dir work_repo output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"
  printf 'local change\n' >"$work_repo/local.txt"
  git -C "$work_repo" add local.txt
  git -C "$work_repo" commit -m "local change" >/dev/null
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$CHECKER")"
  printf '%s\n' "$output" | grep -F 'OK_BASELINE_CONFIRMED' >/dev/null || fail "expected OK output, got: $output"
}

test_tag_not_found() {
  local temp_dir work_repo output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"
  sed -i.bak 's/v0.5.0/v9.9.9/' "$work_repo/apps/lina-core/manifest/config/metadata.yaml"
  rm -f "$work_repo/apps/lina-core/manifest/config/metadata.yaml.bak"
  set +e
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$CHECKER" 2>&1)"
  rc=$?
  set -e
  [ "$rc" -ne 0 ] || fail "tag-not-found should fail"
  printf '%s\n' "$output" | grep -F 'ERR_TAG_NOT_FOUND' >/dev/null || fail "expected ERR_TAG_NOT_FOUND, got: $output"
}

test_head_not_descendant() {
  local temp_dir work_repo output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"
  git -C "$work_repo" checkout --orphan unrelated >/dev/null
  git -C "$work_repo" rm -rf . >/dev/null 2>&1 || true
  mkdir -p "$work_repo/apps/lina-core/manifest/config" "$work_repo/apps/lina-core/manifest/sql"
  cat >"$work_repo/apps/lina-core/manifest/config/metadata.yaml" <<'EOF'
framework:
  version: "v0.5.0"
EOF
  printf 'CREATE TABLE IF NOT EXISTS unrelated (id int);\n' >"$work_repo/apps/lina-core/manifest/sql/001-fixture.sql"
  git -C "$work_repo" add .
  git -C "$work_repo" commit -m "unrelated head" >/dev/null
  set +e
  output="$(LINAPRO_REPO_ROOT="$work_repo" bash "$CHECKER" 2>&1)"
  rc=$?
  set -e
  [ "$rc" -ne 0 ] || fail "head-not-descendant should fail"
  printf '%s\n' "$output" | grep -F 'ERR_HEAD_NOT_DESCENDANT' >/dev/null || fail "expected ERR_HEAD_NOT_DESCENDANT, got: $output"
}

test_non_official_origin_uses_official_url() {
  local temp_dir work_repo remote_repo git_config output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  work_repo="$(clone_fixture "$temp_dir" v0.5.0)"
  remote_repo="$temp_dir/remote.git"
  git -C "$work_repo" remote remove upstream
  git -C "$work_repo" remote set-url origin https://github.com/example/linapro-fork.git
  git_config="$temp_dir/gitconfig"
  cat >"$git_config" <<EOF
[url "file://$remote_repo"]
	insteadOf = https://github.com/linaproai/linapro.git
EOF
  output="$(GIT_CONFIG_GLOBAL="$git_config" LINAPRO_REPO_ROOT="$work_repo" bash "$CHECKER")"
  printf '%s\n' "$output" | grep -F 'OK_BASELINE_CONFIRMED' >/dev/null || fail "expected OK output, got: $output"
  printf '%s\n' "$output" | grep -F 'upstream=https://github.com/linaproai/linapro.git' >/dev/null || fail "expected official upstream URL, got: $output"
}

test_ok
test_tag_not_found
test_head_not_descendant
test_non_official_origin_uses_official_url
printf 'PASS upgrade-baseline-check\n'
