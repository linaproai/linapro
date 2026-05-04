.DEFAULT_GOAL := build

# Update GoFrame and its CLI to the latest stable version.
# 更新 GoFrame 及其 CLI 到最新稳定版本。
.PHONY: up
up: cli.install
	@gf up -a

# Build binary using configuration from hack/config.yaml.
# 使用 hack/config.yaml 配置构建二进制文件。
.PHONY: build
build: cli.install
	@gf build -ew

# Parse API definitions and generate controllers/SDK artifacts.
# 解析 API 定义并生成控制器和 SDK 产物。
.PHONY: ctrl
ctrl: cli.install
	@gf gen ctrl

# Generate Go files for DAO/DO/Entity.
# 生成 DAO/DO/Entity 的 Go 文件。
.PHONY: dao
dao: cli.install
	@gf gen dao

# Parse project Go files and generate enum Go files.
# 解析当前项目 Go 文件并生成枚举 Go 文件。
.PHONY: enums
enums: cli.install
	@gf gen enums

# Generate Go files for services.
# 生成 Service 层 Go 文件。
.PHONY: service
service: cli.install
	@gf gen service


# Build Docker image.
# 构建 Docker 镜像。
.PHONY: image
image: cli.install
	$(eval _TAG  = $(shell git rev-parse --short HEAD))
ifneq (, $(shell git status --porcelain 2>/dev/null))
	$(eval _TAG  = $(_TAG).dirty)
endif
	$(eval _TAG  = $(if ${TAG},  ${TAG}, $(_TAG)))
	$(eval _PUSH = $(if ${PUSH}, ${PUSH}, ))
	@gf docker ${_PUSH} -tn $(DOCKER_NAME):${_TAG};


# Build Docker image and automatically push to the Docker repository.
# 构建 Docker 镜像并自动推送到 Docker 仓库。
.PHONY: image.push
image.push: cli.install
	@make image PUSH=-p;


# Deploy image and YAML manifests to the current kubectl environment.
# 将镜像和 YAML 清单部署到当前 kubectl 环境。
.PHONY: deploy
deploy: cli.install
	$(eval _TAG = $(if ${TAG},  ${TAG}, develop))

	@set -e; \
	mkdir -p $(ROOT_DIR)/temp/kustomize;\
	cd $(ROOT_DIR)/manifest/deploy/kustomize/overlays/${_ENV};\
	kustomize build > $(ROOT_DIR)/temp/kustomize.yaml;\
	kubectl   apply -f $(ROOT_DIR)/temp/kustomize.yaml; \
	if [ $(DEPLOY_NAME) != "" ]; then \
		kubectl patch -n $(NAMESPACE) deployment/$(DEPLOY_NAME) -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"$(shell date +%s)\"}}}}}"; \
	fi;


# Parse protobuf files and generate Go files.
# 解析 protobuf 文件并生成 Go 文件。
.PHONY: pb
pb: cli.install
	@gf gen pb

# Generate protobuf files for database tables.
# 为数据库表生成 protobuf 文件。
.PHONY: pbentity
pbentity: cli.install
	@gf gen pbentity
