## Requirements

### Requirement: 插件清单处理必须与治理副作用分离

系统 SHALL 将插件清单扫描、解析、校验和访问视为清单事实源读取能力。清单处理路径不得隐藏写入治理表或触发菜单同步、权限同步、资源引用同步、hook 分发、运行时节点状态同步或缓存失效。需要治理写入或副作用时，调用方 MUST 在显式编排入口中按顺序调用清单、存储和副作用能力。

### Requirement: 插件治理写入后的副作用调用点必须可追踪

系统 SHALL 在插件显式同步、安装、卸载、启用、禁用、升级、动态包上传和租户供应策略更新等治理路径中保留可追踪的副作用调用点。每个菜单同步、资源引用同步、hook 分发、运行时节点状态同步和缓存失效操作 MUST 由当前编排入口显式触发，且错误处理语义必须与治理写入顺序一致。

### Requirement: 插件清单读取必须复用有界读模型缓存

系统 SHALL 为插件清单事实源提供内部读模型缓存。源码插件 manifest、动态插件 desired manifest、release manifest 与 release YAML 快照 MUST 按各自权威数据源和不可变性分开缓存。缓存读取路径 MUST 不写入插件治理表，不触发副作用。

### Requirement: 插件清单缓存必须支持按插件显式失效

系统 SHALL 在插件同步、动态包上传、安装、卸载、启用、禁用、升级、active release 切换或源码插件同步成功后，按 `pluginID` 或 artifact 路径失效对应清单缓存。全局失效只允许用于全局配置或无法确定影响范围的治理事件。

### Requirement: 插件生命周期编排下沉后必须保持治理语义

系统 SHALL 在将插件生命周期编排迁入 lifecycle 子组件后保持现有安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子的治理语义。平台上下文守卫、依赖检查、反向依赖阻断、host service authorization、SQL migration、资源引用同步、菜单权限同步、hook 分发、runtime state 同步和缓存失效不得因迁移遗漏或改变顺序。

### Requirement: 插件生命周期业务控制参数必须显式传递

系统 SHALL 将生命周期普通业务控制参数作为方法参数、options 或稳定输入结构显式传递。安装 mock data、startup auto-enable 标记、依赖检查结果和类似控制语义 MUST NOT 通过普通请求 context key 隐式改变生命周期行为。

### Requirement: 租户生命周期钩子必须通过 lifecycle 子组件编排

系统 SHALL 由 lifecycle 子组件编排租户删除、租户插件禁用和新租户供应相关的插件 lifecycle precondition 与 notification。根门面和 tenant capability adapter MUST 只依赖窄接口，不得复制租户生命周期扫描和 veto 汇总逻辑。

### Requirement: 升级预览与执行由 lifecycle 编排

### Requirement: 升级预览与执行由 lifecycle 编排

系统 SHALL 通过 lifecycle 编排入口提供源码插件升级与动态插件 runtime 升级预览/执行能力。升级实现可保留独立内部包，但 MUST 由 lifecycle 构造、持有并向根 facade 暴露。

#### Scenario: 管理端请求 runtime 升级预览

- **WHEN** 操作者请求某个待升级插件的 runtime 升级预览
- **THEN** 根 facade 调用 lifecycle 拥有的升级能力
- **AND** 预览结果仍包含版本对比、依赖检查、SQL 摘要与 hostServices 差异
- **AND** 该路径不要求根 facade 直接依赖 upgrade 包构造函数

#### Scenario: 管理端确认执行 runtime 升级

- **WHEN** 操作者确认执行 runtime 升级
- **THEN** lifecycle 拥有的升级能力完成锁、状态迁移、SQL/回调与 release 切换编排
- **AND** 动态副作用仍通过 runtime reconcile / lifecycle callback 执行

### Requirement: 插件清单必须声明分发治理类型

系统 SHALL 支持插件 manifest 的`distribution`字段。缺省值 MUST 归一化为`managed`；合法值仅包含`managed`和`builtin`。非法值 MUST 在 manifest 校验阶段失败。`builtin`仅允许`type=source`，且必须存在同 ID 的编译期源码插件注册绑定。插件注册表和发布 manifest snapshot MUST 保存归一化后的`distribution`。

#### Scenario: 普通插件缺省为 managed

- **WHEN** 插件 manifest 未声明`distribution`
- **THEN** 系统将该插件归一化为`distribution=managed`
- **AND** 同步到`sys_plugin`和发布 manifest snapshot 的值均为`managed`

#### Scenario: 普通可管理插件声明 managed

- **WHEN** 插件 manifest 声明`distribution: managed`
- **THEN** manifest 校验通过
- **AND** 系统将该插件注册表和发布 manifest snapshot 的`distribution`保存为`managed`

#### Scenario: 内建源码插件声明 builtin

- **WHEN** 源码插件 manifest 声明`distribution: builtin`
- **AND** 该插件通过编译期源码插件注册表绑定同一插件 ID
- **THEN** manifest 校验通过
- **AND** 系统将该插件注册表和发布 manifest snapshot 的`distribution`保存为`builtin`

#### Scenario: 动态插件不能声明 builtin

- **WHEN** 动态插件 manifest 声明`distribution: builtin`
- **THEN** manifest 校验失败
- **AND** 系统不得将该插件降级为`managed`继续同步

#### Scenario: 未注册源码插件不能声明 builtin

- **WHEN** 源码插件 manifest 声明`distribution: builtin`
- **AND** 源码插件注册表中不存在同 ID 绑定
- **THEN** manifest 校验失败
- **AND** 错误包含插件 ID 和缺失源码注册绑定的事实

#### Scenario: 旧版 marketplace distribution 被拒绝

- **WHEN** 插件 manifest 声明`distribution: marketplace`
- **THEN** manifest 校验失败
- **AND** 错误列出合法值`managed`和`builtin`

#### Scenario: 非法 distribution 被拒绝

- **WHEN** 插件 manifest 声明`distribution: external`
- **THEN** manifest 校验失败
- **AND** 错误列出合法值`managed`和`builtin`

### Requirement: 插件管理投影必须暴露并过滤分发治理类型

系统 SHALL 在插件列表和详情 API 的服务端投影中返回`distribution`字段。普通插件管理列表默认仅返回`distribution=managed`插件。显式诊断查询包含`builtin`插件时，查询 MUST 仍为只读操作，并不得触发治理表写入或生命周期副作用。

#### Scenario: 普通管理列表隐藏 builtin 插件

- **WHEN** 管理员调用默认插件列表查询
- **THEN** 响应仅包含`distribution=managed`的插件
- **AND** 响应中的每个插件项都包含`distribution`字段

#### Scenario: 诊断查询包含 builtin 插件

- **WHEN** 具备诊断权限的管理员调用包含内建插件的只读查询
- **THEN** 响应可以包含`distribution=builtin`的插件
- **AND** 查询不得安装、启用、升级、同步或修复任何插件治理数据

#### Scenario: 插件详情返回 distribution

- **WHEN** 管理员查询任一插件详情
- **THEN** 响应包含该插件当前注册表或发布投影中的`distribution`
