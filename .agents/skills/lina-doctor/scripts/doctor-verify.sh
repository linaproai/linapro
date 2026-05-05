#!/usr/bin/env bash
# Performs final Lina Doctor verification and prints a concise report.

set -euo pipefail

IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=.claude/skills/lina-doctor/lib/_common.sh
source "$SKILL_DIR/lib/_common.sh"

repo_root="$(find_repo_root)"
cd "$repo_root"

set +e
bash "$SCRIPT_DIR/doctor-check.sh" >/tmp/lina-doctor-check.json
check_rc=$?
set -e

printf 'Lina Doctor verification report\n'
printf 'Status code: %s\n' "$check_rc"
printf '\n'
printf 'Smoke output:\n'

for smoke in 'go version' 'node --version' 'pnpm --version' 'openspec --version' 'gf version'; do
  tool="${smoke%% *}"
  if command_exists "$tool"; then
    printf '$ %s\n' "$smoke"
    bash -lc "$smoke" 2>&1 | head -n 5
  else
    printf '$ %s\nmissing\n' "$smoke"
  fi
done

if [ -d "$repo_root/hack/tests" ] && command_exists pnpm; then
  printf '$ cd hack/tests && pnpm exec playwright --version\n'
  (cd "$repo_root/hack/tests" && pnpm exec playwright --version) 2>&1 | head -n 5 || true
fi

if [ -f "$HOME/.claude/skills/goframe-v2/SKILL.md" ]; then
  printf '$ head -5 ~/.claude/skills/goframe-v2/SKILL.md\n'
  head -5 "$HOME/.claude/skills/goframe-v2/SKILL.md"
else
  printf 'goframe-v2 skill missing at ~/.claude/skills/goframe-v2/SKILL.md\n'
fi

printf '\nNext recommended commands:\n'
printf '  make init\n'
printf '  make dev\n'
printf '  openspec list\n'
printf '  cd hack/tests && pnpm test\n'

if [ "$check_rc" -eq 1 ]; then
  exit 1
fi
exit 0
