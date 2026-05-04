# LinaPro Docker Image Commands
# LinaPro Docker 镜像构建指令
# ==========================

# Arguments passed from Make variables to the Go image builder.
# 将 Make 变量转换为传递给 Go 镜像构建工具的参数。
IMAGE_BUILDER_ARGS :=

ifneq ($(origin image), undefined)
IMAGE_BUILDER_ARGS += --image=$(image)
endif
ifneq ($(origin tag), undefined)
IMAGE_BUILDER_ARGS += --tag=$(tag)
endif
ifneq ($(origin registry), undefined)
IMAGE_BUILDER_ARGS += --registry=$(registry)
endif
ifneq ($(origin push), undefined)
IMAGE_BUILDER_ARGS += --push=$(push)
endif
ifneq ($(origin os), undefined)
IMAGE_BUILDER_ARGS += --os=$(os)
endif
ifneq ($(origin arch), undefined)
IMAGE_BUILDER_ARGS += --arch=$(arch)
endif
ifneq ($(origin platform), undefined)
IMAGE_BUILDER_ARGS += --platform=$(platform)
endif
ifneq ($(origin base_image), undefined)
IMAGE_BUILDER_ARGS += --base-image=$(base_image)
endif
ifneq ($(origin skip_build), undefined)
IMAGE_BUILDER_ARGS += --skip-build=$(skip_build)
endif
ifneq ($(origin verbose), undefined)
IMAGE_BUILDER_ARGS += --verbose=$(verbose)
endif
ifneq ($(origin v), undefined)
IMAGE_BUILDER_ARGS += --verbose=$(v)
endif

# Build the production Docker image using hack/config.yaml and optional overrides.
# 使用 hack/config.yaml 和可选覆盖参数构建生产 Docker 镜像。
## image: Build the production Docker image from hack/config.yaml; supports tag=v0.6.0 registry=ghcr.io/linaproai push=1
.PHONY: image
image:
	@go run ./hack/tools/image-builder $(IMAGE_BUILDER_ARGS)

# Prepare image build artifacts without invoking Docker build.
# 仅准备镜像构建产物，不执行 Docker build。
## image-build: Prepare image build artifacts from hack/config.yaml without running docker build
.PHONY: image-build
image-build:
	@go run ./hack/tools/image-builder --build-only $(IMAGE_BUILDER_ARGS)
