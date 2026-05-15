## Why

当前仓库需要新增一个 `media` 源码插件，用于承载媒体策略、设备/租户策略绑定和流别名管理能力。用户提供的 `media_v2.md` 是 MySQL 表结构，需要转换为 PostgreSQL 并推导出可操作的业务管理逻辑。

## What Changes

- 新增 `apps/lina-plugins/media` 源码插件，显式接入源码插件注册入口。
- 将 `media_v2.md` 中的 MySQL 表结构转换为插件自有 PostgreSQL 安装 SQL 与卸载 SQL。
- 提供媒体策略 CRUD、全局策略切换、设备策略绑定、租户策略绑定、租户设备策略绑定、策略解析预览和流别名 CRUD 接口。
- 提供中文-only 的插件前端管理页，覆盖策略、绑定与流别名管理。
- 新增插件自有 E2E 冒烟用例，验证测试发现链路。
- 本模块按用户明确要求不维护运行时 i18n JSON 或 apidoc i18n JSON，所有可见文案直接使用中文。

## Capabilities

### New Capabilities

- `media-plugin`: 管理媒体策略、策略绑定、策略解析和流别名。

### Modified Capabilities

- 无。

## Impact

- 新增源码插件目录：`apps/lina-plugins/media/`。
- 更新源码插件聚合模块：`apps/lina-plugins/go.mod`、`apps/lina-plugins/lina-plugins.go`、`go.work`。
- 新增插件 SQL、后端 API/Controller/Service、前端页面和插件自有 E2E 文件。
- 不引入新的缓存机制；运行时数据以 PostgreSQL 为权威数据源。
