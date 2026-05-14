#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/lina-core"
CONFIG_PATH="$CORE_DIR/manifest/config/config.yaml"
CONFIG_TEMPLATE_PATH="$CORE_DIR/manifest/config/config.template.yaml"
BACKUP_DIR="$ROOT_DIR/temp/redis-cluster-smoke"
BACKUP_CONFIG_PATH="$BACKUP_DIR/config.yaml.backup"
BACKEND_BINARY="$ROOT_DIR/temp/bin/lina-redis-cluster-smoke"
BACKEND_LOG="$BACKUP_DIR/lina-core.log"
SMOKE_PORT="${LINAPRO_REDIS_CLUSTER_SMOKE_PORT:-18081}"
REDIS_ADDR="${LINAPRO_REDIS_CLUSTER_SMOKE_REDIS_ADDR:-127.0.0.1:6379}"
DB_LINK="${LINAPRO_REDIS_CLUSTER_SMOKE_DB_LINK:-}"
BACKEND_PID=""

# Restore the developer/CI config even when startup, curl assertions, or build
# steps fail. The script rewrites config.yaml to force a temporary cluster mode.
restore_config() {
  if [ -f "$BACKUP_CONFIG_PATH" ]; then
    mv "$BACKUP_CONFIG_PATH" "$CONFIG_PATH"
  fi
}

stop_backend() {
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi
}

cleanup() {
  stop_backend
  restore_config
}

trap cleanup EXIT INT TERM

write_cluster_config() {
  mkdir -p "$BACKUP_DIR"
  if [ -f "$CONFIG_PATH" ]; then
    cp "$CONFIG_PATH" "$BACKUP_CONFIG_PATH"
  else
    cp "$CONFIG_TEMPLATE_PATH" "$BACKUP_CONFIG_PATH"
  fi

  cp "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
  # Keep this smoke isolated from the normal dev backend and from nightly E2E.
  SMOKE_PORT="$SMOKE_PORT" perl -0pi -e 's/address:\s*":\d+"/q{address: ":}.$ENV{SMOKE_PORT}.q{"}/e' "$CONFIG_PATH"
  if [ -n "$DB_LINK" ]; then
    DB_LINK="$DB_LINK" perl -0pi -e 's/link:\s*"[^"]+"/q{link: "}.$ENV{DB_LINK}.q{"}/e' "$CONFIG_PATH"
  fi
  # The core behavior under test is the startup path that enables cluster mode,
  # creates a Redis coordination provider, and wires cache/session/lock services.
  perl -0pi -e 's/(cluster:\n(?:[ \t]*#.*\n)*[ \t]*)enabled:\s*(?:true|false)/${1}enabled: true/' "$CONFIG_PATH"
  REDIS_ADDR="$REDIS_ADDR" perl -0pi -e 's/(redis:\n(?:[ \t]*#.*\n)*[ \t]*address:\s*")[^"]+(")/$1.$ENV{REDIS_ADDR}.$2/e' "$CONFIG_PATH"
}

wait_for_cluster_health() {
  local elapsed=0
  local response=""
  while [ "$elapsed" -lt 90 ]; do
    if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
      echo "Redis cluster smoke startup failed; backend exited early"
      echo "Check log: $BACKEND_LOG"
      exit 1
    fi
    response="$(curl -fsS "http://127.0.0.1:$SMOKE_PORT/api/v1/health" 2>/dev/null || true)"
    # A one-node cluster should become master after Redis leader election
    # acquires the lease. Polling avoids racing the asynchronous election loop.
    if [ -n "$response" ] && HEALTH_RESPONSE="$response" python3 - <<'PY'
import json
import os
import sys

payload = json.loads(os.environ["HEALTH_RESPONSE"])
data = payload.get("data", payload)
if payload.get("code", 0) == 0 and data.get("status") == "ok" and data.get("mode") == "master":
    sys.exit(0)
sys.exit(1)
PY
    then
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  echo "Redis cluster smoke startup timed out waiting for master health mode"
  echo "Last health response: $response"
  echo "Check log: $BACKEND_LOG"
  exit 1
}

assert_login_and_sysinfo() {
  # Login proves the cluster session store can write Redis hot state and the
  # PostgreSQL online-session projection through the normal HTTP path.
  local login_response
  login_response="$(
    curl -fsS \
      -H "Content-Type: application/json" \
      -d '{"username":"admin","password":"admin123"}' \
      "http://127.0.0.1:$SMOKE_PORT/api/v1/auth/login"
  )"

  local token
  token="$(LOGIN_RESPONSE="$login_response" python3 - <<'PY'
import json
import os
import sys

payload = json.loads(os.environ["LOGIN_RESPONSE"])
data = payload.get("data", payload)
if payload.get("code", 0) != 0:
    raise SystemExit(f"login business code is not zero: {payload!r}")
token = data.get("accessToken", "")
if not token:
    raise SystemExit(f"login response did not contain accessToken: {payload!r}")
print(token)
PY
  )"

  local sysinfo_response
  sysinfo_response="$(
    curl -fsS \
      -H "Authorization: Bearer $token" \
      "http://127.0.0.1:$SMOKE_PORT/api/v1/system/info"
  )"

  # System info is the public diagnostics surface for Redis coordination. Keep
  # this assertion focused on wiring and health, not on exhaustive page behavior.
  SYSINFO_RESPONSE="$sysinfo_response" python3 - <<'PY'
import json
import os

payload = json.loads(os.environ["SYSINFO_RESPONSE"])
data = payload.get("data", payload)
if payload.get("code", 0) != 0:
    raise SystemExit(f"system info business code is not zero: {payload!r}")
coordination = data.get("coordination") or {}
if coordination.get("clusterEnabled") is not True:
    raise SystemExit(f"expected clusterEnabled=true, got: {coordination!r}")
if coordination.get("backend") != "redis":
    raise SystemExit(f"expected redis backend, got: {coordination!r}")
if coordination.get("redisHealthy") is not True:
    raise SystemExit(f"expected redisHealthy=true, got: {coordination!r}")
if coordination.get("primary") is not True:
    raise SystemExit(f"expected primary=true, got: {coordination!r}")
if not coordination.get("nodeId"):
    raise SystemExit(f"expected non-empty nodeId, got: {coordination!r}")
PY
}

mkdir -p "$ROOT_DIR/temp/bin" "$BACKUP_DIR"
rm -f "$BACKEND_BINARY" "$BACKEND_LOG"

# This script can drop and rebuild the target database. CI passes a disposable
# service database; local callers must make the same choice explicitly.
if [ -z "$DB_LINK" ]; then
  echo "LINAPRO_REDIS_CLUSTER_SMOKE_DB_LINK is required for Redis cluster smoke" >&2
  exit 1
fi

write_cluster_config

make -C "$ROOT_DIR" init confirm=init rebuild=true
make -C "$CORE_DIR" pack.assets
(
  cd "$CORE_DIR"
  go build -o "$BACKEND_BINARY" .
)

(
  cd "$CORE_DIR"
  "$BACKEND_BINARY"
) >"$BACKEND_LOG" 2>&1 &
BACKEND_PID="$!"

wait_for_cluster_health
assert_login_and_sysinfo

echo "Redis cluster backend smoke passed"
