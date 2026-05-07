# Hack Tools

该目录用于存放仓库级的独立开发工具，这些工具通常以可执行程序的形式维护。

## 目录内容

| 目录 | 用途 |
| --- | --- |
| `build-wasm/` | 从源码插件构建动态插件`Wasm`运行时产物。 |
| `image-builder/` | 根据根目录`hack/config.yaml`构建并按需推送单平台或多平台生产`Docker`镜像。 |
| `runtime-i18n/` | 扫描运行时可见硬编码文案，并校验宿主/插件`i18n`语言包`key`覆盖。 |

## 放置规则

- 当开发工具通过`go run`等命令执行、需要独立`go.mod`或需要聚焦的内部包结构时，应放在`hack/tools/`下。
- 短小的`Shell`、`PowerShell`或`Python`自动化脚本应放在`hack/scripts/`下；长期维护的校验工具如需更强类型、测试和仓库集成，应迁移到`hack/tools/`下。
- `Makefile`拆分片段应放在`hack/makefiles/`下。
- 验证资产与端到端测试代码应放在`hack/tests/`下。
- `hack/tools/`下的每个工具目录都必须同时维护`README.md`与`README.zh_CN.md`，说明用途、参数、示例、输出和验证注意事项。

## 维护说明

- 每个工具都应保持自包含，避免把工具内部实现重新耦合回运行时服务包。
- 当工具目录发生变化时，需要同步更新`go.work`、仓库根`Makefile`以及相关测试入口。
- 框架升级与源码插件升级现在由 `.claude/skills/lina-upgrade/` 下的 `lina-upgrade` 技能处理，不再由本目录中的独立 Go 工具承载。
