# LinaPro Build Commands
# LinaPro 构建目标
# =================

HOST_BINARY_NAME         := lina
HOST_BINARY_PATH         := $(OUTPUT_DIR)/$(HOST_BINARY_NAME)
BUILD_CONFIG_ARGS        :=
verbose ?= 0
ifneq ($(origin v), undefined)
verbose := $(v)
endif

ifneq ($(origin platforms), undefined)
BUILD_CONFIG_ARGS += --platforms=$(platforms)
endif
ifneq ($(origin cgo_enabled), undefined)
BUILD_CONFIG_ARGS += --cgo-enabled=$(cgo_enabled)
endif
ifneq ($(origin output_dir), undefined)
BUILD_CONFIG_ARGS += --output-dir=$(output_dir)
endif
ifneq ($(origin binary_name), undefined)
BUILD_CONFIG_ARGS += --binary-name=$(binary_name)
endif
ifneq ($(origin config), undefined)
BUILD_CONFIG_ARGS += --config=$(config)
endif

# Build frontend assets, packed manifests, dynamic plugins, and the host binary.
# 构建前端资源、嵌入 manifest、动态插件和宿主后端二进制。
## build: Build frontend assets, host manifest assets, runtime wasm plugins, and host binaries using hack/config.yaml or config=<path>; supports platforms=linux/amd64,linux/arm64
.PHONY: build
build:
	@go run ./hack/tools/linactl build $(BUILD_CONFIG_ARGS) verbose=$(verbose)

# Build runtime Wasm plugin artifacts into the shared output directory.
# 将 runtime Wasm 插件产物构建到共享输出目录。
## wasm: Build all runtime wasm plugins under apps/lina-plugins, or use p=<plugin-id> for one plugin; outputs to temp/output, use verbose=1 or v=1 for detailed logs
.PHONY: wasm
wasm:
	@go run ./hack/tools/linactl wasm p="$(p)" out="$(OUTPUT_DIR)" verbose=$(verbose)
