## Why

当前源码插件和动态插件仍存在直接生成或访问宿主`sys_*`核心表的路径，导致插件契约与宿主存储模型、权限模型和缓存状态强耦合。为了让`LinaPro`作为面向可持续交付的`AI`原生全栈框架具备稳定插件扩展边界，需要一次性把插件访问宿主数据的入口收敛为领域能力接口和受治理`hostServices`协议。

## What Changes

- **BREAKING**：插件生产代码不得直接生成、导入或查询宿主核心`sys_*`表、宿主`DAO/DO/Entity`、私有缓存快照或内部服务实现；现有生产调用路径一次性迁移，不保留兼容层。
- **BREAKING**：动态插件`data`服务只允许访问当前插件自有表；`sys_*`表和官方能力插件表不得通过`data`服务对普通插件开放。
- **BREAKING**：现有`orgcap`、`tenantcap`、`ai`和动态`host service`旧方法按统一领域能力模型重整，不保留旧接口、旧协议方法或旧插件调用方式作为生产兼容入口。
- 新增插件宿主领域能力模型，覆盖`usercap`、`authzcap`、`dictcap`、`filecap`、`sessioncap`、`configcap`、`notifycap`、`plugincap`、`jobcap`、`infracap`，并重整现有`orgcap`、`tenantcap`和`ai`能力。
- 新增统一`CapabilityContext`、领域命名`ID`类型、批量缺失语义、`labelKey`/`label`语义、领域方法错误语义、缓存失效模型和数据权限上下文边界。
- 源码插件通过`pluginhost.Services.Admin()`获得完整类型化`AdminService`目录；源码插件不再维护字符串式管理能力授权声明，但管理方法仍必须执行租户边界、目标数据边界、状态机、系统 actor 和审计治理。
- 动态插件通过`plugin.yaml hostServices`声明领域服务和方法，并在安装或启用阶段完成授权；安装授权替代插件级菜单/RBAC 方法校验，但不替代领域数据权限、租户、状态机、数量上限和审计校验。
- 新增治理扫描，阻断插件生产代码重新生成或引用宿主核心表、旧领域接口、旧动态`host service`方法和动态`data`服务核心表授权。
- 一次 OpenSpec 迭代内落地全部领域能力契约和治理规则，任务按领域和迁移面拆分逐步实现。

## Capabilities

### New Capabilities

- `plugin-host-domain-capabilities`：定义插件访问宿主业务数据、治理状态和基础设施原语的领域能力总边界，包括领域接口、上下文、授权、数据权限、缓存一致性、`i18n`标签语义和迁移治理。

### Modified Capabilities

- `plugin-capability-boundary-governance`：强化插件不得依赖宿主核心表、宿主内部实现或其他插件内部实现的边界，并要求源码插件和动态插件共享领域能力语义。
- `plugin-data-service`：将动态`data`服务收窄为当前插件自有表访问，明确`sys_*`核心表和官方能力插件表必须通过领域能力暴露。
- `plugin-host-service-extension`：扩展动态`hostServices`协议以支持领域服务、领域方法、安装授权、`CapabilityContext`和类型化 guest 目录。
- `plugin-permission-governance`：明确动态插件领域管理方法的安装授权模型，以及源码插件`AdminService`目录与插件菜单/RBAC 权限的关系。

## Impact

- 影响`apps/lina-core/pkg/plugin/capability/**`、`apps/lina-core/pkg/plugin/pluginhost/**`、`apps/lina-core/pkg/plugin/pluginbridge/**`、`apps/lina-core/internal/service/plugin/**`和相关宿主领域`internal/service/**`适配器。
- 影响所有直接生成或读取宿主`sys_*`表的源码插件，包括通知、组织、租户、在线用户、操作日志、登录日志等官方插件。
- 影响动态插件`plugin.yaml hostServices`校验、安装授权快照、guest SDK、WASM host service 分发、协议`codec`和治理扫描。
- 可能涉及宿主`manifest/sql/`或插件`manifest/sql/`新增幂等 DDL/Seed DML、插件`backend/hack/config.yaml`生成范围清理、DAO 重新生成和启动装配调整。
- 影响权限、数据权限、租户隔离、插件状态、字典、组织树、运行时配置等关键缓存的一致性策略与验证。
- `i18n`影响：领域能力返回稳定值和`labelKey`，需要返回`label`时按当前请求 locale 解析；新增或修改 API 文档源文本、运行时错误或用户可见标签时必须维护宿主或已启用插件的对应`i18n`资源。
- 开发工具影响：新增或修改治理扫描必须使用跨平台 Go 工具或`linactl`入口，避免新增平台专属脚本。
