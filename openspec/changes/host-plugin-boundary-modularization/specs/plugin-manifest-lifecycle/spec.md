## ADDED Requirements

### Requirement: 官方源码插件使用领域-能力式插件 ID

系统 SHALL 为官方源码插件使用不带 `plugin-` 前缀的领域-能力式 `kebab-case` 标识，以提升可读性并避免语义重复。

#### Scenario: 定义官方源码插件标识
- **WHEN** 团队为开源阶段的官方源码插件命名
- **THEN** 插件 ID 使用 `org-center`、`content-notice`、`monitor-online`、`monitor-server`、`monitor-operlog`、`monitor-loginlog`
- **AND** 不要求使用 `plugin-` 前缀

#### Scenario: 校验插件 ID 合法性
- **WHEN** 宿主解析上述官方插件的 `plugin.yaml`
- **THEN** 这些插件 ID 只需满足全局唯一与 `kebab-case` 规则
- **AND** 不因缺少 `plugin-` 前缀而被视为非法

### Requirement: 源码插件菜单必须挂载到宿主稳定目录

系统 SHALL 要求官方源码插件在 manifest 菜单声明中通过 `parent_key` 指向宿主稳定目录键，保证后台导航结构长期稳定。

#### Scenario: 组织与内容插件声明父级目录
- **WHEN** `org-center` 或 `content-notice` 声明菜单元数据
- **THEN** 其顶层菜单 `parent_key` 分别指向宿主目录键 `org` 与 `content`
- **AND** 插件内部子菜单仍可继续引用同插件已声明的父菜单 key

#### Scenario: 监控插件声明父级目录
- **WHEN** `monitor-online`、`monitor-server`、`monitor-operlog`、`monitor-loginlog` 声明菜单元数据
- **THEN** 其顶层菜单 `parent_key` 指向宿主目录键 `monitor`
- **AND** 宿主按该父级键完成菜单同步与启停联动可见性治理

#### Scenario: 官方插件使用固定父级目录键映射
- **WHEN** 宿主校验官方源码插件 manifest
- **THEN** `org-center` 的顶层 `parent_key` 必须为 `org`
- **AND** `content-notice` 的顶层 `parent_key` 必须为 `content`
- **AND** `monitor-online`、`monitor-server`、`monitor-operlog`、`monitor-loginlog` 的顶层 `parent_key` 必须为 `monitor`

#### Scenario: 官方插件声明了不受支持的顶层挂载键
- **WHEN** 上述官方源码插件在其顶层菜单声明中使用了与约定不一致的 `parent_key`
- **THEN** 宿主拒绝同步该插件菜单
- **AND** 向管理员提供可诊断的挂载校验错误

### Requirement: 源码插件后端目录结构必须收敛到 backend/internal

系统 SHALL 要求源码插件将后端业务实现收敛在 `backend/internal/` 下，避免在 `backend/` 根目录直接暴露业务服务目录，保证插件私有实现边界清晰且与宿主约定一致。

#### Scenario: 规划源码插件标准目录
- **WHEN** 团队创建或重构一个源码插件
- **THEN** 插件后端至少按 `backend/api/`、`backend/plugin.go`、`backend/internal/controller/`、`backend/internal/service/` 组织
- **AND** 插件前端页面保留在 `frontend/pages/`
- **AND** 插件清单与嵌入资源保留在 `plugin.yaml`、`plugin_embed.go`、`manifest/sql/` 与 `manifest/sql/uninstall/`

#### Scenario: 放置插件 service 组件
- **WHEN** 团队为源码插件新增或迁移业务服务
- **THEN** 所有 service 组件 MUST 放在 `backend/internal/service/<component>/`
- **AND** 不得再创建 `backend/service/<component>/`
- **AND** `backend/provider/` 等非 `internal` 目录仅用于稳定 capability provider / adapter，不承载主要业务编排

#### Scenario: 插件需要本地 ORM 工件
- **WHEN** 源码插件需要访问数据库
- **THEN** `backend/hack/config.yaml` 作为该插件本地 `gf gen dao` 配置入口
- **AND** 生成结果落在 `backend/internal/dao/`、`backend/internal/model/do/` 与 `backend/internal/model/entity/`
- **AND** 对宿主共享表的访问也继续沿用该插件本地生成工件
