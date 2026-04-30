#!/usr/bin/env bash
# Produces a structured LinaPro framework upgrade plan for AI review.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${LINAPRO_REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
METADATA_FILE="$REPO_ROOT/apps/lina-core/manifest/config/metadata.yaml"
TARGET_VERSION="${1:-${LINAPRO_TARGET_VERSION:-}}"

read_declared_version() {
  awk '
    /^framework:/ { in_framework=1; next }
    in_framework && /^[^[:space:]]/ { in_framework=0 }
    in_framework && $1 == "version:" {
      value=$2
      gsub(/"/, "", value)
      print value
      exit
    }
  ' "$METADATA_FILE"
}

select_upstream_remote() {
  local official_url origin_url
  official_url="https://github.com/linaproai/linapro.git"
  if git -C "$REPO_ROOT" remote get-url upstream >/dev/null 2>&1; then
    printf 'upstream\n'
    return 0
  fi
  origin_url="$(git -C "$REPO_ROOT" remote get-url origin 2>/dev/null || true)"
  if [ -n "$origin_url" ] && printf '%s\n' "$origin_url" | grep -E 'github\.com[:/]linaproai/linapro(\.git)?$' >/dev/null 2>&1; then
    printf 'origin\n'
    return 0
  fi
  printf '%s\n' "$official_url"
}

classify_changed_files() {
  local baseline_ref="$1"
  local target_ref="$2"
  local file tier
  git -C "$REPO_ROOT" diff --name-only "${baseline_ref}...${target_ref}" -- apps/lina-core apps/lina-vben apps/lina-plugins | while IFS= read -r file; do
    [ -z "$file" ] && continue
    tier="$("$SCRIPT_DIR/upgrade-classify.sh" "$file")"
    printf '%s %s\n' "$tier" "$file"
  done | sort
}

current_max_sql_number() {
  find "$REPO_ROOT/apps/lina-core/manifest/sql" -maxdepth 1 -type f -name '[0-9][0-9][0-9]-*.sql' \
    | sed -E 's#^.*/([0-9]{3})-.*#\1#' \
    | sort -n \
    | tail -n 1
}

new_sql_files() {
  local target_ref="$1"
  local max_number="$2"
  local file number
  git -C "$REPO_ROOT" ls-tree -r --name-only "$target_ref" -- apps/lina-core/manifest/sql \
    | grep -E '/[0-9]{3}-.*\.sql$' \
    | while IFS= read -r file; do
      number="$(printf '%s\n' "$file" | sed -E 's#^.*/([0-9]{3})-.*#\1#')"
      if [ "$number" -gt "$max_number" ]; then
        printf '%s\n' "$file"
      fi
    done
}

print_changelog_summary() {
  local changelog="$REPO_ROOT/CHANGELOG.md"
  if [ ! -f "$changelog" ]; then
    printf 'No CHANGELOG.md found.\n'
    return 0
  fi
  grep -n -E 'BREAKING|\*\*BREAKING\*\*|Tier 1|apps/lina-core/pkg/' "$changelog" | head -n 30 || true
}

print_openspec_breaking_summary() {
  local archive_root="$REPO_ROOT/openspec/changes/archive"
  if [ ! -d "$archive_root" ]; then
    printf 'No OpenSpec archive directory found.\n'
    return 0
  fi
  grep -R -n -E '\*\*BREAKING\*\*|BREAKING|Tier 1' "$archive_root"/*/proposal.md 2>/dev/null | head -n 50 || true
}

main() {
  local baseline upstream baseline_ref target_ref baseline_commit target_commit commits_ahead max_sql

  if [ -z "$TARGET_VERSION" ]; then
    printf 'usage: %s <target-version>\n' "$0" >&2
    exit 1
  fi
  if [ ! -f "$METADATA_FILE" ]; then
    printf 'metadata file not found: %s\n' "$METADATA_FILE" >&2
    exit 1
  fi

  baseline="$(read_declared_version)"
  upstream="$(select_upstream_remote)"
  git -C "$REPO_ROOT" fetch --quiet "$upstream" '+refs/tags/*:refs/tags/*'

  baseline_ref="refs/tags/$baseline"
  target_ref="refs/tags/$TARGET_VERSION"
  baseline_commit="$(git -C "$REPO_ROOT" rev-parse "${baseline_ref}^{commit}")"
  target_commit="$(git -C "$REPO_ROOT" rev-parse "${target_ref}^{commit}")"
  commits_ahead="$(git -C "$REPO_ROOT" rev-list --count "${baseline_commit}..${target_commit}")"
  max_sql="$(current_max_sql_number)"
  max_sql="${max_sql:-000}"

  printf '# LinaPro Upgrade Plan\n\n'
  printf -- "- Baseline version: \`%s\`\n" "$baseline"
  printf -- "- Target version: \`%s\`\n" "$TARGET_VERSION"
  printf -- "- Upstream remote: \`%s\`\n" "$upstream"
  printf -- "- Commits ahead: \`%s\`\n\n" "$commits_ahead"

  printf '## Changelog Highlights\n\n'
  print_changelog_summary
  printf '\n## OpenSpec Breaking Highlights\n\n'
  print_openspec_breaking_summary
  printf '\n## Changed Files by Tier\n\n'
  classify_changed_files "$baseline_commit" "$target_commit"
  printf '\n## New Host SQL Files\n\n'
  new_sql_files "$target_commit" "$max_sql"
}

main "$@"
