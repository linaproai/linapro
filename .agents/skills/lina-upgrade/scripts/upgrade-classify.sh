#!/usr/bin/env bash
# Classifies one repository path into a LinaPro upgrade stability tier.

set -euo pipefail

path="${1:-}"
if [ -z "$path" ]; then
  printf 'unknown\n'
  exit 0
fi

case "$path" in
  apps/lina-core/internal/dao/*|\
  apps/lina-core/internal/model/do/*|\
  apps/lina-core/internal/model/entity/*|\
  apps/lina-core/internal/controller/*|\
  apps/lina-plugins/*/backend/internal/dao/*|\
  apps/lina-plugins/*/backend/internal/model/do/*|\
  apps/lina-plugins/*/backend/internal/model/entity/*|\
  apps/lina-plugins/*/backend/internal/controller/*)
    printf 'tier3\n'
    exit 0
    ;;
  apps/lina-core/pkg/bizerr/*|\
  apps/lina-core/pkg/logger/*|\
  apps/lina-core/pkg/contract/*|\
  apps/lina-core/pkg/pluginbridge/*|\
  apps/lina-core/pkg/plugincontroller/*|\
  apps/lina-core/pkg/plugindb/*|\
  apps/lina-core/pkg/pluginfs/*|\
  apps/lina-core/pkg/pluginhost/*|\
  apps/lina-core/pkg/pluginservice/*|\
  apps/lina-core/pkg/sourceupgrade/contract/*|\
  apps/lina-plugins/*)
    printf 'tier1\n'
    exit 0
    ;;
  apps/lina-core/internal/*|\
  apps/lina-vben/apps/web-antd/src/*|\
  apps/lina-core/manifest/config/*.yaml|\
  apps/lina-core/*|\
  apps/lina-vben/*)
    printf 'tier2\n'
    exit 0
    ;;
  *)
    printf 'unknown\n'
    exit 0
    ;;
esac
