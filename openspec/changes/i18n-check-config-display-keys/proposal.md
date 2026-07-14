# Proposal: i18n.check 覆盖参数设置展示元数据

## 上下文

参数设置页对内置与插件写入的 `sys_config` 行通过 `config.<config_key>.name` / `config.<config_key>.remark` 做运行时投影。插件经 `SetValue` 首次落库时 `name` 等于技术 key；若缺少对应翻译，管理页会直接显示类似 i18n key 的字符串。

现有 `make i18n.check` 只做 locale 对等、`bizerr` messageKey、`plugin.<id>.name/description` 与前端静态 `$t` 覆盖，**不会**校验 `sys_config` 展示键是否齐全，因此无法在 CI 中拦住上述缺口。

## 目标

1. 扩展 `linactl i18n.check` / `make i18n.check`：从宿主 SQL seed 与代码常量、以及 `i18n.enabled: true` 插件的 `SysConfigKey` 常量收集配置 key，强制各运行时 locale 具备 `config.<key>.name` 与 `config.<key>.remark`。
2. 修复宿主现有键名不一致（如 `sys.cron.*`），并补齐插件侧缺失的 config 展示翻译，使门禁可立即生效。
3. 同步 `.agents/rules/i18n.md` 与 linactl 文档，明确该契约。
