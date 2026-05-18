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
ifneq ($(origin platforms), undefined)
IMAGE_BUILDER_ARGS += --platforms=$(platforms)
endif
ifneq ($(origin cgo_enabled), undefined)
IMAGE_BUILDER_ARGS += --cgo-enabled=$(cgo_enabled)
endif
ifneq ($(origin output_dir), undefined)
IMAGE_BUILDER_ARGS += --output-dir=$(output_dir)
endif
ifneq ($(origin binary_name), undefined)
IMAGE_BUILDER_ARGS += --binary-name=$(binary_name)
endif
ifneq ($(origin base_image), undefined)
IMAGE_BUILDER_ARGS += --base-image=$(base_image)
endif
ifneq ($(origin config), undefined)
IMAGE_BUILDER_ARGS += --config=$(config)
endif
ifneq ($(origin verbose), undefined)
IMAGE_BUILDER_ARGS += --verbose=$(verbose)
endif
ifneq ($(origin v), undefined)
IMAGE_BUILDER_ARGS += --verbose=$(v)
endif
ifneq ($(origin plugins), undefined)
IMAGE_BUILDER_ARGS += plugins=$(plugins)
endif

# Build the production Docker image from the standard make build output.
# 基于标准 make build 产物构建生产 Docker 镜像。
## image: Build the production Docker image from make build output and hack/config.yaml or config=<path>; supports tag=v0.6.0 registry=ghcr.io/linaproai push=1 platforms=linux/amd64,linux/arm64
.PHONY: image
image:
	@$(LINACTL) image $(IMAGE_BUILDER_ARGS)

# Prepare image build artifacts without invoking Docker build.
# 仅准备镜像构建产物，不执行 Docker build。
## image.build: Stage image build artifacts from make build output without running docker build
.PHONY: image.build
image.build:
	@$(LINACTL) image.build $(IMAGE_BUILDER_ARGS)
