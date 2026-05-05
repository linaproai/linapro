#!/usr/bin/env bash
# Validates the declared LinaPro framework baseline before an upgrade merge.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${LINAPRO_REPO_ROOT:-$(cd "$SCRIPT_DIR/../../../.." && pwd)}"
METADATA_FILE="$REPO_ROOT/apps/lina-core/manifest/config/metadata.yaml"

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

recent_stable_tags() {
  git -C "$REPO_ROOT" tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 3 | paste -sd ',' -
}

count_sql_files_at_ref() {
  local ref="$1"
  git -C "$REPO_ROOT" ls-tree -r --name-only "$ref" -- apps/lina-core/manifest/sql 2>/dev/null |
    awk '/\.sql$/ { count++ } END { print count + 0 }'
}

count_sql_files_in_worktree() {
  local sql_dir="$REPO_ROOT/apps/lina-core/manifest/sql"
  if [ ! -d "$sql_dir" ]; then
    printf '0\n'
    return 0
  fi
  find "$sql_dir" -maxdepth 1 -type f -name '*.sql' | wc -l | tr -d ' '
}

main() {
  local declared upstream tag_ref tag_commit head_commit commits_ahead core_changed sql_at_tag sql_at_head candidates

  if [ ! -f "$METADATA_FILE" ]; then
    printf 'ERR_METADATA_NOT_FOUND metadata=%s\n' "$METADATA_FILE"
    exit 1
  fi

  declared="$(read_declared_version)"
  if [ -z "$declared" ]; then
    printf 'ERR_METADATA_VERSION_MISSING metadata=%s\n' "$METADATA_FILE"
    exit 1
  fi

  upstream="$(select_upstream_remote)"
  if ! git -C "$REPO_ROOT" fetch --quiet "$upstream" '+refs/tags/*:refs/tags/*'; then
    printf 'ERR_FETCH_TAGS_FAILED upstream=%s\n' "$upstream"
    exit 1
  fi

  tag_ref="refs/tags/$declared"
  if ! git -C "$REPO_ROOT" rev-parse -q --verify "${tag_ref}^{commit}" >/dev/null; then
    candidates="$(recent_stable_tags)"
    printf 'ERR_TAG_NOT_FOUND declared=%s upstream=%s recent_stable_tags=%s\n' "$declared" "$upstream" "${candidates:-none}"
    exit 1
  fi

  tag_commit="$(git -C "$REPO_ROOT" rev-parse "${tag_ref}^{commit}")"
  head_commit="$(git -C "$REPO_ROOT" rev-parse HEAD)"

  if ! git -C "$REPO_ROOT" merge-base --is-ancestor "$tag_commit" HEAD; then
    printf 'ERR_HEAD_NOT_DESCENDANT declared=%s tag_commit=%s head_commit=%s\n' "$declared" "$tag_commit" "$head_commit"
    exit 1
  fi

  commits_ahead="$(git -C "$REPO_ROOT" rev-list --count "${tag_commit}..HEAD")"
  core_changed="$(git -C "$REPO_ROOT" diff --name-only "${tag_commit}...HEAD" -- apps/lina-core apps/lina-vben | wc -l | tr -d ' ')"
  sql_at_tag="$(count_sql_files_at_ref "$tag_commit")"
  sql_at_head="$(count_sql_files_in_worktree)"

  if [ "$sql_at_head" -lt "$sql_at_tag" ]; then
    printf 'WARN_SQL_COUNT_DECREASED sql_at_tag=%s sql_at_head=%s\n' "$sql_at_tag" "$sql_at_head"
  fi

  printf 'OK_BASELINE_CONFIRMED declared=%s upstream=%s tag_commit=%s head_commit=%s commits_ahead=%s core_changed=%s sql_at_tag=%s sql_at_head=%s\n' \
    "$declared" "$upstream" "$tag_commit" "$head_commit" "$commits_ahead" "$core_changed" "$sql_at_tag" "$sql_at_head"
}

main "$@"
