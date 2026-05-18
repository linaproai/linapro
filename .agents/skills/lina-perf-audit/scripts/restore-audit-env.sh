#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/restore-audit-env.sh --run-dir <dir>

Options:
  --run-dir  Audit run directory containing logger-backup.json.
EOF
}

die() {
  printf 'restore-audit-env: %s\n' "$*" >&2
  exit 1
}

require_repo_root() {
  if [[ ! -f "Makefile" || ! -f "apps/lina-core/main.go" || ! -f "apps/lina-core/manifest/config/config.yaml" ]]; then
    die "run this script from the LinaPro repository root"
  fi
}

patch_logger_config() {
  local config_file="$1"
  local logger_path="$2"
  local logger_file="$3"
  python3 - "$config_file" "$logger_path" "$logger_file" <<'PY'
import json
import re
import sys

config_file, logger_path, logger_file = sys.argv[1:4]
lines = open(config_file, encoding="utf-8").read().splitlines(keepends=True)
out = []
in_logger = False
seen_logger = False
seen_path = False
seen_file = False

def yaml_string(value):
    return json.dumps(value, ensure_ascii=False)

for line in lines:
    if re.match(r"^logger:\s*(?:#.*)?$", line):
        in_logger = True
        seen_logger = True
        out.append(line)
        continue
    if in_logger and re.match(r"^[A-Za-z0-9_-]+:\s*", line):
        if not seen_path:
            out.append(f"  path: {yaml_string(logger_path)}\n")
        if not seen_file:
            out.append(f"  file: {yaml_string(logger_file)}\n")
        in_logger = False
    if in_logger:
        if re.match(r"^\s{2}path:\s*", line):
            out.append(f"  path: {yaml_string(logger_path)}\n")
            seen_path = True
            continue
        if re.match(r"^\s{2}file:\s*", line):
            out.append(f"  file: {yaml_string(logger_file)}\n")
            seen_file = True
            continue
    out.append(line)

if in_logger:
    if not seen_path:
        out.append(f"  path: {yaml_string(logger_path)}\n")
    if not seen_file:
        out.append(f"  file: {yaml_string(logger_file)}\n")

if not seen_logger:
    out.append("\nlogger:\n")
    out.append(f"  path: {yaml_string(logger_path)}\n")
    out.append(f"  file: {yaml_string(logger_file)}\n")

with open(config_file, "w", encoding="utf-8", newline="") as f:
    f.writelines(out)
PY
}

read_backup_value() {
  local backup_file="$1"
  local key="$2"
  python3 - "$backup_file" "$key" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
value = payload.get("logger", {}).get(sys.argv[2], "")
if value is None:
    value = ""
print(value)
PY
}

RUN_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir)
      [[ $# -ge 2 ]] || die "--run-dir requires a value"
      RUN_DIR="${2%/}"
      shift 2
      ;;
    --run-dir=*)
      RUN_DIR="${1#--run-dir=}"
      RUN_DIR="${RUN_DIR%/}"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

require_repo_root

if [[ -z "$RUN_DIR" ]]; then
  die "--run-dir is required"
fi

CONFIG_FILE="apps/lina-core/manifest/config/config.yaml"
BACKUP_FILE="$RUN_DIR/logger-backup.json"

if [[ -f "$BACKUP_FILE" ]]; then
  LOGGER_PATH="$(read_backup_value "$BACKUP_FILE" path)"
  LOGGER_FILE="$(read_backup_value "$BACKUP_FILE" file)"
  patch_logger_config "$CONFIG_FILE" "$LOGGER_PATH" "$LOGGER_FILE"
  printf 'Restored logger.path and logger.file from %s\n' "$BACKUP_FILE"
else
  printf 'No logger backup found at %s; skipping config restore\n' "$BACKUP_FILE" >&2
fi

make stop
