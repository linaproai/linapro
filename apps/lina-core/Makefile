ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

.PHONY: prepare-packed-assets
prepare-packed-assets:
	@../../hack/scripts/prepare-packed-assets.sh

build: prepare-packed-assets

define require_confirm
	@if [ "$(confirm)" != "$(1)" ]; then \
		echo "✗ 出于安全原因，执行 make $(1) 需要显式确认"; \
		echo "  请使用: make $(1) confirm=$(1)"; \
		exit 1; \
	fi
endef

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	$(call require_confirm,init)
	@echo "正在初始化数据库..."
	@go run main.go init --confirm=$(confirm) --sql-source=local
	@echo "✓ 数据库初始化完成"

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	$(call require_confirm,mock)
	@echo "正在加载 Mock 数据..."
	@go run main.go mock --confirm=$(confirm) --sql-source=local
	@echo "✓ Mock 数据加载完成"

## upgrade: 统一开发态升级入口（scope=framework|source-plugin；源码插件需配合 plugin=<id|all>）
.PHONY: upgrade
upgrade:
	$(call require_confirm,upgrade)
	@echo "正在执行开发态升级..."
	@$(MAKE) -C ../.. upgrade confirm=$(confirm) \
		$(if $(scope),scope=$(scope),) \
		$(if $(repo),repo=$(repo),) \
		$(if $(target),target=$(target),) \
		$(if $(plugin),plugin=$(plugin),) \
		$(if $(dry_run),dry_run=$(dry_run),)
	@echo "✓ 开发态升级命令执行完成"
