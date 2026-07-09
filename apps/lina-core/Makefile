# LinaPro core backend Makefile.
# This file is intentionally a thin wrapper around the repository-level
# linactl tool so code generation and build behavior stay centralized.

# pack.assets: Prepare host manifest and plugin embedded assets for backend packaging
# pack.assets: 准备宿主 manifest 与插件相关嵌入资源，供后续后端构建打包使用
.PHONY: pack.assets
pack.assets:
	@go run ../../hack/tools/linactl pack.assets

# build: Prepare embedded assets, then build the host backend binary via GoFrame config
# build: 先准备嵌入资源，再按 hack/config.yaml 调用 GoFrame 构建宿主后端二进制
build: pack.assets

# db.init: Initialize host DDL and seed data only; requires confirm=init, supports rebuild=true
# db.init: 初始化数据库，仅执行宿主 DDL 与 Seed 数据；需 confirm=init，可追加 rebuild=true 重建数据库
.PHONY: db.init
db.init:
	@go run ../../hack/tools/linactl db.init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# db.mock: Load mock demo/test data from manifest/sql/mock-data after db.init; requires confirm=mock
# db.mock: 加载 manifest/sql/mock-data 下的 Mock 演示/测试数据；需先执行 db.init，并传入 confirm=mock
.PHONY: db.mock
db.mock:
	@go run ../../hack/tools/linactl db.mock confirm=$(confirm)

# ctrl: Generate host GoFrame controller scaffolding
# ctrl: 生成宿主 GoFrame 控制器骨架
.PHONY: ctrl
ctrl:
	@go run ../../hack/tools/linactl ctrl dir=apps/lina-core

# dao: Generate host DAO/DO/Entity files
# dao: 生成宿主 DAO/DO/Entity 文件
.PHONY: dao
dao:
	@go run ../../hack/tools/linactl dao dir=apps/lina-core
