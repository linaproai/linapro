#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir <dir>

Options:
  --run-dir  Audit run directory containing token.txt.
EOF
}

die() {
  printf 'prepare-builtin-plugins: %s\n' "$*" >&2
  exit 1
}

require_repo_root() {
  if [[ ! -f "Makefile" || ! -d "apps/lina-plugins" || ! -f "apps/lina-core/api/plugin/v1/plugin_sync.go" ]]; then
    die "run this script from the LinaPro repository root"
  fi
}

json_string() {
  python3 - "$1" <<'PY'
import json
import sys

print(json.dumps(sys.argv[1], ensure_ascii=False))
PY
}

bool_json() {
  if [[ "$1" == "true" ]]; then
    printf 'true'
  else
    printf 'false'
  fi
}

assert_envelope_success() {
  local body_file="$1"
  python3 - "$body_file" <<'PY'
import json
import sys

path = sys.argv[1]
try:
    payload = json.load(open(path, encoding="utf-8"))
except Exception as exc:
    raise SystemExit(f"response is not valid JSON: {exc}")

code = payload.get("code")
if code not in (0, "0"):
    message = payload.get("message", "")
    error_code = payload.get("errorCode", "")
    raise SystemExit(f"response code is {code!r}; errorCode={error_code!r}; message={message!r}")
PY
}

api_request() {
  local method="$1"
  local url="$2"
  local body="$3"
  local out_body="$4"
  local out_headers="$5"
  local status
  local curl_args=(-sS -D "$out_headers" -o "$out_body" -w '%{http_code}' -X "$method" -H "Authorization: Bearer ${TOKEN}")

  if [[ -n "$body" ]]; then
    curl_args+=(-H 'Content-Type: application/json' --data "$body")
  fi

  if ! status="$(curl "${curl_args[@]}" "$url")"; then
    return 10
  fi
  if [[ ! "$status" =~ ^2[0-9][0-9]$ ]]; then
    printf 'HTTP %s for %s %s: %s\n' "$status" "$method" "$url" "$(tr '\n' ' ' < "$out_body")" >&2
    return 11
  fi
  if ! assert_envelope_success "$out_body"; then
    printf 'LinaPro envelope failure for %s %s: %s\n' "$method" "$url" "$(tr '\n' ' ' < "$out_body")" >&2
    return 12
  fi
}

discover_plugins() {
  python3 <<'PY'
import glob
import os
import re
def parse_id(path):
    for line in open(path, encoding="utf-8"):
        match = re.match(r"^id\s*:\s*(.+?)\s*$", line)
        if not match:
            continue
        value = match.group(1).split("#", 1)[0].strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
            value = value[1:-1]
        return value.strip()
    return ""

for yaml_path in sorted(glob.glob("apps/lina-plugins/*/plugin.yaml")):
    plugin_dir = os.path.dirname(yaml_path)
    plugin_id = parse_id(yaml_path)
    if not plugin_id:
        raise SystemExit(f"missing top-level id in {yaml_path}")
    mock_dir = os.path.join(plugin_dir, "manifest", "sql", "mock-data")
    has_mock = os.path.isdir(mock_dir)
    api_dir = os.path.join(plugin_dir, "backend", "api")
    has_backend_api = False
    for root, _, files in os.walk(api_dir) if os.path.isdir(api_dir) else []:
        if any(name.endswith(".go") for name in files):
            has_backend_api = True
            break
    print("\t".join([
        plugin_id,
        yaml_path,
        "true" if has_mock else "false",
        "true" if has_backend_api else "false",
    ]))
PY
}

init_plugins_json() {
  local plugins_json="$1"
  python3 - "$plugins_json" <<'PY'
import json
import sys

payload = {
    "sync": {"status": "pending"},
    "failed": False,
    "plugins": [],
}
with open(sys.argv[1], "w", encoding="utf-8") as f:
    json.dump(payload, f, ensure_ascii=False, indent=2)
    f.write("\n")
PY
}

set_sync_status() {
  local status="$1"
  local message="${2:-}"
  python3 - "$PLUGINS_JSON" "$status" "$message" <<'PY'
import json
import sys

path, status, message = sys.argv[1:4]
payload = json.load(open(path, encoding="utf-8"))
payload["sync"] = {"status": status}
if message:
    payload["sync"]["error"] = message
    payload["failed"] = True
with open(path, "w", encoding="utf-8") as f:
    json.dump(payload, f, ensure_ascii=False, indent=2)
    f.write("\n")
PY
}

record_plugin() {
  local plugin_id="$1"
  local manifest_path="$2"
  local has_mock_data="$3"
  local has_backend_api="$4"
  local status="$5"
  local step="$6"
  local error_message="${7:-}"
  python3 - "$PLUGINS_JSON" "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "$status" "$step" "$error_message" <<'PY'
import json
import sys

(
    path,
    plugin_id,
    manifest_path,
    has_mock_data,
    has_backend_api,
    status,
    step,
    error_message,
) = sys.argv[1:9]
payload = json.load(open(path, encoding="utf-8"))
plugins = payload.setdefault("plugins", [])
record = None
for item in plugins:
    if item.get("id") == plugin_id:
        record = item
        break
if record is None:
    record = {"id": plugin_id}
    plugins.append(record)

record.update({
    "manifestPath": manifest_path,
    "hasMockData": has_mock_data == "true",
    "hasBackendApi": has_backend_api == "true",
    "status": status,
    "step": step,
})
if error_message:
    record["error"] = error_message
    payload["failed"] = True
elif "error" in record:
    del record["error"]

with open(path, "w", encoding="utf-8") as f:
    json.dump(payload, f, ensure_ascii=False, indent=2)
    f.write("\n")
PY
}

fail_plugin() {
  local plugin_id="$1"
  local manifest_path="$2"
  local has_mock_data="$3"
  local has_backend_api="$4"
  local step="$5"
  local message="$6"
  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "failed" "$step" "$message"
  die "$message"
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

TOKEN_FILE="$RUN_DIR/token.txt"
PLUGINS_JSON="$RUN_DIR/plugins.json"
API_BASE="http://127.0.0.1:8080/api/v1"

[[ -f "$TOKEN_FILE" ]] || die "token file not found: $TOKEN_FILE"
TOKEN="$(tr -d '\r\n' < "$TOKEN_FILE")"
[[ -n "$TOKEN" ]] || die "token file is empty: $TOKEN_FILE"

mkdir -p "$RUN_DIR/plugin-api-responses"
init_plugins_json "$PLUGINS_JSON"
DISCOVERED_PLUGINS="$RUN_DIR/discovered-plugins.tsv"

SYNC_BODY="$RUN_DIR/plugin-api-responses/sync.json"
SYNC_HEADERS="$RUN_DIR/plugin-api-responses/sync.headers"
if api_request POST "${API_BASE}/plugins/sync" "{}" "$SYNC_BODY" "$SYNC_HEADERS"; then
  set_sync_status "succeeded"
else
  set_sync_status "failed" "failed to sync source plugins"
  die "failed to sync source plugins; see $SYNC_BODY"
fi

discover_plugins > "$DISCOVERED_PLUGINS"

while IFS=$'\t' read -r plugin_id manifest_path has_mock_data has_backend_api; do
  [[ -n "$plugin_id" ]] || continue
  safe_id="${plugin_id//[^A-Za-z0-9_.-]/_}"
  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "pending" "discovered"

  install_mock_data="$(bool_json "$has_mock_data")"
  install_body="$(printf '{"id":%s,"installMockData":%s}' "$(json_string "$plugin_id")" "$install_mock_data")"
  install_response="$RUN_DIR/plugin-api-responses/${safe_id}-install.json"
  install_headers="$RUN_DIR/plugin-api-responses/${safe_id}-install.headers"
  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "running" "install"
  if ! api_request POST "${API_BASE}/plugins/${plugin_id}/install" "$install_body" "$install_response" "$install_headers"; then
    fail_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "install" "failed to install plugin ${plugin_id}; see ${install_response}"
  fi
  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "installed" "install"
done < "$DISCOVERED_PLUGINS"

sort_enable_order() {
  python3 - "$DISCOVERED_PLUGINS" <<'PY'
import sys

rows = []
for line in open(sys.argv[1], encoding="utf-8"):
    row = line.rstrip("\n").split("\t")
    if len(row) == 4 and row[0]:
        rows.append(row)

def priority(row):
    plugin_id, _, _, has_backend_api = row
    if plugin_id == "demo-control":
        return (2, plugin_id)
    if has_backend_api != "true":
        return (1, plugin_id)
    return (0, plugin_id)

for row in sorted(rows, key=priority):
    print("\t".join(row))
PY
}

while IFS=$'\t' read -r plugin_id manifest_path has_mock_data has_backend_api; do
  [[ -n "$plugin_id" ]] || continue
  safe_id="${plugin_id//[^A-Za-z0-9_.-]/_}"
  enable_body="$(printf '{"id":%s}' "$(json_string "$plugin_id")")"
  enable_response="$RUN_DIR/plugin-api-responses/${safe_id}-enable.json"
  enable_headers="$RUN_DIR/plugin-api-responses/${safe_id}-enable.headers"
  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "running" "enable"
  if ! api_request PUT "${API_BASE}/plugins/${plugin_id}/enable" "$enable_body" "$enable_response" "$enable_headers"; then
    fail_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "enable" "failed to enable plugin ${plugin_id}; see ${enable_response}"
  fi

  record_plugin "$plugin_id" "$manifest_path" "$has_mock_data" "$has_backend_api" "succeeded" "enabled"
done < <(sort_enable_order)

printf 'Prepared built-in plugins. Details: %s\n' "$PLUGINS_JSON"
