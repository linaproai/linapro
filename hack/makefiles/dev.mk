# LinaPro Development Server Commands
# LinaPro 开发服务指令
# ================================

DEV_ARGS := backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
ifneq ($(origin plugins), undefined)
DEV_ARGS += plugins=$(plugins)
endif

# Install all prerequisites for development and E2E tests (frontend deps + Playwright browsers).
# 安装开发与 E2E 测试的所有前置依赖（前端依赖 + Playwright 浏览器）。
## dev.setup: Install frontend dependencies and Playwright browsers
.PHONY: dev.setup
dev.setup:
	@$(LINACTL) dev.setup

# Restart both backend and frontend development servers.
# 重启后端和前端开发服务器。
## dev: Restart backend and frontend development servers
.PHONY: dev
dev:
	@$(LINACTL) dev $(DEV_ARGS)

# Stop backend and frontend development servers and clean stale PID files.
# 停止后端和前端开发服务器，并清理残留 PID 文件。
## stop: Stop backend and frontend development servers
.PHONY: stop
stop:
	@$(LINACTL) stop backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)

# Show backend/frontend runtime status and their log file paths.
# 查看前后端运行状态及对应日志文件路径。
## status: Show backend and frontend status with log paths
.PHONY: status
status:
	@$(LINACTL) status backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
