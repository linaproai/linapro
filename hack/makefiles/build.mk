# Lina Build Target
# =================

HOST_BINARY_NAME         := lina
HOST_BINARY_PATH         := $(OUTPUT_DIR)/$(HOST_BINARY_NAME)
verbose ?= 0
ifneq ($(origin v), undefined)
verbose := $(v)
endif

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

## build: 构建单体二进制与插件产物（输出到 temp/output，可追加 verbose=1 或 v=1 查看详细日志）
.PHONY: build
build:
	@echo "准备构建输出目录..."
	@rm -rf $(OUTPUT_DIR)
	@mkdir -p $(OUTPUT_DIR)
	@echo "构建前端..."
	$(call run_build_command,cd $(FRONTEND_DIR) && pnpm run build)
	@rm -rf $(EMBED_DIR)/*
	@mkdir -p $(EMBED_DIR)
	@cp -r $(FRONTEND_DIR)/apps/web-antd/dist/* $(EMBED_DIR)/
	@echo "✓ 宿主前端嵌入资源已生成"
	@./hack/scripts/prepare-packed-assets.sh
	@echo "✓ 宿主 manifest 嵌入资源已生成"
	@echo "构建动态插件产物..."
	$(call run_build_command,$(MAKE) wasm verbose=$(verbose))
	@echo "✓ 动态插件产物已生成"
	@echo "构建后端（嵌入前端静态文件）..."
	$(call run_build_command,cd $(BACKEND_DIR) && go build -o ../../$(HOST_BINARY_PATH) .)
	@echo "✓ 构建完成: $(HOST_BINARY_PATH)"

## wasm: 编译 apps/lina-plugins 下全部或指定 runtime wasm 插件到 temp/output，可追加 verbose=1 或 v=1 查看详细日志
.PHONY: wasm
wasm:
	@mkdir -p $(OUTPUT_DIR)
	$(call run_build_command,$(MAKE) -C apps/lina-plugins wasm p="$(p)" out="../../$(OUTPUT_DIR)")
