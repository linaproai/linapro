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

is_stable_semver() {
  local value="${1#v}"
  [[ "$value" =~ ^[0-9]+[.][0-9]+[.][0-9]+$ ]]
}

semver_gt() {
  local left="${1#v}"
  local right="${2#v}"
  local left_major left_minor left_patch right_major right_minor right_patch

  IFS=. read -r left_major left_minor left_patch <<EOF
$left
EOF
  IFS=. read -r right_major right_minor right_patch <<EOF
$right
EOF

  if [ "$left_major" -gt "$right_major" ]; then
    return 0
  fi
  if [ "$left_major" -lt "$right_major" ]; then
    return 1
  fi
  if [ "$left_minor" -gt "$right_minor" ]; then
    return 0
  fi
  if [ "$left_minor" -lt "$right_minor" ]; then
    return 1
  fi
  [ "$left_patch" -gt "$right_patch" ]
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
  local sql_dir="$REPO_ROOT/apps/lina-core/manifest/sql"
  if [ ! -d "$sql_dir" ]; then
    return 0
  fi
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
    | while IFS= read -r file; do
      case "$file" in
        apps/lina-core/manifest/sql/[0-9][0-9][0-9]-*.sql) ;;
        *) continue ;;
      esac
      number="$(printf '%s\n' "$file" | sed -E 's#^.*/([0-9]{3})-.*#\1#')"
      if [ "$number" -gt "$max_number" ]; then
        printf '%s\n' "$file"
      fi
    done
}

print_changelog_summary() {
  local target_ref="$1"
  if git -C "$REPO_ROOT" cat-file -e "$target_ref:CHANGELOG.md" 2>/dev/null; then
    git -C "$REPO_ROOT" show "$target_ref:CHANGELOG.md" |
      grep -n -E 'BREAKING|\*\*BREAKING\*\*|Tier 1|apps/lina-core/pkg/' | head -n 30 || true
    return 0
  fi
  if [ ! -f "$REPO_ROOT/CHANGELOG.md" ]; then
    printf 'No CHANGELOG.md found.\n'
    return 0
  fi
  grep -n -E 'BREAKING|\*\*BREAKING\*\*|Tier 1|apps/lina-core/pkg/' "$REPO_ROOT/CHANGELOG.md" | head -n 30 || true
}

print_openspec_breaking_summary() {
  local target_ref="$1"
  if git -C "$REPO_ROOT" cat-file -e "$target_ref:openspec/changes/archive" 2>/dev/null; then
    git -C "$REPO_ROOT" grep -n -E '\*\*BREAKING\*\*|BREAKING|Tier 1' "$target_ref" -- 'openspec/changes/archive/*/proposal.md' 2>/dev/null |
      sed -E "s#^${target_ref}:##" |
      head -n 50 || true
    return 0
  fi
  if [ ! -d "$REPO_ROOT/openspec/changes/archive" ]; then
    printf 'No OpenSpec archive directory found.\n'
    return 0
  fi
  grep -R -n -E '\*\*BREAKING\*\*|BREAKING|Tier 1' "$REPO_ROOT/openspec/changes/archive"/*/proposal.md 2>/dev/null | head -n 50 || true
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
  if [ -z "$baseline" ]; then
    printf 'ERR_METADATA_VERSION_MISSING metadata=%s\n' "$METADATA_FILE" >&2
    exit 1
  fi
  if ! is_stable_semver "$baseline"; then
    printf 'ERR_BASELINE_VERSION_INVALID baseline=%s\n' "$baseline" >&2
    exit 1
  fi
  if ! is_stable_semver "$TARGET_VERSION"; then
    printf 'ERR_TARGET_VERSION_INVALID target=%s\n' "$TARGET_VERSION" >&2
    exit 1
  fi
  if ! semver_gt "$TARGET_VERSION" "$baseline"; then
    printf 'ERR_TARGET_NOT_GREATER baseline=%s target=%s\n' "$baseline" "$TARGET_VERSION" >&2
    exit 1
  fi

  upstream="$(select_upstream_remote)"
  if ! git -C "$REPO_ROOT" fetch --quiet "$upstream" '+refs/tags/*:refs/tags/*'; then
    printf 'ERR_FETCH_TAGS_FAILED upstream=%s\n' "$upstream" >&2
    exit 1
  fi

  baseline_ref="refs/tags/$baseline"
  target_ref="refs/tags/$TARGET_VERSION"
  if ! baseline_commit="$(git -C "$REPO_ROOT" rev-parse "${baseline_ref}^{commit}" 2>/dev/null)"; then
    printf 'ERR_BASELINE_TAG_NOT_FOUND baseline=%s upstream=%s\n' "$baseline" "$upstream" >&2
    exit 1
  fi
  if ! target_commit="$(git -C "$REPO_ROOT" rev-parse "${target_ref}^{commit}" 2>/dev/null)"; then
    printf 'ERR_TARGET_TAG_NOT_FOUND target=%s upstream=%s\n' "$TARGET_VERSION" "$upstream" >&2
    exit 1
  fi
  if ! git -C "$REPO_ROOT" merge-base --is-ancestor "$baseline_commit" "$target_commit"; then
    printf 'ERR_TARGET_NOT_DESCENDANT baseline=%s target=%s\n' "$baseline" "$TARGET_VERSION" >&2
    exit 1
  fi

  commits_ahead="$(git -C "$REPO_ROOT" rev-list --count "${baseline_commit}..${target_commit}")"
  max_sql="$(current_max_sql_number)"
  max_sql="${max_sql:-000}"

  printf '# LinaPro Upgrade Plan\n\n'
  printf -- "- Baseline version: \`%s\`\n" "$baseline"
  printf -- "- Target version: \`%s\`\n" "$TARGET_VERSION"
  printf -- "- Upstream remote: \`%s\`\n" "$upstream"
  printf -- "- Commits ahead: \`%s\`\n\n" "$commits_ahead"

  printf '## Changelog Highlights\n\n'
  print_changelog_summary "$target_ref"
  printf '\n## OpenSpec Breaking Highlights\n\n'
  print_openspec_breaking_summary "$target_ref"
  printf '\n## Changed Files by Tier\n\n'
  classify_changed_files "$baseline_commit" "$target_commit"
  printf '\n## New Host SQL Files\n\n'
  new_sql_files "$target_commit" "$max_sql"
}

main "$@"
