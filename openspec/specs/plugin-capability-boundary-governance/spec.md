# plugin-capability-boundary-governance Specification

## Purpose
TBD - created by archiving change refine-plugin-capability-boundaries. Update Purpose after archive.
## Requirements
### Requirement: 插件相关公共组件必须保持单一职责

系统 SHALL 为插件相关公共组件定义清晰职责边界。`pluginhost`只负责源码插件贡献 API；`pluginservice`负责统一插件能力消费目录；`pluginbridge`只负责动态插件 ABI、WASM transport 和协议 facade；`plugindb`只负责动态插件 guest 侧受限数据 DSL 和必要 facade；插件资源扫描、路径治理、runtime cache、source upgrade 和 host-side 执行器等实现细节 MUST 放入职责明确的`internal`组件。

#### Scenario: 开发者定位源码插件贡献入口

- **WHEN** 开发者需要注册源码插件路由、hook、cron、生命周期或 provider factory
- **THEN** 开发者使用`pkg/pluginhost`
- **AND** `pluginhost`不提供宿主业务能力消费实现

#### Scenario: 开发者定位插件消费能力

- **WHEN** 源码插件或动态插件需要访问配置、数据、缓存、通知、鉴权、i18n 或 pluginservice capability
- **THEN** 插件通过`pkg/pluginservice`公开的能力目录或动态 guest client 使用能力
- **AND** 插件不得把`pluginbridge`低层协议包当作业务能力 owner

### Requirement: 不应公开的插件实现必须放入 internal 边界

系统 SHALL 将不属于稳定公共契约的插件实现放入`internal`目录。非公开资源包括 bridge codec、WASM artifact 解析实现、host call dispatcher、host service wire 实现、plugindb typed plan、host DB wrapper、插件资源扫描器、插件路径治理、provider registry、source upgrade 执行器和运行时 cache 实现。

#### Scenario: Bridge 低层实现不再作为公共 API

- **WHEN** 宿主需要编码 bridge envelope 或解析 WASM artifact
- **THEN** 宿主通过`pluginbridge`根 facade 或授权内部包调用
- **AND** 外部插件代码不得 import `pkg/pluginbridge/internal/**`

#### Scenario: 插件资源扫描实现不公开

- **WHEN** 宿主扫描源码插件目录、动态 artifact 或插件 manifest 资源
- **THEN** 扫描器、路径校验和资源索引实现位于宿主`internal`职责包
- **AND** 插件代码不得依赖这些扫描实现作为公共文件系统 API

### Requirement: 插件间运行时调用必须经过稳定能力接缝

系统 SHALL 禁止插件直接调用其他插件的内部实现。插件间协作 MUST 通过`pluginservice`能力目录、事件、hook、版本化 host service、HTTP API 或其他受治理稳定契约完成；插件不得直接 import 其他插件的`backend/internal/**`、provider adapter、DAO、DO、Entity 或缓存实现。

#### Scenario: 插件消费另一个插件提供的租户能力

- **WHEN** 插件`plugin-b`需要使用由`plugin-a`提供的租户能力
- **THEN** `plugin-b`声明对 provider 插件的硬依赖或按可选能力降级，并通过`pluginservice.Services.Tenant()`或等价`tenantcap.Service`调用
- **AND** `plugin-b`不得 import `plugin-a/backend/internal/provider/tenantadapter`

#### Scenario: 静态治理发现跨插件内部导入

- **WHEN** 非测试生产代码 import 其他插件的`backend/internal/**`
- **THEN** 治理验证失败
- **AND** 变更必须改为依赖稳定能力契约或记录受控启动装配例外

### Requirement: 插件能力公开面必须有治理验证

系统 SHALL 提供静态检索、Go 编译门禁或审查记录来验证插件能力公开面。验证 MUST 覆盖公共包导入边界、provider adapter 导入边界、低层实现 internal 化和源码/动态插件统一能力消费路径。

#### Scenario: Provider Adapter 被作为公开契约导入时被拒绝

- **WHEN** 生产代码 import 其他插件的`backend/provider/**`provider adapter
- **THEN** 静态检索或审查记录必须指出该调用方应改为依赖`pluginservice`稳定能力契约
- **AND** 该变更不得通过审查，除非规范明确批准该 adapter 成为稳定公共契约

#### Scenario: 非目标能力契约导入被拒绝

- **WHEN** 新增生产代码继续 import 已迁移的`pkg/frameworkcap`、`pkg/orgcap`、`pkg/tenantcap`或宿主`internal/service/orgcap`、`internal/service/tenantcap`旧路径
- **THEN** 静态检索、Go 编译门禁或审查记录必须指出该代码不符合目标能力契约
- **AND** 代码必须改为使用`pkg/pluginservice/orgcap`或`pkg/pluginservice/tenantcap`能力组件

### Requirement: 插件能力边界不得诱导重复适配和分叉协议

系统 SHALL 确保源码插件和动态插件访问同一宿主能力时使用同一语义契约、授权模型、错误语义和数据边界。动态插件可以通过`pluginbridge`transport 调用，但 host service handler MUST 适配到`pluginservice`统一能力目录；源码插件不得使用另一套平行宿主能力接口。

#### Scenario: 同一配置能力对两类插件语义一致

- **WHEN** 源码插件和动态插件分别读取当前插件作用域配置
- **THEN** 二者通过`pluginservice`配置能力获得一致的 key 作用域、错误语义和授权边界
- **AND** 动态插件的 bridge 调用只作为 transport 差异存在

#### Scenario: 同一框架能力对两类插件语义一致

- **WHEN** 源码插件和动态插件分别消费`framework.org.v1`
- **THEN** 二者最终调用同一个`orgcap.Service`
- **AND** 结果 DTO、降级语义、数据权限边界和错误码保持一致

