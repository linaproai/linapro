#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh [--run-id <id>] [--run-dir <dir>]

Options:
  --run-id   Audit run id. Defaults to YYYYMMDD-HHMMSS.
  --run-dir  Audit run directory. Defaults to temp/lina-perf-audit/<run-id>.
EOF
}

die() {
  printf 'setup-audit-env: %s\n' "$*" >&2
  exit 1
}

require_repo_root() {
  if [[ ! -f "Makefile" || ! -f "apps/lina-core/main.go" || ! -f "apps/lina-core/manifest/config/config.yaml" ]]; then
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

read_logger_config_json() {
  local config_file="$1"
  python3 - "$config_file" <<'PY'
import json
import re
import sys

path = sys.argv[1]
lines = open(path, encoding="utf-8").read().splitlines()
in_logger = False
values = {"path": "", "file": ""}

def parse_scalar(raw):
    value = raw.split("#", 1)[0].strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
        return value[1:-1]
    return value

for line in lines:
    if re.match(r"^logger:\s*(?:#.*)?$", line):
        in_logger = True
        continue
    if in_logger and re.match(r"^[A-Za-z0-9_-]+:\s*", line):
        break
    if not in_logger:
        continue
    match = re.match(r"^\s{2}(path|file):\s*(.*)$", line)
    if match:
        values[match.group(1)] = parse_scalar(match.group(2))

print(json.dumps({"logger": values}, ensure_ascii=False, indent=2))
PY
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

extract_access_token() {
  local body_file="$1"
  python3 - "$body_file" <<'PY'
import json
import sys

payload = json.load(open(sys.argv[1], encoding="utf-8"))
data = payload.get("data")
if not isinstance(data, dict):
    raise SystemExit("response data is not an object")
token = data.get("accessToken")
if not token:
    raise SystemExit("response data.accessToken is empty")
print(token)
PY
}

extract_trace_id() {
  local headers_file="$1"
  python3 - "$headers_file" <<'PY'
import sys

for line in open(sys.argv[1], encoding="utf-8", errors="replace"):
    if line.lower().startswith("trace-id:"):
        print(line.split(":", 1)[1].strip())
        break
PY
}

start_backend() {
  local root_dir="$1"
  local binary="$2"
  local run_dir="$3"
  local pid_file="$4"
  local config_path="$5"
  python3 - "$root_dir" "$binary" "$run_dir" "$pid_file" "$config_path" <<'PY'
import os
import subprocess
import sys

root_dir, binary, run_dir, pid_file, config_path = sys.argv[1:6]
log_path = os.path.join(run_dir, "backend-nohup.log")
env = os.environ.copy()
env["GF_GCFG_PATH"] = config_path
env["GF_GCFG_FILE"] = "config.yaml"

log = open(log_path, "ab", buffering=0)
try:
    process = subprocess.Popen(
        [binary],
        cwd=root_dir,
        stdin=subprocess.DEVNULL,
        stdout=log,
        stderr=subprocess.STDOUT,
        env=env,
        start_new_session=True,
    )
finally:
    log.close()

with open(pid_file, "w", encoding="utf-8") as f:
    f.write(str(process.pid) + "\n")
print(process.pid)
PY
}

curl_json() {
  local method="$1"
  local url="$2"
  local body="$3"
  local out_body="$4"
  local out_headers="$5"
  local status
  local curl_args=(-sS -D "$out_headers" -o "$out_body" -w '%{http_code}' -X "$method")

  if [[ -n "$body" ]]; then
    curl_args+=(-H 'Content-Type: application/json' --data "$body")
  fi

  if ! status="$(curl "${curl_args[@]}" "$url")"; then
    printf 'curl failed for %s %s\n' "$method" "$url" >&2
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

wait_for_health() {
  local url="$1"
  local body_file="$RUN_DIR/health-ready.json"
  local headers_file="$RUN_DIR/health-ready.headers"

  for _ in $(seq 1 60); do
    if curl_json GET "$url" "" "$body_file" "$headers_file" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  die "backend did not become healthy at ${url}; see ${RUN_DIR}/backend-nohup.log"
}

RUN_ID=""
RUN_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-id)
      [[ $# -ge 2 ]] || die "--run-id requires a value"
      RUN_ID="$2"
      shift 2
      ;;
    --run-id=*)
      RUN_ID="${1#--run-id=}"
      shift
      ;;
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

if [[ -z "$RUN_ID" ]]; then
  if [[ -n "$RUN_DIR" ]]; then
    RUN_ID="$(basename "$RUN_DIR")"
  else
    RUN_ID="$(date +%Y%m%d-%H%M%S)"
  fi
fi

if [[ -z "$RUN_DIR" ]]; then
  RUN_DIR="temp/lina-perf-audit/${RUN_ID}"
fi
RUN_DIR="${RUN_DIR%/}"

CONFIG_FILE="apps/lina-core/manifest/config/config.yaml"
BACKUP_FILE="$RUN_DIR/logger-backup.json"
TOKEN_FILE="$RUN_DIR/token.txt"
TRACE_CHECK_FILE="$RUN_DIR/trace-id-check.txt"
BACKEND_BINARY="temp/bin/lina-perf-audit-backend"
BACKEND_PID_FILE="temp/pids/backend.pid"
LOGGER_PATH="${RUN_DIR}/"
LOGGER_FILE="server.log"
HEALTH_URL="http://127.0.0.1:8080/api/v1/health"
LOGIN_URL="http://127.0.0.1:8080/api/v1/auth/login"

mkdir -p "$RUN_DIR" temp/bin temp/pids

make stop

if [[ ! -f "$BACKUP_FILE" ]]; then
  read_logger_config_json "$CONFIG_FILE" > "$BACKUP_FILE"
fi

patch_logger_config "$CONFIG_FILE" "$LOGGER_PATH" "$LOGGER_FILE"

make wasm
./hack/scripts/prepare-packed-assets.sh
go build -o "$BACKEND_BINARY" ./apps/lina-core

rm -f "$BACKEND_PID_FILE"
ROOT_DIR="$(pwd)"
BACKEND_BINARY_ABS="${ROOT_DIR}/${BACKEND_BINARY}"
CONFIG_PATH_ABS="${ROOT_DIR}/apps/lina-core/manifest/config"
start_backend "$ROOT_DIR" "$BACKEND_BINARY_ABS" "$RUN_DIR" "$BACKEND_PID_FILE" "$CONFIG_PATH_ABS" >/dev/null

wait_for_health "$HEALTH_URL"

LOGIN_BODY="$(printf '{"username":%s,"password":%s}' "$(json_string admin)" "$(json_string admin123)")"
if ! curl_json POST "$LOGIN_URL" "$LOGIN_BODY" "$RUN_DIR/login-response.json" "$RUN_DIR/login-response.headers"; then
  die "admin login failed; see $RUN_DIR/login-response.json"
fi
extract_access_token "$RUN_DIR/login-response.json" > "$TOKEN_FILE"
chmod 0600 "$TOKEN_FILE"

if ! curl_json GET "$HEALTH_URL" "" "$RUN_DIR/trace-id-check-response.json" "$RUN_DIR/trace-id-check.headers"; then
  die "Trace-ID health check failed; see $RUN_DIR/trace-id-check-response.json"
fi
TRACE_ID="$(extract_trace_id "$RUN_DIR/trace-id-check.headers")"
if [[ -z "$TRACE_ID" ]]; then
  die "Trace-ID response header is empty for ${HEALTH_URL}"
fi

SERVER_LOG="${RUN_DIR}/${LOGGER_FILE}"
TRACE_FOUND="false"
for _ in $(seq 1 10); do
  if [[ -f "$SERVER_LOG" ]] && grep -F "$TRACE_ID" "$SERVER_LOG" >/dev/null 2>&1; then
    TRACE_FOUND="true"
    break
  fi
  sleep 1
done

{
  printf 'run_id=%s\n' "$RUN_ID"
  printf 'run_dir=%s\n' "$RUN_DIR"
  printf 'health_url=%s\n' "$HEALTH_URL"
  printf 'trace_id=%s\n' "$TRACE_ID"
  printf 'server_log=%s\n' "$SERVER_LOG"
  printf 'trace_id_found_in_server_log=%s\n' "$TRACE_FOUND"
} > "$TRACE_CHECK_FILE"

if [[ "$TRACE_FOUND" != "true" ]]; then
  die "Trace-ID ${TRACE_ID} was not found in ${SERVER_LOG}"
fi

printf 'Audit backend ready.\n'
printf 'run_id=%s\n' "$RUN_ID"
printf 'run_dir=%s\n' "$RUN_DIR"
printf 'token_file=%s\n' "$TOKEN_FILE"
