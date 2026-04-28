# Build Wasm Tool

`build-wasm` builds runtime-managed dynamic plugin artifacts from source plugin directories. It packages the plugin manifest, frontend assets, i18n resources, API doc i18n resources, SQL resources, route contracts, hook specs, resource specs, and the optional `wasip1/wasm` guest runtime into one `.wasm` delivery artifact.

## Usage

Preferred repository entry point:

```bash
cd apps/lina-plugins
make wasm
make wasm p=plugin-demo-dynamic
make wasm p=plugin-demo-dynamic out=../../temp/output
```

Direct tool invocation:

```bash
go run ./hack/tools/build-wasm \
  --plugin-dir apps/lina-plugins/plugin-demo-dynamic \
  --output-dir temp/output
```

## Options

| Option | Required | Description |
| --- | --- | --- |
| `--plugin-dir` | Yes | Source plugin directory that contains `plugin.yaml`. |
| `--output-dir` | No | Directory for the generated artifact. Defaults to repository `temp/output/` when the tool runs inside this repository. |

## Output

- The final dynamic plugin artifact is written as `<output-dir>/<plugin-id>.wasm`.
- If the plugin root contains `main.go`, the guest runtime is built as `wasip1/wasm` under an internal workspace below the output directory before it is embedded into the artifact.

## Notes

- The tool is for dynamic plugins. Use the `apps/lina-plugins/Makefile` wrapper for normal development so only `type: dynamic` plugins are selected.
- The command requires a working Go toolchain that can build `GOOS=wasip1 GOARCH=wasm`.
