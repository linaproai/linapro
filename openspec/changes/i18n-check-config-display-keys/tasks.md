## 1. 门禁实现

- [x] **1.1** 在 `hack/tools/linactl/internal/runtimei18n` 实现 config 展示元数据 key 收集与覆盖校验
- [x] **1.2** 将校验接入 `validateRuntimeI18NMessages` / `i18n.check` 输出文案
- [x] **1.3** 补充 `runtimei18n` 单元测试（宿主 SQL/常量与插件 SysConfigKey 缺译失败、齐全通过）

## 2. 规则与文档

- [x] **2.1** 更新 `.agents/rules/i18n.md`：约定 `config.<key>.name/remark` 与 `i18n.check` 覆盖要求
- [x] **2.2** 更新 linactl README（中英文）中 `i18n.check` 能力列表

## 3. 缺口修复

- [x] **3.1** 对齐宿主 `sys.cron.shell.enabled` / `sys.cron.log.retention` 的 i18n 键并清理错误键
- [x] **3.2** 为声明 `SysConfigKey` 且启用 i18n 的官方插件补齐 `config.plugin.*` 展示翻译

## 4. 验证

- [x] **4.1** 运行 `make i18n.check` 与 `runtimei18n` 相关 Go 测试并通过
- [x] **4.2** `openspec validate i18n-check-config-display-keys --strict`

## Feedback

- [x] **FB-1**: 参数设置页出现类 i18n key 参数名，且原 `i18n.check` 无法发现 — 扩展检查并补齐资源
