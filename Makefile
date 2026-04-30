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

# 引用复杂指令子文件
include hack/makefiles/dev.mk
include hack/makefiles/build.mk
include hack/makefiles/test.mk
include hack/makefiles/i18n.mk

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	@if [ "$(confirm)" != "init" ]; then \
		echo "✗ 出于安全原因，执行 make init 需要显式确认"; \
		echo "  请使用: make init confirm=init"; \
		echo "  如需重建 linapro 数据库: make init confirm=init rebuild=true"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	@if [ "$(confirm)" != "mock" ]; then \
		echo "✗ 出于安全原因，执行 make mock 需要显式确认"; \
		echo "  请使用: make mock confirm=mock"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) mock confirm=$(confirm)

## help: 显示帮助信息
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
