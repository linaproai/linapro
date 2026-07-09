## Purpose
定义代码拥有的源文本命名空间注册机制，确保缺失翻译检查通过注册表识别业务源文本。
## Requirements
### Requirement:i18n 包不得再依赖特定业务模块命名空间前缀

宿主系统 SHALL NOT 允许 `i18n` 包中的任何函数硬编码 `job.handler.`、`job.group.default.` 或其他特定业务模块命名空间前缀。源码文案兜底由拥有模块调用 `Translate(ctx, key, sourceText)` 时自行提供，`i18n` 基础服务不维护业务命名空间注册表，也不通过缺失检查或导出诊断反向判断业务键归属。

#### Scenario:删除 i18n 包对 jobmgmt 的反向依赖

- **当** 审查 `apps/lina-core/internal/service/i18n/` 中的任何源文件时
- **则** 不存在带业务模块特定前缀如 `job.handler.` 或 `job.group.default.` 的硬编码字符串
- **且** 文件不再使用命名空间注册表进行缺失检查判断

