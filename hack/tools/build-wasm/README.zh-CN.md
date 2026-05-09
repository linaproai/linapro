# Build Wasm 工具

`build-wasm`用于从源码插件目录构建运行时托管的动态插件产物。它会把插件清单、前端资源、`i18n`资源、接口文档`i18n`资源、`SQL`资源、路由合同、钩子规格、资源规格以及可选的`wasip1/wasm guest runtime`打包成一个`.wasm`交付产物。

## 使用方式

推荐使用仓库封装入口：

```bash
cd apps/lina-plugins
make wasm
make wasm p=plugin-demo-dynamic
make wasm p=plugin-demo-dynamic out=../../temp/output
```

也可以直接调用工具：

```bash
go run ./hack/tools/build-wasm \
  --plugin-dir apps/lina-plugins/plugin-demo-dynamic \
  --output-dir temp/output
```

## 参数

| 参数 | 是否必填 | 说明 |
| --- | --- | --- |
| `--plugin-dir` | 是 | 包含`plugin.yaml`的源码插件目录。 |
| `--output-dir` | 否 | 生成产物目录。在当前仓库内运行且未指定时，默认写入仓库根目录下的`temp/output/`。 |

## 输出

- 最终动态插件产物写入`<output-dir>/<plugin-id>.wasm`。
- 如果插件根目录包含`main.go`，工具会先在输出目录下的内部工作区构建`wasip1/wasm guest runtime`，再将其嵌入最终产物。

## 注意事项

- 该工具面向动态插件。日常开发优先使用`apps/lina-plugins/Makefile`封装入口，以确保只选择`type: dynamic`插件。
- 命令需要当前`Go`工具链支持构建`GOOS=wasip1 GOARCH=wasm`。
