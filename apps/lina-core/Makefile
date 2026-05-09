ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

# prepare-packed-assets: Prepare host manifest and plugin embedded assets for backend packaging
# prepare-packed-assets: 准备宿主 manifest 与插件相关嵌入资源，供后续后端构建打包使用
.PHONY: prepare-packed-assets
prepare-packed-assets:
	@../../hack/scripts/prepare-packed-assets.sh

# build: Prepare embedded assets, then build the host backend binary via GoFrame config
# build: 先准备嵌入资源，再按 hack/config.yaml 调用 GoFrame 构建宿主后端二进制
build: prepare-packed-assets

# Confirmation helper shared by destructive local targets.
# 本地高风险目标共用的确认辅助逻辑。
define require_confirm
	@if [ "$(confirm)" != "$(1)" ]; then \
		echo "✗ make $(1) requires explicit confirmation for safety"; \
		echo "  Use: make $(1) confirm=$(1)"; \
		exit 1; \
	fi
endef

# init: Initialize host DDL and seed data only; requires confirm=init, supports rebuild=true
# init: 初始化数据库，仅执行宿主 DDL 与 Seed 数据；需 confirm=init，可追加 rebuild=true 重建数据库
.PHONY: init
init:
	$(call require_confirm,init)
	@echo "Initializing database..."
	@tmp_log="$$(mktemp)"; \
	if go run main.go init --confirm=$(confirm) --sql-source=local $(if $(rebuild),--rebuild=$(rebuild),) > "$$tmp_log" 2>&1; then \
		cat "$$tmp_log"; \
		rm -f "$$tmp_log"; \
	else \
		status="$$?"; \
		cat "$$tmp_log"; \
		if grep -Eqi "dial tcp|connection refused|connect: connection|failed to connect|i/o timeout|no such host" "$$tmp_log"; then \
			echo "PostgreSQL is not ready. Start PostgreSQL first."; \
			echo "Local example: docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine"; \
		fi; \
		rm -f "$$tmp_log"; \
		exit "$$status"; \
	fi
	@echo "✓ Database initialization complete"

# mock: Load mock demo/test data from manifest/sql/mock-data after init; requires confirm=mock
# mock: 加载 manifest/sql/mock-data 下的 Mock 演示/测试数据；需先执行 init，并传入 confirm=mock
.PHONY: mock
mock:
	$(call require_confirm,mock)
	@echo "Loading mock data..."
	@go run main.go mock --confirm=$(confirm) --sql-source=local
	@echo "✓ Mock data load complete"
