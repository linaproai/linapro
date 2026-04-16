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

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	@echo "正在初始化数据库..."
	@go run main.go init
	@echo "✓ 数据库初始化完成"

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	@echo "正在加载 Mock 数据..."
	@go run main.go mock
	@echo "✓ Mock 数据加载完成"
