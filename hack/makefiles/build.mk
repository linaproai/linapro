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
ifneq ($(origin plugins), undefined)
BUILD_CONFIG_ARGS += plugins=$(plugins)
endif

WASM_ARGS := p="$(p)" out="$(abspath $(OUTPUT_DIR))" verbose=$(verbose)
ifneq ($(origin dry_run), undefined)
WASM_ARGS += dry_run=$(dry_run)
endif
ifneq ($(origin dry-run), undefined)
WASM_ARGS += dry-run=$(dry-run)
endif

# Build frontend assets, packed manifests, dynamic plugins, and the host binary.
# 构建前端资源、嵌入 manifest、动态插件和宿主后端二进制。
## build: Build host frontend, manifest assets, and host binaries; auto-enables official plugins when apps/lina-plugins contains manifests, or use plugins=0 for host-only
.PHONY: build
build:
	@$(LINACTL) build $(BUILD_CONFIG_ARGS) verbose=$(verbose)

# Prepare host manifest assets for Go embedding.
# 准备宿主 manifest 资源，用于 Go 嵌入。
## pack.assets: Prepare host manifest assets for embedding
.PHONY: pack.assets
pack.assets:
	@$(LINACTL) pack.assets

# Build runtime Wasm plugin artifacts into the shared output directory.
# 将 runtime Wasm 插件产物构建到共享输出目录。
## wasm: Build runtime wasm plugins from the official plugin submodule, or use p=<plugin-id> for one plugin; outputs to temp/output, use verbose=1 or v=1 for detailed logs
.PHONY: wasm
wasm:
	@$(LINACTL) wasm $(WASM_ARGS)
