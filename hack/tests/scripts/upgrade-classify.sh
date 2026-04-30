#!/usr/bin/env bash
# Unit-tests lina-upgrade path tier classification.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLASSIFIER="$REPO_ROOT/.claude/skills/lina-upgrade/scripts/upgrade-classify.sh"

assert_tier() {
  local path="$1"
  local expected="$2"
  local actual
  actual="$(bash "$CLASSIFIER" "$path")"
  if [ "$actual" != "$expected" ]; then
    printf 'expected %s for %s, got %s\n' "$expected" "$path" "$actual" >&2
    exit 1
  fi
}

assert_tier "apps/lina-core/pkg/bizerr/code.go" tier1
assert_tier "apps/lina-core/pkg/pluginservice/auth/service.go" tier1
assert_tier "apps/lina-plugins/plugin-demo-source/backend/plugin.go" tier1
assert_tier "apps/lina-core/internal/service/auth/auth.go" tier2
assert_tier "apps/lina-vben/apps/web-antd/src/views/dashboard/index.vue" tier2
assert_tier "apps/lina-core/manifest/config/config.yaml" tier2
assert_tier "apps/lina-core/internal/dao/sys_user.go" tier3
assert_tier "apps/lina-core/internal/model/entity/sys_user.go" tier3
assert_tier "apps/lina-plugins/org-center/backend/internal/model/do/org_dept.go" tier3
assert_tier "docs/notes.md" unknown

printf 'PASS upgrade-classify\n'
