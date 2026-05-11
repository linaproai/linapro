#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TEST_DIR="$ROOT_DIR/temp/test-stop-dev-services"
TEST_TEMP_DIR="temp/test-stop-dev-services/runtime"
PID_DIR="$TEST_DIR/pids"
BACKEND_PID_FILE="$PID_DIR/backend.pid"
FRONTEND_PID_FILE="$PID_DIR/frontend.pid"
BACKEND_PORT="${LINAPRO_STOP_TEST_BACKEND_PORT:-18080}"
FRONTEND_PORT="${LINAPRO_STOP_TEST_FRONTEND_PORT:-15666}"
MATCHED_FRONTEND_PORT="${LINAPRO_STOP_TEST_MATCHED_FRONTEND_PORT:-15667}"
UNRELATED_PID=""
UNRELATED_PORT_PID=""
MATCHED_BACKEND_PID=""
MATCHED_FRONTEND_PID=""

cleanup() {
  if [ -n "$UNRELATED_PID" ] && kill -0 "$UNRELATED_PID" 2>/dev/null; then
    kill "$UNRELATED_PID" 2>/dev/null || true
    wait "$UNRELATED_PID" 2>/dev/null || true
  fi
  if [ -n "$UNRELATED_PORT_PID" ] && kill -0 "$UNRELATED_PORT_PID" 2>/dev/null; then
    kill "$UNRELATED_PORT_PID" 2>/dev/null || true
    wait "$UNRELATED_PORT_PID" 2>/dev/null || true
  fi
  if [ -n "$MATCHED_BACKEND_PID" ] && kill -0 "$MATCHED_BACKEND_PID" 2>/dev/null; then
    kill "$MATCHED_BACKEND_PID" 2>/dev/null || true
    wait "$MATCHED_BACKEND_PID" 2>/dev/null || true
  fi
  if [ -n "$MATCHED_FRONTEND_PID" ] && kill -0 "$MATCHED_FRONTEND_PID" 2>/dev/null; then
    kill "$MATCHED_FRONTEND_PID" 2>/dev/null || true
    wait "$MATCHED_FRONTEND_PID" 2>/dev/null || true
  fi
  rm -rf "$TEST_DIR"
}

trap cleanup EXIT INT TERM

wait_for_port() {
  local port="$1"
  local elapsed=0
  while [ "$elapsed" -lt 20 ]; do
    if lsof -nP -iTCP:"$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
    elapsed=$((elapsed + 1))
  done
  echo "Timed out waiting for test port $port"
  exit 1
}

assert_running() {
  local pid="$1"
  local label="$2"
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "$label was unexpectedly stopped"
    exit 1
  fi
}

mkdir -p "$PID_DIR"

sleep 60 &
UNRELATED_PID="$!"
printf '%s\n' "$UNRELATED_PID" >"$BACKEND_PID_FILE"

python3 -m http.server "$FRONTEND_PORT" --bind 127.0.0.1 --directory "$TEST_DIR" >/dev/null 2>&1 &
UNRELATED_PORT_PID="$!"
wait_for_port "$FRONTEND_PORT"

ROOT_DIR="$ROOT_DIR" \
  BACKEND_PID="$BACKEND_PID_FILE" \
  FRONTEND_PID="$FRONTEND_PID_FILE" \
  BACKEND_PORT="$BACKEND_PORT" \
  FRONTEND_PORT="$FRONTEND_PORT" \
  bash "$ROOT_DIR/hack/scripts/stop-dev-services.sh"

assert_running "$UNRELATED_PID" "Unrelated PID-file process"
assert_running "$UNRELATED_PORT_PID" "Unrelated port listener"

if [ -f "$BACKEND_PID_FILE" ]; then
  echo "Stale backend PID file was not removed"
  exit 1
fi

mkdir -p "$ROOT_DIR/$TEST_TEMP_DIR/bin" "$PID_DIR"
cat >"$ROOT_DIR/$TEST_TEMP_DIR/bin/lina" <<'SCRIPT'
#!/usr/bin/env bash
sleep 60
SCRIPT
chmod +x "$ROOT_DIR/$TEST_TEMP_DIR/bin/lina"

(
  cd "$ROOT_DIR/apps/lina-core"
  "$ROOT_DIR/$TEST_TEMP_DIR/bin/lina"
) &
MATCHED_BACKEND_PID="$!"
printf '%s\n' "$MATCHED_BACKEND_PID" >"$BACKEND_PID_FILE"

ROOT_DIR="$ROOT_DIR" \
  TEMP_DIR="$TEST_TEMP_DIR" \
  BACKEND_PID="$BACKEND_PID_FILE" \
  FRONTEND_PID="$FRONTEND_PID_FILE" \
  BACKEND_PORT="$BACKEND_PORT" \
  FRONTEND_PORT="$BACKEND_PORT" \
  bash "$ROOT_DIR/hack/scripts/stop-dev-services.sh"

if kill -0 "$MATCHED_BACKEND_PID" 2>/dev/null; then
  echo "Matched LinaPro backend test process was not stopped"
  exit 1
fi
MATCHED_BACKEND_PID=""

mkdir -p "$PID_DIR"
mkdir -p "$TEST_DIR/bin"
cat >"$TEST_DIR/bin/vite" <<'SCRIPT'
#!/usr/bin/env bash
sleep 60
SCRIPT
chmod +x "$TEST_DIR/bin/vite"
(
  cd "$ROOT_DIR/apps/lina-vben/apps/web-antd"
  "$TEST_DIR/bin/vite" --mode development --host 127.0.0.1 --port "$MATCHED_FRONTEND_PORT" --strictPort
) &
MATCHED_FRONTEND_PID="$!"
printf '%s\n' "$MATCHED_FRONTEND_PID" >"$FRONTEND_PID_FILE"

ROOT_DIR="$ROOT_DIR" \
  BACKEND_PID="$BACKEND_PID_FILE" \
  FRONTEND_PID="$FRONTEND_PID_FILE" \
  BACKEND_PORT="$BACKEND_PORT" \
  FRONTEND_PORT="$MATCHED_FRONTEND_PORT" \
  bash "$ROOT_DIR/hack/scripts/stop-dev-services.sh"

if kill -0 "$MATCHED_FRONTEND_PID" 2>/dev/null; then
  echo "Matched LinaPro frontend test process was not stopped"
  exit 1
fi
MATCHED_FRONTEND_PID=""

echo "stop-dev-services safety checks passed"
