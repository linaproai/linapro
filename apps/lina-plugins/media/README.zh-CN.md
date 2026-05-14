# media

`media` 是 LinaPro 的媒体管理源码插件，用于维护媒体策略、策略绑定、流别名、节点配置、设备节点、租户流配置和租户白名单。

本模块按用户要求使用中文-only 文案，不提供运行时 i18n JSON 或 apidoc i18n JSON。

## 能力范围

- 媒体策略增删查改与全局策略设置
- 设备、租户、租户设备三类策略绑定
- 按优先级预览设备当前生效策略
- 流别名增删查改
- 节点、设备节点与租户流配置增删查改
- 租户白名单增删查改

## 开发入口

- 后端入口：`backend/plugin.go`
- API DTO：`backend/api/media/v1/`
- 业务服务：`backend/internal/service/media/`
- 前端页面：`frontend/pages/media-management.vue`
- PostgreSQL 安装 SQL：`manifest/sql/001-media-schema.sql`
- 演示案例数据：`manifest/sql/mock-data/001-media-mock-data.sql`
