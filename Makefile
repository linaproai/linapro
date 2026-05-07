# LinaPro Framework - Root Makefile
# ===========================

BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
TEMP_DIR      := temp
PID_DIR       := $(TEMP_DIR)/pids
BACKEND_PID   := $(PID_DIR)/backend.pid
FRONTEND_PID  := $(PID_DIR)/frontend.pid
BACKEND_PORT  := 8080
FRONTEND_PORT := 5666
BACKEND_LOG   := $(TEMP_DIR)/lina-core.log
FRONTEND_LOG  := $(TEMP_DIR)/lina-vben.log
EMBED_DIR     := $(BACKEND_DIR)/internal/packed/public
OUTPUT_DIR    := $(TEMP_DIR)/output

# Include split makefile targets.
# 引入拆分后的 Makefile 目标文件。
include hack/makefiles/dev.mk
include hack/makefiles/build.mk
include hack/makefiles/image.mk
include hack/makefiles/test.mk
include hack/makefiles/i18n.mk

# Initialize the backend database with schema and required seed data.
# The backend dispatches by database.default.link, for example MySQL or SQLite.
# 初始化后端数据库表结构和系统必需的种子数据。
# 后端会按 database.default.link 自动分发到 MySQL 或 SQLite 等方言。
## init: Initialize the database with DDL and seed data only
.PHONY: init
init:
	@if [ "$(confirm)" != "init" ]; then \
		echo "✗ make init requires explicit confirmation for safety"; \
		echo "  Use: make init confirm=init"; \
		echo "  To rebuild the linapro database: make init confirm=init rebuild=true"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# Load optional mock data for local demos and development verification.
# Mock loading uses the same database.default.link dialect and requires init first.
# 加载用于本地演示和开发验证的可选 Mock 数据。
# Mock 加载使用同一个 database.default.link 方言，并要求先完成 init。
## mock: Load mock demo data after init
.PHONY: mock
mock:
	@if [ "$(confirm)" != "mock" ]; then \
		echo "✗ make mock requires explicit confirmation for safety"; \
		echo "  Use: make mock confirm=mock"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) mock confirm=$(confirm)

# Print the available root Make targets from this file and included target files.
# 打印根 Makefile 及其引入目标文件中可用的 make 目标。
## help: Show help
.PHONY: help
help:
	@set -e; \
	if [ -t 1 ]; then \
		c_title='\033[1;36m'; \
		c_cmd='\033[1;32m'; \
		c_dim='\033[2m'; \
		c_reset='\033[0m'; \
	else \
		c_title=''; \
		c_cmd=''; \
		c_dim=''; \
		c_reset=''; \
	fi; \
	printf "$${c_dim}Usage:$${c_reset} make $${c_cmd}<target>$${c_reset}\n\n"; \
	awk '/^## [^:]+:/ { \
		line=$$0; \
		sub(/^## /, "", line); \
		sep=index(line, ": "); \
		if (sep > 0) { \
			name=substr(line, 1, sep - 1); \
			desc=substr(line, sep + 2); \
			printf "%s\t%s\n", name, desc; \
		} \
	}' $(MAKEFILE_LIST) | sort -k1,1 | \
	awk -F '\t' -v c_cmd="$$c_cmd" -v c_dim="$$c_dim" -v c_reset="$$c_reset" ' \
		{ \
			names[++count]=$$1; \
			descs[count]=$$2; \
			if (length($$1) > max) { \
				max=length($$1); \
			} \
		} \
		END { \
			print c_dim "Available targets:" c_reset; \
			for (i=1; i<=count; i++) { \
				printf "  %s%-*s%s  %s\n", c_cmd, max, names[i], c_reset, descs[i]; \
			} \
		}'
