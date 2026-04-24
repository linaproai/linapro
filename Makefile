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
include hack/makefiles/up.mk
include hack/makefiles/dev.mk
include hack/makefiles/build.mk

## test: 运行完整 E2E 测试套件
.PHONY: test
test:
	@echo "🧪 运行 E2E 测试套件..."
	cd hack/tests && pnpm test

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	@if [ "$(confirm)" != "init" ]; then \
		echo "✗ 出于安全原因，执行 make init 需要显式确认"; \
		echo "  请使用: make init confirm=init"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) init confirm=$(confirm)

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	@if [ "$(confirm)" != "mock" ]; then \
		echo "✗ 出于安全原因，执行 make mock 需要显式确认"; \
		echo "  请使用: make mock confirm=mock"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) mock confirm=$(confirm)

## upgrade: 统一开发态升级入口（scope=framework|source-plugin；源码插件需配合 plugin=<id|all>）
.PHONY: upgrade
upgrade:
	@if [ "$(confirm)" != "upgrade" ]; then \
		echo "✗ 出于安全原因，执行 make upgrade 需要显式确认"; \
		echo "  请使用: make upgrade confirm=upgrade"; \
		exit 1; \
	fi
	@go run ./hack/upgrade-source --confirm=$(confirm) \
		$(if $(scope),--scope=$(scope),) \
		$(if $(repo),--repo=$(repo),) \
		$(if $(target),--target=$(target),) \
		$(if $(plugin),--plugin=$(plugin),) \
		$(if $(dry_run),--dry-run,)

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
