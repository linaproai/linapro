#!/usr/bin/env bash
# Smoke-tests bootstrap.sh against local fixture repositories without network access.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
BOOTSTRAP="$REPO_ROOT/hack/scripts/install/bootstrap.sh"
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

assert_bootstrap_output() {
  local output="$1"
  assert_contains_text "$output" "LinaPro source downloaded successfully."
  assert_contains_text "$output" "lina-doctor"
  assert_contains_text "$output" "make init && make dev"
  assert_not_contains_text "$output" "make mock"
  assert_not_contains_text "$output" "go mod download"
  assert_not_contains_text "$output" "pnpm install"
  assert_not_contains_text "$output" "install-linux.sh"
  assert_not_contains_text "$output" "install-macos.sh"
  assert_not_contains_text "$output" "install-windows.sh"
}

run_bootstrap_case() {
  local name="$1"
  local version="$2"
  local target_mode="$3"
  local temp_dir remote_repo workspace git_config target_dir output
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
    "LINAPRO_VERSION=$version"
  )

  if [ "$target_mode" = "default" ]; then
    target_dir="$workspace/linapro"
    output="$(
      cd "$workspace"
      env "${env_vars[@]}" bash "$BOOTSTRAP"
    )"
  else
    target_dir="$workspace/custom"
    env_vars+=("LINAPRO_DIR=$target_dir")
    output="$(
      cd "$workspace"
      env "${env_vars[@]}" bash "$BOOTSTRAP"
    )"
  fi

  assert_file "$target_dir/VERSION"
  assert_contains_text "$(cat "$target_dir/VERSION")" "$version"
  assert_bootstrap_output "$output"
  printf 'PASS install-bootstrap %s\n' "$name"
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
        bash "$BOOTSTRAP" 2>&1
  )"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "force-home-refused should fail"
  assert_contains_text "$output" "Refusing to overwrite unsafe target directory"
  assert_file "$target_dir/existing.txt"
  printf 'PASS install-bootstrap force-home-refused\n'
}

test_latest_non_tag_refused() {
  local temp_dir fake_bin output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fake_bin="$temp_dir/bin"
  mkdir -p "$fake_bin"
  cat >"$fake_bin/curl" <<'SCRIPT'
#!/usr/bin/env bash
printf 'HTTP/2 302\r\n'
printf 'location: https://github.com/linaproai/linapro/releases\r\n'
SCRIPT
  chmod +x "$fake_bin/curl"

  set +e
  output="$(PATH="$fake_bin:$PATH" bash "$BOOTSTRAP" 2>&1)"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "latest-non-tag should fail"
  assert_contains_text "$output" "not a stable version tag"
  printf 'PASS install-bootstrap latest-non-tag-refused\n'
}

case "$CASE_NAME" in
  default)
    run_bootstrap_case default v0.0.1 default
    ;;
  version-override)
    run_bootstrap_case version-override v0.0.2 custom
    ;;
  force-home-refused)
    test_force_home_refused
    ;;
  latest-non-tag-refused)
    test_latest_non_tag_refused
    ;;
  all)
    run_bootstrap_case default v0.0.1 default
    run_bootstrap_case version-override v0.0.2 custom
    test_force_home_refused
    test_latest_non_tag_refused
    ;;
  *)
    fail "unknown case: $CASE_NAME"
    ;;
esac
