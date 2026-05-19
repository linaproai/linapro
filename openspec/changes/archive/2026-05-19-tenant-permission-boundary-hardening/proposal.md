## Why

`temp/tenant-permission-audit/tenant-permission-audit.md` 确认租户态权限校验存在多处边界缺口：角色授权可以分配平台租户管理权限，菜单、插件策略和任务分组等全局或租户敏感资源未强制校验平台上下文或租户范围。现有实现把“已登录 + 权限字符串”误当成完整授权条件，需要补齐“租户上下文 + 资源归属”的后端强约束。

## What Changes

- 平台控制面接口在校验权限字符串后继续要求平台上下文；租户上下文、代管租户上下文或异常持有平台权限的租户角色均不得调用平台控制面 API。
- 角色授权树和角色创建/更新服务端提交共同使用“当前操作者可分配菜单/权限集合”，租户角色不得获得平台租户管理、平台插件治理和全局菜单治理写权限。
- `sys_menu` 作为全局权限模型时，菜单创建、更新、删除、状态变更等写操作必须要求平台上下文；租户态仅能读取按上下文过滤后的授权投影。
- 插件安装、卸载、启用、禁用、同步、上传、安装模式/新租户自动启用策略等平台治理动作必须要求平台上下文；租户插件自服务继续通过租户插件接口按当前租户执行。
- 任务分组按 `tenant_id` 作为租户资源处理，列表、创建、更新、删除、任务计数和删除迁移均必须限制在当前租户范围内。
- 参数设置和字典的 fallback 行返回来源与动作元数据，前端按 `canEdit` / `canOverride` 隐藏或切换操作，避免“可见但编辑必失败”。
- 补齐后端业务错误、运行时语言包、apidoc i18n、前端类型与 E2E/单元测试覆盖。

## Capabilities

### New Capabilities

- `tenant-platform-access-control`: 定义平台租户控制面 API 的平台上下文访问边界。

### Modified Capabilities

- `role-management`: 角色授权树和角色菜单提交必须执行上下文可分配权限过滤。
- `menu-management`: 菜单治理写操作必须受平台上下文保护，租户态授权树按上下文过滤。
- `plugin-permission-governance`: 平台插件治理权限和租户插件自服务权限必须分离。
- `plugin-manifest-lifecycle`: 插件同步、上传、安装、卸载、启停、安装模式和租户供应策略等生命周期治理必须要求平台上下文。
- `cron-job-management`: 任务分组必须作为租户作用域资源读写和迁移。
- `config-management`: 租户 fallback 参数行必须携带来源与可操作性元数据。
- `dict-management`: 租户 fallback 字典类型和字典数据必须携带来源与可操作性元数据。

## Impact

- 后端影响：`apps/lina-core/internal/service/menu`、`role`、`plugin`、`jobmgmt`、`sysconfig`、`dict`，以及 `apps/lina-plugins/multi-tenant/backend/internal/service/tenant` 等平台租户接口服务。
- API/DTO 影响：角色授权树、菜单树、配置列表/详情、字典类型/数据列表、插件治理写接口、任务分组接口可能新增错误码或响应字段；只读接口保持 RESTful GET 语义。
- 前端影响：角色授权抽屉、菜单管理、插件管理、任务分组、参数设置、字典管理的按钮可见性、错误提示、类型定义和 i18n 文案。
- 数据影响：优先复用 `sys_job_group.tenant_id` 和现有租户字段，不引入新的菜单租户自定义模型；如发现旧任务分组 `tenant_id=0` 数据需要迁移，必须按迁移计划处理。
- 缓存影响：权限拓扑、插件状态、配置/字典 fallback 缓存必须使用现有显式作用域失效和集群协调机制，禁止普通业务路径无理由全量清空。
