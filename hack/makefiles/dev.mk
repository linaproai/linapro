# LinaPro Development Server Commands
# LinaPro 开发服务指令
# ================================

# Restart both backend and frontend development servers.
# 重启后端和前端开发服务器。
## dev: Restart backend and frontend development servers
.PHONY: dev
dev:
	@go run ./hack/tools/linactl dev backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)

# Stop backend and frontend development servers and clean stale PID files.
# 停止后端和前端开发服务器，并清理残留 PID 文件。
## stop: Stop backend and frontend development servers
.PHONY: stop
stop:
	@go run ./hack/tools/linactl stop backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)

# Show backend/frontend runtime status and their log file paths.
# 查看前后端运行状态及对应日志文件路径。
## status: Show backend and frontend status with log paths
.PHONY: status
status:
	@go run ./hack/tools/linactl status backend_port=$(BACKEND_PORT) frontend_port=$(FRONTEND_PORT)
