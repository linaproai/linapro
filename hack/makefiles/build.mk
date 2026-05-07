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
## build: Build frontend assets, host manifest assets, runtime wasm plugins, and host binaries using hack/config.yaml or config=<path>; supports platforms=linux/amd64,linux/arm64
.PHONY: build
build:
	@set -e; \
	eval "$$(go run ./hack/tools/image-builder --print-build-env $(BUILD_CONFIG_ARGS))"; \
	make_cmd="$$(command -v make)"; \
	run_in_dir() { \
		workdir="$$1"; \
		shift; \
		if [ "$(verbose)" = "1" ]; then \
			(cd "$$workdir" && "$$@"); \
		else \
			log_file=$$(mktemp -t lina-build.XXXXXX); \
			if (cd "$$workdir" && "$$@") >"$$log_file" 2>&1; then \
				rm -f "$$log_file"; \
			else \
				cat "$$log_file"; \
				rm -f "$$log_file"; \
				exit 1; \
			fi; \
		fi; \
	}; \
	echo "Preparing build output directory..."; \
	rm -rf "$$BUILD_OUTPUT_DIR"; \
	mkdir -p "$$BUILD_OUTPUT_DIR"; \
	echo "Building frontend..."; \
	run_in_dir "$(FRONTEND_DIR)" pnpm run build; \
	rm -rf $(EMBED_DIR)/*; \
	mkdir -p $(EMBED_DIR); \
	cp -r $(FRONTEND_DIR)/apps/web-antd/dist/* $(EMBED_DIR)/; \
	echo "✓ Host frontend embedded assets generated"; \
	./hack/scripts/prepare-packed-assets.sh; \
	echo "✓ Host manifest embedded assets generated"; \
	echo "Building dynamic plugin artifacts..."; \
	run_in_dir "." "$$make_cmd" -C apps/lina-plugins wasm out="$$PWD/$$BUILD_OUTPUT_DIR"; \
	echo "✓ Dynamic plugin artifacts generated"; \
	for target_platform in $$BUILD_PLATFORMS; do \
		target_os="$${target_platform%%/*}"; \
		target_arch="$${target_platform##*/}"; \
		if [ "$$BUILD_MULTI_PLATFORM" = "1" ]; then \
			target_binary_path="$$BUILD_OUTPUT_DIR/$${target_os}_$${target_arch}/$$BUILD_BINARY_NAME"; \
		else \
			target_binary_path="$$BUILD_BINARY_PATH"; \
		fi; \
		mkdir -p "$$(dirname "$$target_binary_path")"; \
		echo "Building backend with embedded frontend assets for $$target_os/$$target_arch..."; \
		run_in_dir "$(BACKEND_DIR)" env CGO_ENABLED="$$BUILD_CGO_ENABLED" GOOS="$$target_os" GOARCH="$$target_arch" go build -o "../../$$target_binary_path" .; \
		echo "✓ Build complete: $$target_binary_path"; \
	done

# Build runtime Wasm plugin artifacts into the shared output directory.
# 将 runtime Wasm 插件产物构建到共享输出目录。
## wasm: Build all runtime wasm plugins under apps/lina-plugins, or use p=<plugin-id> for one plugin; outputs to temp/output, use verbose=1 or v=1 for detailed logs
.PHONY: wasm
wasm:
	@mkdir -p $(OUTPUT_DIR)
	$(call run_build_command,$(MAKE) -C apps/lina-plugins wasm p="$(p)" out="../../$(OUTPUT_DIR)")
