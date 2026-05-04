# LinaPro Development Server Commands
# LinaPro 开发服务指令
# ================================

# Restart both backend and frontend development servers.
# 重启后端和前端开发服务器。
## dev: Restart backend and frontend development servers
.PHONY: dev
dev: stop
	@set -e; \
	root_dir="$(CURDIR)"; \
	_wait_http() { \
		name="$$1"; \
		pid_file="$$2"; \
		url="$$3"; \
		timeout="$$4"; \
		log_file="$$5"; \
		elapsed=0; \
		while [ "$$elapsed" -lt "$$timeout" ]; do \
			if [ ! -f "$$pid_file" ]; then \
				echo "$$name startup failed: PID file does not exist"; \
				echo "Check log: $$log_file"; \
				exit 1; \
			fi; \
			pid=$$(cat "$$pid_file"); \
			if ! kill -0 "$$pid" 2>/dev/null; then \
				echo "$$name startup failed: process exited"; \
				echo "Check log: $$log_file"; \
				exit 1; \
			fi; \
			if curl -fsS -o /dev/null "$$url" 2>/dev/null; then \
				echo "✓ $$name is ready: $$url"; \
				return 0; \
			fi; \
			sleep 1; \
			elapsed=$$((elapsed + 1)); \
		done; \
		echo "$$name startup timed out ($${timeout}s): $$url"; \
		echo "Check log: $$log_file"; \
		exit 1; \
	}; \
	mkdir -p $(TEMP_DIR) $(PID_DIR) $(TEMP_DIR)/bin; \
	> $(BACKEND_LOG); \
	> $(FRONTEND_LOG); \
	backend_binary="$$root_dir/$(TEMP_DIR)/bin/lina"; \
	echo "Restarting services..."; \
	$(MAKE) wasm; \
	./hack/scripts/prepare-packed-assets.sh; \
	(cd "$$root_dir/$(BACKEND_DIR)" && go build -o "$$backend_binary" .) || { echo "Backend build failed"; exit 1; }; \
	nohup sh -c 'cd "$$1" && exec "$$2"' sh "$$root_dir/$(BACKEND_DIR)" "$$backend_binary" >> $(BACKEND_LOG) 2>&1 < /dev/null & echo $$! > $(BACKEND_PID); \
	nohup sh -c 'cd "'"$$root_dir"'/$(FRONTEND_DIR)/apps/web-antd" && exec ../../node_modules/.bin/vite --mode development --host 127.0.0.1 --port $(FRONTEND_PORT) --strictPort' >> $(FRONTEND_LOG) 2>&1 < /dev/null & echo $$! > $(FRONTEND_PID); \
	_wait_http "Backend" "$(BACKEND_PID)" "http://127.0.0.1:$(BACKEND_PORT)/" 60 "$(BACKEND_LOG)"; \
	_wait_http "Frontend" "$(FRONTEND_PID)" "http://127.0.0.1:$(FRONTEND_PORT)/" 60 "$(FRONTEND_LOG)"; \
	cd "$$root_dir"; \
	$(MAKE) status

# Stop backend and frontend development servers and clean stale PID files.
# 停止后端和前端开发服务器，并清理残留 PID 文件。
## stop: Stop backend and frontend development servers
.PHONY: stop
stop:
	@echo "Stopping services..."
	@_kill_tree() { \
		for child in $$(pgrep -P $$1 2>/dev/null); do \
			_kill_tree $$child; \
		done; \
		kill $$1 2>/dev/null; \
	}; \
	_stop_service() { \
		local name="$$1" pid_file="$$2" port="$$3"; \
		local stopped=false; \
		if [ -f "$$pid_file" ]; then \
			local pid=$$(cat "$$pid_file"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				_kill_tree "$$pid"; \
				stopped=true; \
			fi; \
			rm -f "$$pid_file"; \
		fi; \
		local pids=$$(lsof -ti :"$$port" 2>/dev/null); \
		if [ -n "$$pids" ]; then \
			echo "$$pids" | xargs kill 2>/dev/null; \
			sleep 0.5; \
			pids=$$(lsof -ti :"$$port" 2>/dev/null); \
			if [ -n "$$pids" ]; then \
				echo "$$pids" | xargs kill -9 2>/dev/null; \
			fi; \
			stopped=true; \
		fi; \
		if [ "$$stopped" = true ]; then \
			echo "✓ $$name stopped"; \
		else \
			echo "  $$name is not running"; \
		fi; \
	}; \
	_stop_service "Backend" "$(BACKEND_PID)" "$(BACKEND_PORT)"; \
	_stop_service "Frontend" "$(FRONTEND_PID)" "$(FRONTEND_PORT)"

# Show backend/frontend runtime status and their log file paths.
# 查看前后端运行状态及对应日志文件路径。
## status: Show backend and frontend status with log paths
.PHONY: status
status:
	@echo ""
	@echo "╔══════════════════════════════════════════════╗"
	@echo "║         LinaPro Framework Status             ║"
	@echo "╠══════════════════════════════════════════════╣"
	@if lsof -ti :$(BACKEND_PORT) >/dev/null 2>&1; then \
		echo "║  Backend:  ✓ running  http://localhost:$(BACKEND_PORT)  ║"; \
	else \
		echo "║  Backend:  ✗ stopped  (port $(BACKEND_PORT))            ║"; \
	fi
	@if lsof -ti :$(FRONTEND_PORT) >/dev/null 2>&1; then \
		echo "║  Frontend: ✓ running  http://localhost:$(FRONTEND_PORT)  ║"; \
	else \
		echo "║  Frontend: ✗ stopped  (port $(FRONTEND_PORT))            ║"; \
	fi
	@echo "╠══════════════════════════════════════════════╣"
	@echo "║  Backend log:   temp/lina-core.log           ║"
	@echo "║  Frontend log:  temp/lina-vben.log           ║"
	@echo "╚══════════════════════════════════════════════╝"
	@echo ""
