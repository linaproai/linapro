## Why

当前默认交付同时维护简体中文、繁体中文和英文三套 i18n 资源，增加了宿主、插件、前端运行时语言包和 API 文档资源的同步成本。项目默认只需要保留英文和简体中文，因此需要移除繁体中文默认资源，降低后续内建能力和插件示例的 i18n 维护复杂度。

## What Changes

- **BREAKING**: 默认交付不再提供 `zh-TW` 繁体中文运行时语言、插件 manifest 语言包或 API 文档翻译资源。
- 默认配置中的 `i18n.locales` 仅保留 `en-US` 和 `zh-CN`。
- 默认管理工作台和共享前端语言包仅保留 `en-US` 和 `zh-CN` 静态资源。
- 移除或调整以 `zh-TW` 为目标的 E2E/单元测试断言，保留英文和简体中文语言治理检查。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `framework-i18n-foundation`: 默认内置语言从 `zh-CN`、`en-US`、`zh-TW` 收敛为 `zh-CN` 和 `en-US`，并移除繁体中文运行时语言列表、页面内容、API 文档和测试验收要求。
- `management-workbench-i18n`: 中文浏览器语言标签（包括 `zh-TW`）首次访问时继续统一回退到 `zh-CN`，但默认工作台不再提供 `zh-TW` 静态语言包。

## Impact

- 影响宿主 `apps/lina-core/manifest/i18n` 资源目录和默认配置模板。
- 影响源码插件 `apps/lina-plugins/*/manifest/i18n` 资源目录。
- 影响默认管理工作台和共享前端语言包 `apps/lina-vben/**/locales`。
- 影响繁体中文专项 E2E、前端单元测试、后端 i18n 相关测试和 i18n 静态检查脚本。
- 不新增 REST API、数据库 schema、SQL seed、权限边界或运行时缓存机制。
