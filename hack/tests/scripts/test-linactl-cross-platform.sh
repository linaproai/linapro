#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

cd "$ROOT_DIR"

go run ./hack/tools/linactl help >/dev/null
go run ./hack/tools/linactl status >/dev/null
go run ./hack/tools/linactl wasm dry-run=true >/dev/null

if [ -f make.cmd ]; then
  grep -Fq "go run ./hack/tools/linactl %*" make.cmd
else
  echo "make.cmd is missing"
  exit 1
fi

echo "linactl cross-platform smoke checks passed"
