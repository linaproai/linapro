# Runtime I18n 工具

`runtime-i18n`提供仓库级运行时国际化校验能力。它会扫描源码中的高风险运行时可见硬编码文案，报告生成文件和测试 fixture 的分类统计，并校验宿主与插件范围内的运行时`i18n JSON`的`key`覆盖。

## 使用方式

推荐使用仓库封装入口：

```bash
make i18n.check
```

也可以直接调用工具：

```bash
go run ./hack/tools/runtime-i18n scan
go run ./hack/tools/runtime-i18n scan --format json
go run ./hack/tools/runtime-i18n messages
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `scan` | 扫描`Go`、`Vue`、`TypeScript`文件中的高风险运行时可见硬编码文案。 |
| `messages` | 校验宿主与插件运行时`i18n JSON`的`key`覆盖，并检查重复运行时`key`。 |

`make i18n.check`会同时运行上述两个直接工具命令。

`scan`命令会阻断未被 allowlist 接受的运行时源码命中项，同时输出生成文件和测试 fixture 的非阻断统计，便于审查记录区分源码违规和可接受的生成/测试数据。

## 扫描参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--format` | `text` | 输出格式，支持`text`与`json`。 |
| `--allowlist` | `hack/tools/runtime-i18n/allowlist.json` | 用于记录已接受扫描项的`JSON allowlist`文件。 |

`JSON`输出使用结构化报告：

```json
{
  "summary": {
    "violations": 0,
    "violationFiles": 0,
    "allowlistHits": 0,
    "generatedFiles": 0,
    "generatedItems": 0,
    "testFixtureFiles": 0,
    "testFixtureItems": 0,
    "byCategory": {}
  },
  "findings": [],
  "allowlistHits": []
}
```

## Allowlist 格式

每个接受项都必须记录路径、规则、分类、保留原因和适用范围。只有当接受范围明确限定到某一行时才填写`line`。

```json
{
  "entries": [
    {
      "path": "apps/lina-core/internal/service/example/example.go",
      "rule": "go-string-literal-han",
      "category": "Unclassified",
      "reason": "Stable user data fixture that is not rendered as system copy.",
      "scope": "single fixture value",
      "line": 42
    }
  ]
}
```

## 退出码

- `0`：所选检查通过。
- `1`：工具执行失败、所选检查发现问题或参数非法。

通过`make`调用时，`GNU Make`会把工具的非零退出码表现为`Makefile`执行失败。

## 注意事项

- 运行时 JSON 校验只读取`manifest/i18n/<locale>/*.json`下的直接文件，不递归进入`apidoc/`目录。
- 每个`allowlist`条目都必须包含路径、规则、分类、原因和适用范围。
