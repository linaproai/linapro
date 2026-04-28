# Upgrade Source 工具

`upgrade-source`是仓库根目录`make upgrade`背后的开发态升级命令，支持框架升级和源码插件升级治理。

## 使用方式

推荐使用仓库封装入口：

```bash
make upgrade confirm=upgrade
make upgrade confirm=upgrade scope=framework target=<tag-or-ref>
make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>
make upgrade confirm=upgrade scope=source-plugin plugin=all
make upgrade confirm=upgrade scope=source-plugin plugin=all dry_run=1
```

也可以直接调用工具：

```bash
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=framework --target=<tag-or-ref>
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=source-plugin --plugin=<plugin-id>
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=source-plugin --plugin=all --dry-run
```

## 参数

| 参数 | 适用范围 | 说明 |
| --- | --- | --- |
| `--confirm=upgrade` | 全部 | 必填确认令牌，用于显式确认升级和数据库敏感操作。 |
| `--scope` | 全部 | 升级范围，支持`framework`与`source-plugin`，默认值为`framework`。 |
| `--repo` | `framework` | 可选的上游框架`Git`仓库地址，默认读取`apps/lina-core/hack/config.yaml`中的配置。 |
| `--target` | `framework` | 可选的目标框架`tag`或`Git ref`。 |
| `--plugin` | `source-plugin` | 源码插件`ID`，或使用`all`选择所有源码插件。 |
| `--dry-run` | 全部 | 只输出解析后的计划，不执行代码覆盖、`SQL`或源码插件治理变更。 |

## 行为

- 框架升级模式会检查工作区、解析目标框架版本、输出升级计划，并在需要时同步代码和重放宿主`SQL`文件。
- 源码插件模式会列出已发现的源码插件版本，与宿主当前生效版本对比，拒绝降级场景，并在允许执行时应用已准备的源码插件`release`。
- 命令缺少`--confirm=upgrade`时会拒绝执行。

## 注意事项

- 执行框架升级前，请先提交或`stash`本地变更。工具会执行干净工作区预检查。
- 真实升级前需要先备份代码仓库和数据库。
- 检查框架目标版本或批量源码插件升级时，建议先使用`--dry-run`。
