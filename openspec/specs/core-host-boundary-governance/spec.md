# 核心宿主边界治理规范

## Purpose
待定 - 由归档变更 framework-positioning-core-boundary-and-readme-i18n 创建。归档后更新目的。
## Requirements
### Requirement:lina-core 核心宿主能力保持通用边界
系统 SHALL 将 `apps/lina-core` 视为框架的核心宿主服务，优先保证通用模块接口能力、组件能力、系统治理能力和插件扩展能力的稳定性和可复用性。

#### Scenario:页面需求仅影响工作台展示
- **当** 需求仅变更表格列、筛选项、树选择器、工作台聚合、路由组装或其他特定页面展示结构时
- **则** 系统优先通过工作台适配接口或前端适配层完成变更
- **且** 不因该页面需求直接修改 `lina-core` 的核心领域契约、通用服务语义或存储模型

#### Scenario:计划修改核心接口或模型
- **当** 开发者因某前端页面需求计划修改 `lina-core` 的核心接口、领域模型或持久化结构时
- **则** 修改必须能够证明其服务于框架级通用能力而非单个页面表单
- **且** 如果无法证明，应回退到工作台适配实现方案

### Requirement:工作台适配接口必须显式分类
系统 SHALL 将明确面向工作台组装的输出（如菜单路由投影、当前用户工作台启动数据、树选择器、下拉选项等）显式分类为工作台适配接口，而非通用领域接口。

#### Scenario:接口返回工作台组装数据
- **当** 接口返回菜单路由、宿主工作区启动数据、树选择器节点或下拉选项等工作台组装数据时
- **则** 其接口描述、DTO 注解和相关规范清晰标记输出属于工作台适配语义
- **且** 不将此输出描述为通用领域模型本身

#### Scenario:通用领域能力被多个工作台复用
- **当** 能力需要被不同工作站或不同访问方式同时消费时
- **则** 系统优先保留稳定的通用领域接口
- **且** 不同工作台所需的菜单、路由、展示结构或聚合视图通过独立适配输出组装

### Requirement:开源阶段宿主仅保留框架核心和管理基座

系统 SHALL 在开源阶段将 `apps/lina-core` 收敛为框架核心和管理基座，默认不再内置所有管理后台业务模块。

#### Scenario:规划新增后台模块
- **当** 团队规划新增管理后台模块时
- **则** 先判断能力是否属于宿主基座能力（如认证、权限、菜单、插件管理、任务调度、配置、字典或文件等）
- **且** 如果不属于宿主基座能力，优先采用源码插件设计而非直接纳入宿主

#### Scenario:判断能力是否应保留在宿主
- **当** 能力被多个模块复用且承担框架级统一治理职责时
- **则** 系统将其保留在宿主
- **且** 不因可选业务模块需求继续扩展宿主边界到业务侧

### Requirement:默认后台一级目录由宿主稳定提供

系统 SHALL 由宿主提供默认后台一级目录挂载点，确保开发者长期扩展业务时无需反复调整顶层导航结构。

#### Scenario:插件提供后台功能菜单
- **当** 源码插件需要向默认后台注册菜单时
- **则** 该插件菜单必须挂载到宿主提供的稳定一级目录下
- **且** 插件不得绕过宿主管理自行创建新的一级目录体系

#### Scenario:插件未安装或未启用
- **当** 某级目录下的所有子菜单均来自未安装或未启用的插件时
- **则** 宿主自动隐藏该空目录
- **且** 左侧导航中不保留空壳父目录

### Requirement:宿主稳定目录必须作为真实治理记录存在

系统 SHALL 将默认后台的一级稳定目录作为宿主拥有的稳定菜单记录维护，而非仅在前端投影层临时组装。稳定父级 `menu_key` MUST 至少包含：`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`。系统 MUST NOT 将 `storage` 作为一级宿主稳定目录。

#### Scenario:初始化宿主稳定目录
- **当** 宿主初始化默认后台菜单骨架时
- **则** 宿主创建并维护上述稳定父级 `menu_key`（不含 `storage`）
- **且** 这些目录记录可被插件 `parent_key` 稳定解析

#### Scenario:某目录下无可见子菜单
- **当** `内容管理`、`组织管理` 或 `系统监控` 目录当前没有任何可见子菜单时
- **则** 它们在导航投影中被隐藏
- **且** 宿主不删除对应的稳定目录记录

### Requirement:认证会话内核和统一事件发布能力保留在宿主

系统 SHALL 将认证会话事实源和统一登录事件、统一审计事件的发布能力保留在宿主，而非委托给可选源码插件。

#### Scenario:规划在线用户插件边界
- **当** 团队规划 `linapro-monitor-online` 的能力边界时
- **则** 插件仅承载在线用户查询和强制下线管理
- **且** JWT 验证、会话触碰刷新、超时判定和清理任务仍保留在宿主

#### Scenario:规划日志插件边界
- **当** 团队规划 `linapro-monitor-loginlog` 或 `linapro-monitor-operlog` 的能力边界时
- **则** 宿主在认证链路和请求链路上发布统一事件
- **且** 宿主核心链路不直接依赖这些插件的具体持久化实现

### Requirement:宿主和插件必须通过稳定能力接缝解耦

系统 SHALL 通过`pluginservice`、能力接口、事件 Hook、路由注册器和 Cron 注册器等稳定接缝完成宿主与插件的协作，而非在宿主业务代码中散落插件特定的占位逻辑、大量`if pluginEnabled`分支或对插件内部实现的直接依赖。

#### Scenario:宿主调用可选组织能力
- **当** 用户管理、认证或其他宿主核心模块需要访问部门、岗位、组织树或组织数据范围等可选能力时
- **则** 宿主通过显式注入的`orgcap.Service`或由`pluginservice.Services.Org()`发布的组织能力入口访问这些能力
- **且** `linapro-org-core`的插件状态判断和功能分支不直接散落在宿主实现中
- **且** 宿主仅持有该能力的接口、DTO、消费 service 和空实现，不直接查询或维护`linapro-org-core`的物理表

#### Scenario:宿主调用可选租户能力
- **当** 认证、会话、数据权限或插件 host service 需要访问租户上下文时
- **则** 宿主通过显式注入的`tenantcap.Service`或由`pluginservice.Services.Tenant()`发布的租户能力入口访问该能力
- **且** 租户 provider 的具体实现由满足生命周期条件的插件提供
- **且** 宿主不得在业务路径中直接 import 租户插件的内部 service 或 DAO

#### Scenario:宿主扩展插件日志或监控能力
- **当** 非核心能力拆分为源码插件时
- **则** 宿主仅保留稳定的事件、治理接口和注册入口
- **且** 不在宿主控制器、服务或中间件中为个别插件保留大量功能占位逻辑

### Requirement:宿主不得持有源码插件自有业务存储

系统 SHALL 将源码插件业务表、对应 ORM 产物和演示数据视为插件私有资产，在宿主默认数据库初始化、Mock 加载或宿主源码树中不保留长期副本。

#### Scenario:初始化默认数据库
- **当** 管理员执行宿主默认数据库初始化时
- **则** 宿主仅创建和初始化宿主核心表和必要的 Seed 数据
- **且** 不创建任何源码插件业务表

#### Scenario:迁移业务模块到源码插件
- **当** 某业务模块已迁移到官方源码插件时
- **则** 该模块业务表对应的 `dao`、`do`、`entity` 和直接查表逻辑不再保留在宿主源码中
- **且** 宿主仅通过稳定能力接缝或插件注册入口与插件协作

#### Scenario:加载默认演示数据
- **当** 管理员执行宿主默认 Mock 数据加载时
- **则** 宿主不写入任何源码插件业务表
- **且** 插件演示数据由插件自身的生命周期资源负责

### Requirement: 宿主插件系统私有实现必须收敛到插件服务边界

系统 SHALL 将宿主插件系统的私有实现组织在`apps/lina-core/internal/service/plugin`边界下，并将插件公共契约、SDK、bridge 和 capability 能力继续组织在`apps/lina-core/pkg/plugin`边界下。除明确服务多个宿主领域且不属于插件语义的共享基础组件外，插件 catalog、runtime、host service adapter、runtime cache、lifecycle、integration、frontend、openapi、WASM host service、管理投影和插件治理实现 MUST 不作为`internal/service`根层级的独立 service 组件长期存在。

#### Scenario: 开发者查找宿主插件实现
- **WHEN** 开发者需要理解宿主插件 catalog、runtime、host service adapter、runtime cache、lifecycle 或管理投影实现
- **THEN** 相关私有实现位于`apps/lina-core/internal/service/plugin`及其子目录下
- **AND** 开发者不需要在`internal/service/pluginhostservices`、`internal/service/pluginruntimecache`等平行根组件中继续查找插件系统核心逻辑

#### Scenario: 公共插件契约仍归属 pkg plugin
- **WHEN** 源码插件、动态插件或构建工具需要访问插件公共契约、guest SDK、bridge 协议或 capability 服务接口
- **THEN** 这些稳定契约继续通过`apps/lina-core/pkg/plugin`体系暴露
- **AND** 宿主私有实现不得迁入`pkg/plugin`公共边界

### Requirement: plugin 根包必须作为宿主插件服务 facade

系统 SHALL 将`apps/lina-core/internal/service/plugin`根包维护为宿主内部稳定 facade。根包 MUST 保留`Service`契约、公开投影类型、启动期构造入口、轻量编排和必要适配；具体实现逻辑 MUST 优先下沉到职责明确的同包文件或`plugin/internal/<subcomponent>`子组件。根包不得继续积累可独立测试、跨文件共享状态、缓存协调、插件桥接、运行时升级或管理投影等多职责实现。

#### Scenario: 启动装配构造插件宿主服务
- **WHEN** `internal/cmd`需要构造插件服务、源码插件宿主能力目录或 WASM host service 配置
- **THEN** 它通过`internal/service/plugin`根包的稳定 facade 完成
- **AND** 它不得直接导入`plugin/internal/<subcomponent>`实现包

#### Scenario: 新增插件内部职责
- **WHEN** 变更新增或迁移插件管理列表、运行时升级、平台治理、启动一致性、host service adapter 或 runtime cache 等实现
- **THEN** 实现必须归入职责明确的`plugin/internal/<subcomponent>`或`plugin/runtimecache`等目标子组件
- **AND** 子组件命名必须体现领域职责，不得使用`util`、`common`或`helper`等兜底名称

#### Scenario: 根包保留核心跨组件编排
- **WHEN** 某段逻辑跨越 catalog、runtime、integration、lifecycle、cache 和 i18n 等多个插件子组件
- **THEN** 只有在下沉会引入更多转发接口或循环依赖风险时，该逻辑才可暂留`plugin`根包
- **AND** 审查必须记录暂留原因和后续可继续收敛的判断

### Requirement:破坏性公共契约收敛不得保留兼容 facade

系统 SHALL 在本迭代中直接删除插件公共包中的历史宽契约、误导性 facade 和兼容转发层。系统 MUST NOT 为已删除的`HostServices`、`HostServicesForPlugin`、`HostServices()`、`contract.ProviderEnv.Services`或`pluginbridge`业务能力 client 保留生产兼容入口。

#### Scenario:旧插件能力入口被生产代码引用

- **WHEN** 生产代码继续引用旧`pluginhost.HostServices`、`HostServicesForPlugin`、`HostServices()`、`contract.ProviderEnv.Services`或`pluginbridge.Runtime()`等业务能力入口
- **THEN** 编译或治理扫描必须失败
- **AND** 调用方必须迁移到`capability.Services`、`pluginhost.Services`的`Services()`访问器、强类型 provider env 或`pkg/plugin/capability/guest`

#### Scenario:插件确实需要新增读能力

- **WHEN** 删除旧宽接口后发现插件仍需要组织、租户或宿主能力读数据
- **THEN** 系统只能新增 DTO 化、批量化、只读能力
- **AND** 不得恢复`*gdb.Model`、`*ghttp.Request`、DAO、DO、Entity、写入接口、数据范围注入或宿主内部治理接口

### Requirement:插件公共契约和宿主内部插件运行时必须分离

系统 SHALL 将`apps/lina-core/pkg/plugin`作为插件公共契约命名空间，将`apps/lina-core/internal/service/plugin`作为宿主内部插件运行时治理实现命名空间。公共契约不得导入宿主内部插件运行时实现，插件代码不得直接调用宿主内部插件运行时包。

#### Scenario:公共插件包不能依赖宿主内部运行时

- **WHEN** 开发者在`apps/lina-core/pkg/plugin/**`下实现或修改公共契约
- **THEN** 代码不得导入`lina-core/internal/service/plugin/**`
- **AND** 公共契约只能通过 DTO、接口、协议 envelope 或 provider-facing 契约表达数据边界

#### Scenario:宿主内部运行时使用公共契约

- **WHEN** `apps/lina-core/internal/service/plugin/**`需要执行插件 catalog、runtime、lifecycle、WASM host service 或管理端投影逻辑
- **THEN** 它可以依赖`apps/lina-core/pkg/plugin/**`中的稳定公共契约
- **AND** 不得要求公共包反向暴露宿主内部缓存快照、私有配置、DAO、DO、Entity 或 runtime 状态结构

#### Scenario:插件代码不能导入宿主内部实现

- **WHEN** 源码插件或动态插件业务代码访问宿主能力
- **THEN** 它必须通过`pkg/plugin/capability`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge`或受治理`hostServices`协议完成
- **AND** 不得导入`lina-core/internal/service/**`、`lina-core/internal/dao/**`或`lina-core/internal/model/**`

### Requirement:宿主内部组织和租户治理能力必须通过窄接口注入

系统 SHALL 将组织、租户、数据范围、成员关系、自动开通和启动一致性等宿主内部治理能力拆分为职责明确的窄接口，并通过构造函数显式注入到需要的宿主 service。

#### Scenario:宿主数据范围服务注入组织范围接口

- **WHEN** 宿主`datascope`或其他核心 service 需要按部门或组织关系在数据库查询阶段过滤数据
- **THEN** 该 service 依赖组织范围治理窄接口
- **AND** 不依赖普通插件消费用的`orgcap.Service`宽接口

#### Scenario:宿主用户和角色服务注入租户成员接口

- **WHEN** 宿主用户、角色或通知 service 需要校验、投影或更新用户租户成员关系
- **THEN** 该 service 依赖租户成员治理窄接口
- **AND** 不通过普通插件消费目录获取写入或底层查询注入能力

#### Scenario:测试替身只实现被测窄接口

- **WHEN** 单元测试为组织或租户能力构造替身
- **THEN** 替身只需要实现当前被测 service 实际依赖的窄接口
- **AND** 不得为了满足过宽接口而实现大量无关空方法

### Requirement: 宿主定义插件实现的框架能力必须归属独立 pluginservice 能力组件

系统 SHALL 将“宿主定义接口、插件提供实现”的框架能力按能力领域归属到`pkg/pluginservice/orgcap`和`pkg/pluginservice/tenantcap`等独立组件。每个能力组件必须维护自身公开契约、DTO、消费`Service`、fallback、delegation 和 provider factory facade；共享 provider registry 与激活治理实现必须位于`pkg/pluginservice/internal/capabilityregistry`，不得再保留`pkg/frameworkcap`聚合包、旧`pkg/orgcap`/`pkg/tenantcap`兼容包或宿主`internal/service/orgcap`、`internal/service/tenantcap`双重适配层。

#### Scenario: orgcap 迁移为 pluginservice 独立组件

- **WHEN** 系统维护组织能力契约
- **THEN** 组织能力公开消费契约由`pkg/pluginservice/orgcap`暴露
- **AND** 消费方通过注入的`orgcap.Service`或`pluginservice.Services.Org()`获取组织能力实例
- **AND** 旧`pkg/orgcap`不得作为新代码入口
- **AND** fallback、delegation、provider factory facade 位于`pkg/pluginservice/orgcap`，共享 provider registry 和激活实现位于`pkg/pluginservice/internal/capabilityregistry`

#### Scenario: tenantcap 迁移为 pluginservice 独立组件

- **WHEN** 系统维护租户能力契约
- **THEN** 租户能力公开消费契约由`pkg/pluginservice/tenantcap`暴露
- **AND** 消费方通过注入的`tenantcap.Service`或`pluginservice.Services.Tenant()`获取租户能力实例
- **AND** 旧`pkg/tenantcap`不得作为新代码入口
- **AND** provider 实现由插件通过`tenantcap.Provide(...)`声明并由`pkg/pluginservice/internal/capabilityregistry`中的生命周期治理激活

### Requirement:云对象存储实现由插件扩展且配置挂载到系统设置

系统 SHALL 将对象存储领域契约与内置 local provider 保留在宿主，将具体云厂商对象存储后端实现交付为官方源码插件。云存储配置菜单 MUST 挂载到宿主已有 `setting`（系统设置）稳定目录；MUST NOT 要求单独的 `storage` 一级目录或 `linapro-storage-core` 类壳插件。

#### Scenario:规划云存储插件边界
- **当** 团队规划 `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure` 或 `linapro-storage-s3` 的能力边界时
- **则** 插件仅承载对应云厂商 `storagecap.Provider`、配置 settings 与连通性探测
- **且** `storagecap.Service`、插件/租户 key 作用域与 local provider 仍保留在宿主
- **且** 配置入口挂载在宿主「系统设置」目录下

### Requirement: 邮件协议与连接配置不得进入宿主核心契约

系统 SHALL 将邮件协议实现、Connection/Account 权威存储与邮件管理面保留在 `linapro-mail-core` 及其协议插件中。`apps/lina-core/pkg/plugin` MUST NOT 新增邮件协议 SPI 实现包或邮件 Connection 领域表访问作为长期公共契约。宿主 notify 可依赖邮件 owner 能力完成 email 通道投递，但 MUST NOT 将 SMTP/IMAP/POP3 细节嵌入宿主通用模块。

#### Scenario: 邮件公开契约位于 owner 插件

- **WHEN** 源码插件需要类型化邮件发送或接收能力
- **THEN** 公开契约 MUST 来自 `linapro-mail-core` 的 `backend/cap` 路径
- **AND** MUST NOT 要求从 `lina-core/pkg/plugin/capability` 导入邮件协议实现

#### Scenario: 宿主保持通用边界

- **WHEN** 审查本变更相关代码归属
- **THEN** 邮件 Connection 表与协议客户端 MUST 不落在 `lina-core` 宿主业务表与核心 service 中作为权威实现
- **AND** 宿主插件生命周期仅提供通用全局 Hook 机制，MUST NOT 硬编码邮件 kind 冲突规则

