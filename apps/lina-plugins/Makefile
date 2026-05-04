# wasm: Build all dynamic wasm plugins, or use p=<plugin-id> for one plugin
# wasm: 构建全部或指定 dynamic wasm 插件；可通过 `p=<plugin-id>` 定向构建
.PHONY: wasm
wasm:
	@set -e; \
	repo_root="$$(cd ../.. && pwd)"; \
	builder_dir="$$repo_root/hack/tools/build-wasm"; \
	output_dir="$(out)"; \
	if [ -z "$$output_dir" ]; then \
		output_dir="$$repo_root/temp/output"; \
	elif [ "$${output_dir#/}" = "$$output_dir" ]; then \
		output_dir="$$PWD/$$output_dir"; \
	fi; \
	plugin_list=""; \
	if [ -n "$(p)" ]; then \
		plugin_list="$(p)"; \
	else \
		plugin_list=$$(find . -mindepth 2 -maxdepth 2 -name plugin.yaml | sort | while read -r manifest; do \
			if grep -Eq '^type:[[:space:]]*dynamic[[:space:]]*$$' "$$manifest"; then \
				basename "$$(dirname "$$manifest")"; \
			fi; \
		done); \
	fi; \
	if [ -z "$$plugin_list" ]; then \
		echo "No buildable dynamic wasm plugins found"; \
		exit 0; \
	fi; \
	for plugin in $$plugin_list; do \
		manifest="./$$plugin/plugin.yaml"; \
		if [ ! -f "$$manifest" ]; then \
			echo "Plugin does not exist: $$plugin"; \
			exit 1; \
		fi; \
		if ! grep -Eq '^type:[[:space:]]*dynamic[[:space:]]*$$' "$$manifest"; then \
			echo "Plugin is not dynamic and cannot be built as wasm: $$plugin"; \
			exit 1; \
		fi; \
		echo "Building dynamic wasm plugin: $$plugin"; \
		(cd "$$builder_dir" && go run . --plugin-dir "$$repo_root/apps/lina-plugins/$$plugin" --output-dir "$$output_dir"); \
	done
