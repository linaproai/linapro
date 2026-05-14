DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

# pack.assets: Prepare host manifest and plugin embedded assets for backend packaging
# pack.assets: 准备宿主 manifest 与插件相关嵌入资源，供后续后端构建打包使用
.PHONY: pack.assets
pack.assets:
	@go run ../../hack/tools/linactl prepare-packed-assets

# build: Prepare embedded assets, then build the host backend binary via GoFrame config
# build: 先准备嵌入资源，再按 hack/config.yaml 调用 GoFrame 构建宿主后端二进制
build: pack.assets

# init: Initialize host DDL and seed data only; requires confirm=init, supports rebuild=true
# init: 初始化数据库，仅执行宿主 DDL 与 Seed 数据；需 confirm=init，可追加 rebuild=true 重建数据库
.PHONY: init
init:
	@go run ../../hack/tools/linactl init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# mock: Load mock demo/test data from manifest/sql/mock-data after init; requires confirm=mock
# mock: 加载 manifest/sql/mock-data 下的 Mock 演示/测试数据；需先执行 init，并传入 confirm=mock
.PHONY: mock
mock:
	@go run ../../hack/tools/linactl mock confirm=$(confirm)
