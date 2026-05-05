#!/usr/bin/env bash
# Unit tests for lina-doctor doctor-check.sh.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
DOCTOR_CHECK="$REPO_ROOT/.claude/skills/lina-doctor/scripts/doctor-check.sh"

fail() {
  printf 'FAIL doctor-check: %s\n' "$*" >&2
  exit 1
}

write_fake_tools() {
  local fake_bin="$1"
  mkdir -p "$fake_bin"
  cat >"$fake_bin/go" <<'SCRIPT'
#!/usr/bin/env bash
printf 'go version go1.22.0 darwin/arm64\n'
SCRIPT
  cat >"$fake_bin/node" <<'SCRIPT'
#!/usr/bin/env bash
printf 'v20.19.0\n'
SCRIPT
  cat >"$fake_bin/pnpm" <<'SCRIPT'
#!/usr/bin/env bash
if [ "${1:-}" = "exec" ]; then
  printf 'Version 1.58.2\n'
else
  printf '8.0.0\n'
fi
SCRIPT
  cat >"$fake_bin/git" <<'SCRIPT'
#!/usr/bin/env bash
if [ "${1:-}" = "rev-parse" ] && [ "${2:-}" = "--show-toplevel" ]; then
  exit 1
fi
printf 'git version 2.45.0\n'
SCRIPT
  cat >"$fake_bin/make" <<'SCRIPT'
#!/usr/bin/env bash
printf 'GNU Make 3.81\n'
SCRIPT
  cat >"$fake_bin/openspec" <<'SCRIPT'
#!/usr/bin/env bash
printf 'openspec 1.3.1\n'
SCRIPT
  cat >"$fake_bin/gf" <<'SCRIPT'
#!/usr/bin/env bash
printf 'GoFrame CLI v2.10.0\n'
SCRIPT
  cat >"$fake_bin/npm" <<'SCRIPT'
#!/usr/bin/env bash
if [ "${1:-}" = "config" ] && [ "${2:-}" = "get" ] && [ "${3:-}" = "prefix" ]; then
  printf '%s\n' "${NPM_PREFIX:?}"
elif [ "${1:-}" = "config" ] && [ "${2:-}" = "get" ] && [ "${3:-}" = "registry" ]; then
  printf 'https://registry.example.test/\n'
fi
SCRIPT
  chmod +x "$fake_bin"/*
}

run_case() {
  local name="$1"
  local expected_rc="$2"
  local setup="$3"
  local temp_dir fake_bin output rc
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  fake_bin="$temp_dir/bin"
  write_fake_tools "$fake_bin"
  mkdir -p "$temp_dir/home/Library/Caches/ms-playwright/chromium-123"
  mkdir -p "$temp_dir/home/.claude/skills/goframe-v2"
  mkdir -p "$temp_dir/npm-prefix/bin"
  printf -- '---\nname: goframe-v2\n---\n' >"$temp_dir/home/.claude/skills/goframe-v2/SKILL.md"

  case "$setup" in
    missing-pnpm) rm -f "$fake_bin/pnpm" ;;
    path-warning) mkdir -p "$temp_dir/home/go/bin" && cp "$fake_bin/gf" "$temp_dir/home/go/bin/gf" ;;
    missing-skill) rm -rf "$temp_dir/home/.claude/skills/goframe-v2" ;;
  esac

  set +e
  output="$(
    HOME="$temp_dir/home" \
    PATH="$fake_bin:$temp_dir/npm-prefix/bin:/usr/bin:/bin" \
    NPM_PREFIX="$temp_dir/npm-prefix" \
    GOPROXY="https://proxy.golang.org,direct" \
    PLAYWRIGHT_DOWNLOAD_HOST="https://cdn.example.test" \
    bash "$DOCTOR_CHECK" --check-only
  )"
  rc=$?
  set -e

  [ "$rc" -eq "$expected_rc" ] || fail "$name expected rc $expected_rc, got $rc: $output"
  case "$name" in
    all-ready) printf '%s\n' "$output" | grep -F '"ok": true' >/dev/null || fail "$name missing ok true" ;;
    missing-pnpm) printf '%s\n' "$output" | grep -F '"pnpm": { "present": false' >/dev/null || fail "$name missing pnpm false" ;;
    path-warning) printf '%s\n' "$output" | grep -F '"path_issues": [' >/dev/null || fail "$name missing path issue" ;;
    missing-skill) printf '%s\n' "$output" | grep -F '"goframe-v2": { "present": false' >/dev/null || fail "$name missing goframe-v2 false" ;;
  esac
  printf 'PASS doctor-check %s\n' "$name"
}

run_case all-ready 0 normal
run_case missing-pnpm 1 missing-pnpm
run_case path-warning 2 path-warning
run_case missing-skill 2 missing-skill
