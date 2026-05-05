#!/usr/bin/env bash
# Unit tests for lina-doctor doctor-plan.sh.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
DOCTOR_PLAN="$REPO_ROOT/.claude/skills/lina-doctor/scripts/doctor-plan.sh"

fail() {
  printf 'FAIL doctor-plan: %s\n' "$*" >&2
  exit 1
}

make_check_json() {
  cat <<'JSON'
{"os":"macos","package_manager":"brew","shell":"zsh","repo_root_detected":true,"tools":{"go":{"ok":false},"node":{"ok":false},"pnpm":{"ok":false},"git":{"ok":true},"make":{"ok":true},"openspec":{"ok":false},"gf":{"ok":false},"playwright":{"ok":false},"goframe-v2":{"ok":false}},"path_issues":[],"mirror_hints":[]}
JSON
}

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT
input="$temp_dir/check.json"
make_check_json >"$input"

plan="$(bash "$DOCTOR_PLAN" --input "$input")"

line_of() {
  local tool="$1"
  printf '%s\n' "$plan" | grep -n "\"tool\": \"$tool\"" | cut -d: -f1 | head -n 1
}

go_line="$(line_of go)"
gf_line="$(line_of gf)"
node_line="$(line_of node)"
pnpm_line="$(line_of pnpm)"
goframe_line="$(line_of goframe-v2)"
playwright_line="$(line_of playwright)"

[ "$go_line" -lt "$gf_line" ] || fail "go should be before gf"
[ "$node_line" -lt "$pnpm_line" ] || fail "node should be before pnpm"
[ "$node_line" -lt "$goframe_line" ] || fail "node should be before goframe-v2"
[ "$pnpm_line" -lt "$playwright_line" ] || fail "pnpm should be before playwright"
printf '%s\n' "$plan" | grep -F '"optional": true' >/dev/null || fail "optional targets not marked"

skip_playwright="$(LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1 bash "$DOCTOR_PLAN" --input "$input")"
if printf '%s\n' "$skip_playwright" | grep -F '"tool": "playwright"' >/dev/null; then
  fail "playwright should be skipped"
fi

skip_goframe="$(LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1 bash "$DOCTOR_PLAN" --input "$input")"
if printf '%s\n' "$skip_goframe" | grep -F '"tool": "goframe-v2"' >/dev/null; then
  fail "goframe-v2 should be skipped"
fi

nvm_dir="$temp_dir/nvm"
mkdir -p "$nvm_dir"
printf '# nvm fixture\n' >"$nvm_dir/nvm.sh"
nvm_plan="$(NVM_DIR="$nvm_dir" bash "$DOCTOR_PLAN" --input "$input")"
printf '%s\n' "$nvm_plan" | grep -F 'nvm.sh' >/dev/null || fail "nvm plan should source nvm.sh"
printf '%s\n' "$nvm_plan" | grep -F 'nvm install 20' >/dev/null || fail "nvm plan should install Node 20"

printf 'PASS doctor-plan topology-and-skips\n'
