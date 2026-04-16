#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SOURCE_DIR="$ROOT_DIR/apps/lina-core/manifest"
TARGET_DIR="$ROOT_DIR/apps/lina-core/internal/packed/manifest"

# Rebuild the packed manifest workspace from scratch so each build embeds the
# current distributable assets only, without leaking stale files from prior runs.
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR/config" "$TARGET_DIR/sql"

# Only copy manifest assets that are meant to ship inside the host binary.
# The local developer override config.yaml must stay out of the embedded bundle.
cp "$SOURCE_DIR/config/config.template.yaml" "$TARGET_DIR/config/"
cp "$SOURCE_DIR/config/components.yaml" "$TARGET_DIR/config/"
cp -R "$SOURCE_DIR/sql/." "$TARGET_DIR/sql/"

touch "$TARGET_DIR/.gitkeep"
