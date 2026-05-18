#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/lina-core"
CONFIG_PATH="$CORE_DIR/manifest/config/config.yaml"
CONFIG_TEMPLATE_PATH="$CORE_DIR/manifest/config/config.template.yaml"
BACKUP_DIR="$ROOT_DIR/temp/sqlite-smoke"
BACKUP_CONFIG_PATH="$BACKUP_DIR/config.yaml.backup"
SQLITE_LINK="sqlite::@file(./temp/sqlite/linapro.db)"
SQLITE_DB_PATH="$CORE_DIR/temp/sqlite/linapro.db"
BACKEND_BINARY="$ROOT_DIR/temp/bin/lina-smoke"
BACKEND_LOG="$ROOT_DIR/temp/sqlite-smoke/lina-core.log"
SMOKE_PORT="${LINAPRO_SQLITE_SMOKE_PORT:-18080}"
BACKEND_PID=""

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

write_sqlite_config() {
  mkdir -p "$BACKUP_DIR"
  if [ -f "$CONFIG_PATH" ]; then
    cp "$CONFIG_PATH" "$BACKUP_CONFIG_PATH"
  else
    cp "$CONFIG_TEMPLATE_PATH" "$BACKUP_CONFIG_PATH"
  fi

  cp "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
  SMOKE_PORT="$SMOKE_PORT" perl -0pi -e 's/address:\s*":\d+"/q{address: ":}.$ENV{SMOKE_PORT}.q{"}/e' "$CONFIG_PATH"
  SQLITE_LINK="$SQLITE_LINK" perl -0pi -e 's/link:\s*"[^"]+"/q{link: "}.$ENV{SQLITE_LINK}.q{"}/e' "$CONFIG_PATH"
  perl -0pi -e 's/(cluster:\n(?:[ \t]*#.*\n)*[ \t]*)enabled:\s*(?:true|false)/${1}enabled: true/' "$CONFIG_PATH"
}

wait_for_backend() {
  local elapsed=0
  while [ "$elapsed" -lt 60 ]; do
    if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
      echo "SQLite backend smoke startup failed; backend exited early"
      echo "Check log: $BACKEND_LOG"
      exit 1
    fi
    if curl -fsS -o /dev/null "http://127.0.0.1:$SMOKE_PORT/api/v1/health" 2>/dev/null; then
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  echo "SQLite backend smoke startup timed out"
  echo "Check log: $BACKEND_LOG"
  exit 1
}

assert_log_contains() {
  local expected="$1"
  if ! grep -Fq "$expected" "$BACKEND_LOG"; then
    echo "Expected backend log to contain: $expected"
    echo "Check log: $BACKEND_LOG"
    exit 1
  fi
}

assert_health() {
  local response
  response="$(curl -fsS "http://127.0.0.1:$SMOKE_PORT/api/v1/health")"
  if ! grep -Eq '"code"[[:space:]]*:[[:space:]]*0' <<<"$response"; then
    echo "Expected health response business code 0, got: $response"
    exit 1
  fi
  if ! grep -Eq '"status"[[:space:]]*:[[:space:]]*"ok"' <<<"$response"; then
    echo "Expected health status ok, got: $response"
    exit 1
  fi
  if ! grep -Eq '"mode"[[:space:]]*:[[:space:]]*"single"' <<<"$response"; then
    echo "Expected health mode single, got: $response"
    exit 1
  fi
}

assert_login() {
  local response
  response="$(
    curl -fsS \
      -H "Content-Type: application/json" \
      -d '{"username":"admin","password":"admin123"}' \
      "http://127.0.0.1:$SMOKE_PORT/api/v1/auth/login"
  )"
  if ! grep -Eq '"code"[[:space:]]*:[[:space:]]*0' <<<"$response"; then
    echo "Expected login response business code 0, got: $response"
    exit 1
  fi
  if ! grep -Eq '"accessToken"[[:space:]]*:[[:space:]]*"[^"]+"' <<<"$response"; then
    echo "Expected non-empty accessToken, got: $response"
    exit 1
  fi
}

mkdir -p "$ROOT_DIR/temp/bin" "$BACKUP_DIR" "$(dirname "$SQLITE_DB_PATH")"
rm -f "$SQLITE_DB_PATH" "$SQLITE_DB_PATH-shm" "$SQLITE_DB_PATH-wal" "$BACKEND_LOG"

write_sqlite_config

make -C "$ROOT_DIR" init confirm=init rebuild=true
make -C "$ROOT_DIR" mock confirm=mock
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

wait_for_backend
assert_log_contains "SQLite mode is active"
assert_log_contains "SQLite mode only supports single-node deployment"
assert_log_contains "do not use SQLite mode in production"
assert_health
assert_login

echo "SQLite backend smoke passed"
