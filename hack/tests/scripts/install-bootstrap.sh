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

assert_contains() {
  local path="$1"
  local pattern="$2"
  grep -F "$pattern" "$path" >/dev/null 2>&1 || fail "expected '$path' to contain '$pattern'"
}

create_fixture_remote() {
  local work_dir="$1"
  local source_repo="$work_dir/source"
  local remote_repo="$work_dir/remote.git"

  git init "$source_repo" >/dev/null
  git -C "$source_repo" config user.email "fixture@example.test"
  git -C "$source_repo" config user.name "Fixture"
  mkdir -p "$source_repo/hack/scripts/install"
  cat >"$source_repo/hack/scripts/install/install-linux.sh" <<'SCRIPT'
#!/usr/bin/env bash
set -euo pipefail
printf 'version=%s\n' "$(cat VERSION)" >>"${LINAPRO_INSTALL_FIXTURE_LOG:?}"
printf 'skip_mock=%s\n' "${LINAPRO_SKIP_MOCK:-0}" >>"${LINAPRO_INSTALL_FIXTURE_LOG:?}"
SCRIPT
  cp "$source_repo/hack/scripts/install/install-linux.sh" "$source_repo/hack/scripts/install/install-macos.sh"
  cp "$source_repo/hack/scripts/install/install-linux.sh" "$source_repo/hack/scripts/install/install-windows.sh"
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

run_bootstrap_case() {
  local name="$1"
  local version="$2"
  local target_mode="$3"
  local skip_mock="$4"
  local temp_dir remote_repo workspace git_config log_file target_dir
  local env_vars

  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  remote_repo="$(create_fixture_remote "$temp_dir")"
  workspace="$temp_dir/workspace"
  mkdir -p "$workspace"
  git_config="$temp_dir/gitconfig"
  log_file="$temp_dir/install.log"
  cat >"$git_config" <<EOF
[url "file://$remote_repo"]
	insteadOf = https://github.com/linaproai/linapro.git
EOF

  env_vars=(
    "GIT_CONFIG_GLOBAL=$git_config"
    "LINAPRO_VERSION=$version"
    "LINAPRO_INSTALL_FIXTURE_LOG=$log_file"
  )
  if [ "$skip_mock" = "1" ]; then
    env_vars+=("LINAPRO_SKIP_MOCK=1")
  fi

  if [ "$target_mode" = "default" ]; then
    target_dir="$workspace/linapro"
    (
      cd "$workspace"
      env "${env_vars[@]}" bash "$BOOTSTRAP"
    )
  else
    target_dir="$workspace/custom"
    env_vars+=("LINAPRO_DIR=$target_dir")
    (
      cd "$workspace"
      env "${env_vars[@]}" bash "$BOOTSTRAP"
    )
  fi

  assert_file "$target_dir/VERSION"
  assert_contains "$log_file" "version=$version"
  if [ "$skip_mock" = "1" ]; then
    assert_contains "$log_file" "skip_mock=1"
  fi
  printf 'PASS install-bootstrap %s\n' "$name"
}

test_force_home_refused() {
  local temp_dir remote_repo workspace git_config log_file target_dir output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  remote_repo="$(create_fixture_remote "$temp_dir")"
  workspace="$temp_dir/workspace"
  target_dir="$workspace/linapro"
  mkdir -p "$target_dir"
  printf 'existing\n' >"$target_dir/existing.txt"
  git_config="$temp_dir/gitconfig"
  log_file="$temp_dir/install.log"
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
        "LINAPRO_INSTALL_FIXTURE_LOG=$log_file" \
        bash "$BOOTSTRAP" 2>&1
  )"
  rc=$?
  set -e

  [ "$rc" -ne 0 ] || fail "force-home-refused should fail"
  printf '%s\n' "$output" | grep -F 'Refusing to overwrite unsafe target directory' >/dev/null ||
    fail "expected unsafe target refusal, got: $output"
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
  printf '%s\n' "$output" | grep -F 'not a stable version tag' >/dev/null ||
    fail "expected stable tag refusal, got: $output"
  printf 'PASS install-bootstrap latest-non-tag-refused\n'
}

case "$CASE_NAME" in
  default)
    run_bootstrap_case default v0.0.1 default 0
    ;;
  version-override)
    run_bootstrap_case version-override v0.0.2 custom 0
    ;;
  skip-mock)
    run_bootstrap_case skip-mock v0.0.1 custom 1
    ;;
  force-home-refused)
    test_force_home_refused
    ;;
  latest-non-tag-refused)
    test_latest_non_tag_refused
    ;;
  all)
    run_bootstrap_case default v0.0.1 default 0
    run_bootstrap_case version-override v0.0.2 custom 0
    run_bootstrap_case skip-mock v0.0.1 custom 1
    test_force_home_refused
    test_latest_non_tag_refused
    ;;
  *)
    fail "unknown case: $CASE_NAME"
    ;;
esac
