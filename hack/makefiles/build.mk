# LinaPro Build Commands
# LinaPro 构建目标
# =================

HOST_BINARY_NAME         := lina
HOST_BINARY_PATH         := $(OUTPUT_DIR)/$(HOST_BINARY_NAME)
verbose ?= 0
ifneq ($(origin v), undefined)
verbose := $(v)
endif

# Helper macro that optionally hides noisy build command output.
# 构建命令辅助宏，可按需隐藏详细输出。
define run_build_command
	@set -e; \
	if [ "$(verbose)" = "1" ]; then \
		$(1); \
	else \
		log_file=$$(mktemp -t lina-build.XXXXXX); \
		if $(1) >"$$log_file" 2>&1; then \
			rm -f "$$log_file"; \
		else \
			cat "$$log_file"; \
			rm -f "$$log_file"; \
			exit 1; \
		fi; \
	fi
endef

# Build frontend assets, packed manifests, dynamic plugins, and the host binary.
# 构建前端资源、嵌入 manifest、动态插件和宿主后端二进制。
## build: Build frontend assets, host manifest assets, runtime wasm plugins, and the host single binary into temp/output; use verbose=1 or v=1 for detailed logs
.PHONY: build
build:
	@echo "Preparing build output directory..."
	@rm -rf $(OUTPUT_DIR)
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building frontend..."
	$(call run_build_command,cd $(FRONTEND_DIR) && pnpm run build)
	@rm -rf $(EMBED_DIR)/*
	@mkdir -p $(EMBED_DIR)
	@cp -r $(FRONTEND_DIR)/apps/web-antd/dist/* $(EMBED_DIR)/
	@echo "✓ Host frontend embedded assets generated"
	@./hack/scripts/prepare-packed-assets.sh
	@echo "✓ Host manifest embedded assets generated"
	@echo "Building dynamic plugin artifacts..."
	$(call run_build_command,$(MAKE) wasm verbose=$(verbose))
	@echo "✓ Dynamic plugin artifacts generated"
	@echo "Building backend with embedded frontend assets..."
	$(call run_build_command,cd $(BACKEND_DIR) && go build -o ../../$(HOST_BINARY_PATH) .)
	@echo "✓ Build complete: $(HOST_BINARY_PATH)"

# Build runtime Wasm plugin artifacts into the shared output directory.
# 将 runtime Wasm 插件产物构建到共享输出目录。
## wasm: Build all runtime wasm plugins under apps/lina-plugins, or use p=<plugin-id> for one plugin; outputs to temp/output, use verbose=1 or v=1 for detailed logs
.PHONY: wasm
wasm:
	@mkdir -p $(OUTPUT_DIR)
	$(call run_build_command,$(MAKE) -C apps/lina-plugins wasm p="$(p)" out="../../$(OUTPUT_DIR)")
