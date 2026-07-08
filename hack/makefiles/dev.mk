# LinaPro Development Server Commands
# LinaPro 开发服务指令
# ================================

DEV_ARGS := backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
ifneq ($(origin plugins), undefined)
DEV_ARGS += plugins=$(plugins)
endif
ifneq ($(origin dir), undefined)
DEV_ARGS += dir=$(dir)
endif

# Restart both backend and frontend development servers.
# 重启后端和前端开发服务器。
## dev: Restart backend/frontend services, or use dir=<path> to run the same targeted build logic as make build
.PHONY: dev
dev:
	@$(LINACTL) dev $(DEV_ARGS)

STOP_ARGS := backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
ifneq ($(origin dir), undefined)
STOP_ARGS += dir=$(dir)
endif

# Stop backend and frontend development servers and clean stale PID files.
# 停止后端和前端开发服务器，并清理残留 PID 文件。
## stop: Stop backend/frontend services, or use dir=<path> to run configured stop commands from hack/config.yaml
.PHONY: stop
stop:
	@$(LINACTL) stop $(STOP_ARGS)

STATUS_ARGS := backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
ifneq ($(origin dir), undefined)
STATUS_ARGS += dir=$(dir)
endif

# Show backend/frontend runtime status and their log file paths.
# 查看前后端运行状态及对应日志文件路径。
## status: Show backend/frontend status, or use dir=<path> to run configured status commands from hack/config.yaml
.PHONY: status
status:
	@$(LINACTL) status $(STATUS_ARGS)
