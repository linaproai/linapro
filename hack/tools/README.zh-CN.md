# Hack Tools

该目录用于存放仓库级开发工具。长期维护的`LinaPro`开发命令默认收敛到`linactl`中，只有确需聚焦独立模块边界的工具才继续单独维护。

## 目录内容

| 目录 | 用途 |
| --- | --- |
| `linactl/` | 提供跨平台`LinaPro`开发命令入口，并作为`Makefile`与`Windows make.cmd`薄包装的统一实现，包含镜像构建、动态插件`Wasm`产物构建和运行时`i18n`治理检查能力。 |

## 放置规则

- 当开发工具通过`go run`等命令执行、需要独立`go.mod`或需要聚焦的内部包结构时，应放在`hack/tools/`下。
- 仅服务`LinaPro`的长期维护仓库命令应优先新增到`hack/tools/linactl/internal/<component>`，并通过`linactl`命令包装暴露。
- 短小的`Shell`、`PowerShell`或`Python`自动化脚本应放在`hack/scripts/`下；长期维护的校验工具如需更强类型、测试和仓库集成，应迁移到`hack/tools/`下。
- `Makefile`拆分片段应放在`hack/makefiles/`下。
- 验证资产与端到端测试代码应放在`hack/tests/`下。
- `hack/tools/`下的每个独立工具目录都必须同时维护`README.md`与`README.zh-CN.md`，说明用途、参数、示例、输出和验证注意事项。

## 维护说明

- 每个工具都应保持自包含，避免把工具内部实现重新耦合回运行时服务包。
- 当工具目录发生变化时，需要同步更新`go.work`、仓库根`Makefile`以及相关测试入口。
- 工具生命周期说明应维护在各工具自己的 `README` 中，避免在本目录记录已经删除的命令路径。
