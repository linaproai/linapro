#!/usr/bin/env bash
# Smoke-tests install.sh against local fixture repositories without network access.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
INSTALLER="$REPO_ROOT/hack/scripts/install/install.sh"
CASE_NAME="${1:-all}"

fail() {
  printf 'FAIL install-bootstrap: %s\n' "$*" >&2
  exit 1
}

assert_file() {
  local path="$1"
  [ -f "$path" ] || fail "expected file to exist: $path"
}

assert_contains_text() {
  local text="$1"
  local pattern="$2"
  printf '%s\n' "$text" | grep -F "$pattern" >/dev/null 2>&1 ||
    fail "expected output to contain '$pattern'"
}

assert_not_contains_text() {
  local text="$1"
  local pattern="$2"
  if printf '%s\n' "$text" | grep -F "$pattern" >/dev/null 2>&1; then
    fail "expected output not to contain '$pattern'"
  fi
}

create_fixture_remote() {
  local work_dir="$1"
  local source_repo="$work_dir/source"
  local remote_repo="$work_dir/remote.git"

  git init "$source_repo" >/dev/null
  git -C "$source_repo" config user.email "fixture@example.test"
  git -C "$source_repo" config user.name "Fixture"
  mkdir -p "$source_repo/hack/scripts/install"
  printf 'v0.0.1\n' >"$source_repo/VERSION"
  git -C "$source_repo" add .
  git -C "$source_repo" commit -m "fixture v0.0.1" >/dev/null
  git -C "$source_repo" tag v0.0.1
  printf 'v0.0.2\n' >"$source_repo/VERSION"
  git -C "$source_repo" add VERSION
  git -C "$source_repo" commit -m "fixture v0.0.2" >/dev/null
  git -C "$source_repo" tag v0.0.2
  git clone --bare "$source_repo" "$remote_repo" >/dev/null
  printf '%s\n' "$remote_repo"
}

add_fixture_release() {
  local remote_repo="$1"
  local version="$2"
  local work_dir source_repo

  work_dir="$(mktemp -d)"
  source_repo="$work_dir/source"
  git clone "$remote_repo" "$source_repo" >/dev/null
  git -C "$source_repo" config user.email "fixture@example.test"
  git -C "$source_repo" config user.name "Fixture"
  printf '%s\n' "$version" >"$source_repo/VERSION"
  git -C "$source_repo" add VERSION
  git -C "$source_repo" commit -m "fixture $version" >/dev/null
  git -C "$source_repo" tag "$version"
  git -C "$source_repo" push origin HEAD:main --tags >/dev/null
  rm -rf "$work_dir"
}

assert_bootstrap_output() {
  local output="$1"
  assert_contains_text "$output" "LinaPro source downloaded successfully."
  assert_contains_text "$output" "lina-doctor"
  assert_contains_text "$output" "make init && make dev"
  assert_contains_text "$output" "git fetch --tags --force origin"
  assert_not_contains_text "$output" "make mock"
  assert_not_contains_text "$output" "go mod download"
  assert_not_contains_text "$output" "pnpm install"
  assert_not_contains_text "$output" "install-linux.sh"
  assert_not_contains_text "$output" "install-macos.sh"
  assert_not_contains_text "$output" "install-windows.sh"
}

run_bootstrap_case() {
  local name="$1"
  local expected_version="$2"
  local target_mode="$3"
  local version_env="$4"
  local temp_dir remote_repo workspace git_config target_dir output git_head git_origin
  local env_vars

  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  remote_repo="$(create_fixture_remote "$temp_dir")"
  workspace="$temp_dir/workspace"
  mkdir -p "$workspace"
  git_config="$temp_dir/gitconfig"
  cat >"$git_config" <<EOF
[url "file://$remote_repo"]
	insteadOf = https://github.com/linaproai/linapro.git
EOF

  env_vars=(
    "GIT_CONFIG_GLOBAL=$git_config"
  )
  if [ "$version_env" != "auto" ]; then
    env_vars+=("LINAPRO_VERSION=$version_env")
  fi

  if [ "$target_mode" = "default" ]; then
    target_dir="$workspace/linapro"
    output="$(
      cd "$workspace"
      env "${env_vars[@]}" bash "$INSTALLER"
    )"
  else
    target_dir="$workspace/custom"
    env_vars+=("LINAPRO_DIR=$target_dir")
    output="$(
      cd "$workspace"
      env "${env_vars[@]}" bash "$INSTALLER"
    )"
  fi

  assert_file "$target_dir/VERSION"
  assert_contains_text "$(cat "$target_dir/VERSION")" "$expected_version"
  git_head="$(git -C "$target_dir" describe --tags --exact-match HEAD)"
  [ "$git_head" = "$expected_version" ] || fail "expected HEAD tag $expected_version, got $git_head"
  git_origin="$(git -C "$target_dir" remote get-url origin)"
  [ -n "$git_origin" ] || fail "expected origin remote to remain configured"
  assert_bootstrap_output "$output"
  printf 'PASS install-bootstrap %s\n' "$name"
}

test_git_tag_upgrade() {
  local temp_dir remote_repo workspace git_config target_dir output
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  remote_repo="$(create_fixture_remote "$temp_dir")"
  workspace="$temp_dir/workspace"
  target_dir="$workspace/linapro"
  mkdir -p "$workspace"
  git_config="$temp_dir/gitconfig"
  cat >"$git_config" <<EOF
[url "file://$remote_repo"]
	insteadOf = https://github.com/linaproai/linapro.git
EOF

  output="$(
    cd "$workspace"
    env "GIT_CONFIG_GLOBAL=$git_config" bash "$INSTALLER"
  )"
  assert_contains_text "$output" "Version: v0.0.2"
  assert_contains_text "$(cat "$target_dir/VERSION")" "v0.0.2"

  add_fixture_release "$remote_repo" v0.0.3
  GIT_CONFIG_GLOBAL="$git_config" git -C "$target_dir" fetch --tags --force origin >/dev/null
  git -C "$target_dir" checkout --detach v0.0.3 >/dev/null

  assert_contains_text "$(cat "$target_dir/VERSION")" "v0.0.3"
  printf 'PASS install-bootstrap git-tag-upgrade\n'
}

test_force_home_refused() {
  local temp_dir remote_repo workspace git_config target_dir output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  remote_repo="$(create_fixture_remote "$temp_dir")"
  workspace="$temp_dir/workspace"
  target_dir="$workspace/linapro"
  mkdir -p "$target_dir"
  printf 'existing\n' >"$target_dir/existing.txt"
  git_config="$temp_dir/gitconfig"
  cat >"$git_config" <<EOF
[url "file://$remote_repo"]
	insteadOf = https://github.com/linaproai/linapro.git
EOF

  set +e
  output="$(
    cd "$workspace" &&
      env \
        "GIT_CONFIG_GLOBAL=$git_config" \
        "HOME=$target_dir" \
        "LINAPRO_VERSION=v0.0.1" \
        "LINAPRO_DIR=$target_dir" \
        "LINAPRO_FORCE=1" \
        bash "$INSTALLER" 2>&1
  )"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "force-home-refused should fail"
  assert_contains_text "$output" "Refusing to overwrite unsafe target directory"
  assert_file "$target_dir/existing.txt"
  printf 'PASS install-bootstrap force-home-refused\n'
}

test_latest_non_tag_refused() {
  local temp_dir fake_bin real_git output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  real_git="$(command -v git)"
  fake_bin="$temp_dir/bin"
  mkdir -p "$fake_bin"
  cat >"$fake_bin/git" <<SCRIPT
#!/usr/bin/env bash
if [ "\${1:-}" = "ls-remote" ]; then
  printf '0000000000000000000000000000000000000000\trefs/tags/latest\n'
  exit 0
fi
exec "$real_git" "\$@"
SCRIPT
  chmod +x "$fake_bin/git"

  set +e
  output="$(PATH="$fake_bin:$PATH" bash "$INSTALLER" 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "latest-non-tag should fail"
  assert_contains_text "$output" "Could not resolve the latest stable LinaPro release tag"
  printf 'PASS install-bootstrap latest-non-tag-refused\n'
}

case "$CASE_NAME" in
  default)
    run_bootstrap_case default v0.0.2 default auto
    ;;
  version-override)
    run_bootstrap_case version-override v0.0.1 custom v0.0.1
    ;;
  git-tag-upgrade)
    test_git_tag_upgrade
    ;;
  force-home-refused)
    test_force_home_refused
    ;;
  latest-non-tag-refused)
    test_latest_non_tag_refused
    ;;
  all)
    run_bootstrap_case default v0.0.2 default auto
    run_bootstrap_case version-override v0.0.1 custom v0.0.1
    test_git_tag_upgrade
    test_force_home_refused
    test_latest_non_tag_refused
    ;;
  *)
    fail "unknown case: $CASE_NAME"
    ;;
esac
