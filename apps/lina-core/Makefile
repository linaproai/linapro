ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

## prepare-packed-assets: 准备宿主 manifest 与插件相关嵌入资源，供后续后端构建打包使用
.PHONY: prepare-packed-assets
prepare-packed-assets:
	@../../hack/scripts/prepare-packed-assets.sh

## build: 先准备嵌入资源，再按 hack/config.yaml 调用 GoFrame 构建宿主后端二进制
build: prepare-packed-assets

define require_confirm
	@if [ "$(confirm)" != "$(1)" ]; then \
		echo "✗ 出于安全原因，执行 make $(1) 需要显式确认"; \
		echo "  请使用: make $(1) confirm=$(1)"; \
		exit 1; \
	fi
endef

## init: 初始化数据库，仅执行宿主 DDL 与 Seed 数据；需 confirm=init，可追加 rebuild=true 重建数据库
.PHONY: init
init:
	$(call require_confirm,init)
	@echo "正在初始化数据库..."
	@go run main.go init --confirm=$(confirm) --sql-source=local $(if $(rebuild),--rebuild=$(rebuild),)
	@echo "✓ 数据库初始化完成"

## mock: 加载 manifest/sql/mock-data 下的 Mock 演示/测试数据；需先执行 init，并传入 confirm=mock
.PHONY: mock
mock:
	$(call require_confirm,mock)
	@echo "正在加载 Mock 数据..."
	@go run main.go mock --confirm=$(confirm) --sql-source=local
	@echo "✓ Mock 数据加载完成"

## upgrade: 代理执行根目录开发态升级工具；需 confirm=upgrade，支持 scope=framework|source-plugin、plugin=<id|all>、repo、target、dry_run
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
