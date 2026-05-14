#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

cd "$ROOT_DIR/hack/tools/linactl"
go run . help >/dev/null
go run . status >/dev/null
go run . wasm dry-run=true >/dev/null

cd "$ROOT_DIR"
if [ -f make.cmd ]; then
  grep -Fq 'go run . %*' make.cmd
  if grep -Fq 'GOWORK=off' make.cmd; then
    echo "make.cmd must not force GOWORK=off"
    exit 1
  fi
else
  echo "make.cmd is missing"
  exit 1
fi

echo "linactl cross-platform smoke checks passed"
