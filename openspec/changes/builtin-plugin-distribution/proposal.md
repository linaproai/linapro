## Why

项目内建源码插件需要随 LinaPro 源码一起交付，并在启动时自动安装、启用和安全升级。继续要求部署者维护`plugin.autoEnable`清单会把“项目组成部分”和“部署方托管普通插件”两种语义混在一起，也会让普通插件管理入口暴露不应由页面治理的内建能力。

## What Changes

- 在插件 manifest 中新增`distribution`字段，缺省为`marketplace`，支持`builtin`声明项目内建源码插件。
- 将`distribution`同步到插件注册表和发布 manifest snapshot，列表与详情 API 返回该治理类型。
- 普通插件管理列表默认隐藏`builtin`插件；需要诊断时通过受控只读查询显式包含。
- 对`builtin`插件的安装、启用、禁用、卸载、手动升级和租户供应策略更新等普通管理写操作执行服务端拒绝。
- 新增启动期`builtin`插件收敛流程，在插件接线前自动安装、启用和安全升级内建源码插件，并复用现有生命周期、依赖、迁移、治理同步和缓存刷新路径。
- 前端插件管理页面基于服务端投影隐藏`builtin`插件及其写操作，不把内建插件作为普通管理对象展示。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-manifest-lifecycle`：插件 manifest、注册表、发布快照和管理投影新增`distribution`治理语义。
- `plugin-startup-bootstrap`：启动引导新增 manifest 驱动的内建源码插件自动安装、启用和安全升级收敛，独立于`plugin.autoEnable`。
- `plugin-upgrade-governance`：内建源码插件启动安全升级成为显式例外，并禁止普通管理入口手动升级`builtin`插件。
- `plugin-ui-integration`：插件管理列表默认隐藏`builtin`插件，详情和动作入口对`builtin`保持只读展示。

## Impact

- 影响`apps/lina-core/internal/service/plugin/internal/catalog`、`plugintypes`、`store`、`lifecycle`、`upgrade`、`management`和根插件服务编排。
- 影响`sys_plugin`数据库结构、DAO 生成输入、插件 release snapshot 序列化和启动治理快照。
- 影响`apps/lina-core/api/plugin/v1`响应 DTO、列表查询参数、控制器生成结果和前端插件管理 API 类型。
- 影响插件管理前端页面的列表过滤、详情弹窗和操作可见性。
- 需要新增`bizerr`错误码、宿主运行时错误翻译、宿主 apidoc 翻译和相关单元测试；涉及用户可观察页面变化时补充 E2E 或记录覆盖判断。
