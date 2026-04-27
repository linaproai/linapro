## ADDED Requirements

### Requirement: 配置导出与导入表头必须按当前语言通过翻译键解析
系统 SHALL 让 `Export configs to Excel` 与 `Import configs from Excel` 链路中的列头(`name` / `key` / `value` / `remark` / `createdAt` / `updatedAt`)按当前请求语言通过 `config.field.<name>` 翻译键解析,不得在后端 Go 源码中维护英文/中文字面量映射表。`apps/lina-core/manifest/i18n/<locale>.json` MUST 包含 `config.field.*` 命名空间下的翻译值,新增语言时只需追加资源,不需要修改 Go 代码。

#### Scenario: 阿拉伯语环境导出包含阿拉伯语表头
- **WHEN** 管理员在 `ar-SA` 环境请求 `GET /config/export`
- **THEN** 导出的 Excel 列头按阿拉伯语展示
- **AND** 后端通过 `config.field.name`、`config.field.key`、`config.field.value`、`config.field.remark`、`config.field.createdAt`、`config.field.updatedAt` 翻译键解析得到列头
- **AND** 后端代码中不存在 `englishLabels` / `chineseLabels` 这类硬编码字面量映射

#### Scenario: 加新语言时无需修改后端 Go 代码
- **WHEN** 项目启用新的内置语言并提供该语言的 `manifest/i18n/<locale>.json` 资源
- **AND** 该资源包含 `config.field.*` 命名空间下的翻译值
- **THEN** 配置导入与导出表头自动按该语言展示
- **AND** `apps/lina-core/internal/service/sysconfig/` 内不需要任何源码改动
