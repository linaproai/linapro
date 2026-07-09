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

### Requirement: 插件生产代码不得依赖宿主核心表实现

系统 SHALL 禁止源码插件和动态插件生产代码生成或直接查询宿主核心`sys_*`表、响应宿主私有缓存快照或宿主内部 service 实现。宿主核心数据 MUST 由对应领域 owner 通过领域能力、`pluginhost.Services`或动态`hostServices`协议发布。Go 语言`internal`目录规则已经阻断的宿主`DAO/DO/Entity`导入和类型使用不作为治理扫描规则重复检查。

#### Scenario: 源码插件生成宿主核心表 DAO

- **WHEN** 插件`backend/hack/config.yaml`声明生成`sys_user`、`sys_role`、`sys_dict_data`或其他宿主核心表
- **THEN** 治理验证失败
- **AND** 插件必须改为依赖对应领域能力契约

#### Scenario: 插件生产代码直接查询宿主表

- **WHEN** 插件生产代码调用`g.DB().Model("sys_*")`、`shared.TableSysUser`或等价直接表入口
- **THEN** 治理验证失败
- **AND** 变更不得通过审查，除非该调用位于测试、Mock、安装 SQL 或迁移治理例外边界内

### Requirement: 源码插件和动态插件必须共享领域能力语义

系统 SHALL 要求源码插件和动态插件访问同一宿主领域能力时共享领域 owner、输入输出 DTO、领域 ID 类型、数据权限、缓存一致性、错误语义和 i18n 标签语义。动态插件 hostServices handler 只能作为 transport 适配层，不得成为与源码插件平行的领域语义 owner。

### Requirement: 插件公开能力服务必须归属 *cap 组件包

系统 SHALL 要求`apps/lina-core/pkg/plugin/capability`下对插件公开的具体能力服务接口归属职责明确的领域命名空间或`*cap`组件包。`capability.Services`普通消费面 MUST 只返回各领域命名空间、`*cap.Service`或等价窄接口，不得返回`contract.*Service`具体服务接口。

#### Scenario: 根能力目录返回具体服务

- **WHEN** 开发者查看`capability.Services`
- **THEN** 每个普通能力方法返回对应领域命名空间或`*cap`组件包的服务接口
- **AND** 根目录不得继续暴露`Config()`、`PluginConfig()`、`PluginLifecycle()`或`PluginState()`

#### Scenario: 认证授权能力族入口

- **WHEN** 插件需要访问认证 token handoff 或授权能力
- **THEN** 根能力目录只暴露`Services.Auth()`认证授权能力族入口
- **AND** token 生命周期能力通过`Services.Auth().Token()`访问
- **AND** 授权查询能力通过`Services.Auth().Authz()`访问

### Requirement: Services 方法名必须按领域消费语义命名

系统 SHALL 将`capability.Services`方法名视为插件开发者看到的领域入口。资源集合或领域命名空间 MAY 使用复数入口，例如`Users()`、`Jobs()`、`Plugins()`；单一上下文、配置能力或专有能力 SHOULD 使用单数或专有名词，例如`Tenant()`、`BizCtx()`、`HostConfig()`、`AI()`。

### Requirement: 插件相关能力必须收口到 Plugins 命名空间

系统 SHALL 将插件自身配置、插件状态、插件生命周期和插件治理投影收口到`Services.Plugins()`插件领域命名空间。根`capability.Services` MUST NOT 继续暴露`Config()`、`PluginConfig()`、`PluginLifecycle()`或`PluginState()`。

### Requirement: 配置公开面只能包含插件自身配置和宿主配置

系统 SHALL 将插件公开配置能力限定为两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主授权开放配置。根`Services.Config()` MUST NOT 作为普通插件公开入口存在。

### Requirement: 租户过滤不得进入普通租户消费面

系统 SHALL 将源码插件自有表`tenant_id`过滤接口归属到`tenantcap.PluginTableFilterService`，但该接口 MUST 只通过`pluginhost.Services.TenantFilter()`等源码插件专用受控接缝暴露。普通`capability.Services.Tenant()` MUST 只返回`tenantcap.Service`普通租户消费面。

### Requirement: 公共原语包不得承载具体能力服务

系统 SHALL 允许`capability`维护一个小型公共原语包，用于承载跨领域值对象。该公共原语包 MUST NOT 定义具体能力 Service、AdminService、factory、provider adapter 或 host service handler。

### Requirement: 认证授权子能力必须收敛到 authcap 能力族

系统 SHALL 将认证 token handoff 与授权能力作为`authcap`能力族维护。`authcap`根包只承载聚合入口，子领域`authcap/token`维护租户 token 契约，子领域`authcap/authz`维护权限投影和角色授权管理契约。

### Requirement: 旧 capability/contract 具体服务聚合必须删除

系统 SHALL 删除或清空`capability/contract`作为具体服务聚合包的职责。迁移完成后，生产代码、官方插件和测试替身 MUST 不再导入旧路径获取具体能力服务接口。
