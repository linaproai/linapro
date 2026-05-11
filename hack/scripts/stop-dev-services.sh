#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$(pwd -P)}"
BACKEND_DIR="${BACKEND_DIR:-apps/lina-core}"
FRONTEND_DIR="${FRONTEND_DIR:-apps/lina-vben}"
TEMP_DIR="${TEMP_DIR:-temp}"
BACKEND_PID="${BACKEND_PID:-$ROOT_DIR/$TEMP_DIR/pids/backend.pid}"
FRONTEND_PID="${FRONTEND_PID:-$ROOT_DIR/$TEMP_DIR/pids/frontend.pid}"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-5666}"

# Normalize directories through physical paths so PID and cwd checks compare the
# same workspace even when commands are launched through symlinks.
# 通过物理路径归一化目录，确保即使命令经由符号链接启动，PID 与 cwd 校验也能匹配同一个工作区。
canonical_dir() {
  local path="$1"
  if [ -d "$path" ]; then
    (cd "$path" && pwd -P)
  else
    printf '%s\n' "$path"
  fi
}

# Build a stable absolute file path while tolerating files that may not exist yet.
# 构造稳定的绝对文件路径，同时允许目标文件尚未创建。
canonical_file_path() {
  local path="$1"
  local dir
  dir="$(dirname "$path")"
  printf '%s/%s\n' "$(canonical_dir "$dir")" "$(basename "$path")"
}

ROOT_DIR="$(canonical_dir "$ROOT_DIR")"
BACKEND_CWD="$(canonical_dir "$ROOT_DIR/$BACKEND_DIR")"
FRONTEND_CWD="$(canonical_dir "$ROOT_DIR/$FRONTEND_DIR/apps/web-antd")"
BACKEND_BINARY="$(canonical_file_path "$ROOT_DIR/$TEMP_DIR/bin/lina")"

# Accept only real process IDs and avoid PID 1, which must never be terminated by
# a development cleanup script.
# 只接受真实进程 ID，并排除 PID 1，避免开发环境清理脚本误终止系统根进程。
valid_pid() {
  local pid="$1"
  [[ "$pid" =~ ^[0-9]+$ ]] && ((pid > 1))
}

# Use signal 0 as a side-effect-free liveness check before reading process metadata
# or sending termination signals.
# 使用 signal 0 做无副作用的存活检查，再读取进程元数据或发送终止信号。
process_exists() {
  local pid="$1"
  valid_pid "$pid" && kill -0 "$pid" 2>/dev/null
}

# Resolve the process working directory to prove it belongs to this checkout,
# not just any process that happens to use the same command name.
# 读取进程工作目录，用于确认进程属于当前代码仓库，而不是仅命令名相同的其他进程。
process_cwd() {
  local pid="$1"
  lsof -a -p "$pid" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p' | head -n 1 || true
}

process_args() {
  local pid="$1"
  ps -p "$pid" -o args= 2>/dev/null || true
}

# Backend cleanup is intentionally scoped to the generated LinaPro binary running
# from the backend working directory.
# 后端清理仅作用于从后端工作目录启动的 LinaPro 编译产物，避免误杀其他 Go 服务。
is_backend_process() {
  local pid="$1"
  local cwd args
  cwd="$(process_cwd "$pid")"
  [ "$cwd" = "$BACKEND_CWD" ] || return 1

  args="$(process_args "$pid")"
  [[ "$args" == *"$BACKEND_BINARY"* ]]
}

# Frontend cleanup matches the Vite dev command shape used by make dev, including
# host, port, mode, and strict-port arguments.
# 前端清理匹配 make dev 启动的 Vite 开发命令形态，包括 host、port、mode 与 strictPort 参数。
is_frontend_process() {
  local pid="$1"
  local cwd args
  cwd="$(process_cwd "$pid")"
  [ "$cwd" = "$FRONTEND_CWD" ] || return 1

  args="$(process_args "$pid")"
  [[ "$args" == *vite* ]] || return 1
  [[ "$args" == *"--mode development"* ]] || return 1
  [[ "$args" == *"--host 127.0.0.1"* ]] || return 1
  [[ "$args" == *"--port $FRONTEND_PORT"* ]] || return 1
  [[ "$args" == *"--strictPort"* ]]
}

matches_service() {
  local service="$1"
  local pid="$2"
  case "$service" in
    backend) is_backend_process "$pid" ;;
    frontend) is_frontend_process "$pid" ;;
    *) return 1 ;;
  esac
}

# Terminate child processes before their parent so spawned dev-server helpers do
# not survive after the root process exits.
# 先终止子进程再终止父进程，避免开发服务派生出的辅助进程在根进程退出后残留。
kill_tree_with_signal() {
  local pid="$1"
  local signal="$2"
  local child

  while read -r child; do
    if valid_pid "$child"; then
      kill_tree_with_signal "$child" "$signal"
    fi
  done < <(pgrep -P "$pid" 2>/dev/null || true)

  kill "-$signal" "$pid" 2>/dev/null || true
}

# Try a graceful TERM first, then escalate to KILL only when the root process is
# still alive after a short grace period.
# 优先使用 TERM 优雅退出；若短暂等待后根进程仍存活，再升级为 KILL。
terminate_tree() {
  local pid="$1"
  kill_tree_with_signal "$pid" TERM
  sleep 0.5
  if process_exists "$pid"; then
    kill_tree_with_signal "$pid" KILL
  fi
}

# Port scanning is a fallback for stale or missing PID files, but every candidate
# is still validated against the current workspace before termination.
# 端口扫描用于兜底处理 PID 文件缺失或过期的场景，但终止前仍会校验进程是否属于当前工作区。
port_listener_pids() {
  local port="$1"
  lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null | sort -u || true
}

read_pid_file() {
  local pid_file="$1"
  head -n 1 "$pid_file" 2>/dev/null | tr -d '[:space:]'
}

# Stop one service in two passes: trust the recorded PID when it still matches,
# then inspect the expected port to catch processes left behind by stale PID files.
# 停止单个服务分两步：先处理仍然匹配的 PID 文件记录，再检查预期端口以清理 PID 文件遗漏的残留进程。
stop_service() {
  local display_name="$1"
  local service="$2"
  local pid_file="$3"
  local port="$4"
  local stopped=false
  local pid

  if [ -f "$pid_file" ]; then
    pid="$(read_pid_file "$pid_file")"
    if process_exists "$pid"; then
      if matches_service "$service" "$pid"; then
        # The PID file is authoritative only after workspace and command validation.
        # PID 文件只有在通过工作区和命令校验后才会被视为可信来源。
        terminate_tree "$pid"
        stopped=true
      else
        echo "  $display_name PID $pid does not match this LinaPro workspace; skipped"
      fi
    fi
    rm -f "$pid_file"
  fi

  while read -r pid; do
    [ -n "$pid" ] || continue
    if matches_service "$service" "$pid"; then
      # A matching port listener may be an orphaned dev process, so stop it after
      # the same service-specific validation used for PID files.
      # 匹配端口的监听进程可能是遗留开发进程，因此复用与 PID 文件相同的服务校验后再停止。
      terminate_tree "$pid"
      stopped=true
    else
      echo "  $display_name port $port is used by non-LinaPro process $pid; skipped"
    fi
  done < <(port_listener_pids "$port")

  if [ "$stopped" = true ]; then
    echo "✓ $display_name stopped"
  else
    echo "  $display_name is not running"
  fi
}

echo "Stopping services..."
stop_service "Backend" backend "$BACKEND_PID" "$BACKEND_PORT"
stop_service "Frontend" frontend "$FRONTEND_PID" "$FRONTEND_PORT"
