.DEFAULT_GOAL := build

# Update GoFrame and its CLI to the latest stable version.
# 更新 GoFrame 及其 CLI 到最新稳定版本。
.PHONY: up
up:
	@go run ../../hack/tools/linactl cli.install
	@gf up -a

# Build binary using configuration from hack/config.yaml.
# 使用 hack/config.yaml 配置构建二进制文件。
.PHONY: build
build:
	@go run ../../hack/tools/linactl build

# Parse API definitions and generate controllers/SDK artifacts.
# 解析 API 定义并生成控制器和 SDK 产物。
.PHONY: ctrl
ctrl:
	@go run ../../hack/tools/linactl ctrl

# Generate Go files for DAO/DO/Entity.
# 生成 DAO/DO/Entity 的 Go 文件。
.PHONY: dao
dao:
	@go run ../../hack/tools/linactl dao

# Build Docker image.
# 构建 Docker 镜像。
.PHONY: image
image:
	@go run ../../hack/tools/linactl image tag="$(TAG)" push="$(PUSH)" image="$(DOCKER_NAME)"


# Build Docker image and automatically push to the Docker repository.
# 构建 Docker 镜像并自动推送到 Docker 仓库。
.PHONY: image.push
image.push:
	@go run ../../hack/tools/linactl image tag="$(TAG)" push=1 image="$(DOCKER_NAME)"
