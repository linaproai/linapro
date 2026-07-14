# Design: config 展示元数据 i18n 覆盖检查

## 决策

### 1. 期望键的来源（静态、可复现）

| 范围 | 收集方式 |
|------|----------|
| 宿主 | `apps/lina-core/manifest/sql/**/*.sql` 中 `sys_config` 初始化写入的 key 字面量；以及宿主 Go 中 `SysConfigKey` / 明确的 `sys.*` / `demo.*` 受保护常量 |
| 插件 | 仅 `plugin.yaml` 中 `i18n.enabled: true` 的插件；扫描其模块内 `hostconfigcap.SysConfigKey = "..."`（及等价赋值）字面量 |

不扫描运行时数据库，避免 CI 依赖环境状态。

### 2. 强制翻译键

与 `sysconfig` 列表投影一致：

- `config.<config_key>.name`
- `config.<config_key>.remark`

宿主键必须在宿主 `manifest/i18n/<locale>/` 出现；插件键必须在该插件 `manifest/i18n/<locale>/` 出现（运行时由宿主合并加载）。未启用 i18n 的插件按既有规则跳过。

### 3. 集成点

在 `runtimei18n` 的 `messages` 子检查中追加 `validateConfigDisplayMetadataKeys`，与 bizerr / plugin display metadata 并列，由既有 `RunCheck` 一并执行，无需新 make 目标。

### 4. 与错误键的关系

历史错误写法（如 `config.sys.cron.shellEnabled`）若仍存在于资源中，locale 对等检查仍会要求各语言一致，但**不会**替代正确键；正确键缺失即失败。清理错误键为配套修复，不作为检查器职责。

## 风险

- 首次启用会暴露大量插件缺译；本变更同步补齐官方插件翻译以保持 CI 绿。
- 动态拼接的 key（非常量）无法静态收集；约定插件必须以常量声明 `SysConfigKey`。
