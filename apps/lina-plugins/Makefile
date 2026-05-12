# wasm: Build all dynamic wasm plugins, or use p=<plugin-id> for one plugin
# wasm: 构建全部或指定 dynamic wasm 插件；可通过 `p=<plugin-id>` 定向构建
.PHONY: wasm
wasm:
	@go run ../../hack/tools/linactl wasm p="$(p)" out="$(out)"
